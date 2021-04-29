package main

import (
	"github.com/BurntSushi/toml"
	"time"
)

type Config struct {
	Bind     string
	MondoApi string
	Limiter  Limiter
}

func NewConfig() *Config {
	return &Config{}
}

func (p *Config) Init(cfgFile string) error {
	_, err := toml.DecodeFile(cfgFile, p)
	return err
}

type Limiter struct {
	Interval duration
	Capacity int64
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
