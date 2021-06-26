package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// get value by poolId
func getByPoolId(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := strconv.ParseInt(r.Form.Get("id"), 10, 64)
	if err != nil {
		fmt.Fprintf(w, `{"error": "%v"}`, err)
		return
	}
	values := poolGetById(int(id))
	fmt.Fprintf(w, `{"data": "%v"}`, values)
}

type addBody struct {
	PoolId     int   `json:"poolId"`
	PoolValues []int `json:"poolValues"`
}

type addResponse struct{}

// add a pool-value to pool. Append (if pool already exists) or insert (new pool) the values to the appropriate pool (as per the id)
func add(w http.ResponseWriter, r *http.Request) {
	body := addBody{}
	err := getBodyJson(r, &body)
	if err != nil {
		return
	}
	status, err := poolAdd(body.PoolId, body.PoolValues)
	if err != nil {
		fmt.Fprintf(w, `{"error": "%v"}`, err)
		return
	}
	fmt.Fprintf(w, `{"status": "%s"}`, status)
}

type quantileBody struct {
	PoolId     int     `json:"poolId"`
	Percentile float64 `json:"percentile"`
}

type quantileResponse struct {
	TotalElement int
	Quantile     int
}

// quantile caculator
func quantile(w http.ResponseWriter, r *http.Request) {
	body := quantileBody{}
	err := getBodyJson(r, &body)
	if err != nil {
		return
	}
	quantile, total := poolQuantile(body.PoolId, body.Percentile)
	fmt.Fprintf(w, `{"quantile": %.2f, "totalElement": %d}`, quantile, total)
}

// Util
func getBodyJson(r *http.Request, body interface{}) error {
	bodyByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyByte, body)
}
