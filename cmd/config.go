package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const defaultEnvString = "-1" // default value for fields which can be initialized by environment variables

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
	Login         string `mapstructure:"login"`
	Password      string `mapstructure:"password"`
	Port          int    `mapstructure:"port"`
	DBName        string `mapstructure:"dbname"`
	Host          string `mapstructure:"host"`
	MigrationsDir string `mapstructure:"migrations"`
}

type ServerConf struct {
	Port int `mapstructure:"port"`
}

type RabbitConf struct {
	URL             string `mapstructure:"url"`
	ExchangeName    string `mapstructure:"name"`
	ClickRoutingKey string `mapstructure:"click routing key"`
	ShowRoutingKey  string `mapstructure:"show routing key"`
}

func NewConfig() (Config, error) {
	c := Config{}
	c.DataBase.MigrationsDir = defaultEnvString
	c.DataBase.Login = defaultEnvString
	c.DataBase.DBName = defaultEnvString
	c.DataBase.Password = defaultEnvString
	err := viper.Unmarshal(&c)
	if err != nil {
		log.Errorf("unable to decode into struct")
		return c, err
	}

	if c.DataBase.Login == defaultEnvString {
		ok := false
		c.DataBase.Login, ok = viper.Get("dblogin").(string)
		if !ok {
			return c, fmt.Errorf(
				"database login is not set. Define environment variable 'POSTGRES_USER' "+
					"or \"database\":\"login\" in the config file or check it is not equal '%s'",
				defaultEnvString)
		}
	}
	if c.DataBase.DBName == defaultEnvString {
		ok := false
		c.DataBase.DBName, ok = viper.Get("dbname").(string)
		if !ok {
			return c, fmt.Errorf(
				"database name is not set. Define environment variable 'POSTGRES_DB' "+
					"or \"database\":\"dbname\" in the config file or check it is not equal '%s'",
				defaultEnvString)
		}
	}
	if c.DataBase.Password == defaultEnvString {
		ok := false
		c.DataBase.Password, ok = viper.Get("dbpassword").(string)
		if !ok {
			return c, fmt.Errorf(
				"database login is not set. Define environment variable 'POSTGRES_PASSWORD' "+
					"or \"database\":\"password\" in the config file or check it is not equal '%s'",
				defaultEnvString)
		}
	}
	if c.DataBase.MigrationsDir == defaultEnvString {
		ok := false
		c.DataBase.MigrationsDir, ok = viper.Get("migrations").(string)
		if !ok {
			return c, fmt.Errorf(
				"migrations directory is not set. Define environment variable 'MIGRATIONS_DIRECTORY' "+
					"or \"database\":\"migrations\" in the config file or check it is not equal '%s'",
				defaultEnvString)
		}
	}

	return c, nil
}
