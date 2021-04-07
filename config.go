package main

import (
	"encoding/json"
	"os"
)

type (
	Config struct {
		Deployments []DeploymentConfig `json:"deployments,omitempty"`
	}
	DeploymentConfig struct {
		Quantity  int    `json:"quantity,omitempty"`
		Pods      int    `json:"pods,omitempty"`
		FanIn     int    `json:"fanIn,omitempty"`
		FanOut    int    `json:"fanOut,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}
)

func ReadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
