package main

import (
	"github.com/spf13/viper"
)

type config map[string]interface{}

func NewViper(cfg config) *viper.Viper {
	v := viper.New()

	for key, value := range cfg {
		v.Set(key, value)
	}

	return v
}
