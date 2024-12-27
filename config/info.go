package config

import (
	"encoding/json"
	"github.com/lsclh/gtools/config/internal"
	"github.com/spf13/viper"
)

var cnf any = nil

func SetConfig(ptr any) {
	cnf = ptr
}

func SetUp() {
	internal.Run()
	if err := viper.Unmarshal(cnf); err != nil {
		internal.Println("配置文件解析错误: %s", err.Error())
		return
	}

	d, _ := json.Marshal(cnf)
	internal.Println("配置文件读取成功 %s", string(d))
}
