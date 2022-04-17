package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", RootPath)

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigs)
	go func() {
		for s := range sigs {
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				fmt.Println("Program Exit...", s)
				GracefullExit()
			default:
				fmt.Println("other signal", s)
			}
		}
	}()

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func GracefullExit() {
	fmt.Println("Start Exit...")
	fmt.Println("Execute Clean...")
	fmt.Println("End Exit...")
	os.Exit(0)
}
func RootPath(w http.ResponseWriter, r *http.Request) {

	for k, v := range r.Header {
		for _, value := range v {
			w.Header().Set(k, value)
		}
	}
	w.Header().Set("Server Version", os.Getenv("VERSION"))
	w.WriteHeader(http.StatusOK)
	fmt.Println("Client IP:", r.Host)
	fmt.Println("Return Code:", http.StatusOK)
	w.Write([]byte("Hello World"))
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
