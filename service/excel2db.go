package service

import (
	"context"
	"fmt"
	"math"
	"my-github/users-sync/repository"
	"my-github/users-sync/shared"
	"my-github/users-sync/taskworker"
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
	xlsx, err := excelize.OpenFile("./excel/2022-01-28/Role_280120221615.xlsx")
	if err != nil {
		log.Println(err)
	}

	sheetName := xlsx.GetSheetName(0)
	rows, err := xlsx.Rows(sheetName)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(rows)

	var totalRowsScanned int64
	var skip int64 = 2
	var excelDatas []*ExcelData
	for rows.Next() {
		totalRowsScanned++
		if totalRowsScanned < skip {
			columns, _ := rows.Columns()
			log.WithFields(log.Fields{"header": columns}).Infoln("[Import User] check header")
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
	log.Println("Total Scanned Rows:", len(excelDatas))

	process := &ImportProcess{
		SuccessCount: 0,
		FailedCount:  0,
	}
	start := time.Now()
	processImport(context.Background(), process, excelDatas)
	duration := time.Since(start)
	log.Infoln("Done in:", int(math.Ceil(duration.Seconds())), "seconds")
	log.Infoln("Summary count success:", process.SuccessCount, "failed:", process.FailedCount)
}

func processImport(ctx context.Context, process *ImportProcess, excelDatas []*ExcelData) {
	maxWorker := 10
	worker := taskworker.NewSingleTaskWorker(ctx, uint8(maxWorker), task, len(excelDatas))
	for _, excelData := range excelDatas {
		worker.Do(excelData)
	}

	results := worker.Results()
	for _, result := range results {
		if result.Err != nil {
			process.FailedCount++
		} else {
			process.SuccessCount++
		}
	}
}

func task(ctx context.Context, data interface{}) (interface{}, error) {
	excelData, ok := data.(*ExcelData)
	if !ok {
		return excelData, taskworker.ErrorInvalidObject
	}

	userData := &repository.UserData{
		NIK:         excelData.Nik,
		Name:        excelData.Name,
		Role:        excelData.Role,
		Directorate: excelData.Directorate,
		Status:      int(shared.StatusImported),
		Description: shared.StatusImported.String(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := repository.CreateOrUpdate(ctx, userData)
	if err != nil {
		return excelData, err
	}
	return excelData, nil
}

type ImportProcess struct {
	SuccessCount int64
	FailedCount  int64
}

func parseRow(rows *excelize.Rows) (data *ExcelData, columnCount int, err error) {
	columns, _ := rows.Columns()
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
