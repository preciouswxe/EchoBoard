package setting

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
)

// Conf 全局变量，用于保存程序的所有配置信息
var Conf = new(multipleConfig)

type multipleConfig struct {
	*AppConfig   `mapstructure:"app"`
	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	Mode      string `mapstructure:"mode"`
	Version   string `mapstructure:"version"`
	StartTime string `mapstructure:"start_time"`
	MachineID int64  `mapstructure:"machine_id"`
	Port      int    `mapstructure:"port"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DbName       string `mapstructure:"dbname"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func Init() (err error) {
	// 方式 1: 直接指定配置文件路径（相对路径或者绝对路径）
	// 相对路径：相对执行的可执行文件的相对路径
	//viper.SetConfigFile("./conf/config.yaml")
	// 绝对路径：系统中实际的文件路径
	//viper.SetConfigFile("D:/GOfiles/EchoBoard/EchoBoard_backend/conf/config.yaml")

	// 方式 2: 指定配置文件名和配置文件的位置，viper 自行查找可用的配置文件
	// 配置文件名不需要带后缀
	// 配置文件位置可配置多个
	//viper.SetConfigName("config")	// 指定配置文件名（不带后缀）  ！不要存在同名的不同类型文件
	//viper.AddConfigPath(".")			// 第一次找 指定查找配置文件得到路径（这里使用相对路径） 这步目前可以省略
	//viper.AddConfigPath("./conf")	// 第二次找 指定查找配置文件得到路径（这里使用相对路径）

	// viper.SetConfigType("yaml")      // （专用于远程）指定配置文件类型 支持 yaml / json

	//flag.StringVar(&filePath, "filePath", "config.yaml", "文件路径")
	//// 必须调用 flag.Parse() 来解析命令行参数
	//flag.Parse()

	filePath := os.Getenv("CONFIG_PATH")
	if filePath == "" {
		filePath = "conf/config.yaml" // 默认路径
	}

	fmt.Println("filePath: ", filePath)

	viper.SetConfigFile(filePath)

	err = viper.ReadInConfig() // 开始读取
	if err != nil {
		// 读取配置信息失败
		fmt.Printf("viper.ReadInConfig() failed, err: %v", err)
		return
	}

	// 把读取到的配置信息反序列化到 Conf 变量中
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err: %v\n", err)
	}

	// 支持热加载
	viper.WatchConfig()
	// 回调函数
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改了……")
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err: %v\n", err)
		}
	})

	fmt.Println("[viper.AllSettings] ", viper.AllSettings())
	return
}
