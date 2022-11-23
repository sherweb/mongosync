package src

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
		err := cmd.Help()
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	},
}

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "copy data from one mongodb instance to another",
	Long:  ``,
	Run: copy_handler,
}

var generateConfigCmd = &cobra.Command{
	Use:   "generate-config",
	Short: "generate a config file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		sourceUri, _ := cmd.Flags().GetString("source")
		destinationUri, _ := cmd.Flags().GetString("destination")
		GenerateConfig(cmd, sourceUri, destinationUri)
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(generateConfigCmd)
	copyCmd.Flags().StringP("config", "c", "", "config file location")
	copyCmd.Flags().BoolP("profile", "p", false, "profile the execution")
	generateConfigCmd.Flags().StringP("source", "s", "", "source mongodb instance")
	generateConfigCmd.Flags().StringP("destination", "d", "", "destination mongodb instance")
	copyCmd.Flags().StringP("database", "", "", "database to copy")
	copyCmd.Flags().StringP("collection", "", "", "collection to copy")
	copyCmd.Flags().StringP("log-directory", "", "", "directory to store logs")
	copyCmd.Flags().IntP("batchsize", "b", 1000, "batch size")
	copyCmd.Flags().BoolP("indexes", "i", false, "copy indexes")

}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
