package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type addBody struct {
	PoolId     int64   `json:"poolId"`
	PoolValues []int64 `json:"poolValues"`
}

type addResponse struct{}

// add a pool-value to pool. Append (if pool already exists) or insert (new pool) the values to the appropriate pool (as per the id)
func add(w http.ResponseWriter, r *http.Request) {
	body := addBody{}
	err := getBodyJson(r, &body)
	if err != nil {
		return
	}
	status := poolAdd(body.PoolId, body.PoolValues)
	fmt.Fprintf(w, `{"status": "%s"}`, status)
}

type quantileBody struct {
	PoolId     int64   `json:"poolId"`
	Percentile float64 `json:"percentile"`
}

type quantileResponse struct {
	TotalElement int64
	Quantile     int64
}

// quantile caculator
func quantile(w http.ResponseWriter, r *http.Request) {
	body := quantileBody{}
	err := getBodyJson(r, &body)
	if err != nil {
		return
	}
	quantile, total := poolQuantile(body.PoolId, body.Percentile)
	fmt.Fprintf(w, `{"quantile": "%.2f", "totalElement": %d}`, quantile, total)
}

// Util
func getBodyJson(r *http.Request, body interface{}) error {
	bodyByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyByte, body)
}
