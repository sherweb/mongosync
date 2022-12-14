package cmd

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)



func get_dbs(client *mongo.Client) []string {
	list, err := client.ListDatabases(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var dbs []string
	for _, db := range list.Databases {
		if (db.Name != "admin" && db.Name != "local" && db.Name != "config") {
			dbs = append(dbs, db.Name)
		}
	}
	return dbs
}

func get_collections(dbName string, client *mongo.Client) []string {
	c, _ := client.Database(dbName).ListCollections(context.TODO(), bson.M{})
	var collections []string
	for c.Next(context.TODO()) {
		var elem bson.M
		err := c.Decode(&elem)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		collections = append(collections, elem["name"].(string))
	}
	return collections
}

func get_indexes(dbName string, colName string, client *mongo.Client) []bson.M {
	cur, err := client.Database(dbName).Collection(colName).Indexes().List(context.TODO())
	if err != nil {
		panic(err)
	}
	var indexes []bson.M
	for cur.Next(context.TODO()) {
		var elem bson.M
		err := cur.Decode(&elem)
		if err != nil {
			panic(err)
		}
		indexes = append(indexes, elem)
	}
	return indexes
}
