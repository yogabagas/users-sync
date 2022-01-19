package service

import (
	"context"
	"my-github/users-sync/repository"
	"my-github/users-sync/shared"
	"time"

	"github.com/xuri/excelize/v2"
	"gitlab.sicepat.tech/platform/golib/log"
)

type ExcelData struct {
	Nik         string
	Name        string
	Role        string
	Directorate string
}

func Import() {
	xlsx, err := excelize.OpenFile("role_HR_performa_appraisal_19012021219.xlsx")
	if err != nil {
		log.Println(err)
	}

	sheetName := xlsx.GetSheetName(0)
	rows, err := xlsx.Rows(sheetName)
	if err != nil {
		log.Println(err)
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

		if parsedData.Nik == "" || columnCount < 4 {
			log.Print("end of file")
			break
		}

		excelDatas = append(excelDatas, parsedData)

	}

	log.Println("Scanned data:", len(excelDatas))

	for _, item := range excelDatas {
		data := &repository.UserData{
			NIK:         item.Nik,
			Name:        item.Name,
			Role:        item.Role,
			Directorate: item.Directorate,
			Status:      int(shared.StatusImported),
			Description: shared.StatusImported.String(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		repository.CreateOrUpdate(context.Background(), data)
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
	if length < 4 {
		log.Println("column number should be 4")
		return &ExcelData{}, length, nil
	}

	data = &ExcelData{
		Nik:         columns[0],
		Name:        columns[1],
		Role:        columns[2],
		Directorate: columns[3],
	}
	return data, length, nil
}
