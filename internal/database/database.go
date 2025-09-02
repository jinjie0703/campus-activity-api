package database

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// 数据库初始化
func InitDB() (*sql.DB, error) {
	dsn := "root:2594817591@tcp(127.0.0.1:3306)/campus_activity?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// 设置最大连接时长 3 分钟，10 个同时并发和 10 个空闲连接数
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// 测试连接
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
