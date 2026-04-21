package config

import (
	"time"

	"github.com/BurntSushi/toml"
)

type MysqlConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
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
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	Db       int    `toml:"db"`
}

type KafkaConfig struct {
	MessageMode string        `toml:"messageMode"`
	HostPort    string        `toml:"hostPort"`
	LoginTopic  string        `toml:"loginTopic"`
	ChatTopic   string        `toml:"chatTopic"`
	LogoutTopic string        `toml:"logoutTopic"`
	Timeout     time.Duration `toml:"timeout"`
	Partition   int           `toml:"partition"`
}

type LogConfig struct {
	LogPath  string `toml:"logPath"`
	LogLevel string `toml:"logLevel"`
}

type StaticSrcConfig struct {
	StaticAvatarPath string `toml:"staticAvatarPath"`
	StaticFilePath   string `toml:"staticFilePath"`
}

type Config struct {
	MysqlConfig     `toml:"mysqlConfig"`
	RedisConfig     `toml:"redisConfig"`
	KafkaConfig     `toml:"kafkaConfig"`
	MainConfig      `toml:"mainConfig"`
	LogConfig       `toml:"logConfig"`
	StaticSrcConfig `toml:"staticSrcConfig"`
}

var config *Config

func LoadConfig() error {
	if _, err := toml.DecodeFile("config/config.toml", &config); err != nil {
		return err
	}
	return nil
}

//func LoadConfig() error {
//	// 1. 获取并打印当前程序的工作目录
//	wd, err := os.Getwd()
//	if err != nil {
//		return fmt.Errorf("无法获取当前工作目录: %v", err)
//	}
//	fmt.Printf("[DEBUG] 当前程序工作目录 (Working Directory): %s\n", wd)
//
//	// 2. 定义相对路径（基于你的项目结构）
//	// 注意：如果你在 GoLand 的工作目录设为 D:/Codes/IMChat，那么路径应该是 config/config.toml
//	relPath := "config/config.toml"
//
//	// 3. 获取该文件的绝对路径
//	absPath, _ := filepath.Abs(relPath)
//	fmt.Printf("[DEBUG] 尝试读取的绝对路径: %s\n", absPath)
//
//	// 4. 检查文件是否存在
//	if _, err := os.Stat(absPath); os.IsNotExist(err) {
//		return fmt.Errorf("文件不存在！请检查路径是否正确: %s", absPath)
//	}
//
//	// 5. 正式解码
//	if _, err := toml.DecodeFile(absPath, &config); err != nil {
//		return fmt.Errorf("解码配置文件失败: %v", err)
//	}
//
//	fmt.Println("[SUCCESS] 配置文件加载成功")
//	return nil
//}

func GetConfig() *Config {
	if config == nil {
		config = new(Config)
		err := LoadConfig()
		if err != nil {
			panic(err)
		}
	}
	return config
}
