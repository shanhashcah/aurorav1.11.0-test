package processors

import (
	"github.com/hcnet/go/ingest/io"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/support/errors"
)

type TransactionProcessor struct {
	transactionsQ history.QTransactions
	sequence      uint32
	batch         history.TransactionBatchInsertBuilder
}

func NewTransactionProcessor(transactionsQ history.QTransactions, sequence uint32) *TransactionProcessor {
	return &TransactionProcessor{
		transactionsQ: transactionsQ,
		sequence:      sequence,
		batch:         transactionsQ.NewTransactionBatchInsertBuilder(maxBatchSize),
	}
}

func (p *TransactionProcessor) ProcessTransaction(transaction io.LedgerTransaction) error {
	if err := p.batch.Add(transaction, p.sequence); err != nil {
		return errors.Wrap(err, "Error batch inserting transaction rows")
	}

	return nil
}

func (p *TransactionProcessor) Commit() error {
	if err := p.batch.Exec(); err != nil {
		return errors.Wrap(err, "Error flushing transaction batch")
	}

	return nil
}
