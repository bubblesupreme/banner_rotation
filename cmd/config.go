package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Logger   LoggerConf
	DataBase DBMSConf
	Server   ServerConf
	Rabbit   RabbitConf
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
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

type RabbitConf struct {
	URL             string `mapstructure:"url"`
	ExchangeName    string `mapstructure:"name"`
	ClickRoutingKey string `mapstructure:"click routing key"`
	ShowRoutingKey  string `mapstructure:"show routing key"`
}

func NewConfig() (Config, error) {
	c := Config{}
	err := viper.Unmarshal(&c)
	if err != nil {
		log.Errorf("unable to decode into struct")
		return c, err
	}

	if c.DataBase.Login == "" {
		ok := false
		c.DataBase.Login, ok = viper.Get("dblogin").(string)
		if !ok {
			return c, fmt.Errorf("database login is not set. Define environment variable 'POSTGRES_USER' or \"database\":\"login\" in the config file")
		}
	}
	if c.DataBase.DBName == "" {
		ok := false
		c.DataBase.DBName, ok = viper.Get("dbname").(string)
		if !ok {
			return c, fmt.Errorf("database name is not set. Define environment variable 'POSTGRES_DB' or \"database\":\"dbname\" in the config file")
		}
	}
	if c.DataBase.Password == "" {
		ok := false
		c.DataBase.Password, ok = viper.Get("dbpassword").(string)
		if !ok {
			return c, fmt.Errorf("database login is not set. Define environment variable 'POSTGRES_PASSWORD' or \"database\":\"password\" in the config file")
		}
	}

	return c, nil
}
