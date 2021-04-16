package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var deployments []deployment

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

	go start(deployments)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Serving /metrics on :4191")
	log.Fatal(http.ListenAndServe(":4191", nil))
}
