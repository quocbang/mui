package barcodes

import (
	boombulerBarcode "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf/contrib/barcode"
)

type Generator interface {
	GetSize(boxWidth, boxHeight float64) (width, height float64)
	Generate(id string, width, height int) (string, error)
}

type Code39 struct{}

func (b Code39) GetSize(_, _ float64) (width, height float64) {
	return 40, 40
}

func (b Code39) Generate(id string, width, height int) (string, error) {
	// encode to code39
	barcode39, err := code39.Encode(id, false, true)
	if err != nil {
		return "", err
	}

	// Scaling to avoid broken barcode
	barcodeCode39, err := boombulerBarcode.Scale(barcode39, width, height)
	if err != nil {
		return "", err
	}

	return barcode.Register(barcodeCode39), nil
}

type QRCode struct{}

func (b QRCode) Generate(id string, width, height int) (string, error) {
	// encode to QRCode
	qrCode, err := qr.Encode(id, qr.M, qr.Auto)
	if err != nil {
		return "", err
	}

	// Scaling to avoid broken barcode
	qrCode, err = boombulerBarcode.Scale(qrCode, width, height)
	if err != nil {
		return "", err
	}

	return barcode.Register(qrCode), nil
}

func (b QRCode) GetSize(boxWidth, boxHeight float64) (width, height float64) {
	return boxWidth / 3, boxHeight / 3
}
