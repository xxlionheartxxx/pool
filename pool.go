package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type poolValues struct {
	Values []int64
	Mx     sync.Mutex
}

var (
	pool = map[int64]*poolValues{}
	mx   = sync.Mutex{}
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dat, err := ioutil.ReadFile(fmt.Sprintf("%s/pool.csv", dir))
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(strings.NewReader(string(dat)))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	for _, record := range records {
		poolId, _ := strconv.ParseInt(record[0], 10, 64)
		poolStringValues := strings.Split(record[1], "|")
		values := make([]int64, len(poolStringValues))
		for i, poolStringValue := range poolStringValues {
			value, _ := strconv.ParseInt(poolStringValue, 10, 64)
			values[i] = value
		}
		pool[poolId] = &poolValues{
			Values: values,
			Mx:     sync.Mutex{},
		}
	}
}

func poolAdd(poolId int64, values []int64) (status string) {
	if _, ok := pool[poolId]; !ok {
		mx.Lock()
		pool[poolId] = &poolValues{}
		mx.Unlock()
		status = "INSERTED"
	} else {
		status = "APPENDED"
	}
	pool[poolId].Add(values)
	return
}

func poolQuantile(poolId int64, percentile float64) (quantile float64, totalElement int64) {
	if _, ok := pool[poolId]; !ok {
		return
	}
	return
}

func (r *poolValues) Add(values []int64) {
	r.Mx.Lock()
	r.Values = append(r.Values, values...)
	r.Mx.Unlock()
}
