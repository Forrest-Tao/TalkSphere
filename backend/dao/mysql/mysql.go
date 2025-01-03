package mysql

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"imitation_go-project-demo/setting"
	"strconv"
	"strings"
	"time"
)

var DB *gorm.DB

type MyWriter struct {
	mLog *zap.Logger
}

func (m *MyWriter) Printf(format string, v ...interface{}) {
	logStr := fmt.Sprintf(format, v...)
	//利用zap记录日志
	m.mLog.Info(logStr)
}

func NewMyWriter() *MyWriter {
	logg := zap.L()
	return &MyWriter{mLog: logg}
}

func Init(cfg *setting.MysqlConfig) (err error) {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	//"user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := strings.Join([]string{cfg.User, ":", cfg.PassWord, "@tcp(", cfg.Host, ":", strconv.Itoa(cfg.Port), ")/", cfg.DB,
		"?charset=utf8&parseTime=true&loc=Local"}, "")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			//设置Logger
			NewMyWriter(),
			logger.Config{
				//慢SQL阈值
				SlowThreshold: time.Millisecond,
				//设置日志级别，打印sql
				LogLevel: logger.Info,
			},
		),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 默认不加复数
		}})
	if err != nil {
		zap.L().Fatal("Connect DB failed ", zap.Error(err))
	}
	sqlDB, _ := db.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(20)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(200)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	DB = db
	return nil
}
