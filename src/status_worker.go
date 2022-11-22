package src

import (
	"fmt"
	"time"
)

type Counters struct {
	DBs          *int64
	Collections  *int64
	SourceItems  *int64
	CopyingItems *int64
	CopiedItems  *int64
	Indexes      *int64
}

func StatusWorker(c *Counters, quit chan bool, refresh_rate int) {

	start := time.Now()
	for {
		select {
		case <-quit:
			return
		default:
			t := time.Since(start).Seconds()

			fmt.Printf("DBs: %d, Collections: %d, SourceItems: %d (%02.f/s), CopyingItems: %d, CopiedItems: %d (%02.f/s), Indexes: %d\r", *c.DBs, *c.Collections, *c.SourceItems, (float64(*c.SourceItems) / t), (*c.CopyingItems - *c.CopiedItems), *c.CopiedItems, (float64(*c.CopiedItems) / t), *c.Indexes)

			time.Sleep(time.Duration(refresh_rate) * time.Millisecond)
		}
	}

}