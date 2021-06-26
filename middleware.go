package main

import (
	"fmt"
	"net/http"
)

type handler func(w http.ResponseWriter, r *http.Request)

func RegisterRoute(method, route string, handle handler) {
	// LOG
	fmt.Printf("ROUTE - %s - METHOD - %s\n", route, method)

	// Check Method
	var checkMethod = func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			handle(w, r)
			return
		}
		http.NotFound(w, r)
	}
	http.HandleFunc(route, checkMethod)
}
