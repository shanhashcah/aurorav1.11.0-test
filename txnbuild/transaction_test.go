package txnbuild

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/hcnet/go/keypair"
	"github.com/hcnet/go/network"
	"github.com/hcnet/go/strkey"
	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissingSourceAccount(t *testing.T) {
	_, err := NewTransaction(TransactionParams{})
	assert.EqualError(t, err, "transaction has no source account")
}

func TestIncrementSequenceNum(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), 1)
	inflation := Inflation{}

	_, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&inflation},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), sourceAccount.Sequence)

	_, err = NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&inflation},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), sourceAccount.Sequence)

	_, err = NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: false,
			Operations:           []Operation{&inflation},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), sourceAccount.Sequence)

	_, err = NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: false,
			Operations:           []Operation{&inflation},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), sourceAccount.Sequence)
}

func TestFeeNoOperations(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), 5938436531814403)

	_, err := NewTransaction(
		TransactionParams{
			SourceAccount: &sourceAccount,
			Operations:    []Operation{},
			BaseFee:       MinBaseFee,
			Timebounds:    NewInfiniteTimeout(),
		},
	)
	assert.EqualError(t, err, "transaction has no operations")
}

func TestInflation(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(3556091187167235))

	inflation := Inflation{}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&inflation},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQADKI/AAAABAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAACQAAAAAAAAAB6i5yxQAAAED9zR1l78yiBwd/o44RyE3XP7QT57VmI90qE46TjfncYyqlOaIRWpkh3qouTjV5IRPVGo6+bFWV40H1HE087FgA"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestCreateAccount(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639897))

	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAIiCNAAAAGgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAAHqLnLFAAAAQB7MjKIwNEOTIjbEeV+QIjaQp/ZpV5qpbkbDaU54gkfdTOFOUxZq66lTS5FOfP5fmPIVD8InQ00Usy2SmzFC/wc="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestPayment(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639898))

	payment := Payment{
		Destination: "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H",
		Amount:      "10",
		Asset:       NativeAsset{},
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&payment},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAIiCNAAAAGwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAQAAAAB+Ecs01jX14asC1KAsPdWlpGbYCM2PEgFZCD3NLhVZmAAAAAAAAAAABfXhAAAAAAAAAAAB6i5yxQAAAEDXBkKYzThQi3/XhJqGzfh/EjaAx/4zK3xBT1/JDNtdkk/kxn4qxHVx++xiV72lqZXxiphNwflA8C7mC8Dvim0E"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestPaymentFailsIfNoAssetSpecified(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639898))

	payment := Payment{
		Destination: "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H",
		Amount:      "10",
	}

	_, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&payment},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	expectedErrMsg := "validation failed for *txnbuild.Payment operation: Field: Asset, Error: asset is undefined"
	require.EqualError(t, err, expectedErrMsg, "An asset is required")
}

func TestBumpSequence(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(9606132444168199))

	bumpSequence := BumpSequence{
		BumpTo: 9606132444168300,
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&bumpSequence},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAGQAIiC6AAAACAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAACwAiILoAAABsAAAAAAAAAAHSh2R+AAAAQJ3Y0klngAqW69ETgBCuo8OQsx4i/6wg6WugDtOfq2hw6MElCQXJJMJRLgo2waDvwNOrWTUU9T3q95Yk0K3PHwo="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestAccountMerge(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484298))

	accountMerge := AccountMerge{
		Destination: "GAS4V4O2B7DW5T7IQRPEEVCRXMDZESKISR7DVIGKZQYYV3OSQ5SH5LVP",
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&accountMerge},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAACwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAACAAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAAAAAAAB6i5yxQAAAECf1HDoBOuPhkKcL9Ll12to6yrRXZg7MmemWf7nca8j0vHDQpti+/OIsT2DOF0YJKEAncQt2CvJ+cefgly8668A"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageData(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(3556091187167235))

	manageData := ManageData{
		Name:  "Fruit preference",
		Value: []byte("Apple"),
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&manageData},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQADKI/AAAABAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAACgAAABBGcnVpdCBwcmVmZXJlbmNlAAAAAQAAAAVBcHBsZQAAAAAAAAAAAAAB6i5yxQAAAEDtRCyQRKKgQ8iLEu7kicHtSzoplfxPtPTMhdRv/sq8UoIBVTxIw+S13Jv+jzs3tyLDLiGCVNXreUNlbfX+980K"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageDataRemoveDataEntry(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484309))

	manageData := ManageData{
		Name: "Fruit preference",
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&manageData},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAFgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAACgAAABBGcnVpdCBwcmVmZXJlbmNlAAAAAAAAAAAAAAAB6i5yxQAAAEAfK5BWYLX31E3QgEs8Cd40XDAsx6VW27hW8nuyotnS2qOruXdmks89zNroDSYzRTH0rt4qPWnQqsFSio5NFCUA"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsInflationDestination(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484315))

	setOptions := SetOptions{
		InflationDestination: NewInflationDestination(kp1.Address()),
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAHAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAEAAAAAJcrx2g/Hbs/ohF5CVFG7B5JJSJR+OqDKzDGK7dKHZH4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB6i5yxQAAAEAdES3vQ43R8yzNtsIRY2t2U/ey//NfJb1qZORDkxE6/ZZgx+/wNPxAM3gpEwc2TAotwuqVdT6xga9DSXUaz6MI"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsSetFlags(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484318))

	setOptions := SetOptions{
		SetFlags: []AccountFlag{AuthRequired, AuthRevocable},
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAHwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAAAAAAAAAAAAQAAAAMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAeoucsUAAABAn2E6acbadQNs0m2+lc5DpMpPQ/+8Y2l0cUfmSKoHSt5VpB0EZI8lQY9smiOtSd7a3aewrMCJqbY5Iy6a7dFiDg=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsClearFlags(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484319))

	setOptions := SetOptions{
		ClearFlags: []AccountFlag{AuthRequired, AuthRevocable},
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAIAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAAAAAABAAAAAwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAeoucsUAAABADVzwDfkYL6oxhdJCejMjU4jJ1mhC8Ob2DcMYb/PpotyphljM6IwsXJjAKp4tMwTLBI5fc+x/CU/cdOTpUPZ7Aw=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsMasterWeight(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484320))

	setOptions := SetOptions{
		MasterWeight: NewThreshold(10),
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAIQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAEAAAAKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAeoucsUAAABAiMR9luF2eXzLBuufIXSBMrNp5VUgCtRRI0+RgAxerFhE4RhXPlq5pcOhsCp+mTQJsVVCxIIq3I0MePGmEoBWAw=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsThresholds(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484322))

	setOptions := SetOptions{
		LowThreshold:    NewThreshold(1),
		MediumThreshold: NewThreshold(2),
		HighThreshold:   NewThreshold(2),
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAIwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAQAAAAEAAAACAAAAAQAAAAIAAAAAAAAAAAAAAAAAAAAB6i5yxQAAAEBcEXBW8xLcaMWTrVpTkJXd51ER2boDY+X2hJ3Kb9F/3XK34kFVO5N35E2A7JIlRMRYqu/AgbGAK9Lrr3x+tSEL"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsHomeDomain(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484325))

	setOptions := SetOptions{
		HomeDomain: NewHomeDomain("LovelyLumensLookLuminous.com"),
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAJgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAHExvdmVseUx1bWVuc0xvb2tMdW1pbm91cy5jb20AAAAAAAAAAAAAAAHqLnLFAAAAQLQuB2c70X8qYUYOY45s+Y8wZ/OkgDVwmUufRno0RPC9bgjsYF0hFaIdW/lHrVBIuyTf59RAgRFSa14I9HN+HgY="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsHomeDomainTooLong(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484323))

	setOptions := SetOptions{
		HomeDomain: NewHomeDomain("LovelyLumensLookLuminousLately.com"),
	}

	_, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)

	assert.Error(t, err, "A validation error was expected (home domain > 32 chars)")
}

func TestSetOptionsSigner(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484325))

	setOptions := SetOptions{
		Signer: &Signer{Address: kp1.Address(), Weight: Threshold(4)},
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAJgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAAQAAAAAAAAAAeoucsUAAABAX4JlCvsDY/ETs+/EoNK0NrO5ZrbwOK+XqR5KnPcqMSw6/xkpJoFp3laqCjcVhdCQfS/hqpdfn/DPKdTHBeDLAQ=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMultipleOperations(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(9606132444168199))

	inflation := Inflation{}
	bumpSequence := BumpSequence{
		BumpTo: 9606132444168300,
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&inflation, &bumpSequence},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAMgAIiC6AAAACAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAACQAAAAAAAAALACIgugAAAGwAAAAAAAAAAdKHZH4AAABA5n9wINh8OTXZb8yaaYeCpvmjSsvJH80tRAISFXSicFJzFVoTqX3V0of2npBFXaMV4dvoqKHK8XbZFgGX0t7DBQ=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestChangeTrust(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484348))

	changeTrust := ChangeTrust{
		Line:  CreditAsset{"ABCD", kp1.Address()},
		Limit: "10",
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&changeTrust},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAPQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABgAAAAFBQkNEAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAAAX14QAAAAAAAAAAAeoucsUAAABA+2EndVXXsBHbRFEQGLsgsvHVm8wCxH9byZ/PP4AhEeAjXSL6IzhGnyRIWIc2SYXRu6GvveVI3yPbzCTvKnVjCg=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestChangeTrustNativeAssetNotAllowed(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484348))

	changeTrust := ChangeTrust{
		Line:  NativeAsset{},
		Limit: "10",
	}

	_, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&changeTrust},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)

	expectedErrMsg := "validation failed for *txnbuild.ChangeTrust operation: Field: Line, Error: native (XLM) asset type is not allowed"
	require.EqualError(t, err, expectedErrMsg, "No trustlines for native assets")
}

func TestChangeTrustDeleteTrustline(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484354))

	issuedAsset := CreditAsset{"ABCD", kp1.Address()}
	removeTrust := RemoveTrustlineOp(issuedAsset)

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&removeTrust},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAAQwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABgAAAAFBQkNEAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAAAAAAAAAAAAAAAAAAeoucsUAAABAoHdsJCt+XIr73+jSqbEhQ8iqXcqP3LO8C/kWH2dgQj+3hq1FKbthn0BbX/x5umgcE+pyfnTjU0j158qew6tfCw=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestAllowTrust(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484366))

	issuedAsset := CreditAsset{"ABCD", kp1.Address()}
	allowTrust := AllowTrust{
		Trustor:   kp1.Address(),
		Type:      issuedAsset,
		Authorize: true,
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&allowTrust},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAATwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABwAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAAFBQkNEAAAAAQAAAAAAAAAB6i5yxQAAAEBhgUiorWMaRzTGlVThNgiMpVhSYMKsY4cJyL1mrkkpC2qZ7Q9fBtaTGoS27PC6nK9/nBLOVoyyOHgYculoiYQJ"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestAllowTrustNoIssuer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484366))

	issuedAsset := CreditAsset{Code: "XYZ"}
	allowTrust := AllowTrust{
		Trustor:   kp1.Address(),
		Type:      issuedAsset,
		Authorize: true,
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&allowTrust},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAACS7AAAATwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABwAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAAFYWVoAAAAAAQAAAAAAAAAB6i5yxQAAAEDvJnLIv/kTm6yraPLQAbTfEcFIutdNRagQ08KjEKeITbro8PkhhBWgQmCzcP7uNAxxUUKATYus3ASmwUoPoFcB"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageSellOfferNewOffer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761092))

	selling := NativeAsset{}
	buying := CreditAsset{"ABCD", kp0.Address()}
	sellAmount := "100"
	price := "0.01"
	createOffer, err := CreateOfferOp(selling, buying, sellAmount, price)
	check(err)

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createOffer},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAGQAACVqAAAABQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAwAAAAAAAAABQUJDRAAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAAA7msoAAAAAAQAAAGQAAAAAAAAAAAAAAAAAAAAB0odkfgAAAEAJl3+AZx/G1ocvk58X/u84LIo+6VdG+1wuK6n2FovWSFVGonVj26xYWlo4kG12AdTSncdF44nc5HAIDCJy6g4L"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageSellOfferDeleteOffer(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761105))

	offerID := int64(2921622)
	deleteOffer, err := DeleteOfferOp(offerID)
	check(err)

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&deleteOffer},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAGQAACVqAAAAEgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAwAAAAAAAAABRkFLRQAAAABBB4BkxJWGYvNgJBoiXUo2tjgWlNmhHMMKdwGN7RSdsQAAAAAAAAAAAAAAAQAAAAEAAAAAACyUlgAAAAAAAAAB0odkfgAAAEAUo0X6chACDJ0UDj39QQTsfBxQui5um8cXZY2noJ1LbPEpliRkG2TeWvD0Bszk8BnQSgZPV/XfgSKwVXN5MskO"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageSellOfferUpdateOffer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761097))

	selling := NativeAsset{}
	buying := CreditAsset{"ABCD", kp0.Address()}
	sellAmount := "50"
	price := "0.02"
	offerID := int64(2497628)
	updateOffer, err := UpdateOfferOp(selling, buying, sellAmount, price, offerID)
	check(err)

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&updateOffer},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAGQAACVqAAAACgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAwAAAAAAAAABQUJDRAAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAAAdzWUAAAAAAQAAADIAAAAAACYcXAAAAAAAAAAB0odkfgAAAEAMKloNgv6Hv8x+A92O/8oOUpR6hbxegN4+hkGfTT4d0TqrraLy8gBOtvq718TO4akjc9UbceH6yWjoTmm4egwI"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestCreatePassiveSellOffer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761100))

	createPassiveOffer := CreatePassiveSellOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{"ABCD", kp0.Address()},
		Amount:  "10",
		Price:   "1.0",
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createPassiveOffer},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAGQAACVqAAAADQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABAAAAAAAAAABQUJDRAAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAAAF9eEAAAAAAQAAAAEAAAAAAAAAAdKHZH4AAABAIFA+zNVC+8dptptusks3Eh8SJ3jk+/6/rPxy7IFg4+gpqUotRma5b7QR/gjbnoAsL1tPU0WSYae2y8sJGhQqCg=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestPathPayment(t *testing.T) {
	kp0 := newKeypair0()
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(187316408680450))

	abcdAsset := CreditAsset{"ABCD", kp0.Address()}
	pathPayment := PathPayment{
		SendAsset:   NativeAsset{},
		SendMax:     "10",
		Destination: kp2.Address(),
		DestAsset:   NativeAsset{},
		DestAmount:  "1",
		Path:        []Asset{abcdAsset},
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&pathPayment},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp2,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAB+Ecs01jX14asC1KAsPdWlpGbYCM2PEgFZCD3NLhVZmAAAAGQAAKpdAAAAAwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAgAAAAAAAAAABfXhAAAAAAB+Ecs01jX14asC1KAsPdWlpGbYCM2PEgFZCD3NLhVZmAAAAAAAAAAAAJiWgAAAAAEAAAABQUJDRAAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAAAAAAABLhVZmAAAAEBdpC1C/0aBSMtXJrfhl3Vp9rQ1IyWFd2MBeAPNsyAYamEjuqIDqCzzUbd8PiBggIH0eEPZaWsfsAl1qEBER0sO"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMemoText(t *testing.T) {
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(3556099777101824))

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&BumpSequence{BumpTo: 1}},
			Memo:                 MemoText("Twas brillig"),
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp2,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAB+Ecs01jX14asC1KAsPdWlpGbYCM2PEgFZCD3NLhVZmAAAAGQADKJBAAAAAQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAxUd2FzIGJyaWxsaWcAAAABAAAAAAAAAAsAAAAAAAAAAQAAAAAAAAABLhVZmAAAAECC0/P+zBk5lpH4zIumNt59nFVrPiDGOu8TrJE4r0mXoae8Fmg1yyHQm3Yo5huuPjc/nzwU/R2DKkkQ3C4mWA0N"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMemoID(t *testing.T) {
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(3428320205078528))

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&BumpSequence{BumpTo: 1}},
			Memo:                 MemoID(314159),
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp2,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAB+Ecs01jX14asC1KAsPdWlpGbYCM2PEgFZCD3NLhVZmAAAAGQADC4KAAAAAQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAABMsvAAAAAQAAAAAAAAALAAAAAAAAAAEAAAAAAAAAAS4VWZgAAABAOT/1f1XoeqY14+wp6rVgwE4fCCPnItc9/85jZN++Fy7lS88e40b3ufQCpzzMCD8AyfHF8BCs/Pn2DiJHxCPQCQ=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMemoHash(t *testing.T) {
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(3428320205078528))

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&BumpSequence{BumpTo: 1}},
			Memo:                 MemoHash([32]byte{0x01}),
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp2,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAB+Ecs01jX14asC1KAsPdWlpGbYCM2PEgFZCD3NLhVZmAAAAGQADC4KAAAAAQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAwEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAALAAAAAAAAAAEAAAAAAAAAAS4VWZgAAABAIGrmlKahBhdVXl2LZGINCNfUAtxiVawjzqgxzyHV7xpEPTft1besnyiDdLBP1+Tbg+hYQK0N2ncL2XmjQ4pcDQ=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMemoReturn(t *testing.T) {
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(3428320205078528))

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&BumpSequence{BumpTo: 1}},
			Memo:                 MemoReturn([32]byte{0x01}),
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp2,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAB+Ecs01jX14asC1KAsPdWlpGbYCM2PEgFZCD3NLhVZmAAAAGQADC4KAAAAAQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAABAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAALAAAAAAAAAAEAAAAAAAAAAS4VWZgAAABALixU7p2NPKW1iqJqaHqR3Wsy5q+7nj1EjswOD99/klUSlorvodrZ4DrD/IYGvsKSyV0/Zf9LjEN4s4kVVK4dCg=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageBuyOfferNewOffer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761092))

	buyOffer := ManageBuyOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{"ABCD", kp0.Address()},
		Amount:  "100",
		Price:   "0.01",
		OfferID: 0,
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&buyOffer},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAGQAACVqAAAABQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAADAAAAAAAAAABQUJDRAAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAAA7msoAAAAAAQAAAGQAAAAAAAAAAAAAAAAAAAAB0odkfgAAAEB8LqK1uwbwcCQM/hE0rXng2fVCoaMdctQaiS72iJFkq+azWzqYpo1kMa1DUKMvvsJrWPLYjEr9yW8/A3eEE2kF"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageBuyOfferDeleteOffer(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761105))

	buyOffer := ManageBuyOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{"ABCD", kp1.Address()},
		Amount:  "0",
		Price:   "0.01",
		OfferID: int64(2921622),
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&buyOffer},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAGQAACVqAAAAEgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAADAAAAAAAAAABQUJDRAAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAAAAAAAAAAAAAQAAAGQAAAAAACyUlgAAAAAAAAAB0odkfgAAAECLZ6PnKZlGBb8S3GFWg6J01d3Zr88/tki8yka2KFzqivMAmY3D/5IMzzJl4U7RdrYEPam9KwCGKR/f647WTwYG"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageBuyOfferUpdateOffer(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761097))

	buyOffer := ManageBuyOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{"ABCD", kp1.Address()},
		Amount:  "50",
		Price:   "0.02",
		OfferID: int64(2921622),
	}

	received, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&buyOffer},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp1,
	)
	assert.NoError(t, err)

	expected := "AAAAAgAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAGQAACVqAAAACgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAADAAAAAAAAAABQUJDRAAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAAAdzWUAAAAAAQAAADIAAAAAACyUlgAAAAAAAAAB0odkfgAAAECv7GrE8YDar5M93RmgzslIH2vVAAJlAZoIsmkFNXTJTTb01R9Q+z0Cl5E6KFpm+qiuxHvL2kwhVOoBpkoYQPcB"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestBuildChallengeTx(t *testing.T) {
	kp0 := newKeypair0()

	{
		// 1 minute timebound
		tx, err := BuildChallengeTx(kp0.Seed(), kp0.Address(), "SDF", network.TestNetworkPassphrase, time.Minute)
		assert.NoError(t, err)
		txeBase64, err := tx.Base64()
		assert.NoError(t, err)
		var txXDR xdr.TransactionEnvelope
		err = xdr.SafeUnmarshalBase64(txeBase64, &txXDR)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), txXDR.SeqNum(), "sequence number should be 0")
		assert.Equal(t, uint32(100), txXDR.Fee(), "Fee should be 100")
		assert.Equal(t, 1, len(txXDR.Operations()), "number operations should be 1")
		timeDiff := txXDR.TimeBounds().MaxTime - txXDR.TimeBounds().MinTime
		assert.Equal(t, int64(60), int64(timeDiff), "time difference should be 300 seconds")
		op := txXDR.Operations()[0]
		assert.Equal(t, xdr.OperationTypeManageData, op.Body.Type, "operation type should be manage data")
		assert.Equal(t, xdr.String64("SDF auth"), op.Body.ManageDataOp.DataName, "DataName should be 'SDF auth'")
		assert.Equal(t, 64, len(*op.Body.ManageDataOp.DataValue), "DataValue should be 64 bytes")

	}

	{
		// 5 minutes timebound
		tx, err := BuildChallengeTx(kp0.Seed(), kp0.Address(), "SDF1", network.TestNetworkPassphrase, time.Duration(5*time.Minute))
		assert.NoError(t, err)
		txeBase64, err := tx.Base64()
		assert.NoError(t, err)
		var txXDR1 xdr.TransactionEnvelope
		err = xdr.SafeUnmarshalBase64(txeBase64, &txXDR1)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), txXDR1.SeqNum(), "sequence number should be 0")
		assert.Equal(t, uint32(100), txXDR1.Fee(), "Fee should be 100")
		assert.Equal(t, 1, len(txXDR1.Operations()), "number operations should be 1")

		timeDiff := txXDR1.TimeBounds().MaxTime - txXDR1.TimeBounds().MinTime
		assert.Equal(t, int64(300), int64(timeDiff), "time difference should be 300 seconds")
		op1 := txXDR1.Operations()[0]
		assert.Equal(t, xdr.OperationTypeManageData, op1.Body.Type, "operation type should be manage data")
		assert.Equal(t, xdr.String64("SDF1 auth"), op1.Body.ManageDataOp.DataName, "DataName should be 'SDF1 auth'")
		assert.Equal(t, 64, len(*op1.Body.ManageDataOp.DataValue), "DataValue should be 64 bytes")
	}

	//transaction with infinite timebound
	_, err := BuildChallengeTx(kp0.Seed(), kp0.Address(), "sdf", network.TestNetworkPassphrase, 0)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "provided timebound must be at least 1s (300s is recommended)")
	}
}

func TestHashHex(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639897))

	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, kp0)
	assert.NoError(t, err)

	txeB64, err := tx.Base64()
	assert.NoError(t, err)
	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAIiCNAAAAGgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAAHqLnLFAAAAQB7MjKIwNEOTIjbEeV+QIjaQp/ZpV5qpbkbDaU54gkfdTOFOUxZq66lTS5FOfP5fmPIVD8InQ00Usy2SmzFC/wc="
	assert.Equal(t, expected, txeB64, "Base 64 XDR should match")

	hashHex, err := tx.HashHex(network.TestNetworkPassphrase)
	assert.NoError(t, err)
	expected = "1b3905ba8c3c0ecc68ae812f2d77f27c697195e8daf568740fc0f5662f65f759"
	assert.Equal(t, expected, hashHex, "hex encoded hash should match")

	txEnv, err := tx.TxEnvelope()
	assert.NoError(t, err)
	assert.NotNil(t, txEnv, "transaction xdr envelope should not be nil")
	sourceAccountFromEnv := txEnv.SourceAccount().ToAccountId()
	assert.Equal(t, sourceAccount.AccountID, sourceAccountFromEnv.Address())
	assert.Equal(t, uint32(100), txEnv.Fee())
	assert.Equal(t, sourceAccount.Sequence, int64(txEnv.SeqNum()))
	assert.Equal(t, xdr.MemoTypeMemoNone, txEnv.Memo().Type)
	assert.Len(t, txEnv.Operations(), 1)
}

func TestTransactionFee(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639897))

	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	assert.Equal(t, int64(100), tx.BaseFee(), "Transaction base fee should match")
	assert.Equal(t, int64(100), tx.MaxFee(), "Transaction fee should match")

	tx, err = NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              500,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	assert.Equal(t, int64(500), tx.BaseFee(), "Transaction base fee should match")
	assert.Equal(t, int64(500), tx.MaxFee(), "Transaction fee should match")

	tx, err = tx.Sign(network.TestNetworkPassphrase, kp0)
	assert.NoError(t, err)

	txeB64, err := tx.Base64()
	assert.NoError(t, err)
	expected := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAfQAIiCNAAAAGwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAAHqLnLFAAAAQJ3OvWisOnYNS5R8ZCHrSmbvDrvIYG4+JiAldLYjiXroqvA74r0pQJ4Jw/hZVSGqLZoPIt3RMwYPi3C5xvVLbQU="
	assert.Equal(t, expected, txeB64, "Base 64 XDR should match")
}

func TestPreAuthTransaction(t *testing.T) {
	// Address: GDK3YEHGI3ORGVO7ZEV2XF4SV5JU3BOKHMHPP4QFJ74ZRIIRROZ7ITOJ
	kp0 := newKeypair("SDY4PF6F6OWWERZT6OL2LVNREHUGHKALUI5W4U2JK4GAKPAC2RM43OAU")
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(4353383146192898)) // sequence number is in the future

	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	// build transaction to be submitted in the future.
	txFuture, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	// save the hash of the future transaction.
	txFutureHash, err := txFuture.Hash(network.TestNetworkPassphrase)
	assert.NoError(t, err)

	// sign transaction without keypairs, the hash of the future transaction on the account
	//  will be used for authorisation.
	txFuture, err = txFuture.Sign(network.TestNetworkPassphrase)
	assert.NoError(t, err)

	txeFutureB64, err := txFuture.Base64()
	assert.NoError(t, err)
	expected := "AAAAAgAAAADVvBDmRt0TVd/JK6uXkq9TTYXKOw738gVP+ZihEYuz9AAAAGQAD3dhAAAAAwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAAA="
	assert.Equal(t, expected, txeFutureB64, "Base 64 XDR should match")

	//encode the txFutureHash as a hcnet HashTx signer key.
	preAuth, err := strkey.Encode(strkey.VersionByteHashTx, txFutureHash[:])
	assert.NoError(t, err)

	// set sequence number to the current number.
	sourceAccount.Sequence = int64(4353383146192897)

	// add hash of future transaction as signer to account
	setOptions := SetOptions{
		Signer: &Signer{Address: preAuth, Weight: Threshold(2)},
	}

	// build a transaction to add the hash of the future transaction as a signer on the account.
	txNow, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              500,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	assert.Equal(t, int64(500), txNow.MaxFee(), "Transaction fee should match")
	txNow, err = txNow.Sign(network.TestNetworkPassphrase, kp0)
	assert.NoError(t, err)

	txeNowB64, err := txNow.Base64()
	assert.NoError(t, err)
	expected = "AAAAAgAAAADVvBDmRt0TVd/JK6uXkq9TTYXKOw738gVP+ZihEYuz9AAAAfQAD3dhAAAAAgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAGPXOumKQj5/MjKKSmjQXe4G4g9nK/mkyzmmROMIZnjtQAAAAIAAAAAAAAAARGLs/QAAABAutrV0Cg03KwfFbzkCGiNxAldLsqQZKRjbsqHZyy2Nu4ouEDHQeIOKLWCLymOp21kKmGGqTYekPXVbGHyujh0DA=="
	assert.Equal(t, expected, txeNowB64, "Base 64 XDR should match")
	// Note: txeFutureB64 can be submitted to the network after txeNowB64 has been applied to the account
}

func TestHashXTransaction(t *testing.T) {
	// 256 bit preimage
	preimage := "this is a preimage for hashx transactions on the hcnet network"

	preimageHash := sha256.Sum256([]byte(preimage))

	//encode preimageHash as a hcnet HashX signer key
	hashx, err := strkey.Encode(strkey.VersionByteHashX, preimageHash[:])
	assert.NoError(t, err)

	// add hashx as signer to the account
	setOptions := SetOptions{
		Signer: &Signer{Address: hashx, Weight: Threshold(1)},
	}

	// Address: GDK3YEHGI3ORGVO7ZEV2XF4SV5JU3BOKHMHPP4QFJ74ZRIIRROZ7ITOJ
	kp0 := newKeypair("SDY4PF6F6OWWERZT6OL2LVNREHUGHKALUI5W4U2JK4GAKPAC2RM43OAU")
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(4353383146192899))

	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&setOptions},
			BaseFee:              500,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, int64(500), tx.MaxFee(), "Transaction fee should match")

	tx, err = tx.Sign(network.TestNetworkPassphrase, kp0)
	assert.NoError(t, err)

	txeB64, err := tx.Base64()
	assert.NoError(t, err)

	expected := "AAAAAgAAAADVvBDmRt0TVd/JK6uXkq9TTYXKOw738gVP+ZihEYuz9AAAAfQAD3dhAAAABAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAL7JYG+aCH7iEhT/BWL06rHIhtYklHqyQdwLuk9li6jBQAAAAEAAAAAAAAAARGLs/QAAABAhwcHwm3DsBcqCCy1uzmXo73W7FTxMAes+qHABuHERruvb1ygqwRWA9pjHSUQnoJYCYH4GhY9qrIQYC/MkNeFBw=="
	assert.Equal(t, expected, txeB64, "Base 64 XDR should match")

	// build a transaction to test hashx signer
	payment := Payment{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
		Asset:       NativeAsset{},
	}

	sourceAccount.Sequence = int64(4353383146192902)

	tx, err = NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&payment},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	// sign transaction with the preimage
	tx, err = tx.SignHashX([]byte(preimage))
	assert.NoError(t, err)

	txeB64, err = tx.Base64()
	assert.NoError(t, err)
	expected = "AAAAAgAAAADVvBDmRt0TVd/JK6uXkq9TTYXKOw738gVP+ZihEYuz9AAAAGQAD3dhAAAABwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAQAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAAAAAABfXhAAAAAAAAAAABli6jBQAAAEB0aGlzIGlzIGEgcHJlaW1hZ2UgZm9yIGhhc2h4IHRyYW5zYWN0aW9ucyBvbiB0aGUgc3RlbGxhciBuZXR3b3Jr"
	assert.Equal(t, expected, txeB64, "Base 64 XDR should match")

}

func TestFromXDR(t *testing.T) {
	txeB64 := "AAAAACYWIvM98KlTMs0IlQBZ06WkYpZ+gILsQN6ega0++I/sAAAAZAAXeEkAAAABAAAAAAAAAAEAAAAQMkExVjZKNTcwM0c0N1hIWQAAAAEAAAABAAAAACYWIvM98KlTMs0IlQBZ06WkYpZ+gILsQN6ega0++I/sAAAAAQAAAADMSEvcRKXsaUNna++Hy7gWm/CfqTjEA7xoGypfrFGUHAAAAAAAAAACCPHRAAAAAAAAAAABPviP7AAAAEBu6BCKf4WZHPum5+29Nxf6SsJNN8bgjp1+e1uNBaHjRg3rdFZYgUqEqbHxVEs7eze3IeRbjMZxS3zPf/xwJCEI"

	tx, err := TransactionFromXDR(txeB64)
	assert.NoError(t, err)
	newTx, ok := tx.Transaction()
	assert.True(t, ok)
	_, ok = tx.FeeBump()
	assert.False(t, ok)

	assert.Equal(t, "GATBMIXTHXYKSUZSZUEJKACZ2OS2IYUWP2AIF3CA32PIDLJ67CH6Y5UY", newTx.SourceAccount().AccountID, "source accounts should match")
	assert.Equal(t, int64(100), newTx.BaseFee(), "Base fee should match")
	sa := newTx.SourceAccount()
	assert.Equal(t, int64(6606179392290817), sa.Sequence, "Sequence number should match")
	assert.Equal(t, 1, len(newTx.Operations()), "Operations length should match")
	assert.IsType(t, newTx.Operations()[0], &Payment{}, "Operation types should match")
	paymentOp, ok1 := newTx.Operations()[0].(*Payment)
	assert.Equal(t, true, ok1)
	assert.Equal(t, "GATBMIXTHXYKSUZSZUEJKACZ2OS2IYUWP2AIF3CA32PIDLJ67CH6Y5UY", paymentOp.SourceAccount.GetAccountID(), "Operation source should match")
	assert.Equal(t, "GDGEQS64ISS6Y2KDM5V67B6LXALJX4E7VE4MIA54NANSUX5MKGKBZM5G", paymentOp.Destination, "Operation destination should match")
	assert.Equal(t, "874.0000000", paymentOp.Amount, "Operation amount should match")

	txeB64 = "AAAAAGigiN2q4qBXAERImNEncpaADylyBRtzdqpEsku6CN0xAAABkAAADXYAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAABAAAABm5ldyB0eAAAAAAAAgAAAAEAAAAA+Q2efEMLNGF4i+aYfutUXGMSlf8tNevKeS1Jl/oCVGkAAAAGAAAAAVVTRAAAAAAAaKCI3arioFcAREiY0SdyloAPKXIFG3N2qkSyS7oI3TF//////////wAAAAAAAAAKAAAABHRlc3QAAAABAAAABXZhbHVlAAAAAAAAAAAAAAA="

	tx2, err := TransactionFromXDR(txeB64)
	assert.NoError(t, err)
	newTx2, ok := tx2.Transaction()
	assert.True(t, ok)
	_, ok = tx2.FeeBump()
	assert.False(t, ok)

	assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", newTx2.SourceAccount().AccountID, "source accounts should match")
	assert.Equal(t, int64(200), newTx2.BaseFee(), "Base fee should match")
	assert.Equal(t, int64(14800457302017), newTx2.SourceAccount().Sequence, "Sequence number should match")

	memo, ok := newTx2.Memo().(MemoText)
	assert.Equal(t, true, ok)
	assert.Equal(t, MemoText("new tx"), memo, "memo should match")
	assert.Equal(t, 2, len(newTx2.Operations()), "Operations length should match")
	assert.IsType(t, newTx2.Operations()[0], &ChangeTrust{}, "Operation types should match")
	assert.IsType(t, newTx2.Operations()[1], &ManageData{}, "Operation types should match")
	op1, ok1 := newTx2.Operations()[0].(*ChangeTrust)
	assert.Equal(t, true, ok1)
	assert.Equal(t, "GD4Q3HT4IMFTIYLYRPTJQ7XLKROGGEUV74WTL26KPEWUTF72AJKGSJS7", op1.SourceAccount.GetAccountID(), "Operation source should match")
	assetType, err := op1.Line.GetType()
	assert.NoError(t, err)

	assert.Equal(t, AssetTypeCreditAlphanum4, assetType, "Asset type should match")
	assert.Equal(t, "USD", op1.Line.GetCode(), "Asset code should match")
	assert.Equal(t, "GBUKBCG5VLRKAVYAIREJRUJHOKLIADZJOICRW43WVJCLES52BDOTCQZU", op1.Line.GetIssuer(), "Asset issuer should match")
	assert.Equal(t, "922337203685.4775807", op1.Limit, "trustline limit should match")

	op2, ok2 := newTx2.Operations()[1].(*ManageData)
	assert.Equal(t, true, ok2)
	assert.Equal(t, nil, op2.SourceAccount, "Operation source should match")
	assert.Equal(t, "test", op2.Name, "Name should match")
	assert.Equal(t, "value", string(op2.Value), "Value should match")
}

func TestBuild(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	expectedUnsigned := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAIiCNAAAAGgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAAA="

	expectedSigned := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAIiCNAAAAGgAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAAHqLnLFAAAAQB7MjKIwNEOTIjbEeV+QIjaQp/ZpV5qpbkbDaU54gkfdTOFOUxZq66lTS5FOfP5fmPIVD8InQ00Usy2SmzFC/wc="

	txeB64, err := tx.Base64()
	assert.NoError(t, err)
	assert.Equal(t, expectedUnsigned, txeB64, "tx envelope should match")
	tx, err = tx.Sign(network.TestNetworkPassphrase, kp0)
	assert.NoError(t, err)
	txeB64, err = tx.Base64()
	assert.NoError(t, err)
	assert.Equal(t, expectedSigned, txeB64, "tx envelope should match")
}

func TestFromXDRBuildSignEncode(t *testing.T) {
	expectedUnsigned := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAIiCNAAAAGwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAVuZXd0eAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAAHqLnLFAAAAQAz221zc6QuNPFsmBkLMzd1QPXuNbDabMmdh3EutkV71A7DdAPiFzD0TGgm/loJ9TjOiJGpvaJdDCWDXitAT8Qo="

	expectedSigned := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAIiCNAAAAGwAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAVuZXd0eAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAALqLnLFAAAAQAz221zc6QuNPFsmBkLMzd1QPXuNbDabMmdh3EutkV71A7DdAPiFzD0TGgm/loJ9TjOiJGpvaJdDCWDXitAT8QrqLnLFAAAAQAz221zc6QuNPFsmBkLMzd1QPXuNbDabMmdh3EutkV71A7DdAPiFzD0TGgm/loJ9TjOiJGpvaJdDCWDXitAT8Qo="

	kp0 := newKeypair0()

	// test signing transaction  without modification
	tx, err := TransactionFromXDR(expectedUnsigned)
	assert.NoError(t, err)
	newTx, ok := tx.Transaction()
	assert.True(t, ok)
	_, ok = tx.FeeBump()
	assert.False(t, ok)

	//passphrase is needed for signing
	newTx, err = newTx.Sign(network.TestNetworkPassphrase, kp0)
	assert.NoError(t, err)
	txeB64, err := newTx.Base64()
	assert.NoError(t, err)
	assert.Equal(t, expectedSigned, txeB64, "tx envelope should match")

	// test signing transaction  with modification
	expectedSigned2 := "AAAAAgAAAADg3G3hclysZlFitS+s5zWyiiJD5B0STWy5LXCj6i5yxQAAAGQAIiCNAAAAHAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAVuZXd0eAAAAAAAAAEAAAAAAAAAAAAAAACE4N7avBtJL576CIWTzGCbGPvSlVfMQAOjcYbSsSF2VAAAAAAF9eEAAAAAAAAAAAHqLnLFAAAAQFu7obEnMmrp+1Pnz/8o3IUIOWJ6rVTsJO1dAYapN3/zVjCNW3/JzgewGrKNWjPelF7BNRhk5lx93CFGdHDJ/Ac="
	tx, err = TransactionFromXDR(expectedUnsigned)
	assert.NoError(t, err)
	newTx, ok = tx.Transaction()
	assert.True(t, ok)
	_, ok = tx.FeeBump()
	assert.False(t, ok)

	txeB64, err = newSignedTransaction(
		TransactionParams{
			SourceAccount: &SimpleAccount{
				AccountID: newTx.SourceAccount().AccountID,
				Sequence:  newTx.SourceAccount().Sequence + 1,
			},
			IncrementSequenceNum: false,
			Operations:           newTx.Operations(),
			BaseFee:              newTx.BaseFee(),
			Memo:                 MemoText("newtx"),
			Timebounds:           newTx.Timebounds(),
		},
		network.TestNetworkPassphrase,
		kp0,
	)
	assert.NoError(t, err)
	assert.Equal(t, expectedSigned2, txeB64, "tx envelope should match")
}

func TestSignWithSecretKey(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	txSource := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	tx1Source := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	opSource := NewSimpleAccount(kp1.Address(), 0)
	createAccount := CreateAccount{
		Destination:   "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:        "10",
		SourceAccount: &opSource,
	}

	expected, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0, kp1,
	)
	assert.NoError(t, err)

	tx1, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &tx1Source,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	tx1, err = tx1.SignWithKeyString(
		network.TestNetworkPassphrase,
		"SBPQUZ6G4FZNWFHKUWC5BEYWF6R52E3SEP7R3GWYSM2XTKGF5LNTWW4R", ""+
			"SBMSVD4KKELKGZXHBUQTIROWUAPQASDX7KEJITARP4VMZ6KLUHOGPTYW",
	)
	assert.NoError(t, err)

	actual, err := tx1.Base64()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual, "base64 xdr should match")
}

func TestAddSignatureBase64(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	txSource := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	tx1Source := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	opSource := NewSimpleAccount(kp1.Address(), 0)
	createAccount := CreateAccount{
		Destination:   "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:        "10",
		SourceAccount: &opSource,
	}

	expected, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
		network.TestNetworkPassphrase,
		kp0, kp1,
	)
	assert.NoError(t, err)

	tx1, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &tx1Source,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	tx1, err = tx1.AddSignatureBase64(
		network.TestNetworkPassphrase,
		"GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3",
		"TVogR6tbrWLnOc1BsP/j+Qrxpja2NWNgeRIwujECYscRdMG7AMtnb3dkCT7sqlbSM0TTzlRh7G+BcVocYBtqBw==",
	)
	assert.NoError(t, err)

	tx1, err = tx1.AddSignatureBase64(
		network.TestNetworkPassphrase,
		"GAS4V4O2B7DW5T7IQRPEEVCRXMDZESKISR7DVIGKZQYYV3OSQ5SH5LVP",
		"Iy77JteoW/FbeiuViZpgTyvrHP4BnBOeyVOjrdb5O/MpEMwcSlYXAkCBqPt4tBDil4jIcDDLhm7TsN6aUBkIBg==",
	)

	actual, err := tx1.Base64()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual, "base64 xdr should match")
}

func TestReadChallengeTx_validSignedByServerAndClient(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}

	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP, clientKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.NoError(t, err)
}

func TestReadChallengeTx_validSignedByServer(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.NoError(t, err)
}

func TestReadChallengeTx_invalidNotSignedByServer(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.EqualError(t, err, "transaction not signed by "+serverKP.Address())
}

func TestReadChallengeTx_invalidCorrupted(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	tx64 = strings.ReplaceAll(tx64, "A", "B")
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Nil(t, readTx)
	assert.Equal(t, "", readClientAccountID)
	assert.EqualError(
		t,
		err,
		"could not parse challenge: unable to unmarshal transaction envelope: "+
			"xdr:decode: switch '68174086' is not valid enum value for union",
	)
}

func TestReadChallengeTx_invalidServerAccountIDMismatch(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(newKeypair2().Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, "", readClientAccountID)
	assert.EqualError(t, err, "transaction source account is not equal to server's account")
}

func TestReadChallengeTx_invalidSeqNoNotZero(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), 1234)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, "", readClientAccountID)
	assert.EqualError(t, err, "transaction sequence number must be 0")
}

func TestReadChallengeTx_invalidTimeboundsInfinite(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, "", readClientAccountID)
	assert.EqualError(t, err, "transaction requires non-infinite timebounds")
}

func TestReadChallengeTx_invalidTimeboundsOutsideRange(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimebounds(0, 100),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, "", readClientAccountID)
	assert.Error(t, err)
	assert.Regexp(t, "transaction is not within range of the specified timebounds", err.Error())
}

func TestReadChallengeTx_invalidOperationWrongType(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := BumpSequence{
		SourceAccount: &opSource,
		BumpTo:        0,
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(300),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, "", readClientAccountID)
	assert.EqualError(t, err, "operation type should be manage_data")
}

func TestReadChallengeTx_invalidOperationNoSourceAccount(t *testing.T) {
	serverKP := newKeypair0()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	op := ManageData{
		Name:  "testserver auth",
		Value: []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(300),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	_, _, err = ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.EqualError(t, err, "operation should have a source account")
}

func TestReadChallengeTx_invalidDataValueWrongEncodedLength(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 45))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(300),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.EqualError(t, err, "random nonce encoded as base64 should be 64 bytes long")
}

func TestReadChallengeTx_invalidDataValueCorruptBase64(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA?AAAAAAAAAAAAAAAAAAAAAAAAAA"),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(300),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.EqualError(t, err, "failed to decode random nonce provided in manage_data operation: illegal base64 data at input byte 37")
}

func TestReadChallengeTx_invalidDataValueWrongByteLength(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 47))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	assert.NoError(t, err)

	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.EqualError(t, err, "random nonce before encoding as base64 should be 48 bytes long")
}

func TestReadChallengeTx_acceptsV0AndV1Transactions(t *testing.T) {
	kp0 := newKeypair0()
	tx, err := BuildChallengeTx(
		kp0.Seed(),
		kp0.Address(),
		"SDF",
		network.TestNetworkPassphrase,
		time.Hour,
	)
	assert.NoError(t, err)

	originalHash, err := tx.HashHex(network.TestNetworkPassphrase)
	assert.NoError(t, err)

	v1Challenge, err := marshallBase64(tx.envelope, tx.Signatures())
	assert.NoError(t, err)

	convertToV0(tx)
	v0Challenge, err := marshallBase64(tx.envelope, tx.Signatures())
	assert.NoError(t, err)

	for _, challenge := range []string{v1Challenge, v0Challenge} {
		parsedTx, clientAccountID, err := ReadChallengeTx(
			challenge,
			kp0.Address(),
			network.TestNetworkPassphrase,
			"SDF",
		)
		assert.NoError(t, err)

		assert.Equal(t, kp0.Address(), clientAccountID)

		hash, err := parsedTx.HashHex(network.TestNetworkPassphrase)
		assert.NoError(t, err)
		assert.Equal(t, originalHash, hash)
	}
}

func TestReadChallengeTx_forbidsFeeBumpTransactions(t *testing.T) {
	kp0 := newKeypair0()
	tx, err := BuildChallengeTx(
		kp0.Seed(),
		kp0.Address(),
		"SDF",
		network.TestNetworkPassphrase,
		time.Hour,
	)
	assert.NoError(t, err)

	kp1 := newKeypair1()
	feeBumpTx, err := NewFeeBumpTransaction(
		FeeBumpTransactionParams{
			Inner:      tx,
			FeeAccount: kp1.Address(),
			BaseFee:    3 * MinBaseFee,
		},
	)
	assert.NoError(t, err)

	feeBumpTx, err = feeBumpTx.Sign(network.TestNetworkPassphrase, kp1)
	assert.NoError(t, err)

	challenge, err := feeBumpTx.Base64()
	assert.NoError(t, err)
	_, _, err = ReadChallengeTx(
		challenge,
		kp0.Address(),
		network.TestNetworkPassphrase,
		"SDF",
	)
	assert.EqualError(t, err, "challenge cannot be a fee bump transaction")
}

func TestReadChallengeTx_forbidsMuxedAccounts(t *testing.T) {
	kp0 := newKeypair0()
	tx, err := BuildChallengeTx(
		kp0.Seed(),
		kp0.Address(),
		"SDF",
		network.TestNetworkPassphrase,
		time.Hour,
	)

	env, err := tx.TxEnvelope()
	assert.NoError(t, err)
	aid := xdr.MustAddress(kp0.Address())
	muxedAccount := xdr.MuxedAccount{
		Type: xdr.CryptoKeyTypeKeyTypeMuxedEd25519,
		Med25519: &xdr.MuxedAccountMed25519{
			Id:      0xcafebabe,
			Ed25519: *aid.Ed25519,
		},
	}
	*env.V1.Tx.Operations[0].SourceAccount = muxedAccount

	challenge, err := marshallBase64(env, env.Signatures())
	assert.NoError(t, err)

	_, _, err = ReadChallengeTx(
		challenge,
		kp0.Address(),
		network.TestNetworkPassphrase,
		"SDF",
	)
	errorMessage := "only valid Ed25519 accounts are allowed in challenge transactions"
	assert.Contains(t, err.Error(), errorMessage)
}

func TestReadChallengeTx_doesNotVerifyHomeDomain(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "notvalidated")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.NoError(t, err)
}

func TestReadChallengeTx_allowsAdditionalManageDataOpsWithSourceAccountSetToServerAccount(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := ManageData{
		SourceAccount: &txSource,
		Name:          "a key",
		Value:         []byte("a value"),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.NoError(t, err)
}

func TestReadChallengeTx_disallowsAdditionalManageDataOpsWithoutSourceAccountSetToServerAccount(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := ManageData{
		SourceAccount: &opSource,
		Name:          "a key",
		Value:         []byte("a value"),
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	_, _, err = ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.EqualError(t, err, "subsequent operations are unrecognized")
}

func TestReadChallengeTx_disallowsAdditionalOpsOfOtherTypes(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := BumpSequence{
		SourceAccount: &txSource,
		BumpTo:        0,
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(300),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, serverKP)
	assert.NoError(t, err)
	tx64, err := tx.Base64()
	require.NoError(t, err)
	readTx, readClientAccountID, err := ReadChallengeTx(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.Equal(t, tx, readTx)
	assert.Equal(t, clientKP.Address(), readClientAccountID)
	assert.EqualError(t, err, "operation type should be manage_data")
}

func TestVerifyChallengeTxThreshold_invalidServer(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}

	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		clientKP,
	)
	assert.NoError(t, err)

	threshold := Threshold(1)
	signerSummary := SignerSummary{
		clientKP.Address(): 1,
	}
	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "transaction not signed by "+serverKP.Address())
}

func TestVerifyChallengeTxThreshold_validServerAndClientKeyMeetingThreshold(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	threshold := Threshold(1)
	signerSummary := SignerSummary{
		clientKP.Address(): 1,
	}
	wantSigners := []string{
		clientKP.Address(),
	}

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.ElementsMatch(t, wantSigners, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxThreshold_validServerAndMultipleClientKeyMeetingThreshold(t *testing.T) {
	serverKP := newKeypair0()
	clientKP1 := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP1.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP1, clientKP2,
	)
	assert.NoError(t, err)

	threshold := Threshold(3)
	signerSummary := map[string]int32{
		clientKP1.Address(): 1,
		clientKP2.Address(): 2,
	}
	wantSigners := []string{
		clientKP1.Address(),
		clientKP2.Address(),
	}

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.ElementsMatch(t, wantSigners, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxThreshold_validServerAndMultipleClientKeyMeetingThresholdSomeUnused(t *testing.T) {
	serverKP := newKeypair0()
	clientKP1 := newKeypair1()
	clientKP2 := newKeypair2()
	clientKP3 := keypair.MustRandom()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP1.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	threshold := Threshold(3)
	signerSummary := SignerSummary{
		clientKP1.Address(): 1,
		clientKP2.Address(): 2,
		clientKP3.Address(): 2,
	}
	wantSigners := []string{
		clientKP1.Address(),
		clientKP2.Address(),
	}

	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP1, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.ElementsMatch(t, wantSigners, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxThreshold_validServerAndMultipleClientKeyMeetingThresholdSomeUnusedIgnorePreauthTxHashAndXHash(t *testing.T) {
	serverKP := newKeypair0()
	clientKP1 := newKeypair1()
	clientKP2 := newKeypair2()
	clientKP3 := keypair.MustRandom()
	preauthTxHash := "TAQCSRX2RIDJNHFIFHWD63X7D7D6TRT5Y2S6E3TEMXTG5W3OECHZ2OG4"
	xHash := "XDRPF6NZRR7EEVO7ESIWUDXHAOMM2QSKIQQBJK6I2FB7YKDZES5UCLWD"
	unknownSignerType := "?ARPF6NZRR7EEVO7ESIWUDXHAOMM2QSKIQQBJK6I2FB7YKDZES5UCLWD"
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP1.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	threshold := Threshold(3)
	signerSummary := SignerSummary{
		clientKP1.Address(): 1,
		clientKP2.Address(): 2,
		clientKP3.Address(): 2,
		preauthTxHash:       10,
		xHash:               10,
		unknownSignerType:   10,
	}
	wantSigners := []string{
		clientKP1.Address(),
		clientKP2.Address(),
	}

	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP1, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.ElementsMatch(t, wantSigners, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxThreshold_invalidServerAndMultipleClientKeyNotMeetingThreshold(t *testing.T) {
	serverKP := newKeypair0()
	clientKP1 := newKeypair1()
	clientKP2 := newKeypair2()
	clientKP3 := keypair.MustRandom()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP1.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	threshold := Threshold(10)
	signerSummary := SignerSummary{
		clientKP1.Address(): 1,
		clientKP2.Address(): 2,
		clientKP3.Address(): 2,
	}

	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP1, clientKP2,
	)
	assert.NoError(t, err)

	_, err = VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.EqualError(t, err, "signers with weight 3 do not meet threshold 10")
}

func TestVerifyChallengeTxThreshold_invalidClientKeyUnrecognized(t *testing.T) {
	serverKP := newKeypair0()
	clientKP1 := newKeypair1()
	clientKP2 := newKeypair2()
	clientKP3 := keypair.MustRandom()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP1.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	threshold := Threshold(10)
	signerSummary := map[string]int32{
		clientKP1.Address(): 1,
		clientKP2.Address(): 2,
	}

	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP1, clientKP2, clientKP3,
	)
	assert.NoError(t, err)

	_, err = VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.EqualError(t, err, "transaction has unrecognized signatures")
}

func TestVerifyChallengeTxThreshold_invalidNoSigners(t *testing.T) {
	serverKP := newKeypair0()
	clientKP1 := newKeypair1()
	clientKP2 := newKeypair2()
	clientKP3 := keypair.MustRandom()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP1.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	threshold := Threshold(10)
	signerSummary := SignerSummary{}

	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP1, clientKP2, clientKP3,
	)
	assert.NoError(t, err)

	_, err = VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.EqualError(t, err, "no verifiable signers provided, at least one G... address must be provided")
}

func TestVerifyChallengeTxThreshold_weightsAddToMoreThan8Bits(t *testing.T) {
	serverKP := newKeypair0()
	clientKP1 := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP1.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP1, clientKP2,
	)
	assert.NoError(t, err)

	threshold := Threshold(1)
	signerSummary := SignerSummary{
		clientKP1.Address(): 255,
		clientKP2.Address(): 1,
	}
	wantSigners := []string{
		clientKP1.Address(),
		clientKP2.Address(),
	}

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.ElementsMatch(t, wantSigners, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxThreshold_doesNotVerifyHomeDomain(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	threshold := Threshold(1)
	signerSummary := SignerSummary{
		clientKP.Address(): 1,
	}
	wantSigners := []string{
		clientKP.Address(),
	}

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "notvalidated", threshold, signerSummary)
	assert.ElementsMatch(t, wantSigners, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxThreshold_allowsAdditionalManageDataOpsWithSourceAccountSetToServerAccount(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := ManageData{
		SourceAccount: &txSource,
		Name:          "a key",
		Value:         []byte("a value"),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	threshold := Threshold(1)
	signerSummary := SignerSummary{
		clientKP.Address(): 1,
	}
	wantSigners := []string{
		clientKP.Address(),
	}

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.ElementsMatch(t, wantSigners, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxThreshold_disallowsAdditionalManageDataOpsWithoutSourceAccountSetToServerAccount(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := ManageData{
		SourceAccount: &opSource,
		Name:          "a key",
		Value:         []byte("a value"),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	threshold := Threshold(1)
	signerSummary := SignerSummary{
		clientKP.Address(): 1,
	}

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "subsequent operations are unrecognized")
}

func TestVerifyChallengeTxThreshold_disallowsAdditionalOpsOfOtherTypes(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := BumpSequence{
		SourceAccount: &txSource,
		BumpTo:        0,
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	threshold := Threshold(1)
	signerSummary := SignerSummary{
		clientKP.Address(): 1,
	}

	signersFound, err := VerifyChallengeTxThreshold(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", threshold, signerSummary)
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "operation type should be manage_data")
}

func TestVerifyChallengeTxSigners_invalidServer(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		clientKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address())
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "transaction not signed by "+serverKP.Address())
}

func TestVerifyChallengeTxSigners_validServerAndClientMasterKey(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address())
	assert.Equal(t, []string{clientKP.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_invalidServerAndNoClient(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address())
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "transaction not signed by "+clientKP.Address())
}

func TestVerifyChallengeTxSigners_invalidServerAndUnrecognizedClient(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	unrecognizedKP := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, unrecognizedKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address())
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "transaction not signed by "+clientKP.Address())
}

func TestVerifyChallengeTxSigners_validServerAndMultipleClientSigners(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address(), clientKP2.Address())
	assert.Equal(t, []string{clientKP.Address(), clientKP2.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_validServerAndMultipleClientSignersReverseOrder(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP2, clientKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address(), clientKP2.Address())
	assert.Equal(t, []string{clientKP.Address(), clientKP2.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_validServerAndClientSignersNotMasterKey(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP2.Address())
	assert.Equal(t, []string{clientKP2.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_validServerAndClientSignersIgnoresServerSigner(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", serverKP.Address(), clientKP2.Address())
	assert.Equal(t, []string{clientKP2.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_invalidServerNoClientSignersIgnoresServerSigner(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", serverKP.Address(), clientKP2.Address())
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "transaction not signed by "+clientKP2.Address())
}

func TestVerifyChallengeTxSigners_validServerAndClientSignersIgnoresDuplicateSigner(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP2.Address(), clientKP2.Address())
	assert.Equal(t, []string{clientKP2.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_validIgnorePreauthTxHashAndXHash(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	preauthTxHash := "TAQCSRX2RIDJNHFIFHWD63X7D7D6TRT5Y2S6E3TEMXTG5W3OECHZ2OG4"
	xHash := "XDRPF6NZRR7EEVO7ESIWUDXHAOMM2QSKIQQBJK6I2FB7YKDZES5UCLWD"
	unknownSignerType := "?ARPF6NZRR7EEVO7ESIWUDXHAOMM2QSKIQQBJK6I2FB7YKDZES5UCLWD"
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP2.Address(), preauthTxHash, xHash, unknownSignerType)
	assert.Equal(t, []string{clientKP2.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_invalidServerAndClientSignersIgnoresDuplicateSignerInError(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address(), clientKP.Address())
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "transaction not signed by "+clientKP.Address())
}

func TestVerifyChallengeTxSigners_invalidServerAndClientSignersFailsDuplicateSignatures(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP2, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP2.Address())
	assert.Equal(t, []string{clientKP2.Address()}, signersFound)
	assert.EqualError(t, err, "transaction has unrecognized signatures")
}

func TestVerifyChallengeTxSigners_invalidServerAndClientSignersFailsSignerSeed(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	clientKP2 := newKeypair2()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP2,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP2.Seed())
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "no verifiable signers provided, at least one G... address must be provided")
}

func TestVerifyChallengeTxSigners_invalidNoSigners(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	_, err = VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver")
	assert.EqualError(t, err, "no verifiable signers provided, at least one G... address must be provided")
}

func TestVerifyChallengeTxSigners_doesNotVerifyHomeDomain(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "notvalidated", clientKP.Address())
	assert.Equal(t, []string{clientKP.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_allowsAdditionalManageDataOpsWithSourceAccountSetToServerAccount(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := ManageData{
		SourceAccount: &txSource,
		Name:          "a key",
		Value:         []byte("a value"),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address())
	assert.Equal(t, []string{clientKP.Address()}, signersFound)
	assert.NoError(t, err)
}

func TestVerifyChallengeTxSigners_disallowsAdditionalManageDataOpsWithoutSourceAccountSetToServerAccount(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := ManageData{
		SourceAccount: &opSource,
		Name:          "a key",
		Value:         []byte("a value"),
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address())
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "subsequent operations are unrecognized")
}

func TestVerifyChallengeTxSigners_disallowsAdditionalOpsOfOtherTypes(t *testing.T) {
	serverKP := newKeypair0()
	clientKP := newKeypair1()
	txSource := NewSimpleAccount(serverKP.Address(), -1)
	opSource := NewSimpleAccount(clientKP.Address(), 0)
	op1 := ManageData{
		SourceAccount: &opSource,
		Name:          "testserver auth",
		Value:         []byte(base64.StdEncoding.EncodeToString(make([]byte, 48))),
	}
	op2 := BumpSequence{
		SourceAccount: &txSource,
		BumpTo:        0,
	}
	tx64, err := newSignedTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&op1, &op2},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
		network.TestNetworkPassphrase,
		serverKP, clientKP,
	)
	assert.NoError(t, err)

	signersFound, err := VerifyChallengeTxSigners(tx64, serverKP.Address(), network.TestNetworkPassphrase, "testserver", clientKP.Address())
	assert.Empty(t, signersFound)
	assert.EqualError(t, err, "operation type should be manage_data")
}

func TestVerifyTxSignatureUnsignedTx(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	txSource := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	opSource := NewSimpleAccount(kp1.Address(), 0)
	createAccount := CreateAccount{
		Destination:   "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:        "10",
		SourceAccount: &opSource,
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewTimeout(1000),
		},
	)
	assert.NoError(t, err)

	err = verifyTxSignature(tx, network.TestNetworkPassphrase, kp0.Address())
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "transaction not signed by GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3")
	}
}

func TestVerifyTxSignatureSingle(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	txSource := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	opSource := NewSimpleAccount(kp1.Address(), 0)
	createAccount := CreateAccount{
		Destination:   "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:        "10",
		SourceAccount: &opSource,
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, kp0)
	assert.NoError(t, err)
	err = verifyTxSignature(tx, network.TestNetworkPassphrase, kp0.Address())
	assert.NoError(t, err)
}

func TestVerifyTxSignatureMultiple(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	txSource := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	opSource := NewSimpleAccount(kp1.Address(), 0)
	createAccount := CreateAccount{
		Destination:   "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:        "10",
		SourceAccount: &opSource,
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	// verify tx with multiple signature
	tx, err = tx.Sign(network.TestNetworkPassphrase, kp0, kp1)
	assert.NoError(t, err)
	err = verifyTxSignature(tx, network.TestNetworkPassphrase, kp0.Address())
	assert.NoError(t, err)
	err = verifyTxSignature(tx, network.TestNetworkPassphrase, kp1.Address())
	assert.NoError(t, err)
}
func TestVerifyTxSignatureInvalid(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	txSource := NewSimpleAccount(kp0.Address(), int64(9605939170639897))
	opSource := NewSimpleAccount(kp1.Address(), 0)
	createAccount := CreateAccount{
		Destination:   "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:        "10",
		SourceAccount: &opSource,
	}
	tx, err := NewTransaction(
		TransactionParams{
			SourceAccount:        &txSource,
			IncrementSequenceNum: true,
			Operations:           []Operation{&createAccount},
			BaseFee:              MinBaseFee,
			Timebounds:           NewInfiniteTimeout(),
		},
	)
	assert.NoError(t, err)

	// verify invalid signer
	tx, err = tx.Sign(network.TestNetworkPassphrase, kp0, kp1)
	assert.NoError(t, err)
	err = verifyTxSignature(tx, network.TestNetworkPassphrase, "GATBMIXTHXYKSUZSZUEJKACZ2OS2IYUWP2AIF3CA32PIDLJ67CH6Y5UY")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "transaction not signed by GATBMIXTHXYKSUZSZUEJKACZ2OS2IYUWP2AIF3CA32PIDLJ67CH6Y5UY")
	}
}
