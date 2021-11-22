// Описывает структуру и функции загрузки и сохранения

package main

import (
	"context"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

// Массив из заданий
type Reports struct {
	ctx        context.Context
	db         *pgx.Conn
	loadStmt   *pgconn.StatementDescription
	removeStmt *pgconn.StatementDescription
	reports    []Report
}

// загрузка задания из базы
func (reports *Reports) load() {
	var rows pgx.Rows
	var err error

	if config.Dry {
		if rows, err = reports.db.Query(reports.ctx, `
		SELECT
		  t.entity_uuid AS template_id,
		  report_title,
		  report_query,
		  report_params,
		  template_content,
		  report_fill,
		  coalesce(report_path, '') AS report_path
		FROM reporter.queue q 
		JOIN reporter.templates t ON queue_object = t.entity_uuid
	`); err != nil {
			log.Fatal(err)
		}
	} else {
		if rows, err = reports.db.Query(reports.ctx, `
		WITH r AS (
			  DELETE FROM reporter.queue
			  RETURNING queue_object, report_params
		)
		SELECT
		  t.entity_uuid AS template_id,
		  report_title,
		  report_query,
		  report_params,
		  template_content,
		  report_fill,
          coalesce(report_path, '') AS report_path
		FROM r
		JOIN reporter.templates t ON r.queue_object = t.entity_uuid
	`); err != nil {
			log.Fatal(err)
		}
	}

	defer rows.Close()
	// Цикл по всем полученным строкам
	if rows.Err() == nil {
		for rows.Next() {
			r := Report{db: reports.db, ctx: reports.ctx}
			if err := r.load(rows); err == nil {
				reports.reports = append(reports.reports, r)
			} else {
				log.Error(err)
			}
		}
	} else {
		log.Error(rows.Err())
	}
}

// Исполняет задание
func (reports *Reports) execute() {
	for _, r := range reports.reports {
		r.execute()
	}
}
