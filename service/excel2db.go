package service

import (
	"context"
	"fmt"
	"math"
	"my-github/users-sync/repository"
	"my-github/users-sync/shared"
	"my-github/users-sync/taskworker"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"gitlab.sicepat.tech/platform/golib/log"
)

type ExcelData struct {
	Nik         string
	Name        string
	Role        []string
	Directorate string
}

const (
	sheetOne = "Sheet1"
)

func Import() {
	xlsx, err := excelize.OpenFile("./hpan-20220119.xlsx")
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

	for i := 2; i <= len(xlsx.GetRows(sheetOne)); i++ {
		row := ExcelData{
			Nik:         xlsx.GetCellValue(sheetOne, fmt.Sprintf("A%d", i)),
			Name:        xlsx.GetCellValue(sheetOne, fmt.Sprintf("B%d", i)),
			Role:        strings.Split(xlsx.GetCellValue(sheetOne, fmt.Sprintf("C%d", i)), "/"),
			Directorate: xlsx.GetCellValue(sheetOne, fmt.Sprintf("D%d", i)),
		}
		rows = append(rows, row)
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

	// for _, item := range excelDatas {
	// 	data := &repository.UserData{
	// 		NIK:         item.Nik,
	// 		Name:        item.Name,
	// 		Role:        item.Role,
	// 		Directorate: item.Directorate,
	// 		Status:      int(shared.StatusImported),
	// 		Description: shared.StatusImported.String(),
	// 		CreatedAt:   time.Now(),
	// 		UpdatedAt:   time.Now(),
	// 	}
	// 	repository.CreateOrUpdate(context.Background(), data)
	// 	log.Println(item)
	// }
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

// 	length := len(columns)
// 	if length < 4 {
// 		log.Println("column number should be 4")
// 		return &ExcelData{}, length, nil
// 	}

// 	data = &ExcelData{
// 		Nik:         columns[0],
// 		Name:        columns[1],
// 		Role:        columns[2],
// 		Directorate: columns[3],
// 	}
// 	return data, length, nil
// }
