package main

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Debug        bool   `env:"DEBUG" envDefault:"false"`
	Listen       string `env:"LISTEN" envDefault:"localhost:8085"`
	ImageStorage string `env:"IMAGE_STORAGE" envDefault:"./"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
