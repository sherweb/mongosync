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
	Run: copy,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync data from one mongodb instance to another (keep-alived)",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here

	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(syncCmd)
	copyCmd.Flags().StringP("source", "s", "", "source mongodb instance")
	copyCmd.Flags().StringP("destination", "d", "", "destination mongodb instance")
	copyCmd.Flags().StringP("database", "", "", "database to copy")
	copyCmd.Flags().StringP("collection", "", "", "collection to copy")
	copyCmd.MarkFlagRequired("source")
	copyCmd.MarkFlagRequired("destination")
	syncCmd.Flags().StringP("source", "s", "", "source mongodb instance")
	syncCmd.Flags().StringP("destination", "d", "", "destination mongodb instance")
	syncCmd.MarkFlagRequired("source")
	syncCmd.MarkFlagRequired("destination")
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
