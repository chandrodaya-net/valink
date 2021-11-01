package cmd

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"tendermint-signer/signer"

	tmlog "github.com/tendermint/tendermint/libs/log"
	tmOS "github.com/tendermint/tendermint/libs/os"
	tmService "github.com/tendermint/tendermint/libs/service"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
)

func init() {
	signerCmd.AddCommand(StartSignerCmd())
	rootCmd.AddCommand(signerCmd)
}

var signerCmd = &cobra.Command{
	Use:   "signer",
	Short: "Remote tx signer for TM based nodes.",
}

func StartSignerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start single signer process",
		Args:  validateSignerStart,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			config, err := signer.LoadConfigFromFile(args[0])
			if err != nil {
				log.Fatal(err)
			}

			logger := tmlog.NewTMLogger(
				tmlog.NewSyncWriter(os.Stdout),
			).With("moniker", config.Moniker)
			

			logger.Info(
				"Tendermint Validator",
				"mode", config.Mode,
				"priv-key", config.PrivValKeyFile,
				"priv-state-dir", config.PrivValStateDir,
			)

			// services to stop on shutdown
			var services []tmService.Service

			var pv types.PrivValidator

			chainID := config.ChainID
			if chainID == "" {
				log.Fatal("chain_id option is required")
			}

			logger.Info("Mode: single")
			stateFile := path.Join(config.PrivValStateDir, fmt.Sprintf("%s_priv_validator_state.json", chainID))

			var val types.PrivValidator
			if fileExists(stateFile) {
				val = privval.LoadFilePV(config.PrivValKeyFile, stateFile)
			} else {
				logger.Info("Initializing empty state file", "file", stateFile)
				val = privval.LoadFilePVEmptyState(config.PrivValKeyFile, stateFile)
			}

			pv = &signer.PvGuard{PrivValidator: val}

			pubkey, err := pv.GetPubKey()
			if err != nil {
				log.Fatal(err)
			}
			logger.Info("Signer", "pubkey", pubkey)

			for _, node := range config.Nodes {
				dialer := net.Dialer{Timeout: 30 * time.Second}
				signer := signer.NewReconnRemoteSigner(node.Address, logger, config.ChainID, pv, dialer)

				err := signer.Start()
				if err != nil {
					panic(err)
				}

				services = append(services, signer)
			}

			wg := sync.WaitGroup{}
			wg.Add(1)
			tmOS.TrapSignal(logger, func() {
				for _, service := range services {
					err := service.Stop()
					if err != nil {
						panic(err)
					}
				}
				wg.Done()
			})
			wg.Wait()

			return nil

		},
	}

	return cmd
}

func validateSignerStart(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("wrong num args exp(1) got(%d)", len(args))
	}
	if !tmOS.FileExists(args[0]) {
		return fmt.Errorf("config.toml file(%s) doesn't exist", args[0])
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
