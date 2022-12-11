package legacy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/legacy"
)

const (
	userID = "tester"
)

var (
	principal = &models.Principal{
		ID: userID,
		Roles: []models.Role{
			models.Role(mcomRoles.Role_ADMINISTRATOR),
			models.Role(mcomRoles.Role_LEADER),
		},
	}

	goodBarcodeID     = "B0BARCODEXXXX001"
	badBarcodeID      = "XXXXXXXXXXXXXX001"
	brokenBarcodeID   = "BROKENBARCODE"
	materialProductID = "79700-9"
	materialType      = "201"
	materialStatus    = "AVAL"
	materialQuantity  = "200"

	threeDays               = 3 * 24 * time.Hour
	newMaterialStatus       = "HOLD"
	newReason               = "MTHD"
	newControlArea          = "OtherArea"
	newExtendDay      int64 = 3

	expiredDate = time.Date(2021, 04, 13, 0, 0, 0, 0, time.Local)

	getBarcodeInfoScripts = []mock.Script{
		{
			Name: mock.FuncGetMaterial,
			Input: mock.Input{
				Request: mcom.GetMaterialRequest{
					MaterialID: goodBarcodeID,
				},
			},
			Output: mock.Output{
				Response: mcom.GetMaterialReply{
					MaterialProductID: materialProductID,
					MaterialID:        goodBarcodeID,
					MaterialType:      materialType,
					Status:            materialStatus,
					Quantity:          decimal.NewFromInt(200),
					ExpireDate:        expiredDate,
				},
			},
		},
		{
			Name: mock.FuncGetMaterial,
			Input: mock.Input{
				Request: mcom.GetMaterialRequest{
					MaterialID: badBarcodeID,
				},
			},
			Output: mock.Output{
				Error: &mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
					Details: notFoundBarcode,
				},
			},
		},
		{
			Name: mock.FuncGetMaterial,
			Input: mock.Input{
				Request: mcom.GetMaterialRequest{
					MaterialID: brokenBarcodeID,
				},
			},
			Output: mock.Output{
				Error: errors.New(brokenBarcode),
			},
		},
	}

	getUpdateBarcodeStatusScripts = []mock.Script{
		{
			Name: mock.FuncListChangeableStatus,
			Input: mock.Input{
				Request: mcom.ListChangeableStatusRequest{
					MaterialID: goodBarcodeID,
				},
			},
			Output: mock.Output{
				Response: mcom.ListChangeableStatusReply{
					Codes: []*mcom.Code{
						{
							Code:            "AVAL",
							CodeDescription: "AVAL->AVAL (可用)-暫時用",
						},
						{
							Code:            "MONT",
							CodeDescription: "AVAL->MOUNT (掛載機台)",
						},
						{
							Code:            "HOLD",
							CodeDescription: "AVAL->HOLD (扣留)",
						},
						{
							Code:            "NAVL",
							CodeDescription: "AVAL->NOT AVAILABLE (不可用)",
						},
						{
							Code:            "SHIP",
							CodeDescription: "AVAL->SHIP (出貨)",
						},
						{
							Code:            "TEST",
							CodeDescription: "AVAL->TEST (測試)",
						},
						{
							Code:            "ADD",
							CodeDescription: "AVAL->ADD (摻合)",
						},
					},
				},
			},
		},
		{
			Name: mock.FuncListChangeableStatus,
			Input: mock.Input{
				Request: mcom.ListChangeableStatusRequest{
					MaterialID: badBarcodeID,
				},
			},
			Output: mock.Output{
				Error: &mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
					Details: notFoundBarcode,
				},
			},
		},
		{
			Name: mock.FuncListChangeableStatus,
			Input: mock.Input{
				Request: mcom.ListChangeableStatusRequest{
					MaterialID: brokenBarcodeID,
				},
			},
			Output: mock.Output{
				Error: errors.New(brokenBarcode),
			},
		},
	}

	getExtendDaysScripts = []mock.Script{
		{
			Name: mock.FuncGetMaterialExtendDate,
			Input: mock.Input{
				Request: mcom.GetMaterialExtendDateRequest{
					MaterialID: goodBarcodeID,
				},
			},
			Output: mock.Output{
				Response: mcom.GetMaterialExtendDateReply(threeDays),
			},
		},
		{
			Name: mock.FuncGetMaterialExtendDate,
			Input: mock.Input{
				Request: mcom.GetMaterialExtendDateRequest{
					MaterialID: badBarcodeID,
				},
			},
			Output: mock.Output{
				Error: &mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
					Details: notFoundBarcode,
				},
			},
		},
		{
			Name: mock.FuncGetMaterialExtendDate,
			Input: mock.Input{
				Request: mcom.GetMaterialExtendDateRequest{
					MaterialID: brokenBarcodeID,
				},
			},
			Output: mock.Output{
				Error: errors.New(brokenBarcode),
			},
		},
	}

	getControlAreaScripts = []mock.Script{
		{ // success
			Name: mock.FuncListControlAreas,
			Output: mock.Output{
				Response: mcom.ListControlAreasReply{
					Codes: []*mcom.Code{
						{
							Code:            "KUBB",
							CodeDescription: "建大雲林廠-密煉",
						},
						{
							Code:            "KUCL",
							CodeDescription: "建大雲林廠-蓋膠",
						},
						{
							Code:            "KUSH",
							CodeDescription: "建大雲林廠-成型",
						},
						{
							Code:            "KUTR",
							CodeDescription: "建大雲林廠-後段",
						},
						{
							Code:            "KVBB",
							CodeDescription: "建大越南廠-密煉",
						},
						{
							Code:            "KYBB",
							CodeDescription: "建大員林 廠-密煉",
						},
					},
				},
			},
		},
		{ // user(bad request) error
			Name: mock.FuncListControlAreas,
			Output: mock.Output{
				Error: &mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
					Details: notFoundBarcode,
				},
			},
		},
		{ // internal error
			Name: mock.FuncListControlAreas,
			Output: mock.Output{
				Error: errors.New(internalError),
			},
		},
	}
	getHoldReasonScripts = []mock.Script{
		{
			Name: mock.FuncListControlReasons,
			Output: mock.Output{
				Response: mcom.ListControlReasonsReply{
					Codes: []*mcom.Code{
						{
							Code:            "HDWT",
							CodeDescription: "重量不符",
						},
						{
							Code:            "HDDG",
							CodeDescription: "死膠",
						},
						{
							Code:            "HDEP",
							CodeDescription: "超日限",
						},
						{
							Code:            "HDWD",
							CodeDescription: "寬度不符",
						},
						{
							Code:            "HDTK",
							CodeDescription: "厚度不符",
						},
						{
							Code:            "HDER",
							CodeDescription: "外觀不符",
						},
						{
							Code:            "HDOT",
							CodeDescription: "其他",
						},
						{
							Code:            "HDCL",
							CodeDescription: "捲取不符",
						},
						{
							Code:            "HDFB",
							CodeDescription: "異物",
						},
						{
							Code:            "HDAR",
							CodeDescription: "面積比不符",
						},
					},
				},
			},
		},
		{ // user(bad request) error
			Name: mock.FuncListControlReasons,
			Output: mock.Output{
				Error: &mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
					Details: notFoundBarcode,
				},
			},
		},
		{ // internal error
			Name: mock.FuncListControlReasons,
			Output: mock.Output{
				Error: errors.New(internalError),
			},
		},
	}

	updateBarcodeScripts = []mock.Script{
		{
			Name: mock.FuncUpdateMaterial,
			Input: mock.Input{
				Request: mcom.UpdateMaterialRequest{
					MaterialID:       goodBarcodeID,
					ExtendedDuration: threeDays,
					User:             userID,
					NewStatus:        newMaterialStatus,
					Reason:           newReason,
					ProductCate:      materialType,
					ControlArea:      newControlArea,
				},
			},
		},
		{
			Name: mock.FuncUpdateMaterial,
			Input: mock.Input{
				Request: mcom.UpdateMaterialRequest{
					MaterialID:       badBarcodeID,
					ExtendedDuration: threeDays,
					User:             userID,
					NewStatus:        newMaterialStatus,
					Reason:           newReason,
					ProductCate:      materialType,
					ControlArea:      newControlArea,
				},
			},
			Output: mock.Output{
				Error: &mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
					Details: notFoundBarcode,
				},
			},
		},
		{
			Name: mock.FuncUpdateMaterial,
			Input: mock.Input{
				Request: mcom.UpdateMaterialRequest{
					MaterialID:       brokenBarcodeID,
					ExtendedDuration: threeDays,
					User:             userID,
					NewStatus:        newMaterialStatus,
					Reason:           newReason,
					ProductCate:      materialType,
					ControlArea:      newControlArea,
				},
			},
			Output: mock.Output{
				Error: errors.New(brokenBarcode),
			},
		},
	}
)

var (
	notFoundBarcode = "not found barcode"
	brokenBarcode   = "broken barcode"
	internalError   = "internal error"
)

func TestLegacy_GetBarcodeInfo(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New(getBarcodeInfoScripts)
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/barcode/{ID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    legacy.GetBarcodeInfoParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get barcode success",
			args: args{
				params: legacy.GetBarcodeInfoParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          goodBarcodeID,
				},
				principal: principal,
			},
			want: &legacy.GetBarcodeInfoOK{Payload: &legacy.GetBarcodeInfoOKBody{
				Data: &legacy.GetBarcodeInfoOKBodyData{Material: &models.Material{
					Barcode:   goodBarcodeID,
					ExpiredAt: strfmt.Date(expiredDate),
					Inventory: materialQuantity,
					ProductID: materialProductID,
					Status:    materialStatus,
				}},
			}},
		},
		{
			name: notFoundBarcode,
			args: args{
				params: legacy.GetBarcodeInfoParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          badBarcodeID,
				},
				principal: principal,
			},
			want: legacy.NewGetBarcodeInfoDefault(http.StatusBadRequest).
				WithPayload(&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
					Details: notFoundBarcode,
				}),
		},
		{
			name: brokenBarcode,
			args: args{
				params: legacy.GetBarcodeInfoParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          brokenBarcodeID,
				},
				principal: principal,
			},
			want: legacy.NewGetBarcodeInfoDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{
					Details: brokenBarcode,
				}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetBarcodeInfo(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBarcodeInfo() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetBarcodeInfo(legacy.GetBarcodeInfoParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          goodBarcodeID,
		}, principal).(*legacy.GetBarcodeInfoDefault)
		assert.True(ok)
		assert.Equal(legacy.NewGetBarcodeInfoDefault(http.StatusForbidden), rep)
	}
}

func TestLegacy_GetUpdateBarcodeStatusList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New(getUpdateBarcodeStatusScripts)
	assert.NoError(err)

	httpRequest := httptest.NewRequest("GET", "/barcode/update-status-list/ID/{ID}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    legacy.GetUpdateBarcodeStatusListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get update barcode status success",
			args: args{
				params: legacy.GetUpdateBarcodeStatusListParams{
					HTTPRequest: httpRequest,
					ID:          goodBarcodeID,
				},
				principal: principal,
			},
			want: legacy.NewGetUpdateBarcodeStatusListOK().WithPayload(&legacy.GetUpdateBarcodeStatusListOKBody{
				Data: []*models.CodeDescription{
					{
						Code:        "AVAL",
						Description: "AVAL->AVAL (可用)-暫時用",
					},
					{
						Code:        "MONT",
						Description: "AVAL->MOUNT (掛載機台)",
					},
					{
						Code:        "HOLD",
						Description: "AVAL->HOLD (扣留)",
					},
					{
						Code:        "NAVL",
						Description: "AVAL->NOT AVAILABLE (不可用)",
					},
					{
						Code:        "SHIP",
						Description: "AVAL->SHIP (出貨)",
					},
					{
						Code:        "TEST",
						Description: "AVAL->TEST (測試)",
					},
					{
						Code:        "ADD",
						Description: "AVAL->ADD (摻合)",
					},
				}}),
		},
		{
			name: notFoundBarcode,
			args: args{
				params: legacy.GetUpdateBarcodeStatusListParams{
					HTTPRequest: httpRequest,
					ID:          badBarcodeID,
				},
				principal: principal,
			},
			want: legacy.NewGetUpdateBarcodeStatusListDefault(http.StatusBadRequest).
				WithPayload(&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
					Details: notFoundBarcode,
				}),
		},
		{
			name: brokenBarcode,
			args: args{
				params: legacy.GetUpdateBarcodeStatusListParams{
					HTTPRequest: httpRequest,
					ID:          brokenBarcodeID,
				},
				principal: principal,
			},
			want: legacy.NewGetUpdateBarcodeStatusListDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{
					Details: brokenBarcode,
				}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetUpdateBarcodeStatusList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUpdateBarcodeStatusList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetUpdateBarcodeStatusList(legacy.GetUpdateBarcodeStatusListParams{
			HTTPRequest: httpRequest,
			ID:          goodBarcodeID,
		}, principal).(*legacy.GetUpdateBarcodeStatusListDefault)
		assert.True(ok)
		assert.Equal(legacy.NewGetUpdateBarcodeStatusListDefault(http.StatusForbidden), rep)
	}
}

func TestLegacy_GetExtendDays(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New(getExtendDaysScripts)
	assert.NoError(err)

	httpRequest := httptest.NewRequest("GET", "/barcode/extend-expired-date/ID/{ID}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    legacy.GetExtendDaysParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get extend days success",
			args: args{
				params: legacy.GetExtendDaysParams{
					HTTPRequest: httpRequest,
					ID:          goodBarcodeID,
				},
				principal: principal,
			},
			want: legacy.NewGetExtendDaysOK().WithPayload(&legacy.GetExtendDaysOKBody{
				Data: &legacy.GetExtendDaysOKBodyData{
					ExtendDay: 3,
				},
			}),
		},
		{
			name: notFoundBarcode,
			args: args{
				params: legacy.GetExtendDaysParams{
					HTTPRequest: httpRequest,
					ID:          badBarcodeID,
				},
				principal: principal,
			},
			want: legacy.NewGetExtendDaysDefault(http.StatusBadRequest).
				WithPayload(&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
					Details: notFoundBarcode,
				}),
		},
		{
			name: brokenBarcode,
			args: args{
				params: legacy.GetExtendDaysParams{
					HTTPRequest: httpRequest,
					ID:          brokenBarcodeID,
				},
				principal: principal,
			},
			want: legacy.NewGetExtendDaysDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{
					Details: brokenBarcode,
				}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetExtendDays(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExtendDays() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetExtendDays(legacy.GetExtendDaysParams{
			HTTPRequest: httpRequest,
			ID:          goodBarcodeID,
		}, principal).(*legacy.GetExtendDaysDefault)
		assert.True(ok)
		assert.Equal(legacy.NewGetExtendDaysDefault(http.StatusForbidden), rep)
	}
}

func TestLegacy_GetControlAreaList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New(getControlAreaScripts)
	assert.NoError(err)

	httpRequest := httptest.NewRequest("GET", "/barcode/control-area", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    legacy.GetControlAreaListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get control area list success",
			args: args{
				params: legacy.GetControlAreaListParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: legacy.NewGetControlAreaListOK().WithPayload(&legacy.GetControlAreaListOKBody{
				Data: []*models.CodeDescription{
					{
						Code:        "KUBB",
						Description: "建大雲林廠-密煉",
					},
					{
						Code:        "KUCL",
						Description: "建大雲林廠-蓋膠",
					},
					{
						Code:        "KUSH",
						Description: "建大雲林廠-成型",
					},
					{
						Code:        "KUTR",
						Description: "建大雲林廠-後段",
					},
					{
						Code:        "KVBB",
						Description: "建大越南廠-密煉",
					},
					{
						Code:        "KYBB",
						Description: "建大員林 廠-密煉",
					},
				}}),
		},
		{
			name: notFoundBarcode,
			args: args{
				params: legacy.GetControlAreaListParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: legacy.NewGetControlAreaListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
				Details: notFoundBarcode,
			}),
		},
		{
			name: internalError,
			args: args{
				params: legacy.GetControlAreaListParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: legacy.NewGetControlAreaListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: internalError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetControlAreaList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetControlAreaList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetControlAreaList(legacy.GetControlAreaListParams{
			HTTPRequest: httpRequest,
		}, principal).(*legacy.GetControlAreaListDefault)
		assert.True(ok)
		assert.Equal(legacy.NewGetControlAreaListDefault(http.StatusForbidden), rep)
	}
}

func TestLegacy_GetHoldReasonList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New(getHoldReasonScripts)
	assert.NoError(err)

	httpRequest := httptest.NewRequest("GET", "/barcode/reason-list", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    legacy.GetHoldReasonListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get hold reason list success",
			args: args{
				params: legacy.GetHoldReasonListParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: legacy.NewGetHoldReasonListOK().WithPayload(&legacy.GetHoldReasonListOKBody{
				Data: []*models.CodeDescription{
					{
						Code:        "HDWT",
						Description: "重量不符",
					},
					{
						Code:        "HDDG",
						Description: "死膠",
					},
					{
						Code:        "HDEP",
						Description: "超日限",
					},
					{
						Code:        "HDWD",
						Description: "寬度不符",
					},
					{
						Code:        "HDTK",
						Description: "厚度不符",
					},
					{
						Code:        "HDER",
						Description: "外觀不符",
					},
					{
						Code:        "HDOT",
						Description: "其他",
					},
					{
						Code:        "HDCL",
						Description: "捲取不符",
					},
					{
						Code:        "HDFB",
						Description: "異物",
					},
					{
						Code:        "HDAR",
						Description: "面積比不符",
					},
				}}),
		},
		{
			name: notFoundBarcode,
			args: args{
				params: legacy.GetHoldReasonListParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: legacy.NewGetHoldReasonListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
				Details: notFoundBarcode,
			}),
		},
		{
			name: internalError,
			args: args{
				params: legacy.GetHoldReasonListParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: legacy.NewGetHoldReasonListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: internalError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetHoldReasonList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHoldReasonList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetHoldReasonList(legacy.GetHoldReasonListParams{
			HTTPRequest: httpRequest,
		}, principal).(*legacy.GetHoldReasonListDefault)
		assert.True(ok)
		assert.Equal(legacy.NewGetHoldReasonListDefault(http.StatusForbidden), rep)
	}
}

func TestLegacy_UpdateBarcode(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New(updateBarcodeScripts)
	assert.NoError(err)

	httpRequest := httptest.NewRequest("PUT", "/pda/barcode/{ID}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    legacy.UpdateBarcodeParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "update barcode success",
			args: args{
				params: legacy.UpdateBarcodeParams{
					HTTPRequest: httpRequest,
					ID:          goodBarcodeID,
					Body: &models.UpdateBarcodeBody{
						ControlArea: &newControlArea,
						ExtendDays:  &newExtendDay,
						HoldReason:  &newReason,
						NewStatus:   &newMaterialStatus,
						ProductCate: &materialType,
					},
				},
				principal: principal,
			},
			want: legacy.NewUpdateBarcodeOK(),
		},
		{
			name: notFoundBarcode,
			args: args{
				params: legacy.UpdateBarcodeParams{
					HTTPRequest: httpRequest,
					ID:          badBarcodeID,
					Body: &models.UpdateBarcodeBody{
						ControlArea: &newControlArea,
						ExtendDays:  &newExtendDay,
						HoldReason:  &newReason,
						NewStatus:   &newMaterialStatus,
						ProductCate: &materialType,
					},
				},
				principal: principal,
			},
			want: legacy.NewUpdateBarcodeDefault(http.StatusBadRequest).
				WithPayload(&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
					Details: notFoundBarcode,
				}),
		},
		{
			name: brokenBarcode,
			args: args{
				params: legacy.UpdateBarcodeParams{
					HTTPRequest: httpRequest,
					ID:          brokenBarcodeID,
					Body: &models.UpdateBarcodeBody{
						ControlArea: &newControlArea,
						ExtendDays:  &newExtendDay,
						HoldReason:  &newReason,
						NewStatus:   &newMaterialStatus,
						ProductCate: &materialType,
					},
				},
				principal: principal,
			},
			want: legacy.NewUpdateBarcodeDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{
					Details: brokenBarcode,
				}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.UpdateBarcode(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateBarcode() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewLegacy(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.UpdateBarcode(legacy.UpdateBarcodeParams{
			HTTPRequest: httpRequest,
			ID:          goodBarcodeID,
			Body: &models.UpdateBarcodeBody{
				ControlArea: &newControlArea,
				ExtendDays:  &newExtendDay,
				HoldReason:  &newReason,
				NewStatus:   &newMaterialStatus,
				ProductCate: &materialType,
			},
		}, principal).(*legacy.UpdateBarcodeDefault)
		assert.True(ok)
		assert.Equal(legacy.NewUpdateBarcodeDefault(http.StatusForbidden), rep)
	}
}
