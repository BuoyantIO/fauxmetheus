package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/NYTimes/gziphandler"
)

var deployments []deployment
var counter int

func handler(w http.ResponseWriter, r *http.Request) {
	counter++
	start := time.Now()
	metrics := writeMetrics(w, deployments, counter)
	fmt.Printf("wrote %d timeseries in %s\n", metrics, time.Since(start))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s /path/to/config\n", os.Args[0])
		os.Exit(42)
	}

	config, err := ReadConfig(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	deployments = []deployment{}
	for _, deployConfig := range config.Deployments {
		deployments = append(deployments, makeDeployments(deployConfig)...)
	}
	counter = 0

	http.Handle("/metrics", gziphandler.GzipHandler(http.HandlerFunc(handler)))
	fmt.Println("Serving /metrics on :4191")
	log.Fatal(http.ListenAndServe(":4191", nil))
}
