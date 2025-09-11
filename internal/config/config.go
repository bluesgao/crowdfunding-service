package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Ethereum  EthereumConfig  `mapstructure:"ethereum"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type EthereumConfig struct {
	RPCURL        string `mapstructure:"rpc_url"`
	PrivateKey    string `mapstructure:"private_key"`
	ContractAddr  string `mapstructure:"contract_address"`
	StartBlock    uint64 `mapstructure:"start_block"`
	Confirmations int    `mapstructure:"confirmations"`
}

type SchedulerConfig struct {
	Interval int `mapstructure:"interval"` // 秒
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/cfs")

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "crowdfunding")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("ethereum.start_block", 0)
	viper.SetDefault("ethereum.confirmations", 12)
	viper.SetDefault("scheduler.interval", 60)

	// 自动读取环境变量
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return &config
}
