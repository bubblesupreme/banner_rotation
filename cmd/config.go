package main

import (
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Logger   LoggerConf
	DataBase DBMSConf
	Server   ServerConf
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

type DBMSConf struct {
	Login    string `mapstructure:"login"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	DBName   string `mapstructure:"dbname"`
	Host     string `mapstructure:"host"`
}

type ServerConf struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

func NewConfig() Config {
	c := Config{}
	err := viper.Unmarshal(&c)
	if err != nil {
		log.Warning("Unable to decode into struct: "+err.Error(), nil)
	}
	return c
}
