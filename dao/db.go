package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/cntechpower/utils/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var mySQL *sql.DB
var configDB *gorm.DB
var memDB *gorm.DB
var header *log.Header

func Init(dsn string, persistModels []interface{}, tmpModels []interface{}) {
	header = log.NewHeader("db")
	initGorm(persistModels, tmpModels)
	header.Infof("GORM init finish")
	if dsn != "" {
		go func() {
			time.Sleep(5 * time.Second)
			var err error
			mySQL, err = sql.Open("mysql", dsn)
			if err != nil {
				panic(err)
			}
			mySQL.SetConnMaxLifetime(time.Minute * 120)
			mySQL.SetMaxIdleConns(10)
			header.Infof("MySQL init finish")
			for {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				if err := mySQL.PingContext(ctx); err != nil {
					header.Infof("db ping check error: %v", err)
				}
				cancel()
				time.Sleep(30 * time.Second)
			}
		}()
	}
}

func initGorm(persistModels []interface{}, tmpModels []interface{}) {
	var err error
	if len(persistModels) != 0 {
		configDB, err = gorm.Open(sqlite.Open("proxy.db"), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		_ = configDB.AutoMigrate(persistModels...)
	}

	if len(tmpModels) != 0 {
		memDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		_ = memDB.AutoMigrate(tmpModels...)
	}

}

func Close() {
	if mySQL != nil {
		_ = mySQL.Close()
	}
}

func ConfigDB() *gorm.DB {
	if configDB == nil {
		panic("ConfigDB is not init")
	}
	return configDB
}

func MemDB() *gorm.DB {
	if memDB == nil {
		panic("MemDB is not init")
	}
	return memDB
}

func MySQL() *sql.DB {
	if mySQL == nil {
		panic("MySQL is not init")
	}
	return mySQL
}
