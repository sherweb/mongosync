package src

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/yaml.v3"
)


type RootConfig struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	MaxWorkers	int `yaml:"max_workers"`
	BatchSize int `yaml:"batch_size"`
	RefreshRate int `yaml:"refresh_rate"`
	NoFind bool `yaml:"no_find"`
	ThresholdForSeparateConnection int `yaml:"threshold_for_separate_connection"`
	Databases []*DBConfig `yaml:"databases"`
}

type DBConfig struct {
	Name string `yaml:"name"`
	RenameTo string `yaml:"rename_to"`
	EstimatedCount int `yaml:"estimated_count"`
	BatchSize int `yaml:"batch_size"`
	Enabled bool `yaml:"enabled"`
	NoFind bool `yaml:"no_find"`
	UseSeparateConnection bool `yaml:"use_separate_connection"`
	Collections []*ColConfig `yaml:"collections"`
}

type ColConfig struct {
	Name string `yaml:"name"`
	RenameTo string `yaml:"rename_to"`
	InMemory bool `yaml:"in_memory"`
	TotalCount int `yaml:"total_count"`
	BatchSize int `yaml:"batch_size"`
	CopyIndexes bool `yaml:"copy_indexes"`
	Enabled bool `yaml:"enabled"`
	SourceBatchSize int `yaml:"source_batch_size"`
	UseSeparateConnection bool `yaml:"use_separate_connection"`
	UseMultipleWorkers bool `yaml:"use_multiple_workers"`
	WorkerCount int `yaml:"worker_count"`
	MaxDocsInMemory int `yaml:"max_docs_in_memory"`
	NoFind bool `yaml:"no_find"`
}

func GetBaseConfig() RootConfig {
	return RootConfig{
		Source: "",
		Destination: "",
		MaxWorkers: 2,
		BatchSize: 5000,
		RefreshRate: 50,
		ThresholdForSeparateConnection: 10000,
		NoFind: false,
		Databases: []*DBConfig{},
	}
}

func GetBaseDBConfig() DBConfig {
	return DBConfig{
		Name: "",
		RenameTo: "",
		EstimatedCount: 0,
		BatchSize: 0,
		Enabled: true,
		NoFind: false,
		UseSeparateConnection: false,
		Collections: []*ColConfig{},
	}
}

func GetBaseColConfig() ColConfig {
	return ColConfig{
		Name: "",
		RenameTo: "",
		InMemory: false,
		TotalCount: 0,
		BatchSize: 0,
		SourceBatchSize: 0,
		CopyIndexes: true,
		UseSeparateConnection: false,
		Enabled: true,
		NoFind: false,
		UseMultipleWorkers: false,
		WorkerCount: 0,
		MaxDocsInMemory: 500000,
	}
}

func ReadAndParseConfig(path string) RootConfig {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Couldn't read config")
		fmt.Println(err)
		os.Exit(1)
	}

	var cfg RootConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Println("Couldn't parse config")
		fmt.Println(err)
		os.Exit(1)
	}

	if (cfg.ThresholdForSeparateConnection != 0) {

		for _, db := range cfg.Databases {
			for _, col := range db.Collections {
				if (col.TotalCount > cfg.ThresholdForSeparateConnection && !col.UseSeparateConnection) { 
					col.UseSeparateConnection = true
				}
				ob := col.BatchSize
				if (cfg.BatchSize >= 1) {
					col.BatchSize = cfg.BatchSize
				}
				if (db.BatchSize >= 1) {
					col.BatchSize = db.BatchSize
				}
				if (ob >= 1) {
					col.BatchSize = ob
				}

				if col.BatchSize == 0 {
					col.BatchSize = 5000 //default value
				}

				onf := col.NoFind
				if cfg.NoFind {
					col.NoFind = cfg.NoFind
				}
				if db.NoFind {
					col.NoFind = db.NoFind
				}
				if onf {
					col.NoFind = onf
				}

				if col.MaxDocsInMemory == 0 {
					col.MaxDocsInMemory = 500000 //default value
				}

			}
			if (db.EstimatedCount > cfg.ThresholdForSeparateConnection) {
				db.UseSeparateConnection = true
			}
		}

	}


	return cfg
}


func SaveConfig(cfg RootConfig) {

	data, err := yaml.Marshal(&cfg)

	if err != nil {
			fmt.Println("Couldn't marshal config")
			panic(err)
	}

	err2 := ioutil.WriteFile("config.yml", data, 0644)

	if err2 != nil {

		fmt.Println("Couldn't save config")
		panic(err)
	}

	fmt.Println("Config saved")

}


func GenerateConfig(cmd *cobra.Command, sourceUri string, destinationUri string) {
	if cmd.Flags().Changed("database") && cmd.Flags().Changed("collection") {

	} else if cmd.Flags().Changed("database") && !cmd.Flags().Changed("collection") {
	} else if !cmd.Flags().Changed("database") && cmd.Flags().Changed("collection") {
	} else {

		cfg := GenerateAllConfig(cmd, sourceUri, destinationUri)

		SaveConfig(cfg)

	}
	

}

func GenerateColConfig(colName string, dbName string, conn *mongo.Client) ColConfig {
	colConfig := GetBaseColConfig()
	colConfig.Name = colName
	return colConfig
}

func GenerateDBConfig(dbName string, conn *mongo.Client) DBConfig {

	dbConfig := GetBaseDBConfig()
	dbConfig.Name = dbName
	cols := get_collections(dbName, conn)
	for _, col := range cols {
		colConfig := GenerateColConfig(col, dbName, conn)
		dbConfig.Collections = append(dbConfig.Collections, &colConfig)
	}
	return dbConfig
}

func GenerateAllConfig(cmd *cobra.Command, sourceUri string, destinationUri string) RootConfig {
	config := GetBaseConfig()
	config.Source = sourceUri
	config.Destination = destinationUri
	conn := connect_db(sourceUri)
	dbs := get_dbs(conn)

	for _, db := range dbs {
		dbConfig := GenerateDBConfig(db, conn)
		config.Databases = append(config.Databases, &dbConfig)
	}

	return config
}