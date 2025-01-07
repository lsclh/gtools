package internal

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"runtime"
)

// 是否使用内网配置，不使用配置文件

func Run(fileName string, fileExt string) {
	if fileName == "" {
		fileName = "config"
	}
	if fileExt == "" {
		fileExt = "yaml"
	}
	//加载配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./conf")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../conf")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("../../config")
	viper.AddConfigPath("../../conf")
	viper.AddConfigPath("../../../")
	viper.AddConfigPath("../../../conf")
	viper.AddConfigPath("../../../config")
	viper.AddConfigPath("../../../../")
	viper.AddConfigPath("../../../../conf")
	viper.AddConfigPath("../../../../config")
	viper.AddConfigPath("../../../../../")
	viper.AddConfigPath("../../../../../conf")
	viper.AddConfigPath("../../../../../config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("配置文件读取失败: %s", err.Error())
		return
	}

	log.Println("配置文件读取成功")
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	//配置文件变化 用struct存储的话 可以在这里在 json.Unmarshal() 一下
	//})
	//监控配置文件变化 自动更新
	//viper.WatchConfig()

}

var (
	yellow = string([]byte{27, 91, 51, 51, 109})
	blue2  = string([]byte{27, 91, 51, 54, 109})
	reset  = string([]byte{27, 91, 48, 109})
)

// 控制台输出(生产模式不记录文件 正常输出控制台)
func Println(format string, v ...any) {
	_, file, line, _ := runtime.Caller(2)
	fmt.Println(fmt.Sprintf(" %s | %s | %s ", fmt.Sprint(yellow, "Debug", reset), fmt.Sprint(blue2, fmt.Sprintf(format, v...), reset), fmt.Sprintf("%s:%d", file, line)))
}
