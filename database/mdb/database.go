package mdb

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	sql "github.com/go-sql-driver/mysql"
	oracle "github.com/godoes/gorm-oracle"
	gLog "github.com/lsclh/gtools/log"
	"github.com/mattn/go-colorable"
	"golang.org/x/crypto/ssh"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	ormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type mOptions struct {
	Dbname       string         `json:"dbname"`
	Host         string         `json:"host"`
	User         string         `json:"user"`
	Pass         string         `json:"pass"`
	Port         int            `json:"port"`
	PollMaxOpen  int            `json:"pollMaxOpen"`  //最大打开连接数
	PollMinConns int            `json:"pollMinConns"` //最小保持活跃连接数
	Log          *mOptionLog    `json:"log"`
	Ssh          *mOptionSSH    `json:"ssh"`
	Oracle       *mOptionOracle `json:"oracle"`
}

type mOptionOracle struct {
	ServiceName string `json:"oracle"`
}

type mOptionSSH struct {
	Host      string `json:"host"`
	User      string `json:"user"`
	Pass      string `json:"pass"`
	PublicKey string `json:"publicKey"`
}

type mOptionLog struct {
	Level ormLogger.LogLevel `json:"level"`
	Std   ormLogger.Writer   `json:"std"`
}

var (
	mopt         *mOptions = nil
	databaseType           = "Mysql"
)

func newMDb(o *mOptions) *gorm.DB {
	mopt = o
	if mopt.Oracle != nil {
		databaseType = "Oracle"
	}
	return mdbInit()
}

// Setup : Connect to mysql database
func mdbInit() *gorm.DB {

	//验证参数
	verifyParams()

	//初始化dialector与cnf
	dialector, cnf := buildDialector(sshProxy())

	//设置log
	buildLog(cnf)

	//连接数据库
	db, err := gorm.Open(dialector, cnf)
	if err != nil {
		panic(fmt.Sprintf("%sConnectFail: %s", databaseType, err.Error()))
	}
	//设置会话参数
	setSessionParams(db)

	gLog.Println("%sConnectSuccess", databaseType)
	return db
}

func buildLog(cnf *gorm.Config) {
	if mopt.Log == nil {
		return
	}
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
func verifyParams() {
	if mopt.Host == "" {
		panic(fmt.Sprintf("%sHostFail: empty", databaseType))
	}
	if mopt.Ssh != nil && databaseType != "Mysql" {
		panic(fmt.Sprintf("%sSSHFail: Does not support non-MYSQL databases", databaseType))
	}
}

func buildDialector(network string) (gorm.Dialector, *gorm.Config) {
	switch databaseType {
	case "Mysql":
		return mysql.Open(
			fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
				mopt.User, mopt.Pass, network, mopt.Host, mopt.Port, mopt.Dbname)), gormConfig()
	case "Oracle":
		// oracle://{{user}}:{{pwd}}@{{host}}:{{port}}/{{service}}?CONNECTION TIMEOUT=30&LANGUAGE=SIMPLIFIED+CHINESE&TERRITORY=CHINA
		link := oracle.BuildUrl(mopt.Host, mopt.Port, mopt.Oracle.ServiceName, mopt.User, mopt.Pass, map[string]string{
			"CONNECTION TIMEOUT": "30",
			"LANGUAGE":           "SIMPLIFIED CHINESE",
			"TERRITORY":          "CHINA",
		})
		cnf := gormConfig()
		cnf.NamingStrategy = schema.NamingStrategy{
			TablePrefix:         mopt.Dbname + ".",
			SingularTable:       true,
			NoLowerCase:         true, // 是否不自动转换小写表名
			IdentifierMaxLength: 30,   // Oracle: 30, PostgreSQL:63, MySQL: 64, SQL Server、SQLite、DM: 128
		}
		cnf.SkipDefaultTransaction = true // 是否禁用默认在事务中执行单次创建、更新、删除操作
		cnf.PrepareStmt = false           // 创建并缓存预编译语句，启用后可能会报 ORA-01002 错误
		cnf.CreateBatchSize = 50          // 插入数据默认批处理大小
		return oracle.New(oracle.Config{
			DSN:                       link,
			IgnoreCase:                false,     // 与Oracle默认行为一致，区分查询条件大小写
			NamingCaseSensitive:       true,      // 尊重数据库对象的大小写定义
			VarcharSizeIsCharLength:   true,      // 中文环境更实用，按字符数计算长度
			RowNumberAliasForOracle11: "ROW_NUM", // 保持默认
		}), cnf
	default:
		panic(fmt.Sprintf("%sConnectFail: unknown database type", databaseType))
	}
}

func setSessionParams(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(mopt.PollMinConns)
	sqlDB.SetMaxOpenConns(mopt.PollMaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if databaseType == "Oracle" {
		_, _ = oracle.AddSessionParams(sqlDB, map[string]string{
			"TIME_ZONE":               "+08:00",                       // ALTER SESSION SET TIME_ZONE = '+08:00';
			"NLS_DATE_FORMAT":         "YYYY-MM-DD",                   // ALTER SESSION SET NLS_DATE_FORMAT = 'YYYY-MM-DD';
			"NLS_TIME_FORMAT":         "HH24:MI:SSXFF",                // ALTER SESSION SET NLS_TIME_FORMAT = 'HH24:MI:SS.FF3';
			"NLS_TIMESTAMP_FORMAT":    "YYYY-MM-DD HH24:MI:SSXFF",     // ALTER SESSION SET NLS_TIMESTAMP_FORMAT = 'YYYY-MM-DD HH24:MI:SS.FF3';
			"NLS_TIME_TZ_FORMAT":      "HH24:MI:SS.FF TZR",            // ALTER SESSION SET NLS_TIME_TZ_FORMAT = 'HH24:MI:SS.FF3 TZR';
			"NLS_TIMESTAMP_TZ_FORMAT": "YYYY-MM-DD HH24:MI:SSXFF TZR", // ALTER SESSION SET NLS_TIMESTAMP_TZ_FORMAT = 'YYYY-MM-DD HH24:MI:SS.FF3 TZR';
		})
	}
}

func gormConfig() *gorm.Config {
	return &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键
	}
}

func sshProxy() string {

	if mopt.Ssh == nil {
		return "tcp"
	}

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
			panic(fmt.Sprintf("%sSSHReadPublicKeyFail: %s", databaseType, err.Error()))
		}
		//创建ssh签名
		signer, err := ssh.ParsePrivateKey(k)
		if err != nil {
			panic(fmt.Sprintf("%sSSHReadPublicKeyParseFail: %s", databaseType, err.Error()))
		}
		//设置验证
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}
	// 创建ssh连接
	sshcon, err := ssh.Dial("tcp", mopt.Ssh.Host, sshConfig)
	if err != nil {
		panic(fmt.Sprintf("%sSSHConnectFail: %s", databaseType, err.Error()))
	}
	//注册ssh代理
	sql.RegisterDialContext("mysqlssh", func(ctx context.Context, addr string) (net.Conn, error) {
		return sshcon.Dial("tcp", addr)
	})
	return "mysqlssh"
}
