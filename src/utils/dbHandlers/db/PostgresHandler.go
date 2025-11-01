package db

import (
	"database/sql"
	"fmt"

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

func NewPostgresHandler(host, port, dbName, user, password string) (*PostgresHandler, error) {
	connectionStr := fmt.Sprintf(
		"host=%s user=%s dbname=%s port=%s password=%s sslmode=disable",
		host, user, dbName, port, password,
	)

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres db: %w", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}),
		&gorm.Config{TranslateError: true, PrepareStmt: true})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	return &PostgresHandler{DB: gormDB}, nil
}
