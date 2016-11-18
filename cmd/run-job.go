package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runjobCmd represents the run-job command
var runjobCmd = &cobra.Command{
	Use:   "run-job",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("run-job called")
	},
}

func init() {
	RootCmd.AddCommand(runjobCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runjobCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runjobCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
