package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func copy_handler(cmd *cobra.Command, args []string) {
	sourceUri, _ := cmd.Flags().GetString("source")
	destinationUri, _ := cmd.Flags().GetString("destination")
	fmt.Printf("Connecting to source mongodb instance: %s\n", sourceUri)
	sourceDB := connect_db(sourceUri)
	fmt.Printf("Connecting to destination mongodb instance: %s\n", destinationUri)
	destinationDB := connect_db(destinationUri)
	var wg sync.WaitGroup
	if cmd.Flags().Changed("database") && cmd.Flags().Changed("collection") {
		//Single DB, single col
		copy_col(cmd.Flag("database").Value.String(), cmd.Flag("collection").Value.String(), sourceDB, destinationDB)
	} else if cmd.Flags().Changed("database") && !cmd.Flags().Changed("collection") {
		//Single DB, all cols
		copy_db(cmd.Flag("database").Value.String(), sourceDB, destinationDB, &wg)
	} else if !cmd.Flags().Changed("database") && cmd.Flags().Changed("collection") {
		fmt.Println("You must specify a database to copy a collection")
		os.Exit(1)
	} else {
		//all
		copy_all(sourceDB, destinationDB, &wg)
	}

}

func copy_all(sourceDB *mongo.Client, destDB *mongo.Client, wg *sync.WaitGroup) {
	dbs := get_dbs(sourceDB)
	for _, db := range dbs {
		go copy_db(db, sourceDB, destDB, wg)
	}
	wg.Wait()
}

func copy_db(dbName string, source *mongo.Client, dest *mongo.Client, wg *sync.WaitGroup) {
	wg.Add(1)
	collections := get_collections(dbName, source)
	for _, collection := range collections {
		copy_col(dbName, collection, source, dest)
	}
	defer wg.Done()
}

func copy_col(dbName string, colName string, source *mongo.Client, dest *mongo.Client) {

	count, err := source.Database(dbName).Collection(colName).EstimatedDocumentCount(context.TODO())
	if err != nil {
		panic(err)
	}

	if !db_exists(dbName, dest) {
		err := dest.Database(dbName).CreateCollection(context.TODO(), colName)
		if err != nil {
			fmt.Printf("Error creating collection: %s\n", colName)
			panic(err)
		}
	} else if (!col_exists(dbName, colName, dest)) {
		err := dest.Database(dbName).CreateCollection(context.TODO(), colName)
		if err != nil {
			fmt.Printf("Error creating collection: %s\n", colName)
			panic(err)
		}
	}

	cur, err := source.Database(dbName).Collection(colName).Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}
	models := []mongo.WriteModel{}
	for cur.Next(context.TODO()) {
		var elem bson.D
		err := cur.Decode(&elem)
		if err != nil {
			panic(err)
		}
		exists, match := doc_exists_and_match(dbName, colName, elem, dest);
		if !exists && !match {
			models = append(models, mongo.NewInsertOneModel().SetDocument(elem))
		} else if exists && !match {
			models = append(models, mongo.NewReplaceOneModel().SetFilter(bson.D{{"_id", elem.Map()["_id"]}}).SetReplacement(elem))
		}
	}

	opts := options.BulkWrite().SetOrdered(true)
	if (len(models) > 0) {
		fmt.Printf("[DEST] Database: %s, Collection: %s, Source Items %d, Updating %d items..\n", dbName, colName, count, len(models))
		_, ierr := dest.Database(dbName).Collection(colName).BulkWrite(context.TODO(), models, opts)
		if ierr != nil {
			fmt.Println(ierr)
		}
	} else {
		fmt.Printf("[DEST] Database: %s, Collection: %s, Source Items %d, No updates required\n", dbName, colName, count)
	}
}


