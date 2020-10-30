package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the struct which contains the credentials used
// when stablishing connections.
type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		URI          string `yaml:"uri"`
		DatabaseName string `yaml:"db-name"`
	} `yaml:"mongo-db"`
}

// ReadConfig reads the .yml config
// and decodes it to the passed parameter.
func ReadConfig(cfg *Config) {
	f, err := os.Open("config.yml")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)

	if err != nil {
		panic(err)
	}
}
