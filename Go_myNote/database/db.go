package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// "gorm.io/gorm/logger"
	// "log"
	// "os"
	// "time"
)

func NewDatabaseConnection(dsn string) (*gorm.DB, error) {
	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags),
	// 	logger.Config{
	// 		SlowThreshold: time.Second,
	// 		LogLevel:      logger.Info,
	// 		Colorful:      true,
	// 	},
	// )
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: newLogger,
		SkipDefaultTransaction: true,

	})
}
