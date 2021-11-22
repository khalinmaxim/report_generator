// Главный файл микросервиса

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var config Config

func main() {
	var err error
	var logLevel log.Level
	ctx := context.Background()
	// Чтение конфигурации
	log.SetOutput(os.Stdout)
	config.read()
	if logLevel, err = log.ParseLevel(config.LogLevel); err != nil {
		log.Fatal(err)
	}
	// Настройка журналирования
	log.SetLevel(logLevel)
	if logLevel == log.DebugLevel || logLevel == log.TraceLevel {
		log.SetReportCaller(true)
	}
	// Соединение с БД для ожидания уведомления
	var db *pgx.Conn
	if db, err = config.Database.getConnection(ctx); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := config.Database.disconnect(ctx); err != nil {
			log.Error(err)
		}
	}()
	// Установка канала уведомлений
	if _, err = db.Exec(ctx, fmt.Sprintf("/*NO LOAD BALANCE*/LISTEN \"%s\"", config.Database.Channel)); err != nil {
		log.Fatal(err)
	}
	// Цикл ожидания уведомлений по каналу PostgreSQL
	build(ctx)
	for {
		if db.IsClosed() {
			break
		}
		log.WithFields(log.Fields{
			"channel": config.Database.Channel,
			"timeout": time.Duration(config.Database.Timeout) * time.Second,
		}).Info("Ожидание уведомлений")
		// Ожидание уведомлений
		ctxW, _ := context.WithTimeout(ctx, config.Database.Timeout*time.Second)
		if notification, err := db.WaitForNotification(ctxW); err == nil {
			log.WithFields(log.Fields{
				"notify": notification,
			}).Info("Notify")
			go build(ctx)
		} else {
			log.Warning(err)
			go build(ctx)
		}
	}
}

func build(ctx context.Context) {
	var db *pgx.Conn
	var err error
	if db, err = config.Database.newConnection(ctx); err != nil {
		log.Error(err)
		return
	}
	if tx, err := db.Begin(ctx); err == nil {
		log.WithFields(log.Fields{
			"tx": tx,
		}).Debug("Begin")
		reports := Reports{ctx: ctx, db: db}
		reports.load()
		if err := tx.Commit(ctx); err == nil {
			log.WithFields(log.Fields{
				"tx": tx,
			}).Debug("Commited")
		} else {
			log.Error(err)
		}
		reports.execute()
	}
}
