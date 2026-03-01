package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"supplyservicews/internal/config"
)

type Connections struct {
	Authorization *sql.DB
	Supply        *sql.DB
	Reference     *sql.DB
}

func NewConnections(ctx context.Context, cfg config.DBGroup) (*Connections, error) {
	authDB, err := openAndPing(ctx, cfg.Authorization)
	if err != nil {
		return nil, fmt.Errorf("connect authorization db: %w", err)
	}

	supplyDB, err := openAndPing(ctx, cfg.Supply)
	if err != nil {
		authDB.Close()
		return nil, fmt.Errorf("connect supply db: %w", err)
	}

	referenceDB, err := openAndPing(ctx, cfg.Reference)
	if err != nil {
		authDB.Close()
		supplyDB.Close()
		return nil, fmt.Errorf("connect reference db: %w", err)
	}

	return &Connections{
		Authorization: authDB,
		Supply:        supplyDB,
		Reference:     referenceDB,
	}, nil
}

func (c *Connections) Close() {
	if c == nil {
		return
	}
	if c.Authorization != nil {
		_ = c.Authorization.Close()
	}
	if c.Supply != nil {
		_ = c.Supply.Close()
	}
	if c.Reference != nil {
		_ = c.Reference.Close()
	}
}

func openAndPing(ctx context.Context, cfg config.DBConfig) (*sql.DB, error) {
	mysqlCfg := mysql.Config{
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:                 cfg.User,
		Passwd:               cfg.Password,
		DBName:               cfg.Name,
		ParseTime:            true,
		AllowNativePasswords: true,
		Loc:                  time.UTC,
	}

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
