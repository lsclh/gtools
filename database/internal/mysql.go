package internal

import (
	"context"
	"fmt"
	sql "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-colorable"
	"golang.org/x/crypto/ssh"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	ormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"net"
	"os"
	"time"
)

var (
	Db *gorm.DB
)
var level = map[string]ormLogger.LogLevel{
	"silent": ormLogger.Silent,
	"error":  ormLogger.Error,
	"warn":   ormLogger.Warn,
	"info":   ormLogger.Info,
}

type MOptions struct {
	Dbname       string      `json:"dbname"`
	Host         string      `json:"host"`
	User         string      `json:"user"`
	Pass         string      `json:"pass"`
	Port         int         `json:"port"`
	PollMaxOpen  int         `json:"pollMaxOpen"`  //最大打开连接数
	PollMinConns int         `json:"pollMinConns"` //最小保持活跃连接数
	Log          *MOptionLog `json:"log"`
	Ssh          *MOptionSSH `json:"ssh"`
}

type MOptionSSH struct {
	Host      string `json:"host"`
	User      string `json:"user"`
	Pass      string `json:"pass"`
	PublicKey string `json:"publicKey"`
}

type MOptionLog struct {
	Level ormLogger.LogLevel `json:"level"`
	Std   ormLogger.Writer   `json:"std"`
}

var opt *MOptions = nil

func NewDb(o *MOptions) *MOptions {
	opt = o
	return opt
}

// Setup : Connect to mysql database
func (o *MOptions) Init() {
	//sql输出日志级别
	var err error
	link := ""
	//是否使用ssh代理连接
	if opt.Ssh != nil {

		sshConfig := &ssh.ClientConfig{
			User:            opt.Ssh.User,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		if opt.Ssh.Pass != "" {
			sshConfig.Auth = []ssh.AuthMethod{ssh.Password(opt.Ssh.Pass)}
		} else {
			// 读取秘钥文件
			k, err := os.ReadFile(opt.Ssh.PublicKey)
			if err != nil {
				panic(fmt.Sprintf("MysqlSSHReadPublicKeyFail: %s", err.Error()))
				return
			}
			//创建ssh签名
			signer, err := ssh.ParsePrivateKey(k)
			if err != nil {
				panic(fmt.Sprintf("数据库ssh连接秘钥解析错误: %s", err.Error()))
				return
			}
			//设置验证
			sshConfig.Auth = []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			}
		}
		// 创建ssh连接
		sshcon, err := ssh.Dial("tcp", opt.Ssh.Host, sshConfig)
		if err != nil {
			panic(fmt.Sprintf("数据库ssh连接失败: %s", err.Error()))
			return
		}
		//注册ssh代理
		sql.RegisterDialContext("mysqlssh", func(ctx context.Context, addr string) (net.Conn, error) {
			return sshcon.Dial("tcp", addr)
		})
		link = fmt.Sprintf("%s:%s@mysqlssh(%s)/%s?charset=utf8&parseTime=True&loc=Local", opt.User, opt.Pass, opt.Host, opt.Dbname)
	} else {
		link = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", opt.User, opt.Pass, opt.Host, opt.Dbname)
	}
	cnf := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键
	}

	if opt.Log != nil {
		//默认文件日志 不输出色彩
		var logLevel = ormLogger.Error
		if opt.Log.Level != 0 {
			logLevel = opt.Log.Level
		}
		colorful := false
		//debug仅打印到控制台
		if opt.Log.Std == nil {
			opt.Log.Std = log.New(colorable.NewColorableStdout(), "", log.LstdFlags)
			colorful = true
		}
		cnf.Logger = ormLogger.New(
			opt.Log.Std, // io writer
			ormLogger.Config{
				SlowThreshold:             time.Second * 1, // 慢 SQL 阈值
				LogLevel:                  logLevel,        //logger.Silent //不进行任何打印
				Colorful:                  colorful,        // 色彩打印
				IgnoreRecordNotFoundError: true,            //忽略查询未找到的错误
			},
		)
	}

	Db, err = gorm.Open(mysql.Open(link), cnf)
	if err != nil {
		panic(fmt.Sprintf("MysqlConnectFail: %s", err.Error()))
		return
	} else {
		sqlDB, _ := Db.DB()
		sqlDB.SetMaxIdleConns(opt.PollMinConns)
		sqlDB.SetMaxOpenConns(opt.PollMaxOpen)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
	logger.println("MysqlConnectSuccess")

}
