package main

import (
	"log"
	"os"

	"github.com/spf13/viper"
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

func isValidEnv(env string) bool {
	return env == "dev" || env == "qa" || env == "production"
}

func loadConfig() {
	config.GoEnv = os.ExpandEnv("$GO_ENV")
	if !isValidEnv(config.GoEnv) {
		log.Fatal("Unsupported environment")
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(os.ExpandEnv("$GOPATH/bin/config/$GO_ENV"))
	viper.AddConfigPath(os.ExpandEnv("./config/$GO_ENV"))
	viper.ReadInConfig()

	config.ResourcesPath = os.ExpandEnv(viper.GetString("resourcesPath"))
	config.PublicPath = os.ExpandEnv(viper.GetString("publicPath"))
	config.WebsocketHost = viper.GetString("websocketHost")
	config.ListenPort = viper.GetString("listenPort")
}
