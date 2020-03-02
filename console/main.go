package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

// APP ...
const APP = "accipfs"

// Version ...
const Version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:   APP,
	Short: "accipfs is a very fast ipfs client",
	Long:  `accipfs`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func main() {
	fmt.Println("accipfs starting...")
	rootCmd.AddCommand(daemonRun(), versionCmd())
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of " + APP,
		Long:  `All software has versions. This is ` + APP + `'s`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(APP, "Version:", Version)
		},
	}
	return cmd
}
