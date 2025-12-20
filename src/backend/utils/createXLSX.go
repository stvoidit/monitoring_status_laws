package utils

import (
	"github.com/xuri/excelize/v2"
)

func CreateXLSX(headers []string, rows [][]string) *excelize.File {
	const sheetName = "документы"
	f := excelize.NewFile()
	styleID, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "00000000"},
			{Type: "left", Style: 1, Color: "00000000"},
			{Type: "right", Style: 1, Color: "00000000"},
			{Type: "bottom", Style: 1, Color: "00000000"},
		},
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"},
	})
	f.SetSheetName("Sheet1", sheetName)
	for i := range headers {
		cord, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellStr(sheetName, cord, headers[i])
		f.SetCellStyle(sheetName, cord, cord, styleID)
	}
	for i := range rows {
		for j := range rows[i] {
			cord, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellStr(sheetName, cord, rows[i][j])
			f.SetCellStyle(sheetName, cord, cord, styleID)
		}
	}
	firstCol, _ := excelize.ColumnNumberToName(1)
	lastCol, _ := excelize.ColumnNumberToName(len(headers))
	f.SetColWidth(sheetName, firstCol, lastCol, 20.)
	return f
}
