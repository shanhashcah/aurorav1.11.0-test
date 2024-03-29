package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hcnet/go/keypair"
	"github.com/hcnet/go/services/aurora/internal/test/integration"
	"github.com/hcnet/go/txnbuild"
	"github.com/stretchr/testify/assert"
)

const (
	firstCheckpoint = (64 * (iota + 1)) - 1
	secondCheckpoint
	thirdCheckpoint
)

func TestProtocol14StateVerifier(t *testing.T) {
	itest := integration.NewTest(t, protocol15Config)

	sponsored := keypair.MustRandom()
	sponsoredSource := &txnbuild.SimpleAccount{
		AccountID: sponsored.Address(),
		Sequence:  1,
	}
	signer1 := keypair.MustParseAddress("GAB3CVX6C2KCDZUUS4FIMP5Z2IUDTMKMRKADOFOCNOB437VCPS5DRG3Z")
	signer2 := keypair.MustParseAddress("GBUERII77FW6Z7SPOIMFQQT63PMUQRWTIAARR3QVSXTRULNQSUQVIYRC")
	signer3 := keypair.MustParseAddress("GCNLAKGPBL4H6CQRITHSDTJZ6RLTP3WY2OJZJN4EWKRSNM2A23CV6VD3")

	// The operations below create a sponsorship sandwich, sponsoring an
	// account, its trustlines, offers, data, and claimable balances.
	// Then 3 signers are created with the middle one sponsored.
	master := itest.Master()
	ops := []txnbuild.Operation{
		&txnbuild.BeginSponsoringFutureReserves{
			SponsoredID: sponsored.Address(),
		},
		&txnbuild.CreateAccount{
			Destination: sponsored.Address(),
			Amount:      "100",
		},
		&txnbuild.ChangeTrust{
			SourceAccount: sponsoredSource,
			Line:          txnbuild.CreditAsset{"ABCD", master.Address()},
			Limit:         txnbuild.MaxTrustlineLimit,
		},
		&txnbuild.ManageSellOffer{
			SourceAccount: sponsoredSource,
			Selling:       txnbuild.NativeAsset{},
			Buying:        txnbuild.CreditAsset{"ABCD", master.Address()},
			Amount:        "3",
			Price:         "1",
		},
		&txnbuild.ManageData{
			SourceAccount: sponsoredSource,
			Name:          "test",
			Value:         []byte("test"),
		},
		&txnbuild.CreateClaimableBalance{
			SourceAccount: sponsoredSource,
			Amount:        "2",
			Asset:         txnbuild.NativeAsset{},
			Destinations: []txnbuild.Claimant{
				txnbuild.NewClaimant(keypair.MustRandom().Address(), nil),
			},
		},
		&txnbuild.EndSponsoringFutureReserves{
			SourceAccount: sponsoredSource,
		},
		&txnbuild.SetOptions{
			SourceAccount: sponsoredSource,
			Signer: &txnbuild.Signer{
				Address: signer1.Address(),
				Weight:  3,
			},
		},
		&txnbuild.BeginSponsoringFutureReserves{
			SponsoredID: sponsored.Address(),
		},
		&txnbuild.SetOptions{
			SourceAccount: sponsoredSource,
			Signer: &txnbuild.Signer{
				Address: signer2.Address(),
				Weight:  3,
			},
		},
		&txnbuild.EndSponsoringFutureReserves{
			SourceAccount: sponsoredSource,
		},
		&txnbuild.SetOptions{
			SourceAccount: sponsoredSource,
			Signer: &txnbuild.Signer{
				Address: signer3.Address(),
				Weight:  3,
			},
		},
	}
	txResp, err := itest.SubmitMultiSigOperations(itest.MasterAccount(), []*keypair.Full{master, sponsored}, ops...)
	assert.NoError(t, err)
	assert.True(t, txResp.Successful)

	// Reach the first checkpoint ledger
	// Core will push to history archives *after* checkpoint ledger
	itest.CloseCoreLedgersUntilSequence(firstCheckpoint + 1)
	assert.NoError(t, err)
	for !itest.LedgerIngested(firstCheckpoint) {
		time.Sleep(time.Second)
	}

	verified := waitForStateVerifications(itest, 1)
	if !verified {
		t.Fatal("State verification not run...")
	}

	// Trigger state rebuild to check if ingesting from history archive works
	itest.RunAuroraCLICommand("expingest", "trigger-state-rebuild")

	// Wait for the second checkpoint ledger and state rebuild
	// Core will push to history archives *after* checkpoint ledger
	itest.CloseCoreLedgersUntilSequence(secondCheckpoint + 1)
	assert.NoError(t, err)

	// Wait for the third checkpoint ledger and state verification trigger
	// Core will push to history archives *after* checkpoint ledger
	itest.CloseCoreLedgersUntilSequence(thirdCheckpoint + 1)
	assert.NoError(t, err)
	for !itest.LedgerIngested(thirdCheckpoint) {
		time.Sleep(time.Second)
	}

	verified = waitForStateVerifications(itest, 2)
	if !verified {
		t.Fatal("State verification not run...")
	}
}

func waitForStateVerifications(itest *integration.Test, count int) bool {
	t := itest.CurrentTest()
	// Check metrics until state verification run
	for i := 0; i < 120; i++ {
		t.Logf("Checking metrics (%d attempt)\n", i)
		res, err := http.Get(itest.MetricsURL())
		assert.NoError(t, err)

		metricsBytes, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		assert.NoError(t, err)
		metrics := string(metricsBytes)

		stateInvalid := strings.Contains(metrics, "aurora_ingest_state_invalid 1")
		assert.False(t, stateInvalid, "State is invalid!")

		notVerifiedYet := strings.Contains(
			metrics,
			fmt.Sprintf("aurora_ingest_state_verify_duration_seconds_count %d", count-1),
		)
		if notVerifiedYet {
			time.Sleep(time.Second)
			continue
		}

		return true
	}

	return false
}
