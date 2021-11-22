package main

type ReportFillParam struct {
	Date   ParamFillDate `json:"date"`
	report *Report
}

func (reportFillParam *ReportFillParam) String() string {
	if reportFillParam.Date.Attribute != "" {
		reportFillParam.Date.report = reportFillParam.report
		return reportFillParam.Date.String()
	} else {
		return ""
	}
}
