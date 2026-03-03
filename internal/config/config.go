package config

import "github.com/spf13/viper"

// Database 数据库连接配置
type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Passwd   string `mapstructure:"passwd"`
}

// DBSettings 数据库设置（源数据库和目标数据库）
type DBSettings struct {
	Source Database `mapstructure:"source"`
	Target Database `mapstructure:"target"`
}

// Options 选项配置
type Options struct {
	Databases []string `mapstructure:"databases"`
	DumpTo    string   `mapstructure:"dump_to"`
	SaveTo    string   `mapstructure:"save_to"`
	Threads   int      `mapstructure:"threads"`
	Mode      int      `mapstructure:"mode"`
}

// Config 配置结构体
type Config struct {
	DBSettings DBSettings `mapstructure:"db_settings"`
	Options    Options    `mapstructure:"options"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
