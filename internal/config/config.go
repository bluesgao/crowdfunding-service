package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Ethereum EthereumConfig `mapstructure:"ethereum"`
	Task     TaskConfig     `mapstructure:"task"`
	Log      LogConfig      `mapstructure:"log"`
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
	RpcUrl        string `mapstructure:"rpc_url"`
	PrivateKey    string `mapstructure:"private_key"`
	ContractAddr  string `mapstructure:"contract_address"`
	StartBlock    int64  `mapstructure:"start_block"`
	Confirmations int    `mapstructure:"confirmations"`
}

type TaskConfig struct {
	Interval int `mapstructure:"interval"` // 秒
}

type LogConfig struct {
	Level  string `mapstructure:"level"`  // 日志级别: debug, info, warn, error, fatal
	Output string `mapstructure:"output"` // 输出目标: stdout, stderr, file
	File   string `mapstructure:"file"`   // 日志文件路径（当output为file时使用）
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
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
	viper.SetDefault("task.interval", 60)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.file", "logs/app.log")

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
