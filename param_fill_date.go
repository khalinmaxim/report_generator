package main

import (
	"github.com/goodsign/monday"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"time"
)

type ParamFillDate struct {
	Attribute string `json:"attr"`
	Format    string `json:"format"`
	report    *Report
}

/*
	Возвращает содержимое атрибутов в виде строки в соответствии с форматом
*/

func (param *ParamFillDate) String() string {
	switch param.Attribute {
	case "today":
		return monday.Format(time.Now(), param.Format, monday.LocaleRuRU)
	default:
		// Проверяем есть ли в атрибутах число по regexp
		if res, err := regexp.MatchString(`[0-9]`, param.Attribute); err == nil {
			// если есть (res == True)
			if res {
				// конвертируем Атрибут(string) в int
				if index, err := strconv.Atoi(param.Attribute); err == nil {
					// Подставляем индекс в массив параметров запроса
					if t, err := time.Parse("2006-01-02T15:04:05-07:00", param.report.params.Query[index]); err == nil {
						// получаем дату и форматируем его согласно формату и возвращаем его
						return monday.Format(t, param.Format, monday.LocaleRuRU)
					} else {
						log.Error(err)
					}
				} else {
					log.Error(err)
				}
			}
		} else {
			log.Error(err)
		}
	}
	return ""
}
