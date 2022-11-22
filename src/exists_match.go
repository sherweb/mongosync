package src

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
)
func doc_exists_and_match(dst *mongo.Collection, doc bson.D) (bool, bool) {
	q := bson.D{{Key: "_id", Value: doc.Map()["_id"]}}
	
	res := dst.FindOne(context.TODO(), q)
	if res.Err() != nil {
		return false, false
	} else {
		var result bson.D
		err := res.Decode(&result)
		if err != nil {
			panic(err)
		}
		return true, reflect.DeepEqual(doc.Map(), result.Map())
	}
}
