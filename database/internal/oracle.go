package internal

//
//var (
//	ODb *gorm.DB
//)
//
//type OOptions struct {
//	Dbname       string      `json:"dbname"`
//	Host         string      `json:"host"`
//	User         string      `json:"user"`
//	Pass         string      `json:"pass"`
//	Port         int         `json:"port"`
//	PollMaxOpen  int         `json:"pollMaxOpen"`  //最大打开连接数
//	PollMinConns int         `json:"pollMinConns"` //最小保持活跃连接数
//	Log          *MOptionLog `json:"log"`
//}
//
//var oopt *OOptions = nil
//
//func NewODb(o *OOptions) *gorm.DB {
//	oopt = o
//	return odbInit()
//}
//
//func odbInit() *gorm.DB {
//	//sql输出日志级别
//	var err error
//
//	cnf := &gorm.Config{
//		NamingStrategy: schema.NamingStrategy{
//			TablePrefix:   oopt.Dbname + ".",
//			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
//		},
//		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键
//	}
//	if mopt.Log != nil {
//		//默认文件日志 不输出色彩
//		var logLevel = ormLogger.Error
//		if mopt.Log.Level != 0 {
//			logLevel = mopt.Log.Level
//		}
//		colorful := false
//		//debug仅打印到控制台
//		if mopt.Log.Std == nil {
//			mopt.Log.Std = log.New(colorable.NewColorableStdout(), "", log.LstdFlags)
//			colorful = true
//		}
//		cnf.Logger = ormLogger.New(
//			mopt.Log.Std, // io writer
//			ormLogger.Config{
//				SlowThreshold:             time.Second * 1, // 慢 SQL 阈值
//				LogLevel:                  logLevel,        //logger.Silent //不进行任何打印
//				Colorful:                  colorful,        // 色彩打印
//				IgnoreRecordNotFoundError: true,            //忽略查询未找到的错误
//			},
//		)
//	}
//	ODb, err = gorm.Open(oracle.Open(fmt.Sprintf(
//		"%s/%s@%s:%d/orcl",
//		oopt.User,
//		oopt.Pass,
//		oopt.Host,
//		oopt.Port,
//	)), cnf)
//	if err != nil {
//		panic(fmt.Sprintf("MysqlConnectFail: %s", err.Error()))
//		return nil
//	} else {
//		sqlDB, _ := ODb.DB()
//		sqlDB.SetMaxIdleConns(oopt.PollMinConns)
//		sqlDB.SetMaxOpenConns(oopt.PollMaxOpen)
//		sqlDB.SetConnMaxLifetime(time.Hour)
//	}
//	gLog.Println("OracleConnectSuccess")
//	return ODb
//}
