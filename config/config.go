package config

import (
	"github.com/spf13/viper"
	"go-db/domain"
	"strconv"
)

var PORT = "8080" // PORT is the port number for the API server

var DBMap = map[string]*domain.Database{}

func InitializeConfig() {
	//	LoadConfig from env using viper
	viper.AutomaticEnv()

	// Get PORT from env or use default if not set
	if viper.IsSet("PORT") {
		_port := viper.GetInt("PORT")
		PORT = strconv.Itoa(_port)
	}
}
