package main

import (
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
)

type (
	Config struct {
		Port            int            `toml:"port"`
		ClientOrigin    string         `toml:"client_origin"`
		HTTPSOriginOnly bool           `toml:"https_origin_only"`
		AllowAllOrigins bool           `toml:"allow_all_origins"`
		StreamConfigs   []StreamConfig `toml:"streams"`
	}

	StreamConfig struct {
		Name     string `toml:"name"`
		Id       string `toml:"id"`
		Hostname string `toml:"hostname"`
		Path     string `toml:"path"`
		Port     int    `toml:"port"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	}
)

const exampleConfig = "port = 3000\n" +
	"client_origin = 'localhost'\n" +
	"https_origin_only = false\n" +
	"allow_all_origins = false\n\n" +
	"# example camera definition\n" +
	"# [[streams]]\n" +
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
