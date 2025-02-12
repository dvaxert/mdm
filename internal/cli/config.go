package cli

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Grpc GrpcConfig `yaml:"grpc" env-required:"true"`
}

type GrpcConfig struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

func MustLoadConfig() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		panic("config file does not exists: " + path)
	}

	conf := new(Config)

	err = cleanenv.ReadConfig(path, conf)
	if err != nil {
		panic("failed to read config: " + err.Error())
	}

	return conf
}

func fetchConfigPath() string {
	var result string

	flag.StringVar(&result, "config", "", "path to config file")
	flag.Parse()

	if result == "" {
		result = os.Getenv("CONFIG_PATH")
	}

	return result
}
