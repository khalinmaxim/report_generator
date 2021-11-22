package main

import (
	"bytes"
	"context"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Report Структура задания
type Report struct {
	ctx        context.Context
	db         *pgx.Conn
	id         string // id задания
	title      string // наименование отчета
	params     Params // параметры
	query      string // Запрос
	templateId string // id Шаблона
	template   []byte // Шаблон
	reportfill []ReportFill
	results    []Result
	reportPath string // Пусть сохранения отчета
}

type Params struct {
	Query []string `json:"query"`
	Title string   `json:"title"`
}

type Result struct {
	Sheet   int      `json:"sheet"`
	Row     Row      `json:"row"`
	Columns []Column `json:"columns"`
}

type Row struct {
	Add    bool   `json:"add"`
	Column int    `json:"column"`
	Needle string `json:"needle"`
}

type Column struct {
	Number int     `json:"number"`
	Value  float32 `json:"value"`
	Offset int     `json:"offset"`
}

// Загружает шаблон и параметры отчёта
func (report *Report) load(rows pgx.Rows) error {
	if err := rows.Scan(&report.templateId, &report.title, &report.query, &report.params, &report.template, &report.reportfill, &report.reportPath); err == nil {
		log.WithFields(log.Fields{
			"query":      report.query,
			"template":   len(report.template),
			"params":     report.params,
			"fillobj":    report.reportfill,
			"reportPath": report.reportPath,
		}).Info("Задание")
		return nil
	} else {
		return err
	}

}

// Выполняет задание по отчёту
func (report *Report) execute() {
	// Получаем параметры и запрос
	// Преобразовываем массив параметров в интерфейс для передачи его к запросу
	params := make([]interface{}, len(report.params.Query))
	for i, v := range report.params.Query {
		params[i] = v
	}
	log.WithFields(log.Fields{
		"query":  report.query,
		"params": params,
	}).Info("Выполнение")
	// Точка старта времени выполнения запроса
	timeStart := time.Now().Unix()
	if rows, err := report.db.Query(report.ctx, report.query, params...); err == nil {
		defer rows.Close()
		// Читаем строки результата
		for rows.Next() {
			result := Result{}
			if err = rows.Scan(&result); err == nil {
				report.results = append(report.results, result)
			} else {
				log.Error(err)
			}
		}
		// Точка финиша времени выполнения запроса
		timeFinish := time.Now().Unix()
		// Вывод времени выполнения запроса
		log.WithFields(log.Fields{
			"Time": timeFinish - timeStart,
		}).Info("Время выполнения запроса")
		// Проверка на ошибки внутри Rows
		if err = rows.Err(); err == nil {
			// Вызываем функцию заполнения
			// точка старта заполнения запроса
			timeStart := time.Now().Unix()
			if xlsx, err := report.fill(); err == nil {
				if !config.Dry || config.Test {
					if err := report.saveDataBase(xlsx); err != nil {
						// Ошибка сохранения в Базу
						log.Error(err)
					}
					if err := report.saveFile(xlsx); err != nil {
						// Ошибка сохранения в файл
						log.Error(err)
					}
				}
				//точка финиша выполнения запроса
				timeFinish := time.Now().Unix()
				log.WithFields(log.Fields{
					"Time": timeFinish - timeStart,
				}).Info("Время заполнения шаблона")
			} else {
				log.Error(err)
			}
		} else {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

// Функция поиска строки в которой находится Needle
func (report *Report) findRowIndex(sheetCache [][]string, needle string, column int, add bool) int {
	// Сравниваем needle и ячейку колонки
	for y, needleI := range sheetCache[column] {
		if needleI == needle {
			return y
		}
	}

	// Если не надо добавлять needle в случае его не нахождения возвращаем -1 и needle не заполнится
	if !add {
		return -1
	}

	// Возвращаем последнюю строчку в которой заполнен Needle
	return len(sheetCache[column])
}

// Заполняет шаблон
func (report *Report) fill() (xlsx *excelize.File, err error) {
	// Передаем байтовый шаблон в переменную
	r := bytes.NewReader(report.template)

	// Открываем посредством библиотеки рабочую книгу
	if xlsx, err = excelize.OpenReader(r); err == nil {
		// Блок отвечающий за заполнения дополнительных ячеек
		for x := 0; x < len(report.reportfill); x++ {
			reportFill := report.reportfill[x]
			reportFill.report = report
			if cell, err := excelize.CoordinatesToCellName(reportFill.Column, reportFill.Row); err == nil {
				if err := xlsx.SetCellValue(xlsx.GetSheetName(reportFill.Sheet), cell, reportFill.String()); err != nil {
					log.Error(err)
				}
			} else {
				log.Error(err)
			}
		}
		// Цикл результатов
		// Отправляем в функцию данные и получаем индекс строки с расположением Needle
		sheetCache := make(map[string][][]string)

		for i := 0; i < len(report.results); i++ {
			result := report.results[i]
			// Получаем индекс листа
			listName := xlsx.GetSheetName(result.Sheet)
			if _, ok := sheetCache[listName]; !ok {
				sheetCache[listName], _ = xlsx.GetCols(listName)
				log.Info(sheetCache)
			}
			// Проверка на наличие листа
			if listName == "" {
				log.WithFields(log.Fields{
					"Needle":   result.Row.Needle,
					"Template": report.title,
					"Sheet":    strconv.Itoa(result.Sheet + 1),
					"Column":   strconv.Itoa(result.Row.Column),
					"Time":     time.Now().Format("15:04:05"),
				}).Error("Лист не найден")
			} else {
				// Функция поиска индекса строки в которую мы передаем: наш файл, название листа, Needle
				if rowIndex := report.findRowIndex(sheetCache[listName], result.Row.Needle, result.Row.Column, result.Row.Add); rowIndex != -1 {
					// Функция заполнения
					for x := 0; x < len(result.Columns); x++ {
						// Получаем ячейку
						if cell, err := excelize.CoordinatesToCellName(result.Columns[x].Number, rowIndex+1+result.Columns[x].Offset); err == nil {
							// Заполняем needle даже если он не существует, но требует заполнения
							log.WithFields(log.Fields{
								"Needle":   result.Row.Needle,
								"Template": report.title,
								"Sheet":    strconv.Itoa(result.Sheet + 1),
								"Row":      rowIndex + 1 + result.Columns[x].Offset,
								"Column":   strconv.Itoa(result.Row.Column),
								"Time":     time.Now().Format("15:04:05"),
							}).Debug("Заполнение ячейки")
							if err = xlsx.SetCellValue(listName, cell, result.Columns[x].Value); err != nil {
								log.WithFields(log.Fields{
									"Needle":   result.Row.Needle,
									"Template": report.title,
									"Sheet":    strconv.Itoa(result.Sheet + 1),
									"Column":   strconv.Itoa(result.Row.Column),
								}).Error("Не заполняется ячейка")
							}
						} else {
							log.WithFields(log.Fields{
								"Needle":   result.Row.Needle,
								"Template": report.title,
								"Sheet":    strconv.Itoa(result.Sheet + 1),
								"Column":   strconv.Itoa(result.Columns[x].Number),
								"Row":      strconv.Itoa(rowIndex + 1 + result.Columns[x].Offset),
							}).Error("Строка не найдена")
						}
						// Если Needle не найден - создаем его здесь
						if rowIndex == len(sheetCache[listName][result.Row.Column]) {
							if cell, err := excelize.CoordinatesToCellName(result.Row.Column+1, rowIndex+1); err == nil {
								if err = xlsx.SetCellValue(listName, cell, result.Row.Needle); err != nil {
									log.WithFields(log.Fields{
										"Needle":   result.Row.Needle,
										"Template": report.title,
										"Sheet":    strconv.Itoa(result.Sheet + 1),
										"Column":   strconv.Itoa(result.Row.Column + 1),
									}).Error("Не заполняется ячейка Needle")
								}
							} else {
								log.Error(err)
							}
						}
					}
				}
			}
		}
		return xlsx, nil
	} else {
		return nil, err
	}
}

// Сохраняет отчёт в Базу Данных
func (report *Report) saveDataBase(xlsx *excelize.File) error {

	// Сохраняем в БД
	buff, _ := xlsx.WriteToBuffer()
	if rows, err := report.db.Query(report.ctx, "SELECT reporter.add_report($1::uuid,$2::text,$3::bytea)", report.templateId, report.params.Title, buff.Bytes()); err == nil {
		defer rows.Close()
		log.Info("Файл сохранен в Базу Данных")
	} else {
		log.Error(err)
	}

	return nil
}

// Сохраняем файл в зависимости от режима
func (report *Report) saveFile(xlsx *excelize.File) error {
	buff, _ := xlsx.WriteToBuffer()
	var pathFile string
	if config.Test {
		pathFile = config.Testpath
	} else {
		pathFile = config.Outpath
	}
	var path = filepath.Join(pathFile, report.reportPath)
	if _, err := os.Stat(path); err != nil {
		// Если путь не найден
		if os.IsNotExist(err) {
			// Создаем путь
			if err = os.MkdirAll(path, 0777); err != nil {
				// Путь не создался
				return err
			}
		}
	}

	filepathSave := filepath.Join(path, report.params.Title)

	if err := ioutil.WriteFile(filepathSave, buff.Bytes(), 0777); err == nil {
		log.WithFields(log.Fields{
			"File": filepathSave,
		}).Info("Файл сохранен")
	} else {
		return err
	}

	return nil
}
