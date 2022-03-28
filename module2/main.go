package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", RootPath)

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func RootPath(w http.ResponseWriter, r *http.Request) {

	for k, v := range r.Header {
		for _, value := range v {
			w.Header().Set(k, value)
		}
	}
	w.Header().Set("Server Version", os.Getenv("VERSION"))
	retcode := 200
	w.WriteHeader(retcode)
	fmt.Println("Client IP:", r.Host)
	fmt.Println("Return Code:", retcode)

}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
