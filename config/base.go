package config

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"

	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout)

type BackendConfig struct {
	ConnectionString string `json:"connection_string" yaml:"connection_string"`
	Debug            bool   `json:"debug" yaml:"debug"`
}

type HttpServerConfig struct {
	Port           int    `json:"port" yaml:"port"`
	Name           string `json:"name" yaml:"name"`
	PublicEndpoint string `json:"public_endpoint" yaml:"public_endpoint"`
}

type LastFmConfig struct {
	ApiKey       string `json:"api_key" yaml:"api_key"`
	SharedSecret string `json:"shared_secret" yaml:"shared_secret"`
}

type ServicesConfig struct {
	LastFm LastFmConfig `json:"last_fm,omitempty" yaml:"last_fm"`
}
type FrontendConfig struct {
	Url string `json:"url" yaml:"url"`
}

type Config struct {
	Backend   BackendConfig    `json:"backend" yaml:"backend"`
	Server    HttpServerConfig `json:"server" yaml:"server"`
	SecretKey string           `json:"secret_key" yaml:"secret_key"`
	HashSalt  string           `json:"hash_salt" yaml:"hash_salt"`
	Services  ServicesConfig   `json:"services,omitempty" yaml:"services"`
	Frontend  FrontendConfig   `json:"frontend,omitempty" yaml:"frontend"`
}

func ParseFromFile(path string) Config {
	/* ConfigFromFile creates */
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Fatal(err)
		return Config{}
	}
	var cfg Config
	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(rawData, &cfg)
	} else {
		err = yaml.Unmarshal(rawData, &cfg)
	}

	if err != nil {
		logger.Fatal(err)
		return cfg
	}
	return cfg

}
