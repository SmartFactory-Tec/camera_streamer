package main

import (
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
)

type (
	Config struct {
		Port    int
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

const exampleConfig = "port = 3000\n\n" +
	"# example camera definition\n" +
	"# [[cameras]]\n" +
	"# name = 'example'\n" +
	"# id = 'cam1'\n" +
	"# hostname = 'localhost'\n" +
	"# path = '/cam/1'\n" +
	"# port = 80\n" +
	"# user = 'admin'\n" +
	"# password = 'admin'\n"

func NewConfig() (Config, error) {

	var ConfigPath string

	if ServerConfigPath, found := os.LookupEnv("SERVER_CONFIG_PATH"); found {
		ConfigPath = ServerConfigPath
	} else if XdgConfigHome, found := os.LookupEnv("XDG_CONFIG_HOME"); found {
		ConfigPath = filepath.Join(XdgConfigHome, "camera_server")
	} else {
		HomePath := os.Getenv("HOME")
		ConfigPath = filepath.Join(HomePath, ".config", "camera_server")
	}

	var config Config

	configFilePath := filepath.Join(ConfigPath, "config.toml")

	for {
		if _, err := toml.DecodeFile(configFilePath, &config); err != nil {
			if err := os.Mkdir(ConfigPath, os.ModePerm); err != nil {
				return Config{}, err
			}

			if err := os.WriteFile(configFilePath, []byte(exampleConfig), 0666); err != nil {
				return Config{}, err
			}
		} else {
			break
		}
	}

	return config, nil
}
