package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	RegisterRoute("POST", "/add", add)
	RegisterRoute("GET", "/", getByPoolId)
	RegisterRoute("POST", "/quantile", quantile)

	fmt.Println("SERVER LISTEN ON http://locahost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
