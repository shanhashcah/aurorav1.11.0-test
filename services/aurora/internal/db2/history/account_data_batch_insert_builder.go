package history

import (
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

func (i *accountDataBatchInsertBuilder) Add(entry xdr.LedgerEntry) error {
	data := entry.Data.MustData()
	// Add ledger_key only when inserting rows
	key, err := dataEntryToLedgerKeyString(entry)
	if err != nil {
		return errors.Wrap(err, "Error running dataEntryToLedgerKeyString")
	}

	return i.builder.Row(map[string]interface{}{
		"ledger_key":           key,
		"account_id":           data.AccountId.Address(),
		"name":                 data.DataName,
		"value":                AccountDataValue(data.DataValue),
		"last_modified_ledger": entry.LastModifiedLedgerSeq,
		"sponsor":              ledgerEntrySponsorToNullString(entry),
	})
}

func (i *accountDataBatchInsertBuilder) Exec() error {
	return i.builder.Exec()
}
