package internal

import (
	"context"
	"fmt"
	sql "github.com/go-sql-driver/mysql"
	gLog "github.com/lsclh/gtools/log"
	"github.com/mattn/go-colorable"
	"golang.org/x/crypto/ssh"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	ormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	Db *gorm.DB
)

var mOnce sync.Once

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

var mopt *MOptions = nil

func NewMDb(o *MOptions) *gorm.DB {
	mopt = o
	if Db == nil {
		return mdbInit()
	}
	return Db
}

// Setup : Connect to mysql database
func mdbInit() *gorm.DB {
	//sql输出日志级别
	var err error
	link := ""
	//是否使用ssh代理连接
	if mopt.Ssh != nil {

		sshConfig := &ssh.ClientConfig{
			User:            mopt.Ssh.User,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		if mopt.Ssh.Pass != "" {
			sshConfig.Auth = []ssh.AuthMethod{ssh.Password(mopt.Ssh.Pass)}
		} else {
			// 读取秘钥文件
			k, err := os.ReadFile(mopt.Ssh.PublicKey)
			if err != nil {
				panic(fmt.Sprintf("MysqlSSHReadPublicKeyFail: %s", err.Error()))
				return nil
			}
			//创建ssh签名
			signer, err := ssh.ParsePrivateKey(k)
			if err != nil {
				panic(fmt.Sprintf("数据库ssh连接秘钥解析错误: %s", err.Error()))
				return nil
			}
			//设置验证
			sshConfig.Auth = []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			}
		}
		// 创建ssh连接
		sshcon, err := ssh.Dial("tcp", mopt.Ssh.Host, sshConfig)
		if err != nil {
			panic(fmt.Sprintf("数据库ssh连接失败: %s", err.Error()))
			return nil
		}
		//注册ssh代理
		sql.RegisterDialContext("mysqlssh", func(ctx context.Context, addr string) (net.Conn, error) {
			return sshcon.Dial("tcp", addr)
		})
		link = fmt.Sprintf("%s:%s@mysqlssh(%s)/%s?charset=utf8&parseTime=True&loc=Local", mopt.User, mopt.Pass, mopt.Host, mopt.Dbname)
	} else {
		link = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", mopt.User, mopt.Pass, mopt.Host, mopt.Dbname)
	}
	cnf := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键
	}

	if mopt.Log != nil {
		//默认文件日志 不输出色彩
		var logLevel = ormLogger.Error
		if mopt.Log.Level != 0 {
			logLevel = mopt.Log.Level
		}
		colorful := false
		//debug仅打印到控制台
		if mopt.Log.Std == nil {
			mopt.Log.Std = log.New(colorable.NewColorableStdout(), "", log.LstdFlags)
			colorful = true
		}
		cnf.Logger = ormLogger.New(
			mopt.Log.Std, // io writer
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
		return nil
	} else {
		sqlDB, _ := Db.DB()
		sqlDB.SetMaxIdleConns(mopt.PollMinConns)
		sqlDB.SetMaxOpenConns(mopt.PollMaxOpen)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
	gLog.Println("MysqlConnectSuccess")
	return Db
}
