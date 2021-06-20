package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout)

type BackendConfig struct {
	ConnectionString string `json:"connection_string"`
	Debug            bool   `json:"debug"`
}

type HttpServerConfig struct {
	Port int `json:"port"`
	Name string `json:"name"`
	PublicEndpoint string `json:"public_endpoint"`
}

type LastFmConfig struct {
	ApiKey string `json:"api_key"`
	SharedSecret string `json:"shared_secret"`
}

type ServicesConfig struct {
	LastFm LastFmConfig `json:"last_fm,omitempty"`
}
type FrontendConfig struct {
	Url string `json:"url"`
}

type Config struct {
	Backend   BackendConfig    `json:"backend"`
	Server    HttpServerConfig `json:"server"`
	SecretKey string           `json:"secret_key"`
	HashSalt string `json:"hash_salt"`
	Services ServicesConfig `json:"services,omitempty"`
	Frontend FrontendConfig `json:"frontend,omitempty"`
}

func ParseFromFile(path string) Config {
	/* ConfigFromFile creates */
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Fatal(err)
		return Config{}
	}
	var cfg Config
	err = json.Unmarshal(rawData, &cfg)
	if err != nil {
		logger.Fatal(err)
		return cfg
	}
	return cfg

}
