package config

import (
	"github.com/blues/cfs/internal/logger"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Chain    ChainConfig    `mapstructure:"chain"`
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

// ChainConfig 单链配置
type ChainConfig struct {
	ChainType  string                    `mapstructure:"chain_type"`  // 链类型 (ethereum, polygon, etc.)
	ChainId    int64                     `mapstructure:"chain_id"`    // 链ID
	RpcUrl     string                    `mapstructure:"rpc_url"`     // RPC节点URL
	PrivateKey string                    `mapstructure:"private_key"` // 私钥
	Contracts  map[string]ContractConfig `mapstructure:"contracts"`   // 该链上的合约配置
}

// ContractConfig 单个合约配置
type ContractConfig struct {
	Address  string `mapstructure:"address"`   // 合约地址
	ABIPath  string `mapstructure:"abi_path"`  // ABI文件路径
	Enabled  bool   `mapstructure:"enabled"`   // 是否启用此合约
	BlockNum int64  `mapstructure:"block_num"` // 合约部署区块号
}

type TaskConfig struct {
	Interval int `mapstructure:"interval"` // 秒
}

type LogConfig struct {
	Level  string `mapstructure:"level"`  // 日志级别: debug, info, warn, error, fatal
	Output string `mapstructure:"output"` // 输出目标: stdout, stderr, file
	File   string `mapstructure:"file"`   // 日志文件路径（当output为file时使用）
}

// GetLevel 实现 logger.LogConfig 接口
func (l LogConfig) GetLevel() string {
	return l.Level
}

// GetOutput 实现 logger.LogConfig 接口
func (l LogConfig) GetOutput() string {
	return l.Output
}

// GetFile 实现 logger.LogConfig 接口
func (l LogConfig) GetFile() string {
	return l.File
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
		logger.Warn("Warning: Could not read config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		logger.Fatal("Unable to decode config into struct: %v", err)
	}

	return &config
}
