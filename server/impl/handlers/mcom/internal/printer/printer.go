package printer

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/barcode"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils/barcodes"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
)

type PrintData struct {
	StationID      string
	NextStationID  string
	ProductID      string
	ProductionDate time.Time
	ExpiryDate     time.Time
	Quantity       decimal.Decimal
	ResourceID     string
}

func CreateResourcesPDF(ctx context.Context, fieldName models.MaterialResourceLabelFieldName, dataIn PrintData, generator barcodes.Generator, fontPath string) (io.ReadCloser, error) {
	if (fieldName == models.MaterialResourceLabelFieldName{}) {
		fieldName = models.MaterialResourceLabelFieldName{
			Station:        "工程機台別",
			NextStation:    "後工程機台別",
			ProductID:      "產品代號",
			ProductionDate: "製造日期",
			ExpiryDate:     "有效期限",
			Quantity:       "收料數量",
			ResourceID:     "收料條碼",
		}
	}
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size: gofpdf.SizeType{
			Wd: 182,
			Ht: 128,
		},
	})

	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	font := float64(12)

	// use font that can display the local language.
	// like use kaiu.ttf(標楷體) to show chinese.
	// if use gofpdf.AddUTF8Font import fonts, you can not set styleStr(ex:boldataIn...)
	// gofpdf can not use *.ttc file
	pdf.AddUTF8Font("font", "", fontPath)
	pdf.SetFont("Helvetica", "BI", font*2)

	pageWidth, pageHeight := pdf.GetPageSize()
	boxWidth, boxHeight := math.Abs(pageWidth/2), math.Abs(pageHeight/3)

	barcodeWidth, barcodeHeight := generator.GetSize(boxWidth, boxHeight)

	baseX, baseY := boxWidth*0.2, barcodeHeight/3
	x, y := baseX, baseY
	pdf.AddPage()

	//Kenda title
	pdf.SetXY(x, y)
	pdf.CellFormat(pageWidth*0.8, font, "KENDA", "1", 2, "CM", false, 0, "")

	columnNameWidth := barcodeWidth
	dataWidth := pageWidth*0.8 - barcodeWidth

	pdf.SetFont("font", "", font)
	//Station
	y += font
	pdf.SetXY(x, y)
	pdf.CellFormat(columnNameWidth, font/2, fieldName.Station, "1", 0, "LM", false, 0, "")
	pdf.CellFormat(dataWidth, font/2, dataIn.StationID, "1", 0, "LM", false, 0, "")

	//BackStation
	y += font / 2
	pdf.SetXY(x, y)
	pdf.CellFormat(columnNameWidth, font/2, fieldName.NextStation, "1", 0, "LM", false, 0, "")
	pdf.CellFormat(dataWidth, font/2, dataIn.NextStationID, "1", 0, "LM", false, 0, "")

	//ProductID
	y += font / 2
	pdf.SetXY(x, y)
	pdf.CellFormat(columnNameWidth, font/2, fieldName.ProductID, "1", 0, "LM", false, 0, "")
	pdf.CellFormat(dataWidth, font/2, dataIn.ProductID, "1", 0, "LM", false, 0, "")

	//ProductionDate
	y += font / 2
	pdf.SetXY(x, y)
	pdf.CellFormat(columnNameWidth, font/2, fieldName.ProductionDate, "1", 0, "LM", false, 0, "")
	productionDate := fmt.Sprintf("%d/%02d/%02d",
		dataIn.ProductionDate.Year(), dataIn.ProductionDate.Month(), dataIn.ProductionDate.Day())
	pdf.CellFormat(dataWidth, font/2, productionDate, "1", 0, "LM", false, 0, "")

	//ExpiryDate
	y += font / 2
	pdf.SetXY(x, y)
	pdf.CellFormat(columnNameWidth, font/2, fieldName.ExpiryDate, "1", 0, "LM", false, 0, "")
	expiryDate := fmt.Sprintf("%d/%02d/%02d %02d:00:00",
		dataIn.ExpiryDate.Year(), dataIn.ExpiryDate.Month(), dataIn.ExpiryDate.Day(), dataIn.ExpiryDate.Hour())
	pdf.CellFormat(dataWidth, font/2, expiryDate, "1", 0, "LM", false, 0, "")

	//Quantity
	y += font / 2
	pdf.SetXY(x, y)
	pdf.CellFormat(columnNameWidth, font/2, fieldName.Quantity, "1", 0, "LM", false, 0, "")
	pdf.CellFormat(dataWidth, font/2, dataIn.Quantity.String(), "1", 0, "LM", false, 0, "")

	//Barcode
	y += font / 2
	pdf.SetXY(x, y)
	pdf.CellFormat(columnNameWidth, font*2, fieldName.ResourceID, "1", 0, "LM", false, 0, "")
	pdf.CellFormat(dataWidth, font*2, "", "RB", 0, "CT", false, 0, "")
	pdf.SetXY(x+barcodeWidth, y+font*1.6)
	pdf.CellFormat(dataWidth, font*2, dataIn.ResourceID, "", 0, "CT", false, 0, "")

	//Generate Barcode
	key, err := generator.Generate(dataIn.ResourceID, int(barcodeWidth)*10, int(font)*2)
	if err != nil {
		return nil, err
	}
	x += barcodeWidth * 1.25
	y += 1
	pdf.SetXY(x, y)

	barcode.Barcode(pdf, key, x, y, dataWidth*0.8, font*1.5, false)

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		if err := pdf.OutputAndClose(pipeWriter); err != nil {
			commonsCtx.Logger(ctx).Warn("failed to output and close pdf file", zap.Error(err))
		}
	}()

	return ioutil.NopCloser(pipeReader), nil
}
