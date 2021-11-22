package main

import (
	"fmt"
)

type ReportFill struct {
	Row              int               `json:"row"`
	Column           int               `json:"column"`
	Format           string            `json:"format"`
	ReportFillParams []ReportFillParam `json:"params"`
	Sheet            int               `json:"sheet"`
	report           *Report
}

/*
	Возвращает содержимое параметров в виде строки в соответствии с форматом
*/
func (reportFill *ReportFill) String() string {
	params := make([]interface{}, len(reportFill.ReportFillParams))
	for i, v := range reportFill.ReportFillParams {
		v.report = reportFill.report
		params[i] = v.String()
	}
	return fmt.Sprintf(reportFill.Format, params...)
}
