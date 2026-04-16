package config

import (
	"github.com/BurntSushi/toml"
)

type MysqlConfig struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

type MainConfig struct {
	AppName string `toml:"appName"`
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
}

type RedisConfig struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	Password string `toml:"password"`
	Db       int    `toml:"db"`
}

type KafkaConfig struct {
	MessageMode string `toml:"messageMode"`
	HostPort    string `toml:"hostPort"`
	LoginTopic  string `toml:"loginTopic"`
	ChatTopic   string `toml:"chatTopic"`
	LogoutTopic string `toml:"logoutTopic"`
	Timeout     int    `toml:"timeout"`
	Partition   int    `toml:"partition"`
}

type LogConfig struct {
	LogPath  string `toml:"logPath"`
	LogLevel string `toml:"logLevel"`
}

type Config struct {
	MysqlConfig `toml:"mysqlConfig"`
	RedisConfig `toml:"redisConfig"`
	KafkaConfig `toml:"kafkaConfig"`
	MainConfig  `toml:"mainConfig"`
	LogConfig   `toml:"logConfig"`
}

var config *Config

func LoadConfig() error {
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		return err
	}
	return nil
}

func GetConfig() *Config {
	if config == nil {
		config = new(Config)
		_ = LoadConfig()
	}
	return config
}
