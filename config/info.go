package config

import (
	"encoding/json"
	"github.com/lsclh/gtools/config/internal"
	"github.com/spf13/viper"
	"github.com/lsclh/gtools/log"
)

func SetUp(ptr any, fileName string, fileExt string) {
	internal.Run(fileName, fileExt)
	if err := viper.Unmarshal(ptr); err != nil {
		log.Println("配置文件解析错误: %s", err.Error())
		return
	}

	d, _ := json.Marshal(ptr)
	log.Println("配置文件读取成功 %s", string(d))
}
