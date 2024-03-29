package cmd

import (
	"fmt"
	"go/types"
	"net/http"
	_ "net/http/pprof"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/hcnet/go/historyarchive"
	aurora "github.com/hcnet/go/services/aurora/internal"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/ingest"
	support "github.com/hcnet/go/support/config"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/log"
)

var ingestCmd = &cobra.Command{
	Use:   "expingest",
	Short: "ingestion related commands",
}

var ingestVerifyFrom, ingestVerifyTo, ingestVerifyDebugServerPort uint32
var ingestVerifyState bool

var ingestVerifyRangeCmdOpts = []*support.ConfigOption{
	{
		Name:        "from",
		ConfigKey:   &ingestVerifyFrom,
		OptType:     types.Uint32,
		Required:    true,
		FlagDefault: uint32(0),
		Usage:       "first ledger of the range to ingest",
	},
	{
		Name:        "to",
		ConfigKey:   &ingestVerifyTo,
		OptType:     types.Uint32,
		Required:    true,
		FlagDefault: uint32(0),
		Usage:       "last ledger of the range to ingest",
	},
	{
		Name:        "verify-state",
		ConfigKey:   &ingestVerifyState,
		OptType:     types.Bool,
		Required:    false,
		FlagDefault: false,
		Usage:       "[optional] verifies state at the last ledger of the range when true",
	},
	{
		Name:        "debug-server-port",
		ConfigKey:   &ingestVerifyDebugServerPort,
		OptType:     types.Uint32,
		Required:    false,
		FlagDefault: uint32(0),
		Usage:       "[optional] opens a net/http/pprof server at given port",
	},
}

var ingestVerifyRangeCmd = &cobra.Command{
	Use:   "verify-range",
	Short: "runs ingestion pipeline within a range. warning! requires clean DB.",
	Long:  "runs ingestion pipeline between X and Y sequence number (inclusive)",
	Run: func(cmd *cobra.Command, args []string) {
		for _, co := range ingestVerifyRangeCmdOpts {
			co.Require()
			co.SetValue()
		}

		aurora.ApplyFlags(config, flags)

		if ingestVerifyDebugServerPort != 0 {
			go func() {
				log.Infof("Starting debug server at: %d", ingestVerifyDebugServerPort)
				err := http.ListenAndServe(
					fmt.Sprintf("localhost:%d", ingestVerifyDebugServerPort),
					nil,
				)
				if err != nil {
					log.Error(err)
				}
			}()
		}

		auroraSession, err := db.Open("postgres", config.DatabaseURL)
		if err != nil {
			log.Fatalf("cannot open Aurora DB: %v", err)
		}

		if !historyarchive.IsCheckpoint(ingestVerifyFrom) && ingestVerifyFrom != 1 {
			log.Fatal("`--from` must be a checkpoint ledger")
		}

		if ingestVerifyState && !historyarchive.IsCheckpoint(ingestVerifyTo) {
			log.Fatal("`--to` must be a checkpoint ledger when `--verify-state` is set.")
		}

		ingestConfig := ingest.Config{
			NetworkPassphrase:     config.NetworkPassphrase,
			HistorySession:        auroraSession,
			HistoryArchiveURL:     config.HistoryArchiveURLs[0],
			EnableCaptiveCore:     config.EnableCaptiveCoreIngestion,
			HcnetCoreBinaryPath: config.HcnetCoreBinaryPath,
			RemoteCaptiveCoreURL:  config.RemoteCaptiveCoreURL,
		}

		if !ingestConfig.EnableCaptiveCore {
			if config.HcnetCoreDatabaseURL == "" {
				log.Fatalf("flag --%s cannot be empty", aurora.HcnetCoreDBURLFlagName)
			}

			coreSession, dbErr := db.Open("postgres", config.HcnetCoreDatabaseURL)
			if dbErr != nil {
				log.Fatalf("cannot open Core DB: %v", dbErr)
			}
			ingestConfig.CoreSession = coreSession
		}

		system, err := ingest.NewSystem(ingestConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = system.VerifyRange(
			ingestVerifyFrom,
			ingestVerifyTo,
			ingestVerifyState,
		)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Range run successfully!")
	},
}

var stressTestNumTransactions, stressTestChangesPerTransaction int

var stressTestCmdOpts = []*support.ConfigOption{
	{
		Name:        "transactions",
		ConfigKey:   &stressTestNumTransactions,
		OptType:     types.Int,
		Required:    false,
		FlagDefault: int(1000),
		Usage:       "total number of transactions to ingest (at most 1000)",
	},
	{
		Name:        "changes",
		ConfigKey:   &stressTestChangesPerTransaction,
		OptType:     types.Int,
		Required:    false,
		FlagDefault: int(4000),
		Usage:       "changes per transaction to ingest (at most 4000)",
	},
}

var ingestStressTestCmd = &cobra.Command{
	Use:   "stress-test",
	Short: "runs ingestion pipeline on a ledger with many changes. warning! requires clean DB.",
	Long:  "runs ingestion pipeline on a ledger with many changes. warning! requires clean DB.",
	Run: func(cmd *cobra.Command, args []string) {
		for _, co := range stressTestCmdOpts {
			co.Require()
			co.SetValue()
		}

		aurora.ApplyFlags(config, flags)

		auroraSession, err := db.Open("postgres", config.DatabaseURL)
		if err != nil {
			log.Fatalf("cannot open Aurora DB: %v", err)
		}

		if stressTestNumTransactions <= 0 {
			log.Fatal("`--transactions` must be positive")
		}

		if stressTestChangesPerTransaction <= 0 {
			log.Fatal("`--changes` must be positive")
		}

		ingestConfig := ingest.Config{
			NetworkPassphrase: config.NetworkPassphrase,
			HistorySession:    auroraSession,
			HistoryArchiveURL: config.HistoryArchiveURLs[0],
			EnableCaptiveCore: config.EnableCaptiveCoreIngestion,
		}

		if config.EnableCaptiveCoreIngestion {
			ingestConfig.HcnetCoreBinaryPath = config.HcnetCoreBinaryPath
			ingestConfig.RemoteCaptiveCoreURL = config.RemoteCaptiveCoreURL
		} else {
			if config.HcnetCoreDatabaseURL == "" {
				log.Fatalf("flag --%s cannot be empty", aurora.HcnetCoreDBURLFlagName)
			}

			coreSession, dbErr := db.Open("postgres", config.HcnetCoreDatabaseURL)
			if dbErr != nil {
				log.Fatalf("cannot open Core DB: %v", dbErr)
			}
			ingestConfig.CoreSession = coreSession
		}

		system, err := ingest.NewSystem(ingestConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = system.StressTest(
			stressTestNumTransactions,
			stressTestChangesPerTransaction,
		)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Stress test completed successfully!")
	},
}

var ingestTriggerStateRebuildCmd = &cobra.Command{
	Use:   "trigger-state-rebuild",
	Short: "updates a database to trigger state rebuild, state will be rebuilt by a running Aurora instance, DO NOT RUN production DB, some endpoints will be unavailable until state is rebuilt",
	Run: func(cmd *cobra.Command, args []string) {
		aurora.ApplyFlags(config, flags)

		auroraSession, err := db.Open("postgres", config.DatabaseURL)
		if err != nil {
			log.Fatalf("cannot open Aurora DB: %v", err)
		}

		historyQ := &history.Q{auroraSession}
		err = historyQ.UpdateExpIngestVersion(0)
		if err != nil {
			log.Fatalf("cannot trigger state rebuild: %v", err)
		}

		log.Info("Triggered state rebuild")
	},
}

var ingestInitGenesisStateCmd = &cobra.Command{
	Use:   "init-genesis-state",
	Short: "ingests genesis state (ledger 1)",
	Run: func(cmd *cobra.Command, args []string) {
		aurora.ApplyFlags(config, flags)

		auroraSession, err := db.Open("postgres", config.DatabaseURL)
		if err != nil {
			log.Fatalf("cannot open Aurora DB: %v", err)
		}

		historyQ := &history.Q{auroraSession}

		lastIngestedLedger, err := historyQ.GetLastLedgerExpIngestNonBlocking()
		if err != nil {
			log.Fatalf("cannot get last ledger value: %v", err)
		}

		if lastIngestedLedger != 0 {
			log.Fatalf("cannot run on non-empty DB")
		}

		ingestConfig := ingest.Config{
			NetworkPassphrase: config.NetworkPassphrase,
			HistorySession:    auroraSession,
			HistoryArchiveURL: config.HistoryArchiveURLs[0],
			EnableCaptiveCore: config.EnableCaptiveCoreIngestion,
		}

		if config.EnableCaptiveCoreIngestion {
			ingestConfig.HcnetCoreBinaryPath = config.HcnetCoreBinaryPath
		} else {
			if config.HcnetCoreDatabaseURL == "" {
				log.Fatalf("flag --%s cannot be empty", aurora.HcnetCoreDBURLFlagName)
			}

			coreSession, dbErr := db.Open("postgres", config.HcnetCoreDatabaseURL)
			if dbErr != nil {
				log.Fatalf("cannot open Core DB: %v", dbErr)
			}
			ingestConfig.CoreSession = coreSession
		}

		system, err := ingest.NewSystem(ingestConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = system.BuildGenesisState()
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Genesis ledger stat successfully ingested!")
	},
}

func init() {
	for _, co := range ingestVerifyRangeCmdOpts {
		err := co.Init(ingestVerifyRangeCmd)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	for _, co := range stressTestCmdOpts {
		err := co.Init(ingestStressTestCmd)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	viper.BindPFlags(ingestVerifyRangeCmd.PersistentFlags())

	rootCmd.AddCommand(ingestCmd)
	ingestCmd.AddCommand(
		ingestVerifyRangeCmd,
		ingestStressTestCmd,
		ingestTriggerStateRebuildCmd,
		ingestInitGenesisStateCmd,
	)
}
