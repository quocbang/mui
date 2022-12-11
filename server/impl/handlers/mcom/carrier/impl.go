package carrier

import (
	"context"
	"io"
	"io/ioutil"
	"math"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/barcode"
	"go.uber.org/zap"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils/barcodes"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/carrier"
)

// Carrier definitions.
type Carrier struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewCarrier returns Carrier service.
func NewCarrier(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.Carrier {
	return Carrier{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// GetCarrierList implementation
func (c Carrier) GetCarrierList(params carrier.GetCarrierListParams, principal *models.Principal) middleware.Responder {
	if !c.hasPermission(kenda.FunctionOperationID_LIST_CARRIER, principal.Roles) {
		return carrier.NewGetCarrierListDefault(http.StatusForbidden)
	}

	pageRequest := mcom.PaginationRequest{}
	if params.Page != nil && params.Limit != nil {
		pageRequest = mcom.PaginationRequest{
			PageCount:      uint(*params.Page),
			ObjectsPerPage: uint(*params.Limit),
		}
	}

	orderRequest := parseOrderRequest(params.Body.OrderRequest, getCarrierListInfoByTypeDefaultOrderFunc)

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	reply, err := c.dm.ListCarriers(ctx, mcom.ListCarriersRequest{
		DepartmentID: params.DepartmentOID,
	}.WithPagination(pageRequest).
		WithOrder(orderRequest...))

	if err != nil {
		return utils.ParseError(ctx, carrier.NewGetCarrierListDefault(0), err)
	}

	data := make([]*models.CarrierData, len(reply.Info))
	for i, carrier := range reply.Info {
		allowedMaterial := carrier.AllowedMaterial
		data[i] = &models.CarrierData{
			ID:              carrier.ID,
			AllowedMaterial: &allowedMaterial,
			UpdateAt:        strfmt.DateTime(carrier.UpdateAt),
			UpdateBy:        carrier.UpdateBy,
		}
	}

	return carrier.NewGetCarrierListOK().WithPayload(&carrier.GetCarrierListOKBody{
		Data: &carrier.GetCarrierListOKBodyData{
			Items: data,
			Total: reply.AmountOfData},
	})
}

// CreateCarrier implementation
func (c Carrier) CreateCarrier(params carrier.CreateCarrierParams, principal *models.Principal) middleware.Responder {
	if !c.hasPermission(kenda.FunctionOperationID_CREATE_CARRIER, principal.Roles) {
		return carrier.NewCreateCarrierDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	if err := c.dm.CreateCarrier(ctx, mcom.CreateCarrierRequest{
		DepartmentID:    *params.Body.DepartmentOID,
		IDPrefix:        *params.Body.IDPrefix,
		Quantity:        int32(*params.Body.Quantity),
		AllowedMaterial: *params.Body.AllowedMaterial,
	}); err != nil {
		return utils.ParseError(ctx, carrier.NewCreateCarrierDefault(0), err)
	}

	return carrier.NewCreateCarrierOK()
}

// UpdateCarrier implementation
func (c Carrier) UpdateCarrier(params carrier.UpdateCarrierParams, principal *models.Principal) middleware.Responder {
	if !c.hasPermission(kenda.FunctionOperationID_UPDATE_CARRIER, principal.Roles) {
		return carrier.NewUpdateCarrierDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	if err := c.dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
		ID: params.ID,
		Action: mcom.UpdateProperties{
			AllowedMaterial: *params.Body.AllowedMaterial,
		},
	}); err != nil {
		return utils.ParseError(ctx, carrier.NewUpdateCarrierDefault(0), err)
	}

	return carrier.NewUpdateCarrierOK()
}

// DeleteCarrier implementation
func (c Carrier) DeleteCarrier(params carrier.DeleteCarrierParams, principal *models.Principal) middleware.Responder {
	if !c.hasPermission(kenda.FunctionOperationID_DELETE_CARRIER, principal.Roles) {
		return carrier.NewDeleteCarrierDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	if err := c.dm.DeleteCarrier(ctx, mcom.DeleteCarrierRequest{
		ID: params.ID,
	}); err != nil {
		return utils.ParseError(ctx, carrier.NewDeleteCarrierDefault(0), err)
	}

	return carrier.NewDeleteCarrierOK()
}

// DownloadCode39 to generate code39 barcodes and save in pdf file
func (c Carrier) DownloadCode39(params carrier.DownloadCode39Params) middleware.Responder {
	f, err := createBarcodesPDF(params.HTTPRequest.Context(), params.Body, barcodes.Code39{})
	if err != nil {
		return carrier.NewDownloadCode39Default(http.StatusInternalServerError).WithPayload(&models.Error{
			Details: err.Error(),
		})
	}

	return carrier.NewDownloadCode39OK().WithPayload(f)
}

// DownloadQRCode to generate QRCode barcodes and save in pdf file
func (c Carrier) DownloadQRCode(params carrier.DownloadQRCodeParams) middleware.Responder {
	f, err := createBarcodesPDF(params.HTTPRequest.Context(), params.Body, barcodes.QRCode{})
	if err != nil {
		return carrier.NewDownloadQRCodeDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Details: err.Error(),
		})
	}

	return carrier.NewDownloadQRCodeOK().WithPayload(f)
}

func createBarcodesPDF(ctx context.Context, ids []string, generator barcodes.Generator) (io.ReadCloser, error) {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size: gofpdf.SizeType{
			Wd: 100,
			Ht: 150,
		},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetFont("Helvetica", "B", 10)

	pageWidth, pageHeight := pdf.GetPageSize()
	boxWidth, boxHeight := math.Abs(pageWidth/2), math.Abs(pageHeight/3)

	barcodeWidth, barcodeHeight := generator.GetSize(boxWidth, boxHeight)

	baseX, baseY := boxWidth/2-barcodeWidth/2, boxHeight/2-barcodeHeight/2
	for i, code := range ids {
		// set each pdf page maximum 6 barcodes
		if i%6 == 0 {
			pdf.AddPage()
		}

		x, y := baseX, baseY
		if (i+1)%2 == 0 {
			x += boxWidth
		}
		y += float64((i%6)/2) * boxHeight

		key, err := generator.Generate(code, int(barcodeWidth)*100, int(barcodeHeight)*10)
		if err != nil {
			return nil, err
		}

		barcode.Barcode(pdf, key, x, y, barcodeWidth, barcodeHeight, false)

		// print barcode string below barcode
		y += barcodeHeight + 1
		pdf.SetXY(x, y)
		pdf.CellFormat(barcodeWidth, 8, code, "0", 0, "CT", false, 0, "")
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		if err := pdf.OutputAndClose(pipeWriter); err != nil {
			commonsCtx.Logger(ctx).Warn("failed to output and close pdf file", zap.Error(err))
		}
	}()

	return ioutil.NopCloser(pipeReader), nil
}

func parseOrderRequest(dataIn []*carrier.GetCarrierListParamsBodyOrderRequestItems0, defaultOrderFunc func() []mcom.Order) []mcom.Order {
	length := len(dataIn)
	if length == 0 {
		return defaultOrderFunc()
	}
	dataOut := make([]mcom.Order, length)
	for i, d := range dataIn {
		dataOut[i] = mcom.Order{
			Name:       d.OrderName,
			Descending: d.Descending,
		}
	}
	return dataOut
}

func getCarrierListInfoByTypeDefaultOrderFunc() []mcom.Order {
	return []mcom.Order{{
		Name:       "id_prefix",
		Descending: false,
	}, {
		Name:       "serial_number",
		Descending: false,
	}}
}
