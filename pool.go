package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Int64 []int64

func (a Int64) Len() int           { return len(a) }
func (a Int64) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Int64) Less(i, j int) bool { return a[i] < a[j] }

type poolValues struct {
	Values []int
	Mx     sync.Mutex
}

var (
	pool                = map[int]*poolValues{}
	mx                  = sync.Mutex{}
	lsnMx               = sync.RWMutex{}
	currentDir          = ""
	checkPointLSN int64 = 0
	isChange      bool  = false
	mainTypeFile        = "main"
	walTypeFile         = "wal"
	csvSeparate         = "|"
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	currentDir = dir

	// Read from disk
	loadFromFile(fmt.Sprintf("%s/data/pool.csv", currentDir), mainTypeFile)

	// Read from WAL
	files, err := os.ReadDir(fmt.Sprintf("%s/wals", currentDir))
	if err != nil {
		panic(err)
	}
	fileNames := make([]int64, len(files))
	for i, file := range files {
		fileNames[i], _ = strconv.ParseInt(file.Name(), 10, 64)
	}
	sort.Sort(Int64(fileNames))
	for _, fileName := range fileNames {
		if checkPointLSN < fileName {
			isChange = true
			loadFromFile(fmt.Sprintf("%s/wals/%d", currentDir, fileName), walTypeFile)
		}
	}
	// Start storage writer
	go writeToStorage()
}

func loadFromFile(fileName, typeFile string) {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(strings.NewReader(string(dat)))
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	for i, record := range records {
		if i == 0 && typeFile == "main" {
			checkPointLSN, _ = strconv.ParseInt(record[0], 10, 64)
			continue
		}
		poolId, _ := strconv.ParseInt(record[0], 10, 64)
		poolStringValues := strings.Split(record[1], csvSeparate)
		values := make([]int, len(poolStringValues))
		for i, poolStringValue := range poolStringValues {
			value, _ := strconv.ParseInt(poolStringValue, 10, 64)
			values[i] = int(value)
		}
		if poolV, ok := pool[int(poolId)]; ok {
			poolV.Values = append(poolV.Values, values...)
		} else {
			pool[int(poolId)] = &poolValues{
				Values: values,
				Mx:     sync.Mutex{},
			}
		}
	}
}

func poolGetById(poolId int) []int {
	poolValues, ok := pool[poolId]
	if ok {
		return poolValues.Values
	}
	return []int{}
}

func poolAdd(poolId int, values []int) (status string, err error) {
	if _, ok := pool[poolId]; !ok {
		mx.Lock()
		pool[poolId] = &poolValues{}
		mx.Unlock()
		status = "INSERTED"
	} else {
		status = "APPENDED"
	}
	err = pool[poolId].Add(values, poolId)
	return status, err
}

func poolQuantile(poolId int, percentile float64) (quantile float64, totalElement int) {
	percentile = percentile / 100
	if _, ok := pool[poolId]; !ok {
		return
	}
	values := pool[poolId].Values
	totalElement = len(values)

	// Sort
	sort.Ints(values)
	index := percentile * float64(totalElement-1)
	lhs := int(index)
	delta := index - float64(lhs)
	if len(values) == 0 {
		return 0.0, 0
	}

	if lhs == totalElement-1 {
		quantile = float64(values[lhs])
	} else {
		quantile = (1-delta)*float64(values[lhs]) + delta*float64(values[lhs+1])
	}
	return
}

func (r *poolValues) Add(values []int, poolId int) error {
	// Lock when write to storage
	lsnMx.RLock()

	// Lock per poolId
	r.Mx.Lock()
	defer func() {
		r.Mx.Unlock()
		lsnMx.RUnlock()
	}()

	// Write ahead log
	stringValues := make([]string, len(values))
	for i, value := range values {
		stringValues[i] = fmt.Sprintf("%d", value)
	}
	d1 := []byte(fmt.Sprintf("%d,%s", poolId, strings.Join(stringValues, csvSeparate)))
	err := ioutil.WriteFile(fmt.Sprintf("%s/wals/%d", currentDir, time.Now().UnixNano()), d1, 0644)
	if err != nil {
		return err
	}
	r.Values = append(r.Values, values...)
	isChange = true
	return nil
}
