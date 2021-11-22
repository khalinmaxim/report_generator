package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"time"
)

type Database struct {
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Name     string        `json:"name"`
	UserName string        `json:"username"`
	Password string        `json:"password"`
	Channel  string        `json:"channel"`
	Timeout  time.Duration `json:"timeout"`
	Limit    int           `json:"limit"`
	conn     *pgx.Conn
}

func (db *Database) getConnection(ctx context.Context) (*pgx.Conn, error) {
	var err error
	if db.conn == nil {
		db.conn, err = db.newConnection(ctx)
	}
	return db.conn, err
}

func (db *Database) disconnect(ctx context.Context) error {
	var err error
	if !db.conn.IsClosed() {
		if err := db.conn.Close(ctx); err != nil {
			log.Error(err)
		}
	}
	db.conn = nil
	return err
}

func (db *Database) newConnection(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx,
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Database.Host,
			config.Database.Port,
			config.Database.UserName,
			config.Database.Password,
			config.Database.Name,
		),
	)
	if err == nil {
		log.WithFields(log.Fields{
			"connection": conn,
		}).Debug("Новое соединение")
		return conn, nil
	} else {
		log.Error(err)
		return nil, err
	}
}
