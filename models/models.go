package models

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

type ServerCfg struct {
	Timeout time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"10s"`
	Host    string        `yaml:"host" env:"Host" env-default:":8080"`
}

func MustLoad(path string) *ServerCfg {
	conf := &ServerCfg{}
	if err := cleanenv.ReadConfig(path, conf); err != nil {
		log.Fatal("Can't read the common config")
		return nil
	}
	return conf
}
