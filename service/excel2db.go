package service

import (
	"github.com/xuri/excelize/v2"
	"gitlab.sicepat.tech/platform/golib/log"
)

type ExcelData struct {
	Nik  string
	Name string
	Role string
}

func Import() {
	xlsx, err := excelize.OpenFile("file.xlsx")
	if err != nil {
		log.Errorln(err)
	}

	sheetName := xlsx.GetSheetName(0)
	rows, err := xlsx.Rows(sheetName)
	if err != nil {
		log.Errorln(err)
	}

	var totalRowsScanned int64
	var skip int64 = 2
	var excelDatas []*ExcelData
	for rows.Next() {
		totalRowsScanned++
		if totalRowsScanned < skip {
			log.Println("skip header")
			continue
		}

		parsedData, columnCount, err := parseRow(rows)
		if err != nil {
			log.Errorln(err)
		}

		if parsedData.Nik == "" || columnCount < 3 {
			log.Print("end of file")
			break
		}

		excelDatas = append(excelDatas, parsedData)
	}

	log.Println("Scanned data:", len(excelDatas))

	for _, item := range excelDatas {
		log.Println(item)
	}
}

func parseRow(rows *excelize.Rows) (data *ExcelData, columnCount int, err error) {
	columns, err := rows.Columns()
	if err != nil {
		log.WithError(err).Errorln("[parseRow] error when read columns")
		return &ExcelData{}, 0, err
	}

	length := len(columns)
	if length < 3 {
		log.Println("column number should be 3")
		return &ExcelData{}, length, nil
	}

	data = &ExcelData{
		Nik:  columns[0],
		Name: columns[1],
		Role: columns[2],
	}
	return data, length, nil
}
