package ingest

import (
	"context"
	"reflect"
	"testing"

	"github.com/guregu/null"
	"github.com/hcnet/go/ingest/adapters"
	"github.com/hcnet/go/ingest/io"
	"github.com/hcnet/go/ingest/ledgerbackend"
	"github.com/hcnet/go/network"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/ingest/processors"
	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessorRunnerRunHistoryArchiveIngestionGenesis(t *testing.T) {
	maxBatchSize := 100000

	q := &mockDBQ{}

	// Batches
	mockOffersBatchInsertBuilder := &history.MockOffersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockOffersBatchInsertBuilder)
	mockOffersBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQOffers.On("NewOffersBatchInsertBuilder", maxBatchSize).
		Return(mockOffersBatchInsertBuilder).Once()

	mockAccountDataBatchInsertBuilder := &history.MockAccountDataBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountDataBatchInsertBuilder)
	mockAccountDataBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQData.On("NewAccountDataBatchInsertBuilder", maxBatchSize).
		Return(mockAccountDataBatchInsertBuilder).Once()

	mockClaimableBalancesBatchInsertBuilder := &history.MockClaimableBalancesBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockClaimableBalancesBatchInsertBuilder)
	mockClaimableBalancesBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQClaimableBalances.On("NewClaimableBalancesBatchInsertBuilder", maxBatchSize).
		Return(mockClaimableBalancesBatchInsertBuilder).Once()

	q.MockQAccounts.On("UpsertAccounts", []xdr.LedgerEntry{
		{
			LastModifiedLedgerSeq: 1,
			Data: xdr.LedgerEntryData{
				Type: xdr.LedgerEntryTypeAccount,
				Account: &xdr.AccountEntry{
					AccountId:  xdr.MustAddress("GAAZI4TCR3TY5OJHCTJC2A4QSY6CJWJH5IAJTGKIN2ER7LBNVKOCCWN7"),
					Balance:    xdr.Int64(1000000000000000000),
					SeqNum:     xdr.SequenceNumber(0),
					Thresholds: [4]byte{1, 0, 0, 0},
				},
			},
		},
	}).Return(nil).Once()

	mockAccountSignersBatchInsertBuilder := &history.MockAccountSignersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountSignersBatchInsertBuilder)
	mockAccountSignersBatchInsertBuilder.On("Add", history.AccountSigner{
		Account: "GAAZI4TCR3TY5OJHCTJC2A4QSY6CJWJH5IAJTGKIN2ER7LBNVKOCCWN7",
		Signer:  "GAAZI4TCR3TY5OJHCTJC2A4QSY6CJWJH5IAJTGKIN2ER7LBNVKOCCWN7",
		Weight:  1,
		Sponsor: null.String{},
	}).Return(nil).Once()
	mockAccountSignersBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQSigners.On("NewAccountSignersBatchInsertBuilder", maxBatchSize).
		Return(mockAccountSignersBatchInsertBuilder).Once()

	q.MockQAssetStats.On("InsertAssetStats", []history.ExpAssetStat{}, 100000).
		Return(nil)

	runner := ProcessorRunner{
		config: Config{
			NetworkPassphrase: network.PublicNetworkPassphrase,
		},
		historyQ: q,
	}

	_, err := runner.RunHistoryArchiveIngestion(1)
	assert.NoError(t, err)
}

func TestProcessorRunnerRunHistoryArchiveIngestionHistoryArchive(t *testing.T) {
	maxBatchSize := 100000

	config := Config{
		NetworkPassphrase: network.PublicNetworkPassphrase,
	}

	q := &mockDBQ{}
	defer mock.AssertExpectationsForObjects(t, q)
	historyAdapter := &adapters.MockHistoryArchiveAdapter{}
	defer mock.AssertExpectationsForObjects(t, historyAdapter)
	ledgerBackend := &ledgerbackend.MockDatabaseBackend{}
	defer mock.AssertExpectationsForObjects(t, ledgerBackend)

	historyAdapter.On("BucketListHash", uint32(63)).
		Return(xdr.Hash([32]byte{0, 1, 2}), nil).Once()

	historyAdapter.
		On(
			"GetState",
			mock.AnythingOfType("*context.emptyCtx"),
			uint32(63),
		).
		Return(
			&io.GenesisLedgerStateReader{
				NetworkPassphrase: network.PublicNetworkPassphrase,
			},
			nil,
		).Once()

	ledgerBackend.On("GetLedger", uint32(63)).
		Return(
			true,
			xdr.LedgerCloseMeta{
				V0: &xdr.LedgerCloseMetaV0{
					LedgerHeader: xdr.LedgerHeaderHistoryEntry{
						Header: xdr.LedgerHeader{
							BucketListHash: xdr.Hash([32]byte{0, 1, 2}),
						},
					},
				},
			},
			nil,
		).Twice() // 2nd time for protocol version check

	// Batches
	mockOffersBatchInsertBuilder := &history.MockOffersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockOffersBatchInsertBuilder)
	mockOffersBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQOffers.On("NewOffersBatchInsertBuilder", maxBatchSize).
		Return(mockOffersBatchInsertBuilder).Once()

	mockAccountDataBatchInsertBuilder := &history.MockAccountDataBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountDataBatchInsertBuilder)
	mockAccountDataBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQData.On("NewAccountDataBatchInsertBuilder", maxBatchSize).
		Return(mockAccountDataBatchInsertBuilder).Once()

	mockClaimableBalancesBatchInsertBuilder := &history.MockClaimableBalancesBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockClaimableBalancesBatchInsertBuilder)
	mockClaimableBalancesBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQClaimableBalances.On("NewClaimableBalancesBatchInsertBuilder", maxBatchSize).
		Return(mockClaimableBalancesBatchInsertBuilder).Once()

	q.MockQAccounts.On("UpsertAccounts", []xdr.LedgerEntry{
		xdr.LedgerEntry{
			LastModifiedLedgerSeq: 1,
			Data: xdr.LedgerEntryData{
				Type: xdr.LedgerEntryTypeAccount,
				Account: &xdr.AccountEntry{
					AccountId:  xdr.MustAddress("GAAZI4TCR3TY5OJHCTJC2A4QSY6CJWJH5IAJTGKIN2ER7LBNVKOCCWN7"),
					Balance:    xdr.Int64(1000000000000000000),
					SeqNum:     xdr.SequenceNumber(0),
					Thresholds: [4]byte{1, 0, 0, 0},
				},
			},
		},
	}).Return(nil).Once()

	mockAccountSignersBatchInsertBuilder := &history.MockAccountSignersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountSignersBatchInsertBuilder)
	mockAccountSignersBatchInsertBuilder.On("Add", history.AccountSigner{
		Account: "GAAZI4TCR3TY5OJHCTJC2A4QSY6CJWJH5IAJTGKIN2ER7LBNVKOCCWN7",
		Signer:  "GAAZI4TCR3TY5OJHCTJC2A4QSY6CJWJH5IAJTGKIN2ER7LBNVKOCCWN7",
		Weight:  1,
	}).Return(nil).Once()
	mockAccountSignersBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQSigners.On("NewAccountSignersBatchInsertBuilder", maxBatchSize).
		Return(mockAccountSignersBatchInsertBuilder).Once()

	q.MockQAssetStats.On("InsertAssetStats", []history.ExpAssetStat{}, 100000).
		Return(nil)

	runner := ProcessorRunner{
		ctx:            context.Background(),
		config:         config,
		historyQ:       q,
		historyAdapter: historyAdapter,
		ledgerBackend:  ledgerBackend,
	}

	_, err := runner.RunHistoryArchiveIngestion(63)
	assert.NoError(t, err)
}

func TestProcessorRunnerRunHistoryArchiveIngestionProtocolVersionNotSupported(t *testing.T) {
	maxBatchSize := 100000

	config := Config{
		NetworkPassphrase: network.PublicNetworkPassphrase,
	}

	q := &mockDBQ{}
	defer mock.AssertExpectationsForObjects(t, q)
	historyAdapter := &adapters.MockHistoryArchiveAdapter{}
	defer mock.AssertExpectationsForObjects(t, historyAdapter)
	ledgerBackend := &ledgerbackend.MockDatabaseBackend{}
	defer mock.AssertExpectationsForObjects(t, ledgerBackend)

	ledgerBackend.On("GetLedger", uint32(100)).
		Return(
			true,
			xdr.LedgerCloseMeta{
				V0: &xdr.LedgerCloseMetaV0{
					LedgerHeader: xdr.LedgerHeaderHistoryEntry{
						Header: xdr.LedgerHeader{
							LedgerVersion: 200,
						},
					},
				},
			},
			nil,
		).Once()

	// Batches
	mockOffersBatchInsertBuilder := &history.MockOffersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockOffersBatchInsertBuilder)
	q.MockQOffers.On("NewOffersBatchInsertBuilder", maxBatchSize).
		Return(mockOffersBatchInsertBuilder).Once()

	mockAccountDataBatchInsertBuilder := &history.MockAccountDataBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountDataBatchInsertBuilder)
	q.MockQData.On("NewAccountDataBatchInsertBuilder", maxBatchSize).
		Return(mockAccountDataBatchInsertBuilder).Once()

	mockClaimableBalancesBatchInsertBuilder := &history.MockClaimableBalancesBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockClaimableBalancesBatchInsertBuilder)
	q.MockQClaimableBalances.On("NewClaimableBalancesBatchInsertBuilder", maxBatchSize).
		Return(mockClaimableBalancesBatchInsertBuilder).Once()

	mockAccountSignersBatchInsertBuilder := &history.MockAccountSignersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountSignersBatchInsertBuilder)
	q.MockQSigners.On("NewAccountSignersBatchInsertBuilder", maxBatchSize).
		Return(mockAccountSignersBatchInsertBuilder).Once()

	q.MockQAssetStats.On("InsertAssetStats", []history.ExpAssetStat{}, 100000).
		Return(nil)

	runner := ProcessorRunner{
		ctx:            context.Background(),
		config:         config,
		historyQ:       q,
		historyAdapter: historyAdapter,
		ledgerBackend:  ledgerBackend,
	}

	_, err := runner.RunHistoryArchiveIngestion(100)
	assert.EqualError(t, err, "Error while checking for supported protocol version: This Aurora version does not support protocol version 200. The latest supported protocol version is 15. Please upgrade to the latest Aurora version.")
}

func TestProcessorRunnerBuildChangeProcessor(t *testing.T) {
	maxBatchSize := 100000

	q := &mockDBQ{}
	defer mock.AssertExpectationsForObjects(t, q)

	// Twice = checking ledgerSource and historyArchiveSource
	q.MockQOffers.On("NewOffersBatchInsertBuilder", maxBatchSize).
		Return(&history.MockOffersBatchInsertBuilder{}).Twice()
	q.MockQData.On("NewAccountDataBatchInsertBuilder", maxBatchSize).
		Return(&history.MockAccountDataBatchInsertBuilder{}).Twice()
	q.MockQSigners.On("NewAccountSignersBatchInsertBuilder", maxBatchSize).
		Return(&history.MockAccountSignersBatchInsertBuilder{}).Twice()

	runner := ProcessorRunner{
		historyQ: q,
	}

	stats := &io.StatsChangeProcessor{}
	processor := runner.buildChangeProcessor(stats, ledgerSource, 123)
	assert.IsType(t, groupChangeProcessors{}, processor)

	assert.IsType(t, &statsChangeProcessor{}, processor.(groupChangeProcessors)[0])
	assert.IsType(t, &processors.AccountDataProcessor{}, processor.(groupChangeProcessors)[1])
	assert.IsType(t, &processors.AccountsProcessor{}, processor.(groupChangeProcessors)[2])
	assert.IsType(t, &processors.OffersProcessor{}, processor.(groupChangeProcessors)[3])
	assert.IsType(t, &processors.AssetStatsProcessor{}, processor.(groupChangeProcessors)[4])
	assert.True(t, reflect.ValueOf(processor.(groupChangeProcessors)[4]).
		Elem().FieldByName("useLedgerEntryCache").Bool())
	assert.IsType(t, &processors.SignersProcessor{}, processor.(groupChangeProcessors)[5])
	assert.True(t, reflect.ValueOf(processor.(groupChangeProcessors)[5]).
		Elem().FieldByName("useLedgerEntryCache").Bool())
	assert.IsType(t, &processors.TrustLinesProcessor{}, processor.(groupChangeProcessors)[6])

	runner = ProcessorRunner{
		historyQ: q,
	}

	processor = runner.buildChangeProcessor(stats, historyArchiveSource, 456)
	assert.IsType(t, groupChangeProcessors{}, processor)

	assert.IsType(t, &statsChangeProcessor{}, processor.(groupChangeProcessors)[0])
	assert.IsType(t, &processors.AccountDataProcessor{}, processor.(groupChangeProcessors)[1])
	assert.IsType(t, &processors.AccountsProcessor{}, processor.(groupChangeProcessors)[2])
	assert.IsType(t, &processors.OffersProcessor{}, processor.(groupChangeProcessors)[3])
	assert.IsType(t, &processors.AssetStatsProcessor{}, processor.(groupChangeProcessors)[4])
	assert.False(t, reflect.ValueOf(processor.(groupChangeProcessors)[4]).
		Elem().FieldByName("useLedgerEntryCache").Bool())
	assert.IsType(t, &processors.SignersProcessor{}, processor.(groupChangeProcessors)[5])
	assert.False(t, reflect.ValueOf(processor.(groupChangeProcessors)[5]).
		Elem().FieldByName("useLedgerEntryCache").Bool())
	assert.IsType(t, &processors.TrustLinesProcessor{}, processor.(groupChangeProcessors)[6])
}

func TestProcessorRunnerBuildTransactionProcessor(t *testing.T) {
	maxBatchSize := 100000

	q := &mockDBQ{}
	defer mock.AssertExpectationsForObjects(t, q)

	q.MockQOperations.On("NewOperationBatchInsertBuilder", maxBatchSize).
		Return(&history.MockOperationsBatchInsertBuilder{}).Twice() // Twice = with/without failed
	q.MockQTransactions.On("NewTransactionBatchInsertBuilder", maxBatchSize).
		Return(&history.MockTransactionsBatchInsertBuilder{}).Twice()

	runner := ProcessorRunner{
		config:   Config{},
		historyQ: q,
	}

	stats := &io.StatsLedgerTransactionProcessor{}
	ledger := xdr.LedgerHeaderHistoryEntry{}
	processor := runner.buildTransactionProcessor(stats, ledger)
	assert.IsType(t, groupTransactionProcessors{}, processor)

	assert.IsType(t, &statsLedgerTransactionProcessor{}, processor.(groupTransactionProcessors)[0])
	assert.IsType(t, &processors.EffectProcessor{}, processor.(groupTransactionProcessors)[1])
	assert.IsType(t, &processors.LedgersProcessor{}, processor.(groupTransactionProcessors)[2])
	assert.IsType(t, &processors.OperationProcessor{}, processor.(groupTransactionProcessors)[3])
	assert.IsType(t, &processors.TradeProcessor{}, processor.(groupTransactionProcessors)[4])
	assert.IsType(t, &processors.ParticipantsProcessor{}, processor.(groupTransactionProcessors)[5])
	assert.IsType(t, &processors.TransactionProcessor{}, processor.(groupTransactionProcessors)[6])
}

func TestProcessorRunnerRunAllProcessorsOnLedger(t *testing.T) {
	maxBatchSize := 100000

	config := Config{
		NetworkPassphrase: network.PublicNetworkPassphrase,
	}

	q := &mockDBQ{}
	defer mock.AssertExpectationsForObjects(t, q)
	ledgerBackend := &ledgerbackend.MockDatabaseBackend{}
	defer mock.AssertExpectationsForObjects(t, ledgerBackend)

	ledger := xdr.LedgerHeaderHistoryEntry{
		Header: xdr.LedgerHeader{
			BucketListHash: xdr.Hash([32]byte{0, 1, 2}),
		},
	}

	// Method called 4 times:
	// - Protocol version check,
	// - Changes reader,
	// - Transactions reader (includes protocol check again because it's a public method).
	ledgerBackend.On("GetLedger", uint32(63)).
		Return(
			true,
			xdr.LedgerCloseMeta{
				V0: &xdr.LedgerCloseMetaV0{
					LedgerHeader: ledger,
				},
			},
			nil,
		).Times(4)

	// Batches
	mockOffersBatchInsertBuilder := &history.MockOffersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockOffersBatchInsertBuilder)
	mockOffersBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQOffers.On("NewOffersBatchInsertBuilder", maxBatchSize).
		Return(mockOffersBatchInsertBuilder).Once()

	mockAccountDataBatchInsertBuilder := &history.MockAccountDataBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountDataBatchInsertBuilder)
	mockAccountDataBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQData.On("NewAccountDataBatchInsertBuilder", maxBatchSize).
		Return(mockAccountDataBatchInsertBuilder).Once()

	mockAccountSignersBatchInsertBuilder := &history.MockAccountSignersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountSignersBatchInsertBuilder)
	q.MockQSigners.On("NewAccountSignersBatchInsertBuilder", maxBatchSize).
		Return(mockAccountSignersBatchInsertBuilder).Once()

	mockOperationsBatchInsertBuilder := &history.MockOperationsBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockOperationsBatchInsertBuilder)
	mockOperationsBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQOperations.On("NewOperationBatchInsertBuilder", maxBatchSize).
		Return(mockOperationsBatchInsertBuilder).Twice()

	mockTransactionsBatchInsertBuilder := &history.MockTransactionsBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockTransactionsBatchInsertBuilder)
	mockTransactionsBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQTransactions.On("NewTransactionBatchInsertBuilder", maxBatchSize).
		Return(mockTransactionsBatchInsertBuilder).Twice()

	mockClaimableBalancesBatchInsertBuilder := &history.MockClaimableBalancesBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockClaimableBalancesBatchInsertBuilder)
	mockClaimableBalancesBatchInsertBuilder.On("Exec").Return(nil).Once()
	q.MockQClaimableBalances.On("NewClaimableBalancesBatchInsertBuilder", maxBatchSize).
		Return(mockClaimableBalancesBatchInsertBuilder).Once()

	q.MockQLedgers.On("InsertLedger", ledger, 0, 0, 0, 0, CurrentVersion).
		Return(int64(1), nil).Once()

	runner := ProcessorRunner{
		ctx:           context.Background(),
		config:        config,
		historyQ:      q,
		ledgerBackend: ledgerBackend,
	}

	_, _, err := runner.RunAllProcessorsOnLedger(63)
	assert.NoError(t, err)
}

func TestProcessorRunnerRunAllProcessorsOnLedgerProtocolVersionNotSupported(t *testing.T) {
	maxBatchSize := 100000

	config := Config{
		NetworkPassphrase: network.PublicNetworkPassphrase,
	}

	q := &mockDBQ{}
	defer mock.AssertExpectationsForObjects(t, q)
	ledgerBackend := &ledgerbackend.MockDatabaseBackend{}
	defer mock.AssertExpectationsForObjects(t, ledgerBackend)

	ledger := xdr.LedgerHeaderHistoryEntry{
		Header: xdr.LedgerHeader{
			LedgerVersion: 200,
		},
	}

	ledgerBackend.On("GetLedger", uint32(63)).
		Return(
			true,
			xdr.LedgerCloseMeta{
				V0: &xdr.LedgerCloseMetaV0{
					LedgerHeader: ledger,
				},
			},
			nil,
		).Once()

	// Batches
	mockOffersBatchInsertBuilder := &history.MockOffersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockOffersBatchInsertBuilder)
	q.MockQOffers.On("NewOffersBatchInsertBuilder", maxBatchSize).
		Return(mockOffersBatchInsertBuilder).Once()

	mockAccountDataBatchInsertBuilder := &history.MockAccountDataBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountDataBatchInsertBuilder)
	q.MockQData.On("NewAccountDataBatchInsertBuilder", maxBatchSize).
		Return(mockAccountDataBatchInsertBuilder).Once()

	mockAccountSignersBatchInsertBuilder := &history.MockAccountSignersBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockAccountSignersBatchInsertBuilder)
	q.MockQSigners.On("NewAccountSignersBatchInsertBuilder", maxBatchSize).
		Return(mockAccountSignersBatchInsertBuilder).Once()

	mockOperationsBatchInsertBuilder := &history.MockOperationsBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockOperationsBatchInsertBuilder)
	q.MockQOperations.On("NewOperationBatchInsertBuilder", maxBatchSize).
		Return(mockOperationsBatchInsertBuilder).Twice()

	mockTransactionsBatchInsertBuilder := &history.MockTransactionsBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockTransactionsBatchInsertBuilder)
	q.MockQTransactions.On("NewTransactionBatchInsertBuilder", maxBatchSize).
		Return(mockTransactionsBatchInsertBuilder).Twice()

	mockClaimableBalancesBatchInsertBuilder := &history.MockClaimableBalancesBatchInsertBuilder{}
	defer mock.AssertExpectationsForObjects(t, mockClaimableBalancesBatchInsertBuilder)
	q.MockQClaimableBalances.On("NewClaimableBalancesBatchInsertBuilder", maxBatchSize).
		Return(mockClaimableBalancesBatchInsertBuilder).Once()

	runner := ProcessorRunner{
		ctx:           context.Background(),
		config:        config,
		historyQ:      q,
		ledgerBackend: ledgerBackend,
	}

	_, _, err := runner.RunAllProcessorsOnLedger(63)
	assert.EqualError(t, err, "Error while checking for supported protocol version: This Aurora version does not support protocol version 200. The latest supported protocol version is 15. Please upgrade to the latest Aurora version.")
}
