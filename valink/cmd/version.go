package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// application's version string
	Version = "v2.0.0"
	// commit
	Commit = ""
	// sdk version
	SDKVersion = ""
	// tendermint version
	TMVersion = ""
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Info defines the application version information.
type Info struct {
	Version           string `json:"version" yaml:"version"`
	GitCommit         string `json:"commit" yaml:"commit"`
	GoVersion         string `json:"go_version" yaml:"go_version"`
  CosmosSdkVersion  string `json:"cosmos_sdk_version" yaml:"cosmos_sdk_version"`
	TendermintVersion string `json:"tendermint_version" yaml:"tendermint_version"`
}

func NewInfo() Info {
	return Info{
		Version:           Version,
		GitCommit:         Commit,
		GoVersion:         fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		CosmosSdkVersion:  SDKVersion,
		TendermintVersion: TMVersion,
	}
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version information for valink",
	RunE: func(cmd *cobra.Command, args []string) error {

    // version.TMCoreSemVer
		bz, err := json.MarshalIndent(NewInfo(), "", "  ")
		if err != nil {
			return err
		}
		cmd.Println(string(bz))
		return nil
	},
}
