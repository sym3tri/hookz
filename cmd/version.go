package cmd

import (
	"github.com/spf13/cobra"
)

var Version = "undefined"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("version:", Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
