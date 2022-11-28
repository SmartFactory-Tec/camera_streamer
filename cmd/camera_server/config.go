package main

import "github.com/BurntSushi/toml"

type (
	Config struct {
		Cameras []CameraConfig
	}

	CameraConfig struct {
		Name     string
		Id       string
		Hostname string
		Path     string
		Port     int
		User     string
		Password string
	}
)

func NewConfig(configFilePath string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(configFilePath, &config)

	if err != nil {
		return Config{}, err
	}

	return config, nil
}
