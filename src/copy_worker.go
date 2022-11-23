package src

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
)

type WorkerState struct {
	mu sync.Mutex
	Active int
}

type WorkerQueue struct {
	Workers []*ColCopyWorker
	Current int // index of current worker
	State *WorkerState
	Counters *Counters
	Done chan bool
}

func (wq *WorkerQueue) GetNextWorker() *ColCopyWorker {
	if wq.Current >= len(wq.Workers) {
		return nil
	}
	w := wq.Workers[wq.Current]
	wq.Current++
	return w
}

func (wq *WorkerQueue) Run(cfg *RootConfig) {
	statusDoneChan := make(chan bool)

	go StatusWorker(wq.Counters, statusDoneChan, cfg.RefreshRate, wq.State, cfg)

	for {
		select {
		case <-wq.Done:
			statusDoneChan <- true
			return
		default:
			if wq.State.Active < cfg.MaxWorkers {
				nw := wq.GetNextWorker()
				if nw != nil {
					go wq.RunWorker(nw, wq.Counters)
				} else {
					if wq.State.Active == 0 {
						time.Sleep(50 * time.Millisecond)
						statusDoneChan <- true
						time.Sleep(500 * time.Millisecond)
						fmt.Println("")
						fmt.Println("All workers are done")

						return
					}
				}
			}
			time.Sleep(time.Millisecond * 5)
		}
	}


}

func (wq *WorkerQueue) RunWorker(w *ColCopyWorker, c *Counters) {
	wq.State.mu.Lock()
	wq.State.Active++
	wq.State.mu.Unlock()
	w.Copy(c)
	wq.State.mu.Lock()
	wq.State.Active--
	wq.State.mu.Unlock()
}

func Copy(cfg *RootConfig, dbc *DBConnector, cmd *cobra.Command) {
	state := &WorkerState{
		Active: 0,
	}
	queue := &WorkerQueue{
		Current: 0,
		State: state,
		Counters: &Counters{
			DBs:          new(int64),
			Collections:  new(int64),
			SourceItems:  new(int64),
			CopyingItems: new(int64),
			CopiedItems:  new(int64),
			Indexes:      new(int64),
		},
	}

	for _, db := range cfg.Databases {

		var srcDBConn *mongo.Database
		var destDBConn *mongo.Database
		var destDBName string

		if db.RenameTo != "" {
			destDBName = db.RenameTo
		} else {
			destDBName = db.Name
		}

		if db.UseSeparateConnection {
			srcDBConn = connect_db(cfg.Source).Database(db.Name)
			destDBConn = connect_db(cfg.Destination).Database(destDBName)
		} else {
			srcDBConn = dbc.srcConn.Database(db.Name)
			destDBConn = dbc.destConn.Database(destDBName)
		}

		for _, col := range db.Collections {
			var srcColConn *mongo.Collection
			var destColConn *mongo.Collection
			var destColName string

			if col.RenameTo != "" {
				destColName = col.RenameTo
			} else {
				destColName = col.Name
			}

			if col.UseSeparateConnection {
				srcColConn = connect_db(cfg.Source).Database(db.Name).Collection(col.Name)
				destColConn = connect_db(cfg.Destination).Database(destDBName).Collection(destColName)
			} else {
				srcColConn = srcDBConn.Collection(col.Name)
				destColConn = destDBConn.Collection(destColName)	
			}

			cw := &ColCopyWorker{
				SRC: srcColConn,
				DST: destColConn,
				DBName: db.Name,
				ColName: col.Name,
				Config: col,
				BatchSize: col.BatchSize,
			}

			if col.Enabled {
				queue.Workers = append(queue.Workers, cw)
			}

		}

	}

	queue.Run(cfg)

	//TODO: Add save-logs

}