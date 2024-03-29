package ingest

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hcnet/go/ingest/adapters"
	"github.com/hcnet/go/ingest/io"
	"github.com/hcnet/go/ingest/ledgerbackend"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/ingest/processors"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

type ingestionSource int

const (
	_                    = iota
	historyArchiveSource = ingestionSource(iota)
	ledgerSource         = ingestionSource(iota)
	logFrequency         = 100000
)

type auroraTransactionProcessor interface {
	io.LedgerTransactionProcessor
	// TODO maybe rename to Flush()
	Commit() error
}

type statsChangeProcessor struct {
	*io.StatsChangeProcessor
}

func (statsChangeProcessor) Commit() error {
	return nil
}

type statsLedgerTransactionProcessor struct {
	*io.StatsLedgerTransactionProcessor
}

func (statsLedgerTransactionProcessor) Commit() error {
	return nil
}

type ProcessorRunnerInterface interface {
	SetLedgerBackend(ledgerBackend ledgerbackend.LedgerBackend)
	SetHistoryAdapter(historyAdapter adapters.HistoryArchiveAdapterInterface)
	EnableMemoryStatsLogging()
	DisableMemoryStatsLogging()
	RunHistoryArchiveIngestion(checkpointLedger uint32) (io.StatsChangeProcessorResults, error)
	RunTransactionProcessorsOnLedger(sequence uint32) (io.StatsLedgerTransactionProcessorResults, error)
	RunAllProcessorsOnLedger(sequence uint32) (
		io.StatsChangeProcessorResults,
		io.StatsLedgerTransactionProcessorResults,
		error,
	)
}

var _ ProcessorRunnerInterface = (*ProcessorRunner)(nil)

type ProcessorRunner struct {
	config Config

	ctx            context.Context
	historyQ       history.IngestionQ
	historyAdapter adapters.HistoryArchiveAdapterInterface
	ledgerBackend  ledgerbackend.LedgerBackend
	logMemoryStats bool
}

func (s *ProcessorRunner) SetLedgerBackend(ledgerBackend ledgerbackend.LedgerBackend) {
	s.ledgerBackend = ledgerBackend
}

func (s *ProcessorRunner) SetHistoryAdapter(historyAdapter adapters.HistoryArchiveAdapterInterface) {
	s.historyAdapter = historyAdapter
}

func (s *ProcessorRunner) EnableMemoryStatsLogging() {
	s.logMemoryStats = true
}

func (s *ProcessorRunner) DisableMemoryStatsLogging() {
	s.logMemoryStats = false
}

func (s *ProcessorRunner) buildChangeProcessor(
	changeStats *io.StatsChangeProcessor,
	source ingestionSource,
	sequence uint32,
) auroraChangeProcessor {
	statsChangeProcessor := &statsChangeProcessor{
		StatsChangeProcessor: changeStats,
	}

	useLedgerCache := source == ledgerSource
	return groupChangeProcessors{
		statsChangeProcessor,
		processors.NewAccountDataProcessor(s.historyQ),
		processors.NewAccountsProcessor(s.historyQ),
		processors.NewOffersProcessor(s.historyQ, sequence),
		processors.NewAssetStatsProcessor(s.historyQ, useLedgerCache),
		processors.NewSignersProcessor(s.historyQ, useLedgerCache),
		processors.NewTrustLinesProcessor(s.historyQ),
		processors.NewClaimableBalancesProcessor(s.historyQ),
	}
}

func (s *ProcessorRunner) buildTransactionProcessor(
	ledgerTransactionStats *io.StatsLedgerTransactionProcessor,
	ledger xdr.LedgerHeaderHistoryEntry,
) auroraTransactionProcessor {
	statsLedgerTransactionProcessor := &statsLedgerTransactionProcessor{
		StatsLedgerTransactionProcessor: ledgerTransactionStats,
	}

	sequence := uint32(ledger.Header.LedgerSeq)
	return groupTransactionProcessors{
		statsLedgerTransactionProcessor,
		processors.NewEffectProcessor(s.historyQ, sequence),
		processors.NewLedgerProcessor(s.historyQ, ledger, CurrentVersion),
		processors.NewOperationProcessor(s.historyQ, sequence),
		processors.NewTradeProcessor(s.historyQ, ledger),
		processors.NewParticipantsProcessor(s.historyQ, sequence),
		processors.NewTransactionProcessor(s.historyQ, sequence),
	}
}

// checkIfProtocolVersionSupported checks if this Aurora version supports the
// protocol version of a ledger with the given sequence number.
func (s *ProcessorRunner) checkIfProtocolVersionSupported(ledgerSequence uint32) error {
	exists, ledgerCloseMeta, err := s.ledgerBackend.GetLedger(ledgerSequence)
	if err != nil {
		return errors.Wrap(err, "Error getting ledger")
	}

	if !exists {
		return errors.New("cannot check if protocol version supported: ledger does not exist")
	}

	ledgerVersion := ledgerCloseMeta.V0.LedgerHeader.Header.LedgerVersion

	if ledgerVersion > MaxSupportedProtocolVersion {
		return fmt.Errorf(
			"This Aurora version does not support protocol version %d. "+
				"The latest supported protocol version is %d. Please upgrade to the latest Aurora version.",
			ledgerVersion,
			MaxSupportedProtocolVersion,
		)
	}

	return nil
}

// validateBucketList validates if the bucket list hash in history archive
// matches the one in corresponding ledger header in hcnet-core backend.
// This gives you full security if data in hcnet-core backend can be trusted
// (ex. you run it in your infrastructure).
// The hashes of actual buckets of this HAS file are checked using
// historyarchive.XdrStream.SetExpectedHash (this is done in
// SingleLedgerStateReader).
func (s *ProcessorRunner) validateBucketList(ledgerSequence uint32) error {
	historyBucketListHash, err := s.historyAdapter.BucketListHash(ledgerSequence)
	if err != nil {
		return errors.Wrap(err, "Error getting bucket list hash")
	}

	exists, ledgerCloseMeta, err := s.ledgerBackend.GetLedger(ledgerSequence)
	if err != nil {
		return errors.Wrap(err, "Error getting ledger")
	}

	if !exists {
		return fmt.Errorf(
			"cannot validate bucket hash list. Checkpoint ledger (%d) must exist in Hcnet-Core database.",
			ledgerSequence,
		)
	}

	ledgerBucketHashList := ledgerCloseMeta.V0.LedgerHeader.Header.BucketListHash

	if !bytes.Equal(historyBucketListHash[:], ledgerBucketHashList[:]) {
		return fmt.Errorf(
			"Bucket list hash of history archive and ledger header does not match: %#x %#x",
			historyBucketListHash,
			ledgerBucketHashList,
		)
	}

	return nil
}

func (s *ProcessorRunner) RunHistoryArchiveIngestion(checkpointLedger uint32) (io.StatsChangeProcessorResults, error) {
	changeStats := io.StatsChangeProcessor{}
	changeProcessor := s.buildChangeProcessor(&changeStats, historyArchiveSource, checkpointLedger)

	var changeReader io.ChangeReader
	var err error

	if checkpointLedger == 1 {
		changeReader = &io.GenesisLedgerStateReader{
			NetworkPassphrase: s.config.NetworkPassphrase,
		}
	} else {
		if err = s.checkIfProtocolVersionSupported(checkpointLedger); err != nil {
			return changeStats.GetResults(), errors.Wrap(err, "Error while checking for supported protocol version")
		}

		if err = s.validateBucketList(checkpointLedger); err != nil {
			return changeStats.GetResults(), errors.Wrap(err, "Error validating bucket list from HAS")
		}

		changeReader, err = s.historyAdapter.GetState(s.ctx, checkpointLedger)
		if err != nil {
			return changeStats.GetResults(), errors.Wrap(err, "Error creating HAS reader")
		}
	}
	defer changeReader.Close()

	log.WithField("ledger", checkpointLedger).
		Info("Processing entries from History Archive Snapshot")

	err = io.StreamChanges(changeProcessor, newloggingChangeReader(
		changeReader,
		"historyArchive",
		checkpointLedger,
		logFrequency,
		s.logMemoryStats,
	))
	if err != nil {
		return changeStats.GetResults(), errors.Wrap(err, "Error streaming changes from HAS")
	}

	err = changeProcessor.Commit()
	if err != nil {
		return changeStats.GetResults(), errors.Wrap(err, "Error commiting changes from processor")
	}

	return changeStats.GetResults(), nil
}

func (s *ProcessorRunner) runChangeProcessorOnLedger(
	changeProcessor auroraChangeProcessor, ledger uint32,
) error {
	var changeReader io.ChangeReader
	var err error
	changeReader, err = io.NewLedgerChangeReader(s.ledgerBackend, s.config.NetworkPassphrase, ledger)
	if err != nil {
		return errors.Wrap(err, "Error creating ledger change reader")
	}
	changeReader = newloggingChangeReader(
		changeReader,
		"ledger",
		ledger,
		logFrequency,
		s.logMemoryStats,
	)
	if err = io.StreamChanges(changeProcessor, changeReader); err != nil {
		return errors.Wrap(err, "Error streaming changes from ledger")
	}

	err = changeProcessor.Commit()
	if err != nil {
		return errors.Wrap(err, "Error commiting changes from processor")
	}

	return nil
}

func (s *ProcessorRunner) RunTransactionProcessorsOnLedger(ledger uint32) (io.StatsLedgerTransactionProcessorResults, error) {
	ledgerTransactionStats := io.StatsLedgerTransactionProcessor{}

	transactionReader, err := io.NewLedgerTransactionReader(s.ledgerBackend, s.config.NetworkPassphrase, ledger)
	if err != nil {
		return ledgerTransactionStats.GetResults(), errors.Wrap(err, "Error creating ledger reader")
	}

	if err = s.checkIfProtocolVersionSupported(ledger); err != nil {
		return ledgerTransactionStats.GetResults(), errors.Wrap(err, "Error while checking for supported protocol version")
	}

	txProcessor := s.buildTransactionProcessor(&ledgerTransactionStats, transactionReader.GetHeader())
	err = io.StreamLedgerTransactions(txProcessor, transactionReader)
	if err != nil {
		return ledgerTransactionStats.GetResults(), errors.Wrap(err, "Error streaming changes from ledger")
	}

	err = txProcessor.Commit()
	if err != nil {
		return ledgerTransactionStats.GetResults(), errors.Wrap(err, "Error commiting changes from processor")
	}

	return ledgerTransactionStats.GetResults(), nil
}

func (s *ProcessorRunner) RunAllProcessorsOnLedger(sequence uint32) (io.StatsChangeProcessorResults, io.StatsLedgerTransactionProcessorResults, error) {
	changeStats := io.StatsChangeProcessor{}
	var statsLedgerTransactionProcessorResults io.StatsLedgerTransactionProcessorResults

	if err := s.checkIfProtocolVersionSupported(sequence); err != nil {
		return changeStats.GetResults(), statsLedgerTransactionProcessorResults, errors.Wrap(err, "Error while checking for supported protocol version")
	}

	err := s.runChangeProcessorOnLedger(
		s.buildChangeProcessor(&changeStats, ledgerSource, sequence),
		sequence,
	)
	if err != nil {
		return changeStats.GetResults(), statsLedgerTransactionProcessorResults, err
	}

	statsLedgerTransactionProcessorResults, err = s.RunTransactionProcessorsOnLedger(sequence)
	if err != nil {
		return changeStats.GetResults(), statsLedgerTransactionProcessorResults, err
	}

	return changeStats.GetResults(), statsLedgerTransactionProcessorResults, nil
}
