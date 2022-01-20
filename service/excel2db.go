package service

import (
	"context"
	"fmt"
	"my-github/users-sync/repository"
	"my-github/users-sync/shared"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"gitlab.sicepat.tech/platform/golib/log"
)

type ExcelData struct {
	Nik         string
	Name        string
	Role        string
	Directorate string
}

const (
	sheetOne = "Sheet1"
)

func Import() {
	xlsx, err := excelize.OpenFile("role_HR_performa_appraisal_19012021219.xlsx")
	if err != nil {
		log.Println(err)
	}

	rows := make([]ExcelData, 0)

	log.Println(xlsx.GetRows(sheetOne))

	for i := 2; i <= len(xlsx.GetRows(sheetOne)); i++ {
		row := ExcelData{
			Nik:         xlsx.GetCellValue(sheetOne, fmt.Sprintf("A%d", i)),
			Name:        xlsx.GetCellValue(sheetOne, fmt.Sprintf("B%d", i)),
			Role:        xlsx.GetCellValue(sheetOne, fmt.Sprintf("C%d", i)),
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
