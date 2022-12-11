package carrier

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/carrier"
)

const (
	userID = "tester"

	testInternalServerError = "internal error"
)

var (
	principal = &models.Principal{
		ID: userID,
		Roles: []models.Role{
			models.Role(mcomRoles.Role_ADMINISTRATOR),
			models.Role(mcomRoles.Role_LEADER),
		},
	}

	testDepartmentOID          = "M2110"
	testIDPrefix               = "LA"
	testEmpty                  = ""
	testQuantity         int64 = 11
	testCarrierID1             = "CARRIER_A"
	testAllowedMaterial1       = "CARA"
	testCarrierID2             = "CARRIER_B"
	testAllowedMaterial2       = "CARD"

	testTime = time.Date(2021, 8, 9, 16, 45, 23, 500, time.Local)
)

func TestCarrier_List(t *testing.T) {

	var (
		testTotal = 10
		testPage  = int64(1)
		testLimit = int64(20)

		testPageRequest = mcom.PaginationRequest{
			PageCount:      uint(testPage),
			ObjectsPerPage: uint(testLimit),
		}

		testOrderRequest = []mcom.Order{
			{
				Name:       "id_prefix",
				Descending: false,
			},
			{
				Name:       "serial_number",
				Descending: false,
			},
		}

		testOrder = []*carrier.GetCarrierListParamsBodyOrderRequestItems0{
			{
				OrderName:  "id_prefix",
				Descending: false,
			},
			{
				OrderName:  "serial_number",
				Descending: false,
			}}
	)
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListCarriers,
			Input: mock.Input{
				Request: mcom.ListCarriersRequest{
					DepartmentID: testDepartmentOID,
				}.WithPagination(testPageRequest).WithOrder(testOrderRequest...),
			},
			Output: mock.Output{
				Response: mcom.ListCarriersReply{
					Info: []mcom.CarrierInfo{
						{
							ID:              testCarrierID1,
							AllowedMaterial: testAllowedMaterial1,
							UpdateBy:        userID,
							UpdateAt:        testTime,
						},
						{
							ID:              testCarrierID2,
							AllowedMaterial: testAllowedMaterial2,
							UpdateBy:        userID,
							UpdateAt:        testTime,
						},
					},
					PaginationReply: mcom.PaginationReply{
						AmountOfData: int64(testTotal),
					},
				},
			},
		},
		{
			Name: mock.FuncListCarriers,
			Input: mock.Input{
				Request: mcom.ListCarriersRequest{
					DepartmentID: "",
				}.WithPagination(testPageRequest).WithOrder(testOrderRequest...),
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "empty department oid",
				},
			},
		},
		{
			Name: mock.FuncListCarriers,
			Input: mock.Input{
				Request: mcom.ListCarriersRequest{
					DepartmentID: testDepartmentOID,
				}.WithPagination(testPageRequest).WithOrder(testOrderRequest...),
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/carrier/department-oid/{departmentOID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    carrier.GetCarrierListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "success",
			args: args{
				params: carrier.GetCarrierListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: carrier.GetCarrierListBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: carrier.NewGetCarrierListOK().WithPayload(&carrier.GetCarrierListOKBody{
				Data: &carrier.GetCarrierListOKBodyData{
					Items: []*models.CarrierData{{
						ID:              testCarrierID1,
						AllowedMaterial: &testAllowedMaterial1,
						UpdateAt:        strfmt.DateTime(testTime),
						UpdateBy:        userID,
					},
						{
							ID:              testCarrierID2,
							AllowedMaterial: &testAllowedMaterial2,
							UpdateAt:        strfmt.DateTime(testTime),
							UpdateBy:        userID,
						},
					},
					Total: int64(testTotal),
				},
			}),
		},
		{
			name: "insufficient request",
			args: args{
				params: carrier.GetCarrierListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: "",
					Limit:         &testLimit,
					Page:          &testPage,
					Body: carrier.GetCarrierListBody{
						OrderRequest: []*carrier.GetCarrierListParamsBodyOrderRequestItems0{
							{
								OrderName:  "id_prefix",
								Descending: false,
							},
							{
								OrderName:  "serial_number",
								Descending: false,
							}},
					},
				},
				principal: principal,
			},
			want: carrier.NewGetCarrierListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "empty department oid",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: carrier.GetCarrierListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: carrier.GetCarrierListBody{
						OrderRequest: []*carrier.GetCarrierListParamsBodyOrderRequestItems0{
							{
								OrderName:  "id_prefix",
								Descending: false,
							},
							{
								OrderName:  "serial_number",
								Descending: false,
							}},
					},
				},
				principal: principal,
			},
			want: carrier.NewGetCarrierListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCarrier(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := c.GetCarrierList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		c := NewCarrier(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := c.GetCarrierList(carrier.GetCarrierListParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentOID: testDepartmentOID,
		}, principal).(*carrier.GetCarrierListDefault)
		assert.True(ok)
		assert.Equal(carrier.NewGetCarrierListDefault(http.StatusForbidden), rep)
	}
}

func TestCarrier_Create(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncCreateCarrier,
			Input: mock.Input{
				Request: mcom.CreateCarrierRequest{
					DepartmentID:    testDepartmentOID,
					IDPrefix:        testIDPrefix,
					Quantity:        int32(testQuantity),
					AllowedMaterial: testAllowedMaterial1,
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncCreateCarrier,
			Input: mock.Input{
				Request: mcom.CreateCarrierRequest{
					DepartmentID:    testDepartmentOID,
					IDPrefix:        testEmpty,
					Quantity:        int32(testQuantity),
					AllowedMaterial: testAllowedMaterial1,
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "insufficient request",
				},
			},
		},
		{
			Name: mock.FuncCreateCarrier,
			Input: mock.Input{
				Request: mcom.CreateCarrierRequest{
					DepartmentID:    testDepartmentOID,
					IDPrefix:        testIDPrefix,
					Quantity:        int32(testQuantity),
					AllowedMaterial: testAllowedMaterial1,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("POST", "/carrier", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    carrier.CreateCarrierParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "success",
			args: args{
				params: carrier.CreateCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					Body: carrier.CreateCarrierBody{
						AllowedMaterial: &testAllowedMaterial1,
						DepartmentOID:   &testDepartmentOID,
						IDPrefix:        &testIDPrefix,
						Quantity:        &testQuantity,
					},
				},
				principal: principal,
			},
			want: carrier.NewCreateCarrierOK(),
		},
		{
			name: "bad request",
			args: args{
				params: carrier.CreateCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					Body: carrier.CreateCarrierBody{
						AllowedMaterial: &testAllowedMaterial1,
						DepartmentOID:   &testDepartmentOID,
						IDPrefix:        &testEmpty,
						Quantity:        &testQuantity,
					},
				},
				principal: principal,
			},
			want: carrier.NewCreateCarrierDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "insufficient request",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: carrier.CreateCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					Body: carrier.CreateCarrierBody{
						AllowedMaterial: &testAllowedMaterial1,
						DepartmentOID:   &testDepartmentOID,
						IDPrefix:        &testIDPrefix,
						Quantity:        &testQuantity,
					},
				},
				principal: principal,
			},
			want: carrier.NewCreateCarrierDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCarrier(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := c.CreateCarrier(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		c := NewCarrier(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := c.CreateCarrier(carrier.CreateCarrierParams{
			HTTPRequest: httpRequestWithHeader,
			Body: carrier.CreateCarrierBody{
				AllowedMaterial: &testAllowedMaterial1,
				DepartmentOID:   &testDepartmentOID,
				IDPrefix:        &testIDPrefix,
				Quantity:        &testQuantity,
			},
		}, principal).(*carrier.CreateCarrierDefault)
		assert.True(ok)
		assert.Equal(carrier.NewCreateCarrierDefault(http.StatusForbidden), rep)
	}
}

func TestCarrier_Update(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncUpdateCarrier,
			Input: mock.Input{
				Request: mcom.UpdateCarrierRequest{
					ID: testCarrierID1,
					Action: mcom.UpdateProperties{
						AllowedMaterial: testAllowedMaterial1,
					},
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncUpdateCarrier,
			Input: mock.Input{
				Request: mcom.UpdateCarrierRequest{
					ID: "",
					Action: mcom.UpdateProperties{
						AllowedMaterial: testAllowedMaterial1,
					},
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "insufficient request",
				},
			},
		},
		{
			Name: mock.FuncUpdateCarrier,
			Input: mock.Input{
				Request: mcom.UpdateCarrierRequest{
					ID: testCarrierID1,
					Action: mcom.UpdateProperties{
						AllowedMaterial: testAllowedMaterial1,
					},
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("PUT", "/carrier/{ID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    carrier.UpdateCarrierParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "success",
			args: args{
				params: carrier.UpdateCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testCarrierID1,
					Body: &models.CarrierData{
						AllowedMaterial: &testAllowedMaterial1,
					},
				},
				principal: principal,
			},
			want: carrier.NewUpdateCarrierOK(),
		},
		{
			name: "insufficient request",
			args: args{
				params: carrier.UpdateCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          "",
					Body: &models.CarrierData{
						AllowedMaterial: &testAllowedMaterial1,
					},
				},
				principal: principal,
			},
			want: carrier.NewUpdateCarrierDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "insufficient request",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: carrier.UpdateCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testCarrierID1,
					Body: &models.CarrierData{
						AllowedMaterial: &testAllowedMaterial1,
					},
				},
				principal: principal,
			},
			want: carrier.NewUpdateCarrierDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCarrier(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := c.UpdateCarrier(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		c := NewCarrier(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := c.UpdateCarrier(carrier.UpdateCarrierParams{
			HTTPRequest: httpRequestWithHeader,
			Body: &models.CarrierData{
				AllowedMaterial: &testAllowedMaterial1,
			},
		}, principal).(*carrier.UpdateCarrierDefault)
		assert.True(ok)
		assert.Equal(carrier.NewUpdateCarrierDefault(http.StatusForbidden), rep)
	}
}

func TestCarrier_Delete(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncDeleteCarrier,
			Input: mock.Input{
				Request: mcom.DeleteCarrierRequest{
					ID: testCarrierID1,
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncDeleteCarrier,
			Input: mock.Input{
				Request: mcom.DeleteCarrierRequest{
					ID: "",
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "insufficient request",
				},
			},
		},
		{
			Name: mock.FuncDeleteCarrier,
			Input: mock.Input{
				Request: mcom.DeleteCarrierRequest{
					ID: testCarrierID1,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("DELETE", "/carrier/{ID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    carrier.DeleteCarrierParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "success",
			args: args{
				params: carrier.DeleteCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testCarrierID1,
				},
				principal: principal,
			},
			want: carrier.NewDeleteCarrierOK(),
		},
		{
			name: "insufficient request",
			args: args{
				params: carrier.DeleteCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          "",
				},
				principal: principal,
			},
			want: carrier.NewDeleteCarrierDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "insufficient request",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: carrier.DeleteCarrierParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testCarrierID1,
				},
				principal: principal,
			},
			want: carrier.NewDeleteCarrierDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCarrier(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := c.DeleteCarrier(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Delete() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		c := NewCarrier(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := c.DeleteCarrier(carrier.DeleteCarrierParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testCarrierID1,
		}, principal).(*carrier.DeleteCarrierDefault)
		assert.True(ok)
		assert.Equal(carrier.NewDeleteCarrierDefault(http.StatusForbidden), rep)
	}
}
