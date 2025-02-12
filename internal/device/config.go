package device

import (
	"flag"
	"os"
	"time"

	"github.com/dvaxert/mdm/internal/domain/models"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string            `yaml:"env" env-default:"prod"` // local dev prod
	Uuid       string            `yaml:"uuid" env-required:"true"`
	DeviceType models.DeviceType `yaml:"device_type"`
	Grpc       GrpcConfig        `yaml:"grpc" env-required:"true"`
	PingPeriod time.Duration     `yaml:"ping_period" env-required:"true"`
	Location   string            `yaml:"location" env-required:"true"`
	Battery    int               `yaml:"battery" env-required:"true"`
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
