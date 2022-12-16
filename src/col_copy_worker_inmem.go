package src

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)



func (cw *ColCopyWorker) CopyMultiThreadInMem(c *Counters, cur *mongo.Cursor, dcur *mongo.Cursor) {
	
	cfg := cw.Config
	maxInMem := 500000

	if (cfg.MaxDocsInMemory > 0) {
		maxInMem = cfg.MaxDocsInMemory
	}

	queue := goconcurrentqueue.NewFixedFIFO(maxInMem)

	chans := make([]chan bool, cfg.WorkerCount)

	dstMap := make(map[string]bson.D)

	wg := sync.WaitGroup{}

	for wc := 0; wc < cfg.WorkerCount; wc++ {
		wg.Add(1)
		go func() {
			for dcur.Next(context.TODO()) {
				var elem bson.D
				err := dcur.Decode(&elem)
				if err != nil {
					panic(err)
				}		
		
				dstMap[elem.Map()["_id"].(string)] = elem
				atomic.AddInt64(c.InMemItems, 1)
			}
		
		}()

	}

	wg.Wait()

	for wc := 0; wc < cfg.WorkerCount; wc++ {
		chans[wc] = make(chan bool)
		go cw.CopyMultiWorker(c, queue, chans[wc])
	}

	for cur.Next(context.TODO()) {
		var elem bson.D
		err := cur.Decode(&elem)
		if err != nil {
			panic(err)
		}
		queued := false
		for !queued {
			if (queue.GetLen() < maxInMem) {
				queue.Enqueue(elem)
				queued = true
			}
		}
	}

	doneProcessing := false 
	slept := false
	//Done queuing docs, now wait for workers to finish
	for !doneProcessing {
		if (queue.GetLen() == 0 && !slept) {
			time.Sleep(time.Second * 1) //safety sleep
			slept = true
		} else if (queue.GetLen() == 0 && slept) {
			go func() {
				for _, c := range chans {
					c <- true
				}
			}()
			doneProcessing = true
		}
	}

	if cfg.CopyIndexes == true {
		copy_index(cw.SRC, cw.DST, c)
	}
	cw.Done = true

}

func (cw *ColCopyWorker) CopyMultiWorkerInMem(c *Counters, q *goconcurrentqueue.FixedFIFO, dstMap *map[string]bson.D, done chan bool) {
	cfg := cw.Config
	
	models := []mongo.WriteModel{}
	batchCount := 0
	totalCount := 0

	for {
		select {
		case <-done:
			opts := options.BulkWrite().SetOrdered(true)
			if (len(models) > 0) {
				atomic.AddInt64(c.CopyingItems, int64(len(models)))
				_, ierr := cw.DST.BulkWrite(context.TODO(), models, opts)
				if ierr != nil {
					fmt.Println(ierr)
				}
				atomic.AddInt64(c.CopiedItems, int64(len(models)))
			} else {
			}
			return
		default:
			dequeued := false
			for !dequeued {

				elem, err := q.DequeueOrWaitForNextElement()
				if err == nil {
					atomic.AddInt64(c.SourceItems, 1)
					dequeued = true
					if cfg.NoFind {
						models = append(models, mongo.NewInsertOneModel().SetDocument(elem))
						batchCount++
						totalCount++
					} else {

						if v, ok := (*dstMap)[elem.(bson.D).Map()["_id"].(string)]; ok {
							//Item found
							if !reflect.DeepEqual(v, elem) {
								models = append(models, mongo.NewReplaceOneModel().SetFilter(bson.D{{Key: "_id", Value: elem.(bson.D).Map()["_id"]}}).SetReplacement(elem))
								batchCount++
								totalCount++
							}
						} else {
							//Item not found
							models = append(models, mongo.NewInsertOneModel().SetDocument(elem))
							batchCount++
							totalCount++
						}

					}

					if batchCount >= cfg.BatchSize {
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
				
			}

		}
	}


}


func (cw *ColCopyWorker) CopySingleThread(c *Counters, cur *mongo.Cursor) {
	cfg := cw.Config
	
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