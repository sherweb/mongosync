package src

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ColCopyWorker struct {
	SRC *mongo.Collection
	SRCDocCount int64
	DST *mongo.Collection
	DBName string
	ColName string
	Done bool
	Logs []string
	Config *ColConfig
	BatchSize int
}

func (cw *ColCopyWorker) ColCopyWorker() {
	cw.Done = false
}

func (cw *ColCopyWorker) GetDocCount(db *mongo.Database) int64 {
	count, err := db.Collection(cw.ColName).EstimatedDocumentCount(context.TODO())
	if err != nil {
		panic(err)
	}
	cw.SRCDocCount = count
	return count
}


func (cw *ColCopyWorker) Copy(c *Counters) {
	cfg := cw.Config

	cur, err := cw.SRC.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}

	models := []mongo.WriteModel{}
	batchCount := 0
	totalCount := 0
	

	for cur.Next(context.TODO()) {
		atomic.AddInt64(c.SourceItems, 1)
		var elem bson.D
		err := cur.Decode(&elem)
		if err != nil {
			panic(err)
		}
		exists, match := doc_exists_and_match(cw.DST, elem);
		if !exists && !match {
			models = append(models, mongo.NewInsertOneModel().SetDocument(elem))
			batchCount++
			totalCount++
		} else if exists && !match {
			models = append(models, mongo.NewReplaceOneModel().SetFilter(bson.D{{Key: "_id", Value: elem.Map()["_id"]}}).SetReplacement(elem))
			batchCount++
			totalCount++
		}

		if batchCount >= cfg.BatchSize {
			cw.Logs = append(cw.Logs, fmt.Sprintf("DB: %s, Col: %s, Reached batch threshold of %d, writing documents %d/%d\n", cw.DBName, cw.ColName, cfg.BatchSize, totalCount, cw.SRCDocCount))
			atomic.AddInt64(c.CopyingItems, int64(batchCount))
			opts := options.BulkWrite().SetOrdered(true)
			_, ierr := cw.DST.BulkWrite(context.TODO(), models, opts)
			if ierr != nil {
				fmt.Println(ierr)
			}
			atomic.AddInt64(c.CopiedItems, int64(batchCount))
			models = []mongo.WriteModel{}
			batchCount = 0
		}

	}

	opts := options.BulkWrite().SetOrdered(true)
	if (len(models) > 0) {
		atomic.AddInt64(c.CopyingItems, int64(len(models)))
		cw.Logs = append(cw.Logs, fmt.Sprintf("[DEST] DB: %s, COL: %s, SRC ITEMS %d, Updating %d items..\n", cw.DBName, cw.ColName, cw.SRCDocCount, len(models)))
		_, ierr := cw.DST.BulkWrite(context.TODO(), models, opts)
		if ierr != nil {
			fmt.Println(ierr)
		}
		atomic.AddInt64(c.CopiedItems, int64(len(models)))
	} else {
		cw.Logs = append(cw.Logs, fmt.Sprintf("[DEST] DB: %s, COL: %s, SRC ITEMS %d, No updates required\n", cw.DBName, cw.ColName, cw.SRCDocCount))
	}

	if cfg.CopyIndexes == true {
		copy_index(cw.SRC, cw.DST, c)
	}
	cw.Done = true
}

func copy_index(src *mongo.Collection, dst *mongo.Collection, c *Counters) {

	sourceIndexes := get_indexes(src)
	destIndexes := get_indexes(dst)

	count := 0
	for _, sourceIndex := range sourceIndexes {
		exists := false
		for _, destIndex := range destIndexes {
			if (sourceIndex["name"] == destIndex["name"]) {
				exists = true
			}
		}
		if (!exists) {
			i := mongo.IndexModel{
				Keys:	sourceIndex["key"],
				Options: options.Index().SetName(sourceIndex["name"].(string)),
			}
			_, err := dst.Indexes().CreateOne(context.TODO(), i)
			if err != nil {
				fmt.Printf("[DEST] Error creating index: %s\n", sourceIndex["name"])
				fmt.Println(err)
			} else {
				count++
			}
		}
	}

	if (count > 0) {
		atomic.AddInt64(c.Indexes, int64(count))
	}

}