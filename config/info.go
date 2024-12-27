package config

import (
	"encoding/json"
	"github.com/lsclh/gtools/config/internal"
	"github.com/spf13/viper"
)

func SetUp(ptr any) {
	internal.Run()
	if err := viper.Unmarshal(ptr); err != nil {
		internal.Println("配置文件解析错误: %s", err.Error())
		return
	}

	d, _ := json.Marshal(ptr)
	internal.Println("配置文件读取成功 %s", string(d))
}
