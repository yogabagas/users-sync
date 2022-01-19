package service

import (
	"context"
	"fmt"
	"my-github/users-sync/repository"
	"my-github/users-sync/shared"
	"strings"
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

	rows := make([]ExcelData, 0)

	for i := 2; i <= len(xlsx.GetRows(sheetOne)); i++ {
		row := ExcelData{
			Nik:         xlsx.GetCellValue(sheetOne, fmt.Sprintf("A%d", i)),
			Name:        xlsx.GetCellValue(sheetOne, fmt.Sprintf("B%d", i)),
			Role:        strings.Split(xlsx.GetCellValue(sheetOne, fmt.Sprintf("C%d", i)), "/"),
			Directorate: xlsx.GetCellValue(sheetOne, fmt.Sprintf("D%d", i)),
		}
		rows = append(rows, row)
	}

	log.Println("Scanned data:", len(rows))

	for _, item := range rows {
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

// func parseRow(rows *excelize.Rows) (data *ExcelData, columnCount int, err error) {
// 	columns, err := rows.Columns()
// 	if err != nil {
// 		log.WithError(err).Errorln("[parseRow] error when read columns")
// 		return &ExcelData{}, 0, err
// 	}

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
