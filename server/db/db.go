package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cntechpower/utils/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var MySQL *sql.DB
var ConfigDB *gorm.DB
var MemDB *gorm.DB
var header *log.Header

func Init(dsn string, persistModels []interface{}, tmpModels []interface{}) {
	if dsn == "" {
		panic(fmt.Errorf("mysql dsn is empty"))
	}
	header = log.NewHeader("db")
	initGorm(persistModels, tmpModels)
	header.Infof("GORM init finish")
	go func() {
		time.Sleep(5 * time.Second)
		var err error
		MySQL, err = sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		MySQL.SetConnMaxLifetime(time.Minute * 120)
		MySQL.SetMaxIdleConns(10)
		header.Infof("MySQL init finish")
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			if err := MySQL.PingContext(ctx); err != nil {
				header.Infof("db ping check error: %v", err)
			}
			cancel()
			time.Sleep(30 * time.Second)
		}
	}()
}

func initGorm(persistModels []interface{}, tmpModels []interface{}) {
	var err error
	ConfigDB, err = gorm.Open(sqlite.Open("proxy.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	_ = ConfigDB.AutoMigrate(persistModels...)

	MemDB, err = gorm.Open(sqlite.Open("memory.db?cache=shared&mode=memory"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	_ = MemDB.AutoMigrate(tmpModels...)
}

func Close() {
	if MySQL != nil {
		_ = MySQL.Close()
	}
}