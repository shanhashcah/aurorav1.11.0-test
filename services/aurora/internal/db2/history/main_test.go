package history

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/test"
)

func TestLatestLedger(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	var seq int
	err := q.LatestLedger(&seq)

	if tt.Assert.NoError(err) {
		tt.Assert.Equal(3, seq)
	}
}

func TestGetLatestLedgerEmptyDB(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	test.ResetAuroraDB(t, tt.AuroraDB)
	q := &Q{tt.AuroraSession()}

	value, err := q.GetLatestLedger()
	tt.Assert.NoError(err)
	tt.Assert.Equal(uint32(0), value)
}

func TestElderLedger(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	var seq int
	err := q.ElderLedger(&seq)

	if tt.Assert.NoError(err) {
		tt.Assert.Equal(1, seq)
	}
}
