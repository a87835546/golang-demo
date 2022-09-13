package consts

import (
	"github.com/spf13/viper"
)

const (
	JWTSalt             = "OG_AGENT_MEMBER_LOGIN"
	MEMBER_PREFIX       = "member-login:"
	MEMBER_TOKEN_PREFIX = "member-token:"
)

var Conf *Config

type Config struct {
	Server     ServerConfig
	MySQL      MysqlConfig
	Redis      RedisConfig
	Beanstalkd BeanstalkdConfig
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	Password string `yaml:"password"`
}
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
}

type BeanstalkdConfig struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
}

type MysqlConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	Database string `yaml:"database"`
}

type ResourceConfig struct {
	Path string `yaml:"path"`
}

func InitYaml() {
	viper.SetConfigName("proj-dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./doraemon/")

	err := viper.ReadInConfig()
	if err != nil {
		//log.Println("Config file not found: ", err)
		panic("Failed to open config file")
	}

	viper.Unmarshal(&Conf)
}
