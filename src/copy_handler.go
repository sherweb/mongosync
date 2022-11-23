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

	if !cmd.Flags().Changed("config") {
		fmt.Println("No config file provided, using command line arguments --config filepath.yml to provide a config file")
		fmt.Println("If you haven't generated a config file by running monogsync generate-config, you can do it now")
		os.Exit(1)
	}

	configFile, _ := cmd.Flags().GetString("config")

	cfg := ReadAndParseConfig(configFile)

	fmt.Printf("Connecting to source mongodb instance: %s\n", cfg.Source)
	sourceDB := connect_db(cfg.Source)
	fmt.Printf("Connecting to destination mongodb instance: %s\n", cfg.Destination)
	destinationDB := connect_db(cfg.Destination)

	
	dbc := DBConnector{
		SourceURI: cfg.Source,
		DestURI:   cfg.Destination,
		srcConn:   sourceDB,
		destConn:  destinationDB,
	}

	Copy(&cfg, &dbc, cmd)

}