package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	ticker *time.Ticker = time.NewTicker(time.Second * 5)
)

func writeToStorage() {
	for {
		select {
		case <-ticker.C:
			if !isChange {
				continue
			}
			lsnMx.Lock()
			currentTime := time.Now().UnixNano()
			data := [][]string{
				{fmt.Sprintf("%d", currentTime), ""},
			}
			for poolId, poolValues := range pool {
				row := make([]string, 2)
				row[0] = fmt.Sprintf("%d", poolId)
				values := make([]string, len(poolValues.Values))
				for i, poolValue := range poolValues.Values {
					values[i] = fmt.Sprintf("%d", poolValue)
				}
				row[1] = strings.Join(values, csvSeparate)
				data = append(data, row)
			}

			f, err := os.Create(fmt.Sprintf("%s/data/pool.csv", currentDir))
			if err != nil {
				panic(err)
			}
			w := csv.NewWriter(f)

			for _, record := range data {
				if err := w.Write(record); err != nil {
					log.Fatalln("error writing record to csv:", err)
				}
			}

			// Write any buffered data to the underlying writer (standard output).
			w.Flush()
			isChange = false
			lsnMx.Unlock()
		}
	}
}
