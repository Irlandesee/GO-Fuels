package db

import (
	"Irlandesee/GO-Fuels/src/models"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresHandler struct {
	DB *gorm.DB
}

func (h *PostgresHandler) Close() error {
	sqlDB, err := h.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func Migrate(db *gorm.DB) {

}

func NewPostgresHandler() (db *gorm.DB) {
	dsn := fmt.Sprintf("host=%s, port=%s, user=%s, database=%s, password=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_DATABASE"),
		os.Getenv("DB_PASSWORD"),
	)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}
	err = db.AutoMigrate(models.FuelData{})
	if err != nil {
		panic(err)
	}

	return db
}
