package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"tendermint-signer/signer"

	"github.com/tendermint/tendermint/crypto/ed25519"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/os"
	tmOS "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/privval"
	tsed25519 "gitlab.com/polychainlabs/threshold-ed25519/pkg"
)

func init() {
	rootCmd.AddCommand(CreateCosignerSharesCmd())
}

// CreateCosignerSharesCmd is a cobra command for creating cosigner shares from a priv validator
func CreateCosignerSharesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-shares [priv_validator.json] [threshold] [shares]",
		Aliases: []string{"shard", "shares"},
		Args:    validateCreateCosignerShares,
		Short:   "Create  cosigner shares",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			keyFilePath := args[0]
			threshold, _ := strconv.ParseInt(args[1], 10, 64)
			total, _ := strconv.ParseInt(args[2], 10, 64)

			keyJSONBytes, err := ioutil.ReadFile(keyFilePath)
			if err != nil {
				tmOS.Exit(err.Error())
			}
			pvKey := privval.FilePVKey{}
			err = tmjson.Unmarshal(keyJSONBytes, &pvKey)
			if err != nil {
				tmOS.Exit(fmt.Sprintf("Error reading PrivValidator key from %v: %v\n", keyFilePath, err))
			}

			privKeyBytes := [64]byte{}

			// extract the raw private key bytes from the loaded key
			// we need this to compute the expanded secret
			switch ed25519Key := pvKey.PrivKey.(type) {
			case ed25519.PrivKey:
				if len(ed25519Key) != len(privKeyBytes) {
					panic("Key length inconsistency")
				}
				copy(privKeyBytes[:], ed25519Key[:])
				break
			default:
				panic("Not an ed25519 private key")
			}

			// generate shares from secret
			shares := tsed25519.DealShares(tsed25519.ExpandSecret(privKeyBytes[:32]), uint8(threshold), uint8(total))

			// generate all rsa keys
			rsaKeys := make([]*rsa.PrivateKey, len(shares))
			pubkeys := make([]*rsa.PublicKey, len(shares))
			for idx := range shares {
				bitSize := 4096
				rsaKey, err := rsa.GenerateKey(rand.Reader, bitSize)
				if err != nil {
					panic(err)
				}
				rsaKeys[idx] = rsaKey
				pubkeys[idx] = &rsaKey.PublicKey
			}

			// write shares and keys to private share files
			for idx, share := range shares {
				shareID := idx + 1

				privateFilename := fmt.Sprintf("private_share_%d.json", shareID)

				cosignerKey := signer.CosignerKey{
					PubKey:       pvKey.PubKey,
					ShareKey:     share,
					ID:           shareID,
					RSAKey:       *rsaKeys[idx],
					CosignerKeys: pubkeys,
				}

				jsonBytes, err := json.MarshalIndent(&cosignerKey, "", "  ")
				if err != nil {
					panic(err)
				}

				err = ioutil.WriteFile(privateFilename, jsonBytes, 0644)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Created Share %d\n", shareID)
			}
			return nil
		},
	}
	return cmd
}

func validateCreateCosignerShares(cmd *cobra.Command, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("wrong num args exp(3) got(%d)", len(args))
	}
	if !os.FileExists(args[0]) {
		return fmt.Errorf("priv_validator.json file(%s) doesn't exist", args[0])
	}
	if _, err := strconv.ParseInt(args[1], 10, 64); err != nil {
		return fmt.Errorf("shards must be an integer got(%s)", args[1])
	}
	if _, err := strconv.ParseInt(args[2], 10, 64); err != nil {
		return fmt.Errorf("threshold must be an integer got(%s)", args[2])
	}
	return nil
}