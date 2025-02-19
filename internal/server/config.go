package server

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string     `yaml:"env" env-default:"prod"` // local dev prod
	StoragePath string     `yaml:"storage_path" env-required:"true"`
	Grpc        GrpcConfig `yaml:"grpc" env-required:"true"`
}

type GrpcConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
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
