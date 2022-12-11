package plan

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

	"gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/dm/rs"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/plan"
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

	testProductA                        = "PRODUCT-A"
	testProductTypeA                    = "RUBBER"
	testProductADailyQuantity           = "200"
	testProductADailyQuantityDecimal    = decimal.NewFromInt(200)
	testProductAWeekQuantity            = "1000"
	testProductAWeekQuantityDecimal     = decimal.NewFromInt(1000)
	testProductAReservedQuantity        = "100"
	testProductAReservedQuantityDecimal = decimal.NewFromInt(100)
	testProductAStockQuantity           = "0"
	testProductAStockQuantityDecimal    = decimal.NewFromInt(0)

	testProductB                        = "PRODUCT-B"
	testProductBDailyQuantity           = "100"
	testProductBDailyQuantityDecimal    = decimal.NewFromInt(100)
	testProductBWeekQuantity            = "700"
	testProductBWeekQuantityDecimal     = decimal.NewFromInt(700)
	testProductBReservedQuantity        = "20"
	testProductBReservedQuantityDecimal = decimal.NewFromInt(20)
	testProductBStockQuantity           = "10"
	testProductBStockQuantityDecimal    = decimal.NewFromInt(10)

	testBrokenQuantity = "T_T"

	testDepartmentOID = "ABCDEFGHIJKLMOPQRSTUVWXYZ123456789"

	testDate = strfmt.Date(testPlanDate)

	testPlanDate    = time.Date(2021, 8, 9, 0, 0, 0, 0, time.Local)
	testProductType = rs.ProductType_RUBBER.String()

	listScripts = []mock.Script{
		{ // success
			Name: mock.FuncListProductPlans,
			Input: mock.Input{
				Request: mcom.ListProductPlansRequest{
					Date:         testPlanDate,
					DepartmentID: testDepartmentOID,
					ProductType:  testProductType,
				},
			},
			Output: mock.Output{
				Response: mcom.ListProductPlansReply{
					ProductPlans: []mcom.ProductPlan{
						{
							ProductID: testProductA,
							Quantity: mcom.ProductPlanQuantity{
								Daily:    testProductADailyQuantityDecimal,
								Week:     testProductAWeekQuantityDecimal,
								Stock:    testProductAStockQuantityDecimal,
								Reserved: testProductAReservedQuantityDecimal,
							},
						},
						{
							ProductID: testProductB,
							Quantity: mcom.ProductPlanQuantity{
								Daily:    testProductBDailyQuantityDecimal,
								Week:     testProductBWeekQuantityDecimal,
								Stock:    testProductBStockQuantityDecimal,
								Reserved: testProductBReservedQuantityDecimal,
							},
						},
					},
				},
			},
		},
		{ // user request error
			Name: mock.FuncListProductPlans,
			Input: mock.Input{
				Request: mcom.ListProductPlansRequest{
					Date:         testPlanDate,
					DepartmentID: testDepartmentOID,
					ProductType:  testProductType,
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_PRODUCTION_PLAN_NOT_FOUND,
				},
			},
		},
		{ // internal error
			Name: mock.FuncListProductPlans,
			Input: mock.Input{
				Request: mcom.ListProductPlansRequest{
					Date:         testPlanDate,
					DepartmentID: testDepartmentOID,
					ProductType:  testProductType,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}

	createScripts = []mock.Script{
		{
			Name: mock.FuncCreateProductPlan,
			Input: mock.Input{
				Request: mcom.CreateProductionPlanRequest{
					Date: testPlanDate,
					Product: mcom.Product{
						ID:   testProductA,
						Type: testProductTypeA,
					},
					DepartmentID: testDepartmentOID,
					Quantity:     testProductADailyQuantityDecimal,
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncCreateProductPlan,
			Input: mock.Input{
				Request: mcom.CreateProductionPlanRequest{
					Date: testPlanDate,
					Product: mcom.Product{
						ID:   testProductA,
						Type: testProductTypeA,
					},
					DepartmentID: testDepartmentOID,
					Quantity:     testProductADailyQuantityDecimal,
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_PRODUCTION_PLAN_EXISTED,
				},
			},
		},
		{
			Name: mock.FuncCreateProductPlan,
			Input: mock.Input{
				Request: mcom.CreateProductionPlanRequest{
					Date: testPlanDate,
					Product: mcom.Product{
						ID:   testProductA,
						Type: testProductTypeA,
					},
					DepartmentID: testDepartmentOID,
					Quantity:     testProductADailyQuantityDecimal,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}
)

func TestPlan_List(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New(listScripts)
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/plans/department-oid/{testDepartmentOID}/product-type/{testProductType}/date/{date}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    plan.GetPlanListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get plan list success",
			args: args{
				params: plan.GetPlanListParams{
					HTTPRequest:   httpRequestWithHeader,
					Date:          strfmt.Date(testPlanDate),
					DepartmentOID: testDepartmentOID,
					ProductType:   testProductType,
				},
				principal: principal,
			},
			want: plan.NewGetPlanListOK().WithPayload(&plan.GetPlanListOKBody{
				Data: []*models.PlanData{
					{
						ProductID:         testProductA,
						DayQuantity:       testProductADailyQuantity,
						WeekQuantity:      testProductAWeekQuantity,
						ScheduledQuantity: testProductAReservedQuantity,
						StockQuantity:     testProductAStockQuantity,
					},
					{
						ProductID:         testProductB,
						DayQuantity:       testProductBDailyQuantity,
						WeekQuantity:      testProductBWeekQuantity,
						ScheduledQuantity: testProductBReservedQuantity,
						StockQuantity:     testProductBStockQuantity,
					},
				},
			}),
		},
		{
			name: "not found data",
			args: args{
				params: plan.GetPlanListParams{
					HTTPRequest:   httpRequestWithHeader,
					Date:          strfmt.Date(testPlanDate),
					DepartmentOID: testDepartmentOID,
					ProductType:   testProductType,
				},
				principal: principal,
			},
			want: plan.NewGetPlanListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_PRODUCTION_PLAN_NOT_FOUND),
			}),
		},
		{
			name: "internal error",
			args: args{
				params: plan.GetPlanListParams{
					HTTPRequest:   httpRequestWithHeader,
					Date:          strfmt.Date(testPlanDate),
					DepartmentOID: testDepartmentOID,
					ProductType:   testProductType,
				},
				principal: principal,
			},
			want: plan.NewGetPlanListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlan(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetPlanList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewPlan(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetPlanList(plan.GetPlanListParams{
			HTTPRequest:   httpRequestWithHeader,
			Date:          strfmt.Date(testPlanDate),
			DepartmentOID: testDepartmentOID,
			ProductType:   testProductType,
		}, principal).(*plan.GetPlanListDefault)
		assert.True(ok)
		assert.Equal(plan.NewGetPlanListDefault(http.StatusForbidden), rep)
	}
}

func TestPlan_Create(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New(createScripts)
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("POST", "/plan", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    plan.AddPlanParams
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
				params: plan.AddPlanParams{
					HTTPRequest: httpRequestWithHeader,
					Body: plan.AddPlanBody{
						Date:          &testDate,
						DayQuantity:   &testProductADailyQuantity,
						DepartmentOID: &testDepartmentOID,
						ProductID:     &testProductA,
						ProductType:   &testProductType,
					},
				},
				principal: principal,
			},
			want: plan.NewAddPlanOK(),
		},
		{
			name: "plan existed",
			args: args{
				params: plan.AddPlanParams{
					HTTPRequest: httpRequestWithHeader,
					Body: plan.AddPlanBody{
						Date:          &testDate,
						DayQuantity:   &testProductADailyQuantity,
						DepartmentOID: &testDepartmentOID,
						ProductID:     &testProductA,
						ProductType:   &testProductType,
					},
				},
				principal: principal,
			},
			want: plan.NewAddPlanDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_PRODUCTION_PLAN_EXISTED),
			}),
		},
		{
			name: "invalid number",
			args: args{
				params: plan.AddPlanParams{
					HTTPRequest: httpRequestWithHeader,
					Body: plan.AddPlanBody{
						Date:          &testDate,
						DayQuantity:   &testBrokenQuantity,
						DepartmentOID: &testDepartmentOID,
						ProductID:     &testProductA,
						ProductType:   &testProductType,
					},
				},
				principal: principal,
			},
			want: plan.NewAddPlanDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: "invalid_number=" + testBrokenQuantity,
			}),
		},
		{
			name: "internal error",
			args: args{
				params: plan.AddPlanParams{
					HTTPRequest: httpRequestWithHeader,
					Body: plan.AddPlanBody{
						Date:          &testDate,
						DayQuantity:   &testProductADailyQuantity,
						DepartmentOID: &testDepartmentOID,
						ProductID:     &testProductA,
						ProductType:   &testProductType,
					},
				},
				principal: principal,
			},
			want: plan.NewAddPlanDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlan(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.AddPlan(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewPlan(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.AddPlan(plan.AddPlanParams{
			HTTPRequest: httpRequestWithHeader,
			Body: plan.AddPlanBody{
				Date:          &testDate,
				DayQuantity:   &testProductADailyQuantity,
				DepartmentOID: &testDepartmentOID,
				ProductID:     &testProductA,
				ProductType:   &testProductType,
			},
		}, principal).(*plan.AddPlanDefault)
		assert.True(ok)
		assert.Equal(plan.NewAddPlanDefault(http.StatusForbidden), rep)
	}
}
