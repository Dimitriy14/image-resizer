package postgres

import (
	"fmt"

	"github.com/Dimitriy14/image-resizing/config"
	"github.com/Dimitriy14/image-resizing/logger"
	"github.com/Dimitriy14/image-resizing/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const dbInfo = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"

var Client *PGClient

type PGClient struct {
	Session *gorm.DB
}

const pg = "postgres"

func Load() error {
	url := fmt.Sprintf(
		dbInfo,
		config.Conf.Postgres.Host,
		config.Conf.Postgres.Port,
		config.Conf.Postgres.User,
		config.Conf.Postgres.Password,
		config.Conf.Postgres.DBName,
	)

	fmt.Println(url)

	db, err := gorm.Open(pg, url)
	if err != nil {
		return fmt.Errorf("connecting to postgress: %s", err)
	}

	Client = &PGClient{Session: db}
	db.SetLogger(logger.NewGormLogger(logger.Log))
	db.LogMode(true)

	db.AutoMigrate(&models.Images{})
	return nil
}
