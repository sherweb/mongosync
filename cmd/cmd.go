package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mongosync",
	Short: "mongosync is an utility to sync two different mongodb instances",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "copy data from one mongodb instance to another",
	Long:  ``,
	Run: copy_handler,
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringP("source", "s", "", "source mongodb instance")
	copyCmd.Flags().StringP("destination", "d", "", "destination mongodb instance")
	copyCmd.Flags().StringP("database", "", "", "database to copy")
	copyCmd.Flags().StringP("collection", "", "", "collection to copy")
	copyCmd.Flags().BoolP("indexes", "i", false, "copy indexes")
	copyCmd.MarkFlagRequired("source")
	copyCmd.MarkFlagRequired("destination")
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
