package xlsx

import (
	"fmt"
	"github.com/tealeg/xlsx/v3"
	"github.com/tony-bondarenko/tradetools"
)

const (
	limitColor   = "FF00FFFF"
	limitOffset  = 8
	limitPerRow  = 2
	offsetTicker = 1
	offsetPrice  = 4
	offsetLots   = 5
)

type LimitReader struct {
	file       *xlsx.File
	rowPtr     int
	limitIndex int
}

func CreateLimitReader(filePath string) (*LimitReader, error) {
	file, err := xlsx.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	if len(file.Sheets) != 1 {
		return nil, fmt.Errorf("file have more then 1 sheet")
	}
	return &LimitReader{file, 0, 0}, nil
}

func (reader *LimitReader) NextLimit() (*tradetools.Limit, error) {
	sheet := reader.file.Sheets[0]
	for ; reader.rowPtr < sheet.MaxRow; reader.rowPtr++ {
		row, err := sheet.Row(reader.rowPtr)
		if err != nil {
			return nil, err
		}

		for ; reader.limitIndex < limitPerRow; reader.limitIndex++ {
			cell := row.GetCell(reader.limitIndex*limitOffset + offsetTicker)
			if cell.GetStyle().Fill.BgColor == limitColor {
				limit := new(tradetools.Limit)
				limit.Ticker = cell.Value

				cell = row.GetCell(reader.limitIndex*limitOffset + offsetPrice)
				limit.Price, err = cell.Float()
				if err != nil {
					return nil, err
				}

				cell = row.GetCell(reader.limitIndex*limitOffset + offsetLots)
				limit.Lots, err = cell.Int()
				if err != nil {
					return nil, err
				}

				reader.limitIndex++
				if reader.limitIndex >= limitPerRow {
					reader.rowPtr++
					reader.limitIndex = 0
				}
				return limit, nil
			}
		}
		reader.limitIndex = 0
	}
	return nil, nil
}
