package src

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"
)



func copy_handler(cmd *cobra.Command, args []string) {
	if (cmd.Flags().Changed("profile")) {
		f, perr := os.Create("cpu.pprof")
		if perr != nil {
			panic(perr)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}


	sourceUri, _ := cmd.Flags().GetString("source")
	destinationUri, _ := cmd.Flags().GetString("destination")
	fmt.Printf("Connecting to source mongodb instance: %s\n", sourceUri)
	sourceDB := connect_db(sourceUri)
	fmt.Printf("Connecting to destination mongodb instance: %s\n", destinationUri)
	destinationDB := connect_db(destinationUri)

	if !cmd.Flags().Changed("config") {
		fmt.Println("No config file provided, using command line arguments --config filepath.yml to provide a config file")
		fmt.Println("If you haven't generate a config file by running monogsync generate-config, you can do it now")
		os.Exit(1)
	}

	configFile, _ := cmd.Flags().GetString("config")

	cfg := ReadAndParseConfig(configFile)
	
	dbc := DBConnector{
		SourceURI: sourceUri,
		DestURI:   destinationUri,
		srcConn:   sourceDB,
		destConn:  destinationDB,
	}

	Copy(&cfg, &dbc)

}