package aurora

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/ingest"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

var (
	data1 = xdr.LedgerEntry{
		LastModifiedLedgerSeq: 100,
		Data: xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeData,
			Data: &xdr.DataEntry{
				AccountId: xdr.MustAddress("GAOQJGUAB7NI7K7I62ORBXMN3J4SSWQUQ7FOEPSDJ322W2HMCNWPHXFB"),
				DataName:  "name1",
				// This also tests if base64 encoding is working as 0 is invalid UTF-8 byte
				DataValue: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
		},
		Ext: xdr.LedgerEntryExt{
			V: 1,
			V1: &xdr.LedgerEntryExtensionV1{
				SponsoringId: xdr.MustAddressPtr("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
			},
		},
	}

	data2 = xdr.LedgerEntry{
		LastModifiedLedgerSeq: 100,
		Data: xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeData,
			Data: &xdr.DataEntry{
				AccountId: xdr.MustAddress("GAOQJGUAB7NI7K7I62ORBXMN3J4SSWQUQ7FOEPSDJ322W2HMCNWPHXFB"),
				DataName:  "name ",
				DataValue: []byte("it got spaces!"),
			},
		},
	}
)

func TestDataActions_Show(t *testing.T) {
	ht := StartHTTPTestWithoutScenario(t)
	defer ht.Finish()
	test.ResetAuroraDB(t, ht.AuroraDB)
	q := &history.Q{ht.AuroraSession()}

	// Makes StateMiddleware happy
	err := q.UpdateLastLedgerExpIngest(100)
	ht.Assert.NoError(err)
	err = q.UpdateExpIngestVersion(ingest.CurrentVersion)
	ht.Assert.NoError(err)
	_, err = q.InsertLedger(xdr.LedgerHeaderHistoryEntry{
		Header: xdr.LedgerHeader{
			LedgerSeq: 100,
		},
	}, 0, 0, 0, 0, 0)
	ht.Assert.NoError(err)

	rows, err := q.InsertAccountData(data1)
	assert.NoError(t, err)
	ht.Assert.Equal(int64(1), rows)

	rows, err = q.InsertAccountData(data2)
	assert.NoError(t, err)
	ht.Assert.Equal(int64(1), rows)

	prefix := "/accounts/GAOQJGUAB7NI7K7I62ORBXMN3J4SSWQUQ7FOEPSDJ322W2HMCNWPHXFB"
	result := map[string]string{}

	// json
	w := ht.Get(prefix + "/data/name1")
	if ht.Assert.Equal(200, w.Code) {
		err := json.Unmarshal(w.Body.Bytes(), &result)
		ht.Assert.NoError(err)
		decoded, err := base64.StdEncoding.DecodeString(result["value"])
		ht.Assert.NoError(err)
		ht.Assert.Equal([]byte(data1.Data.Data.DataValue), decoded)
		ht.Assert.Equal("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML", result["sponsor"])
	}

	// raw
	w = ht.Get(prefix+"/data/name1", test.RequestHelperRaw)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.Equal([]byte(data1.Data.Data.DataValue), w.Body.Bytes())
	}

	result = map[string]string{}
	// regression: https://github.com/hcnet/aurora/issues/325
	// names with special characters do not work
	w = ht.Get(prefix + "/data/name%20")
	if ht.Assert.Equal(200, w.Code) {
		err := json.Unmarshal(w.Body.Bytes(), &result)
		ht.Assert.NoError(err)

		decoded, err := base64.StdEncoding.DecodeString(result["value"])
		ht.Assert.NoError(err)
		ht.Assert.Equal([]byte(data2.Data.Data.DataValue), decoded)
		ht.Assert.Equal("", result["sponsor"])
	}

	w = ht.Get(prefix+"/data/name%20", test.RequestHelperRaw)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.Equal("it got spaces!", w.Body.String())
	}

	// missing
	w = ht.Get(prefix + "/data/missing")
	ht.Assert.Equal(404, w.Code)

	w = ht.Get(prefix+"/data/missing", test.RequestHelperRaw)
	ht.Assert.Equal(404, w.Code)

	// Too long
	w = ht.Get(prefix+"/data/01234567890123456789012345678901234567890123456789012345678901234567890123456789", test.RequestHelperRaw)
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.Contains(w.Body.String(), "does not validate as length(1|64)")
	}

}
