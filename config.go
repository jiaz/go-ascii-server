package main

import (
	"os"
)

var (
	config struct {
		GoEnv         string
		ResourcesPath string
		PublicPath    string
		WebsocketHost string
		ListenPort    string
	}
)

func loadConfig() {
	config.ResourcesPath = os.ExpandEnv("./resources")
	config.PublicPath = os.ExpandEnv("./public")
	config.WebsocketHost = "localhost:8080"
	config.ListenPort = "8080"
}
