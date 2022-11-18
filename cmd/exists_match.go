package cmd

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
)

func db_exists(dbName string, client *mongo.Client) bool {
	dbs := get_dbs(client)
	return contains(dbs, dbName)
}

func col_exists(dbName string, colName string, client *mongo.Client) bool {
	cols := get_collections(dbName, client)
	return contains(cols, colName)
}

func doc_exists(dbName string, colName string, doc_id string, client *mongo.Client) bool {
	res := client.Database(dbName).Collection(colName).FindOne(context.TODO(), bson.M{"_id": doc_id})
	if res.Err() != nil {
		return true
	} else {
		return true
	}
	
}

func doc_exists_and_match(dbName string, colName string, doc bson.D, client *mongo.Client) (bool, bool) {
	/** if reflect.TypeOf(doc["_id"]).String() == "primitive.M" {
		fmt.Println("Using slow scan")
		res, err := client.Database(dbName).Collection(colName).Find(context.TODO(), bson.M{})
		if err != nil {
			fmt.Println(err)
		}
		for res.Next(context.TODO()) {
			var elem bson.M
			err := res.Decode(&elem)
			if err != nil {
				fmt.Println(err)
			}
			for k, v := range doc["_id"].(bson.M) {
				if elem["_id"].(bson.M)[k] != v {
					fmt.Println("match")
					return true, false
				}
			}

			if reflect.DeepEqual(doc["_id"].(bson.M), elem["_id"].(bson.M)) {
				return true, true
			} else {
				return true, false
			}
		}
	} else { **/
		q := bson.D{{"_id", doc.Map()["_id"]}}
		
		res := client.Database(dbName).Collection(colName).FindOne(context.TODO(), q)
		if res.Err() != nil {
			return false, false
		} else {
			var result bson.D
			res.Decode(&result)
			return true, reflect.DeepEqual(doc.Map(), result.Map())
		}
	//}
}