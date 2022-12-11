package workorder

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
	mcomWorkOrder "gitlab.kenda.com.tw/kenda/mcom/utils/workorder"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/work_order"
)

const (
	userID                  = "tester"
	remarkNone              = 0
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

	testDepartmentOID = "ABCDEFGHIJKLMOPQRSTUVWXYZ123456789"
	testStationID     = "STATION-A"

	testBrokenQuantity = "T_T"

	testWorkOrder1            = "WORKORDERID001"
	testWorkOrder1Batches1    = decimal.NewFromInt(10)
	testWorkOrder1Batches2    = decimal.NewFromInt(25)
	testWorkOrder1ProductA    = "PRODUCT-A"
	testWorkOrder1ProductType = "RUBBER"
	testWorkOrder1RecipeID    = "RECIPE001"
	testWorkOrderProcessNAme  = "PROCESS-A"
	testParentWorkOrder1      = "PARENTWORKORDERID001"

	testProcessOID  = "PROCESS001"
	testProcessName = "PROCESS-A"
	testProcessType = "PRODUCTION"

	falseToAbort = false
	trueToAbort  = true

	updateSequence int64 = 7

	scheduleDate       = strfmt.Date(testSchedulingDate)
	testSchedulingDate = time.Date(2021, 8, 9, 0, 0, 0, 0, time.Local)
	testUpdateDate     = time.Date(2021, 8, 9, 16, 45, 23, 500, time.Local)

	testStationA  = "STATION-A"
	testSiteName1 = "TESTSITENAME1"

	testPlanQuantity = float64(35)

	testWorkOrderProcessOID  = "PROCESS-A-OID"
	testWorkOrderProcessName = "PROCESS-A"
	testWorkOrderProcessType = "PROCESS-A-TYPE"

	testToolID          = "ToolID"
	testSequence        = 99
	testWorkOrderError1 = "WORKORDERERRORID"
)

func TestWorkOrder_GetStationScheduling(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("GET", "/schedulings/station/{station}/date/{date}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    work_order.GetStationSchedulingParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success batchSize 0",
			args: args{
				params: work_order.GetStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Date:        strfmt.Date(testSchedulingDate),
					Station:     testStationID,
				},
				principal: principal,
			},
			want: work_order.NewGetStationSchedulingOK().WithPayload(&work_order.GetStationSchedulingOKBody{
				Data: models.WorkOrders{
					{
						ID:              testWorkOrder1,
						BatchSize:       int64(mcomWorkOrder.BatchSize_PER_BATCH_QUANTITIES),
						BatchesQuantity: []string{testWorkOrder1Batches1.String(), testWorkOrder1Batches2.String()},
						DepartmentOID:   testDepartmentOID,
						PlanDate:        strfmt.Date(testSchedulingDate),
						ProductID:       testWorkOrder1ProductA,
						Recipe: &models.Recipe{
							ProcessName: testProcessName,
							ProcessType: testProcessType,
							ProcessOID:  testProcessOID,
							ID:          testWorkOrder1RecipeID,
						},
						Sequence: 1,
						Station:  testStationID,
						Status:   0,
						UpdateAt: strfmt.DateTime(testUpdateDate),
						UpdateBy: userID,
						ParentID: testParentWorkOrder1,
					},
				},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate,
							Until:   testSchedulingDate,
							Station: testStationID,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Response: mcom.ListWorkOrdersByDurationReply{
							Contents: []mcom.GetWorkOrderReply{
								{
									ID: testWorkOrder1,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										OID:  testProcessOID,
										Name: testWorkOrderProcessNAme,
										Type: testProcessType,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       0,
									DepartmentID: testDepartmentOID,
									Station:      testStationID,
									Sequence:     1,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcom.NewQuantityPerBatch(
										[]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}).Detail(),
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
									Parent:     testParentWorkOrder1,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "success batchSize 1",
			args: args{
				params: work_order.GetStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Date:        strfmt.Date(testSchedulingDate),
					Station:     testStationID,
				},
				principal: principal,
			},
			want: work_order.NewGetStationSchedulingOK().WithPayload(&work_order.GetStationSchedulingOKBody{
				Data: models.WorkOrders{
					{
						ID:            testWorkOrder1,
						DepartmentOID: testDepartmentOID,
						PlanDate:      strfmt.Date(testSchedulingDate),
						BatchSize:     int64(mcomWorkOrder.BatchSize_FIXED_QUANTITY),
						BatchCount:    2,
						PlanQuantity:  fmt.Sprint(testPlanQuantity),
						ProductID:     testWorkOrder1ProductA,
						Recipe: &models.Recipe{
							ProcessOID:  testProcessOID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
							ID:          testWorkOrder1RecipeID,
						},
						Sequence: 1,
						Station:  testStationID,
						Status:   0,
						UpdateAt: strfmt.DateTime(testUpdateDate),
						UpdateBy: userID,
						ParentID: testParentWorkOrder1,
					},
				},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate,
							Until:   testSchedulingDate,
							Station: testStationID,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Response: mcom.ListWorkOrdersByDurationReply{
							Contents: []mcom.GetWorkOrderReply{
								{
									ID: testWorkOrder1,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										OID:  testProcessOID,
										Name: testWorkOrderProcessNAme,
										Type: testProcessType,
									},
									RecipeID:             testWorkOrder1RecipeID,
									Status:               0,
									DepartmentID:         testDepartmentOID,
									Station:              testStationID,
									Sequence:             1,
									Date:                 testSchedulingDate,
									BatchQuantityDetails: mcom.NewFixedQuantity(2, decimal.NewFromFloat(testPlanQuantity)).Detail(),
									UpdatedBy:            userID,
									UpdatedAt:            testUpdateDate,
									InsertedBy:           userID,
									InsertedAt:           testUpdateDate,
									Parent:               testParentWorkOrder1,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "not found",
			args: args{
				params: work_order.GetStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Date:        strfmt.Date(testSchedulingDate),
					Station:     testStationID,
				},
				principal: principal,
			},
			want: work_order.NewGetStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate,
							Until:   testSchedulingDate,
							Station: testStationID,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Error: &mcomErrors.Error{
							Code: mcomErrors.Code_WORKORDER_NOT_FOUND, // for test case, in fact it won't return this code
						},
					},
				},
			},
		},
		{
			name: "BatchSize internal error",
			args: args{
				params: work_order.GetStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Date:        strfmt.Date(testSchedulingDate),
					Station:     testStationID,
				},
				principal: principal,
			},
			want: work_order.NewGetStationSchedulingDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Details: fmt.Sprintf("no implementation with %d of BatchSize", 4),
				}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate,
							Until:   testSchedulingDate,
							Station: testStationID,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Response: mcom.ListWorkOrdersByDurationReply{
							Contents: []mcom.GetWorkOrderReply{
								{
									ID: testWorkOrder1,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										Name: testWorkOrderProcessNAme,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       0,
									DepartmentID: testDepartmentOID,
									Station:      testStationID,
									Sequence:     1,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcomModels.BatchQuantityDetails{
										BatchQuantityType:  4,
										QuantityForBatches: []decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2},
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
									Parent:     testParentWorkOrder1,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: work_order.GetStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Date:        strfmt.Date(testSchedulingDate),
					Station:     testStationID,
				},
				principal: principal,
			},
			want: work_order.NewGetStationSchedulingDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate,
							Until:   testSchedulingDate,
							Station: testStationID,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Error: errors.New(testInternalServerError),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, err := mock.New(tt.script)
			assert.NoErrorf(err, tt.name)

			s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.GetStationScheduling(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("GetStationScheduling() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.GetStationScheduling(work_order.GetStationSchedulingParams{
			HTTPRequest: httpRequestWithHeader,
			Date:        strfmt.Date(testSchedulingDate),
			Station:     testStationID,
		}, principal).(*work_order.GetStationSchedulingDefault)
		assert.True(ok)
		assert.Equal(work_order.NewGetStationSchedulingDefault(http.StatusForbidden), rep)
	}
}

func TestWorkOrder_CreateStationScheduling(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("POST", "/schedulings", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    work_order.CreateStationSchedulingParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success batchSize 0",
			args: args{
				params: work_order.CreateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*models.CreateWorkOrder{
						{
							BatchesQuantity: []string{testWorkOrder1Batches1.String(), testWorkOrder1Batches2.String()},
							DepartmentOID:   &testDepartmentOID,
							PlanDate:        &scheduleDate,
							Recipe: &models.Recipe{
								ProcessName: testProcessName,
								ProcessOID:  testProcessOID,
								ProcessType: testProcessType,
								ID:          testWorkOrder1RecipeID,
							},
							BatchSize: int64(mcomWorkOrder.BatchSize_PER_BATCH_QUANTITIES),
							Station:   &testStationID,
							ParentID:  testParentWorkOrder1,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewCreateStationSchedulingOK().WithPayload(&work_order.CreateStationSchedulingOKBody{
				Data: []string{testWorkOrder1},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncCreateWorkOrders,
					Input: mock.Input{
						Request: mcom.CreateWorkOrdersRequest{
							WorkOrders: []mcom.CreateWorkOrder{
								{
									ProcessOID:      testProcessOID,
									ProcessName:     testProcessName,
									ProcessType:     testProcessType,
									RecipeID:        testWorkOrder1RecipeID,
									DepartmentID:    testDepartmentOID,
									Station:         testStationID,
									Date:            testSchedulingDate,
									BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}),
									Parent:          testParentWorkOrder1,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.CreateWorkOrdersReply{
							IDs: []string{testWorkOrder1},
						},
					},
				},
			},
		},
		{
			name: "success batchSize 1",
			args: args{
				params: work_order.CreateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*models.CreateWorkOrder{
						{
							DepartmentOID: &testDepartmentOID,
							PlanDate:      &scheduleDate,
							Recipe: &models.Recipe{
								ProcessOID:  testProcessOID,
								ProcessName: testProcessName,
								ProcessType: testProcessType,
								ID:          testWorkOrder1RecipeID,
							},
							Station:      &testStationID,
							ParentID:     testParentWorkOrder1,
							BatchSize:    int64(mcomWorkOrder.BatchSize_FIXED_QUANTITY),
							BatchCount:   2,
							PlanQuantity: fmt.Sprint(testPlanQuantity),
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewCreateStationSchedulingOK().WithPayload(&work_order.CreateStationSchedulingOKBody{
				Data: []string{testWorkOrder1},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncCreateWorkOrders,
					Input: mock.Input{
						Request: mcom.CreateWorkOrdersRequest{
							WorkOrders: []mcom.CreateWorkOrder{
								{
									ProcessOID:      testProcessOID,
									ProcessName:     testProcessName,
									ProcessType:     testProcessType,
									RecipeID:        testWorkOrder1RecipeID,
									DepartmentID:    testDepartmentOID,
									Station:         testStationID,
									Date:            testSchedulingDate,
									BatchesQuantity: mcom.NewFixedQuantity(2, decimal.NewFromFloat(testPlanQuantity)),
									Parent:          testParentWorkOrder1,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.CreateWorkOrdersReply{
							IDs: []string{testWorkOrder1},
						},
					},
				},
			},
		},
		{
			name: "invalid numbers",
			args: args{
				params: work_order.CreateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*models.CreateWorkOrder{
						{
							BatchesQuantity: []string{testWorkOrder1Batches1.String(), testBrokenQuantity},
							DepartmentOID:   &testDepartmentOID,
							PlanDate:        &scheduleDate,
							Recipe: &models.Recipe{
								ProcessOID:  testProcessOID,
								ProcessName: testProcessName,
								ProcessType: testProcessType,
								ID:          testWorkOrder1RecipeID,
							},
							Station:  &testStationID,
							ParentID: testParentWorkOrder1,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewCreateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: fmt.Sprintf("invalid_numbers=%v", []string{testWorkOrder1Batches1.String(), testBrokenQuantity}),
			}),
			script: []mock.Script{},
		},
		{
			name: "batchSize invalid",
			args: args{
				params: work_order.CreateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*models.CreateWorkOrder{
						{
							BatchesQuantity: []string{testWorkOrder1Batches1.String()},
							BatchSize:       4,
							DepartmentOID:   &testDepartmentOID,
							PlanDate:        &scheduleDate,
							Recipe: &models.Recipe{
								ProcessOID:  testProcessOID,
								ProcessName: testProcessName,
								ProcessType: testProcessType,
								ID:          testWorkOrder1RecipeID,
							},
							Station: &testStationID,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewCreateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: fmt.Sprintf("no implementation with %d of BatchSize", 4),
			}),
			script: []mock.Script{},
		},
		{
			name: "not found",
			args: args{
				params: work_order.CreateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*models.CreateWorkOrder{
						{
							BatchSize:       int64(mcomWorkOrder.BatchSize_PER_BATCH_QUANTITIES),
							BatchesQuantity: []string{testWorkOrder1Batches1.String(), testWorkOrder1Batches2.String()},
							DepartmentOID:   &testDepartmentOID,
							PlanDate:        &scheduleDate,
							Recipe: &models.Recipe{
								ProcessOID:  testProcessOID,
								ProcessName: testProcessName,
								ProcessType: testProcessType,
								ID:          testWorkOrder1RecipeID,
							},
							Station: &testStationID,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewCreateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncCreateWorkOrders,
					Input: mock.Input{
						Request: mcom.CreateWorkOrdersRequest{
							WorkOrders: []mcom.CreateWorkOrder{
								{
									ProcessOID:      testProcessOID,
									ProcessName:     testProcessName,
									ProcessType:     testProcessType,
									RecipeID:        testWorkOrder1RecipeID,
									DepartmentID:    testDepartmentOID,
									Station:         testStationID,
									Date:            testSchedulingDate,
									BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}),
								},
							},
						},
					},
					Output: mock.Output{
						Error: &mcomErrors.Error{
							Code: mcomErrors.Code_WORKORDER_NOT_FOUND, // for test case, in fact it won't return this code
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: work_order.CreateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*models.CreateWorkOrder{
						{
							BatchSize:       int64(mcomWorkOrder.BatchSize_PER_BATCH_QUANTITIES),
							BatchesQuantity: []string{testWorkOrder1Batches1.String(), testWorkOrder1Batches2.String()},
							DepartmentOID:   &testDepartmentOID,
							PlanDate:        &scheduleDate,
							Recipe: &models.Recipe{
								ProcessName: testProcessName,
								ProcessOID:  testProcessOID,
								ProcessType: testProcessType,
								ID:          testWorkOrder1RecipeID,
							},
							Station: &testStationID,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewCreateStationSchedulingDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncCreateWorkOrders,
					Input: mock.Input{
						Request: mcom.CreateWorkOrdersRequest{
							WorkOrders: []mcom.CreateWorkOrder{
								{
									ProcessOID:      testProcessOID,
									ProcessName:     testProcessName,
									ProcessType:     testProcessType,
									RecipeID:        testWorkOrder1RecipeID,
									DepartmentID:    testDepartmentOID,
									Station:         testStationID,
									Date:            testSchedulingDate,
									BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}),
								},
							},
						},
					},
					Output: mock.Output{
						Error: errors.New(testInternalServerError),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, err := mock.New(tt.script)
			assert.NoErrorf(err, tt.name)

			s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.CreateStationScheduling(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("CreateStationScheduling() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.CreateStationScheduling(work_order.CreateStationSchedulingParams{
			HTTPRequest: httpRequestWithHeader,
			Body: []*models.CreateWorkOrder{
				{
					BatchesQuantity: []string{testWorkOrder1Batches1.String(), testWorkOrder1Batches2.String()},
					DepartmentOID:   &testDepartmentOID,
					PlanDate:        &scheduleDate,
					Recipe: &models.Recipe{
						ProcessName: testProcessName,
						ProcessOID:  testProcessOID,
						ProcessType: testProcessType,
						ID:          testWorkOrder1RecipeID,
					},
					Station: &testStationID,
				},
			},
		}, principal).(*work_order.CreateStationSchedulingDefault)
		assert.True(ok)
		assert.Equal(work_order.NewCreateStationSchedulingDefault(http.StatusForbidden), rep)
	}
}

func TestWorkOrder_UpdateStationScheduling(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncUpdateWorkOrders,
			Input: mock.Input{
				Request: mcom.UpdateWorkOrdersRequest{
					Orders: []mcom.UpdateWorkOrder{
						{
							ID:       testWorkOrder1,
							Sequence: 7,
						},
					},
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncUpdateWorkOrders,
			Input: mock.Input{
				Request: mcom.UpdateWorkOrdersRequest{
					Orders: []mcom.UpdateWorkOrder{
						{
							ID:     testWorkOrder1,
							Status: workorder.Status_SKIPPED,
						},
					},
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncUpdateWorkOrders,
			Input: mock.Input{
				Request: mcom.UpdateWorkOrdersRequest{
					Orders: []mcom.UpdateWorkOrder{
						{
							ID:     testWorkOrder1,
							Status: workorder.Status_SKIPPED,
						},
					},
				},
			},
			Output: mock.Output{
				Error: &mcomErrors.Error{
					Code: mcomErrors.Code_WORKORDER_NOT_FOUND,
				},
			},
		},
		{
			Name: mock.FuncUpdateWorkOrders,
			Input: mock.Input{
				Request: mcom.UpdateWorkOrdersRequest{
					Orders: []mcom.UpdateWorkOrder{
						{
							ID:     testWorkOrder1,
							Status: workorder.Status_SKIPPED,
						},
					},
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("PUT", "/schedulings", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    work_order.UpdateStationSchedulingParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "success update sequence",
			args: args{
				params: work_order.UpdateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*work_order.UpdateStationSchedulingParamsBodyItems0{
						{
							ID:           &testWorkOrder1,
							ForceToAbort: &falseToAbort,
							Sequence:     &updateSequence,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewUpdateStationSchedulingOK(),
		},
		{
			name: "success aborting work order",
			args: args{
				params: work_order.UpdateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*work_order.UpdateStationSchedulingParamsBodyItems0{
						{
							ID:           &testWorkOrder1,
							ForceToAbort: &trueToAbort,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewUpdateStationSchedulingOK(),
		},
		{
			name: "not found",
			args: args{
				params: work_order.UpdateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*work_order.UpdateStationSchedulingParamsBodyItems0{
						{
							ID:           &testWorkOrder1,
							ForceToAbort: &trueToAbort,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewUpdateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
			}),
		},
		{
			name: "internal error",
			args: args{
				params: work_order.UpdateStationSchedulingParams{
					HTTPRequest: httpRequestWithHeader,
					Body: []*work_order.UpdateStationSchedulingParamsBodyItems0{
						{
							ID:           &testWorkOrder1,
							ForceToAbort: &trueToAbort,
						},
					},
				},
				principal: principal,
			},
			want: work_order.NewUpdateStationSchedulingDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.UpdateStationScheduling(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.UpdateStationScheduling(work_order.UpdateStationSchedulingParams{
			HTTPRequest: httpRequestWithHeader,
			Body: []*work_order.UpdateStationSchedulingParamsBodyItems0{
				{
					ID:           &testWorkOrder1,
					ForceToAbort: &trueToAbort,
				},
			},
		}, principal).(*work_order.UpdateStationSchedulingDefault)
		assert.True(ok)
		assert.Equal(work_order.NewUpdateStationSchedulingDefault(http.StatusForbidden), rep)
	}
}

func TestWorkOrder_ListWorkOrders(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("GET", "/production-flow/work-orders/station/{stationID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	var (
		testWorkOrder2 = "WORKORDERID002"
		testWorkOrder3 = "WORKORDERID003"
		testWorkOrder4 = "WORKORDERID004"
		testWorkOrder5 = "WORKORDERID005"
	)

	workOrderList := []*work_order.ListWorkOrdersOKBodyDataItems0{
		{
			WorkOrderID:     testWorkOrder1,
			ProductID:       testWorkOrder1ProductA,
			ProductType:     testWorkOrder1ProductType,
			RecipeID:        testWorkOrder1RecipeID,
			PlanQuantity:    fmt.Sprint(testPlanQuantity),
			Date:            strfmt.Date(testSchedulingDate),
			WorkOrderStatus: models.WorkOrderStatus(workorder.Status_PENDING),
		},
		{
			WorkOrderID:     testWorkOrder2,
			ProductID:       testWorkOrder1ProductA,
			ProductType:     testWorkOrder1ProductType,
			RecipeID:        testWorkOrder1RecipeID,
			PlanQuantity:    fmt.Sprint(testPlanQuantity),
			Date:            strfmt.Date(testSchedulingDate),
			WorkOrderStatus: models.WorkOrderStatus(workorder.Status_ACTIVE),
		},
		{
			WorkOrderID:     testWorkOrder5,
			ProductID:       testWorkOrder1ProductA,
			ProductType:     testWorkOrder1ProductType,
			RecipeID:        testWorkOrder1RecipeID,
			PlanQuantity:    fmt.Sprint(testPlanQuantity),
			Date:            strfmt.Date(testSchedulingDate),
			WorkOrderStatus: models.WorkOrderStatus(workorder.Status_CLOSING),
		},
	}

	type args struct {
		params    work_order.ListWorkOrdersParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success",
			args: args{
				params: work_order.ListWorkOrdersParams{
					HTTPRequest: httpRequestWithHeader,
					WorkDate:    strfmt.Date(testSchedulingDate),
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: work_order.NewListWorkOrdersOK().WithPayload(&work_order.ListWorkOrdersOKBody{
				Data: workOrderList,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate.AddDate(0, 0, -9),
							Station: testStationA,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Response: mcom.ListWorkOrdersByDurationReply{
							Contents: []mcom.GetWorkOrderReply{
								{
									ID: testWorkOrder1,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										OID:  testWorkOrderProcessOID,
										Name: testWorkOrderProcessName,
										Type: testWorkOrderProcessType,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       workorder.Status_PENDING,
									DepartmentID: testDepartmentOID,
									Station:      testStationA,
									Sequence:     1,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcom.NewQuantityPerBatch(
										[]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}).Detail(),
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
									Parent:     testParentWorkOrder1,
								},
								{
									ID: testWorkOrder2,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										OID:  testWorkOrderProcessOID,
										Name: testWorkOrderProcessName,
										Type: testWorkOrderProcessType,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       workorder.Status_ACTIVE,
									DepartmentID: testDepartmentOID,
									Station:      testStationA,
									Sequence:     2,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcom.NewFixedQuantity(
										2, decimal.NewFromFloat(testPlanQuantity)).Detail(),
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
									Parent:     testParentWorkOrder1,
								},
								{
									ID: testWorkOrder3,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										OID:  testWorkOrderProcessOID,
										Name: testWorkOrderProcessName,
										Type: testWorkOrderProcessType,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       workorder.Status_CLOSED,
									DepartmentID: testDepartmentOID,
									Station:      testStationA,
									Sequence:     2,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcom.NewQuantityPerBatch(
										[]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}).Detail(),
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
									Parent:     testParentWorkOrder1,
								},
								{
									ID: testWorkOrder4,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										OID:  testWorkOrderProcessOID,
										Name: testWorkOrderProcessName,
										Type: testWorkOrderProcessType,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       workorder.Status_SKIPPED,
									DepartmentID: testDepartmentOID,
									Station:      testStationA,
									Sequence:     2,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcom.NewQuantityPerBatch(
										[]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}).Detail(),
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
									Parent:     testParentWorkOrder1,
								},
								{
									ID: testWorkOrder5,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										OID:  testWorkOrderProcessOID,
										Name: testWorkOrderProcessName,
										Type: testWorkOrderProcessType,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       workorder.Status_CLOSING,
									DepartmentID: testDepartmentOID,
									Station:      testStationA,
									Sequence:     2,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcom.NewPlanQuantity(
										2, decimal.NewFromFloat(testPlanQuantity)).Detail(),
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
									Parent:     testParentWorkOrder1,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "not found",
			args: args{
				params: work_order.ListWorkOrdersParams{
					HTTPRequest: httpRequestWithHeader,
					WorkDate:    strfmt.Date(testSchedulingDate),
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: work_order.NewListWorkOrdersDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate.AddDate(0, 0, -9),
							Station: testStationA,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Error: &mcomErrors.Error{
							Code: mcomErrors.Code_WORKORDER_NOT_FOUND,
						},
					},
				},
			},
		},
		{
			name: "BatchSize internal error",
			args: args{
				params: work_order.ListWorkOrdersParams{
					HTTPRequest: httpRequestWithHeader,
					WorkDate:    strfmt.Date(testSchedulingDate),
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: work_order.NewListWorkOrdersDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Details: fmt.Sprintf("no implementation with %d of BatchSize", 4),
				}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate.AddDate(0, 0, -9),
							Station: testStationA,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Response: mcom.ListWorkOrdersByDurationReply{
							Contents: []mcom.GetWorkOrderReply{
								{
									ID: testWorkOrder1,
									Product: mcom.Product{
										ID:   testWorkOrder1ProductA,
										Type: testWorkOrder1ProductType,
									},
									Process: mcom.WorkOrderProcess{
										OID:  testWorkOrderProcessOID,
										Name: testWorkOrderProcessName,
										Type: testWorkOrderProcessType,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       workorder.Status_PENDING,
									DepartmentID: testDepartmentOID,
									Station:      testStationA,
									Sequence:     1,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcomModels.BatchQuantityDetails{
										BatchQuantityType:  4,
										QuantityForBatches: []decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2},
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
									Parent:     testParentWorkOrder1,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: work_order.ListWorkOrdersParams{
					HTTPRequest: httpRequestWithHeader,
					WorkDate:    strfmt.Date(testSchedulingDate),
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: work_order.NewListWorkOrdersDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:   testSchedulingDate.AddDate(0, 0, -9),
							Station: testStationA,
						}.WithOrder(
							mcom.Order{
								Name:       "reserved_date",
								Descending: false,
							},
							mcom.Order{
								Name:       "reserved_sequence",
								Descending: false,
							},
						),
					},
					Output: mock.Output{
						Error: errors.New(testInternalServerError),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, err := mock.New(tt.script)
			assert.NoErrorf(err, tt.name)

			s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.ListWorkOrders(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("ListWorkOrders() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}

	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.ListWorkOrders(work_order.ListWorkOrdersParams{
			HTTPRequest: httpRequestWithHeader,
			WorkDate:    strfmt.Date(testSchedulingDate),
			StationID:   testStationA,
		}, principal).(*work_order.ListWorkOrdersDefault)
		assert.True(ok)
		assert.Equal(work_order.NewListWorkOrdersDefault(http.StatusForbidden), rep)
	}
}

func TestWorkOrder_ListWorkOrdersRate(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("GET", "/work-orders-rate/department/{departmentID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")
	var (
		testWorkOrder2        = "WORKORDERID002"
		dateTest       string = testSchedulingDate.Format("2006-01-02")
		dateTest1      string = ""
		testTotal             = 10
		testPage              = int64(1)
		testLimit             = int64(10)

		testPageRequest = mcom.PaginationRequest{
			PageCount:      uint(testPage),
			ObjectsPerPage: uint(testLimit),
		}

		testOrderRequest = []mcom.Order{
			{
				Name:       "reserved_date",
				Descending: false,
			},
			{
				Name:       "reserved_sequence",
				Descending: false,
			},
		}

		testOrder = []*work_order.ListWorkOrdersRateParamsBodyOrderRequestItems0{
			{
				OrderName:  "reserved_date",
				Descending: false,
			},
			{
				OrderName:  "reserved_sequence",
				Descending: false,
			},
		}
	)

	workOrderList := []*models.WorkOrderRateData{
		{
			DepartmentID:      "B2200",
			WorkOrderID:       testWorkOrder1,
			ProductID:         testWorkOrder1ProductA,
			Station:           testStationA,
			PlanQuantity:      "10",
			CollectedQuantity: "15",
			Ratio:             "50.00%",
			ProductionTime:    &dateTest,
			ProductionEndTime: &dateTest,
			UpdateBy:          userID,
			CreatedBy:         userID,
			RecipeID:          testWorkOrder1RecipeID,
		},
		{
			DepartmentID:      "B2200",
			WorkOrderID:       testWorkOrder2,
			ProductID:         testWorkOrder1ProductA,
			Station:           testStationA,
			PlanQuantity:      "10",
			CollectedQuantity: "0",
			Ratio:             "0.00%",
			ProductionTime:    &dateTest1,
			ProductionEndTime: &dateTest1,
			UpdateBy:          userID,
			CreatedBy:         userID,
			RecipeID:          testWorkOrder1RecipeID,
		},
	}

	type args struct {
		params    work_order.ListWorkOrdersRateParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success",
			args: args{
				params: work_order.ListWorkOrdersRateParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentID:  "B2200",
					WorkEndDate:   strfmt.Date(testSchedulingDate),
					WorkStartDate: strfmt.Date(testSchedulingDate),
					Limit:         &testLimit,
					Page:          &testPage,
					Body: work_order.ListWorkOrdersRateBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: work_order.NewListWorkOrdersRateOK().WithPayload(&work_order.ListWorkOrdersRateOKBody{
				Data: &work_order.ListWorkOrdersRateOKBodyData{
					Items: workOrderList,
					Total: int64(testTotal),
				},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:        testSchedulingDate,
							Until:        testSchedulingDate,
							DepartmentID: "B2200",
						}.WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},

					Output: mock.Output{
						Response: mcom.ListWorkOrdersByDurationReply{
							Contents: []mcom.GetWorkOrderReply{
								{
									ID: testWorkOrder1,
									Product: mcom.Product{
										ID: testWorkOrder1ProductA,
									},
									RecipeID:          testWorkOrder1RecipeID,
									Status:            workorder.Status_CLOSED,
									DepartmentID:      "B2200",
									Station:           testStationA,
									Date:              testSchedulingDate,
									CurrentBatch:      1,
									CollectedSequence: 1,
									CollectedQuantity: decimal.NewFromFloat(15),
									BatchQuantityDetails: mcomModels.BatchQuantityDetails{
										BatchQuantityType: 1,
										FixedQuantity: &mcomModels.FixedQuantity{
											BatchCount:   2,
											PlanQuantity: testWorkOrder1Batches1,
										},
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
								},
								{
									ID: testWorkOrder2,
									Product: mcom.Product{
										ID: testWorkOrder1ProductA,
									},
									RecipeID:     testWorkOrder1RecipeID,
									Status:       workorder.Status_ACTIVE,
									DepartmentID: "B2200",
									Station:      testStationA,
									Date:         testSchedulingDate,
									BatchQuantityDetails: mcomModels.BatchQuantityDetails{
										BatchQuantityType: 1,
										FixedQuantity: &mcomModels.FixedQuantity{
											BatchCount:   2,
											PlanQuantity: testWorkOrder1Batches1,
										},
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testUpdateDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(1),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(1),
								Status:    int32(2),
								Records: []mcomModels.FeedRecord{
									mcomModels.FeedRecord{
										Time: testSchedulingDate,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "not found",
			args: args{
				params: work_order.ListWorkOrdersRateParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentID:  "B2200",
					WorkEndDate:   strfmt.Date(testSchedulingDate),
					WorkStartDate: strfmt.Date(testSchedulingDate),
					Limit:         &testLimit,
					Page:          &testPage,
					Body: work_order.ListWorkOrdersRateBody{
						OrderRequest: []*work_order.ListWorkOrdersRateParamsBodyOrderRequestItems0{
							{
								OrderName:  "reserved_date",
								Descending: false,
							},
							{
								OrderName:  "reserved_sequence",
								Descending: false,
							}},
					},
				},
				principal: principal,
			},
			want: work_order.NewListWorkOrdersRateDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:        testSchedulingDate,
							Until:        testSchedulingDate,
							DepartmentID: "B2200",
						}.WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Error: &mcomErrors.Error{
							Code: mcomErrors.Code_WORKORDER_NOT_FOUND,
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: work_order.ListWorkOrdersRateParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentID:  "B2200",
					WorkEndDate:   strfmt.Date(testSchedulingDate),
					WorkStartDate: strfmt.Date(testSchedulingDate),
					Limit:         &testLimit,
					Page:          &testPage,
					Body: work_order.ListWorkOrdersRateBody{
						OrderRequest: []*work_order.ListWorkOrdersRateParamsBodyOrderRequestItems0{
							{
								OrderName:  "reserved_date",
								Descending: false,
							},
							{
								OrderName:  "reserved_sequence",
								Descending: false,
							}},
					},
				},
				principal: principal,
			},
			want: work_order.NewListWorkOrdersRateDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListWorkOrdersByDuration,
					Input: mock.Input{
						Request: mcom.ListWorkOrdersByDurationRequest{
							Since:        testSchedulingDate,
							Until:        testSchedulingDate,
							DepartmentID: "B2200",
						}.WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Error: errors.New(testInternalServerError),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, err := mock.New(tt.script)
			assert.NoErrorf(err, tt.name)

			s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.ListWorkOrdersRate(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("ListWorkOrdersRate() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.ListWorkOrdersRate(work_order.ListWorkOrdersRateParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentID:  "B2200",
			WorkEndDate:   strfmt.Date(testSchedulingDate),
			WorkStartDate: strfmt.Date(testSchedulingDate),
		}, principal).(*work_order.ListWorkOrdersRateDefault)
		assert.True(ok)
		assert.Equal(work_order.NewListWorkOrdersRateDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestWorkOrder_ChangeWorkOrderStatus(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("PUT", "/production-flow/status/work-order/{workOrderID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    work_order.ChangeWorkOrderStatusParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			// active work order
			name: "success",
			args: args{
				params: work_order.ChangeWorkOrderStatusParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
					Body: work_order.ChangeWorkOrderStatusBody{
						Type:   startWorkOrder,
						Remark: remarkNone,
					},
				},
				principal: principal,
			},
			want: work_order.NewChangeWorkOrderStatusOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							ID:     testWorkOrder1,
							Status: workorder.Status_PENDING,
						},
					},
				},
				{
					Name: mock.FuncUpdateWorkOrders,
					Input: mock.Input{
						Request: mcom.UpdateWorkOrdersRequest{
							Orders: []mcom.UpdateWorkOrder{
								{
									ID:          testWorkOrder1,
									Status:      workorder.Status_ACTIVE,
									Abnormality: mcomWorkOrder.Abnormality_NONE,
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			// closed work order
			name: "success",
			args: args{
				params: work_order.ChangeWorkOrderStatusParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
					Body: work_order.ChangeWorkOrderStatusBody{
						Type:   closeWorkOrder,
						Remark: remarkNone,
					},
				},
				principal: principal,
			},
			want: work_order.NewChangeWorkOrderStatusOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							ID:     testWorkOrder1,
							Status: workorder.Status_CLOSING,
						},
					},
				},
				{
					Name: mock.FuncUpdateWorkOrders,
					Input: mock.Input{
						Request: mcom.UpdateWorkOrdersRequest{
							Orders: []mcom.UpdateWorkOrder{
								{
									ID:          testWorkOrder1,
									Status:      workorder.Status_CLOSED,
									Abnormality: mcomWorkOrder.Abnormality_NONE,
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "not found",
			args: args{
				params: work_order.ChangeWorkOrderStatusParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrderError1,
					Body: work_order.ChangeWorkOrderStatusBody{
						Type:   startWorkOrder,
						Remark: remarkNone,
					},
				},
				principal: principal,
			},
			want: work_order.NewChangeWorkOrderStatusDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrderError1,
						},
					},
					Output: mock.Output{
						Error: &mcomErrors.Error{
							Code: mcomErrors.Code_WORKORDER_NOT_FOUND,
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: work_order.ChangeWorkOrderStatusParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
					Body: work_order.ChangeWorkOrderStatusBody{
						Type:   closeWorkOrder,
						Remark: remarkNone,
					},
				},
				principal: principal,
			},
			want: work_order.NewChangeWorkOrderStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Error: errors.New(testInternalServerError),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, err := mock.New(tt.script)
			assert.NoErrorf(err, tt.name)

			s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.ChangeWorkOrderStatus(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("ChangeWorkOrderStatus() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.ChangeWorkOrderStatus(work_order.ChangeWorkOrderStatusParams{
			HTTPRequest: httpRequestWithHeader,
			WorkOrderID: testWorkOrder1,
			Body: work_order.ChangeWorkOrderStatusBody{
				Type: 0,
			},
		}, principal).(*work_order.ChangeWorkOrderStatusDefault)
		assert.True(ok)
		assert.Equal(work_order.NewChangeWorkOrderStatusDefault(http.StatusForbidden), rep)
	}
}

func TestWorkOrder_GetWorkOrderInformation(t *testing.T) {
	var (
		testProcessAOID  = "PROCESS001OID"
		testProcessA     = "PROCESS-A"
		testProcessAType = "PROCESS"

		testStationB = "STATION-B"

		testMaterialAID    = "MATERIAL-A"
		testMaterialAGrade = "X"
		testMaterialBID    = "MATERIAL-B"

		testBatchSizeDecimal = types.Decimal.NewFromInt16(10)
		testCurrentBatch     = 10
		testCurrentQuantity  = decimal.Decimal(decimal.NewFromFloat(7921.8))
		testMaterialAValue   = types.Decimal.NewFromInt16(10)

		testMaterialAMaxValue = types.Decimal.NewFromInt16(15)
		testMaterialAMinValue = types.Decimal.NewFromFloat32(1.5)
		testUnit              = "TESTUNIT"
	)
	assert := assert.New(t)
	httpRequestWithHeader := httptest.NewRequest("GET", "/production-flow/work-order/{workOrderID}/information", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    work_order.GetWorkOrderInformationParams
		principal *models.Principal
	}

	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationOK().WithPayload(&work_order.GetWorkOrderInformationOKBody{
				Data: &work_order.GetWorkOrderInformationOKBodyData{
					WorkOrderID:     testWorkOrder1,
					ProductID:       testWorkOrder1ProductA,
					ProductType:     testWorkOrder1ProductType,
					RecipeID:        testWorkOrder1RecipeID,
					Date:            strfmt.Date(testSchedulingDate),
					WorkOrderStatus: int64(workorder.Status_PENDING),
					CollectSequence: int64(testSequence),
					CurrentQuantity: testCurrentQuantity.InexactFloat64(),
					CurrentBatch:    int64(testCurrentBatch),
					PlanQuantity:    fmt.Sprint(testPlanQuantity),
					Recipe: &work_order.GetWorkOrderInformationOKBodyDataRecipe{
						Tools: []*work_order.GetWorkOrderInformationOKBodyDataRecipeToolsItems0{
							{
								ID:        testToolID,
								Necessity: true,
							},
						},
						Materials: []*work_order.GetWorkOrderInformationOKBodyDataRecipeMaterialsItems0{
							{
								ID:            testMaterialAID,
								SiteName:      testSiteName1,
								StandardValue: testMaterialAValue.InexactFloat64() * 2,
							},
							{
								ID:            testMaterialBID,
								SiteName:      testSiteName1,
								StandardValue: testMaterialAValue.InexactFloat64(),
							},
						},
					},
				},
			}),
			script: []mock.Script{
				{ // success
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply(
							mcom.GetWorkOrderReply{
								ID: testWorkOrder1,
								Product: mcom.Product{
									ID:   testWorkOrder1ProductA,
									Type: testWorkOrder1ProductType,
								},
								Process: mcom.WorkOrderProcess{
									Name: testProcessA,
									Type: testProcessAType,
								},
								RecipeID:     testWorkOrder1RecipeID,
								Status:       workorder.Status_PENDING,
								DepartmentID: testDepartmentOID,
								Station:      testStationA,
								Sequence:     int32(testSequence),
								Date:         testSchedulingDate,
								BatchQuantityDetails: mcom.NewQuantityPerBatch(
									[]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}).Detail(),
								UpdatedBy:         userID,
								UpdatedAt:         testUpdateDate,
								InsertedBy:        userID,
								InsertedAt:        testUpdateDate,
								Parent:            testParentWorkOrder1,
								CurrentBatch:      testCurrentBatch,
								CollectedQuantity: testCurrentQuantity,
								CollectedSequence: testSequence,
							}),
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testWorkOrder1RecipeID,
							ProcessName: testProcessA,
							ProcessType: testProcessAType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								OID:  testProcessAOID,
								Name: testProcessA,
								Type: testProcessAType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations:  []string{testStationA, testStationB},
										BatchSize: testBatchSizeDecimal,
										CommonControls: []*mcom.RecipeProperty{
											{
												Name: "CONTROL",
												Param: &mcom.RecipePropertyParameter{
													High: testMaterialAMaxValue,
													Mid:  testMaterialAValue,
													Low:  testMaterialAMinValue,
												},
											},
											{
												Name:  "CONTROL2",
												Param: &mcom.RecipePropertyParameter{},
											},
										},
										Tools: []*mcom.RecipeTool{
											{
												ID:       testToolID,
												Required: true,
											},
										},
										Steps: []*mcom.RecipeProcessStep{
											{
												Controls: []*mcom.RecipeProperty{
													{
														Name: "CONTROL",
														Param: &mcom.RecipePropertyParameter{
															High: testMaterialAMaxValue,
															Mid:  testMaterialAValue,
															Low:  testMaterialAMinValue,
														},
													},
													{
														Name: "CONTROL2",
														Param: &mcom.RecipePropertyParameter{
															High: testMaterialAMaxValue,
															Low:  testMaterialAMinValue,
														},
													},
												},
												Materials: []*mcom.RecipeMaterial{
													{
														Name:  testMaterialAID,
														Grade: testMaterialAGrade,
														Value: mcom.RecipeMaterialParameter{
															High: testMaterialAMaxValue,
															Mid:  testMaterialAValue,
															Low:  testMaterialAMinValue,
															Unit: testUnit,
														},
														Site:             testSiteName1,
														RequiredRecipeID: testWorkOrder1RecipeID,
													},
													{
														Name: testMaterialBID,
														Value: mcom.RecipeMaterialParameter{
															High: testMaterialAMaxValue,
															Mid:  testMaterialAValue,
															Low:  testMaterialAMinValue,
															Unit: testUnit,
														},
														Site:             testSiteName1,
														RequiredRecipeID: testWorkOrder1RecipeID,
													},
												},
											},
											{
												Materials: []*mcom.RecipeMaterial{
													{
														Name:  testMaterialAID,
														Grade: testMaterialAGrade,
														Value: mcom.RecipeMaterialParameter{
															High: testMaterialAMaxValue,
															Mid:  testMaterialAValue,
															Low:  testMaterialAMinValue,
															Unit: testUnit,
														},
														Site:             testSiteName1,
														RequiredRecipeID: testWorkOrder1RecipeID,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "success with sites",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationOK().WithPayload(&work_order.GetWorkOrderInformationOKBody{
				Data: &work_order.GetWorkOrderInformationOKBodyData{
					WorkOrderID:     testWorkOrder1,
					ProductID:       testWorkOrder1ProductA,
					ProductType:     testWorkOrder1ProductType,
					RecipeID:        testWorkOrder1RecipeID,
					Date:            strfmt.Date(testSchedulingDate),
					WorkOrderStatus: int64(workorder.Status_PENDING),
					CollectSequence: int64(testSequence),
					CurrentQuantity: testCurrentQuantity.InexactFloat64(),
					CurrentBatch:    int64(testCurrentBatch),
					PlanQuantity:    fmt.Sprint(testPlanQuantity),
					Recipe: &work_order.GetWorkOrderInformationOKBodyDataRecipe{
						Tools: []*work_order.GetWorkOrderInformationOKBodyDataRecipeToolsItems0{
							{
								ID:        testToolID,
								Necessity: true,
							},
						},
						Materials: []*work_order.GetWorkOrderInformationOKBodyDataRecipeMaterialsItems0{
							{
								ID:            testMaterialAID,
								SiteName:      testSiteName1,
								StandardValue: testMaterialAValue.InexactFloat64() * 2,
							},
							{
								ID:            testMaterialBID,
								SiteName:      testSiteName1,
								StandardValue: testMaterialAValue.InexactFloat64(),
							},
						},
					},
				},
			}),
			script: []mock.Script{
				{ // success
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply(
							mcom.GetWorkOrderReply{
								ID: testWorkOrder1,
								Product: mcom.Product{
									ID:   testWorkOrder1ProductA,
									Type: testWorkOrder1ProductType,
								},
								Process: mcom.WorkOrderProcess{
									Name: testProcessA,
									Type: testProcessAType,
								},
								RecipeID:             testWorkOrder1RecipeID,
								Status:               workorder.Status_PENDING,
								DepartmentID:         testDepartmentOID,
								Station:              testStationA,
								Sequence:             int32(testSequence),
								Date:                 testSchedulingDate,
								BatchQuantityDetails: mcom.NewPlanQuantity(2, decimal.NewFromFloat(testPlanQuantity)).Detail(),
								UpdatedBy:            userID,
								UpdatedAt:            testUpdateDate,
								InsertedBy:           userID,
								InsertedAt:           testUpdateDate,
								Parent:               testParentWorkOrder1,
								CurrentBatch:         testCurrentBatch,
								CollectedQuantity:    testCurrentQuantity,
								CollectedSequence:    testSequence,
							}),
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testWorkOrder1RecipeID,
							ProcessName: testProcessA,
							ProcessType: testProcessAType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								OID:  testProcessAOID,
								Name: testProcessA,
								Type: testProcessAType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations:  []string{testStationA, testStationB},
										BatchSize: testBatchSizeDecimal,
										CommonControls: []*mcom.RecipeProperty{
											{
												Name: "CONTROL",
												Param: &mcom.RecipePropertyParameter{
													High: testMaterialAMaxValue,
													Mid:  testMaterialAValue,
													Low:  testMaterialAMinValue,
												},
											}},
										Tools: []*mcom.RecipeTool{
											{
												ID:       testToolID,
												Required: true,
											},
										},
										Steps: []*mcom.RecipeProcessStep{
											{
												Controls: []*mcom.RecipeProperty{
													{
														Name: "CONTROL",
														Param: &mcom.RecipePropertyParameter{
															High: testMaterialAMaxValue,
															Mid:  testMaterialAValue,
															Low:  testMaterialAMinValue,
														},
													},
												},
												Materials: []*mcom.RecipeMaterial{
													{
														Name:  testMaterialAID,
														Grade: testMaterialAGrade,
														Value: mcom.RecipeMaterialParameter{
															High: testMaterialAMaxValue,
															Mid:  testMaterialAValue,
															Low:  testMaterialAMinValue,
															Unit: testUnit,
														},
														Site:             testSiteName1,
														RequiredRecipeID: testWorkOrder1RecipeID,
													},
													{
														Name: testMaterialBID,
														Value: mcom.RecipeMaterialParameter{
															High: testMaterialAMaxValue,
															Mid:  testMaterialAValue,
															Low:  testMaterialAMinValue,
															Unit: testUnit,
														},
														Site:             testSiteName1,
														RequiredRecipeID: testWorkOrder1RecipeID,
													},
												},
											},
											{
												Materials: []*mcom.RecipeMaterial{
													{
														Name:  testMaterialAID,
														Grade: testMaterialAGrade,
														Value: mcom.RecipeMaterialParameter{
															High: testMaterialAMaxValue,
															Mid:  testMaterialAValue,
															Low:  testMaterialAMinValue,
															Unit: testUnit,
														},
														Site:             testSiteName1,
														RequiredRecipeID: testWorkOrder1RecipeID,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "some step standard value is nil",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Details: "some step standard value is nil",
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply(
							mcom.GetWorkOrderReply{
								ID: testWorkOrder1,
								Product: mcom.Product{
									ID:   testWorkOrder1ProductA,
									Type: testWorkOrder1ProductType,
								},
								Process: mcom.WorkOrderProcess{
									Name: testProcessA,
									Type: testProcessAType,
								},
								RecipeID:     testWorkOrder1RecipeID,
								Status:       workorder.Status_PENDING,
								DepartmentID: testDepartmentOID,
								Station:      testStationA,
								Sequence:     int32(testSequence),
								Date:         testSchedulingDate,
								BatchQuantityDetails: mcom.NewQuantityPerBatch(
									[]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}).Detail(),
								UpdatedBy:         userID,
								UpdatedAt:         testUpdateDate,
								InsertedBy:        userID,
								InsertedAt:        testUpdateDate,
								Parent:            testParentWorkOrder1,
								CurrentBatch:      testCurrentBatch,
								CollectedQuantity: testCurrentQuantity,
								CollectedSequence: testSequence,
							}),
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testWorkOrder1RecipeID,
							ProcessName: testProcessA,
							ProcessType: testProcessAType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								OID:  testProcessAOID,
								Name: testProcessA,
								Type: testProcessAType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations:  []string{testStationA, testStationB},
										BatchSize: testBatchSizeDecimal,
										Tools: []*mcom.RecipeTool{
											{
												ID:       testToolID,
												Required: true,
											},
										},
										Steps: []*mcom.RecipeProcessStep{
											{
												Materials: []*mcom.RecipeMaterial{
													{
														Name:  testMaterialAID,
														Grade: testMaterialAGrade,
														Value: mcom.RecipeMaterialParameter{
															High: testMaterialAMaxValue,
															Low:  testMaterialAMinValue,
															Unit: testUnit,
														},
														Site:             testSiteName1,
														RequiredRecipeID: testWorkOrder1RecipeID,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insufficient request getWorkOrder",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
			script: []mock.Script{
				{ // insufficient request FuncGetWorkOrder
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_INSUFFICIENT_REQUEST,
						},
					},
				},
			},
		},
		{
			name: "insufficient request getProcessDefinition",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
			script: []mock.Script{
				{ // insufficient request FuncProcessDefinition
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply(
							mcom.GetWorkOrderReply{
								ID: testWorkOrder1,
								Product: mcom.Product{
									ID:   testWorkOrder1ProductA,
									Type: testWorkOrder1ProductType,
								},
								Process: mcom.WorkOrderProcess{
									Name: testProcessA,
									Type: testProcessAType,
								},
								Status:       workorder.Status_PENDING,
								DepartmentID: testDepartmentOID,
								Station:      testStationA,
								Sequence:     int32(testSequence),
								Date:         testSchedulingDate,
								BatchQuantityDetails: mcom.NewQuantityPerBatch(
									[]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}).Detail(),
								UpdatedBy:         userID,
								UpdatedAt:         testUpdateDate,
								InsertedBy:        userID,
								InsertedAt:        testUpdateDate,
								Parent:            testParentWorkOrder1,
								CollectedSequence: testSequence,
							}),
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							ProcessName: testProcessA,
							ProcessType: testProcessAType,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_INSUFFICIENT_REQUEST,
						},
					},
				},
			},
		},
		{
			name: "workorder not found",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: "ERROR",
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
			}),
			script: []mock.Script{
				{ // bad request FuncGetWorkOrder
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: "ERROR",
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_WORKORDER_NOT_FOUND,
						},
					},
				},
			},
		},
		{
			name: "getWorkOrder internal error",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{ // internal error FuncGetWorkOrder
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Error: errors.New(testInternalServerError),
					},
				},
			},
		},
		{
			name: "BatchSize internal error",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: fmt.Sprintf("no implementation with %d of BatchSize", 4),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							ID: testWorkOrder1,
							Product: mcom.Product{
								ID:   testWorkOrder1ProductA,
								Type: testWorkOrder1ProductType,
							},
							Process: mcom.WorkOrderProcess{
								OID:  testWorkOrderProcessOID,
								Name: testWorkOrderProcessName,
								Type: testWorkOrderProcessType,
							},
							RecipeID:     testWorkOrder1RecipeID,
							Status:       workorder.Status_PENDING,
							DepartmentID: testDepartmentOID,
							Station:      testStationA,
							Sequence:     1,
							Date:         testSchedulingDate,
							BatchQuantityDetails: mcomModels.BatchQuantityDetails{
								BatchQuantityType:  4,
								QuantityForBatches: []decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2},
							},
							UpdatedBy:  userID,
							UpdatedAt:  testUpdateDate,
							InsertedBy: userID,
							InsertedAt: testUpdateDate,
							Parent:     testParentWorkOrder1,
						},
					},
				},
			},
		},
		{
			name: "getProcessDefinition internal error",
			args: args{
				params: work_order.GetWorkOrderInformationParams{
					HTTPRequest: httpRequestWithHeader,
					WorkOrderID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewGetWorkOrderInformationDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{ // internal error FuncGetProcessDefinition
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply(
							mcom.GetWorkOrderReply{
								ID: testWorkOrder1,
								Product: mcom.Product{
									ID:   testWorkOrder1ProductA,
									Type: testWorkOrder1ProductType,
								},
								Process: mcom.WorkOrderProcess{
									Name: testProcessA,
									Type: "",
								},
								RecipeID:     testWorkOrder1RecipeID,
								Status:       workorder.Status_PENDING,
								DepartmentID: testDepartmentOID,
								Station:      testStationA,
								Sequence:     int32(testSequence),
								Date:         testSchedulingDate,
								BatchQuantityDetails: mcom.NewQuantityPerBatch(
									[]decimal.Decimal{testWorkOrder1Batches1, testWorkOrder1Batches2}).Detail(),
								UpdatedBy:  userID,
								UpdatedAt:  testUpdateDate,
								InsertedBy: userID,
								InsertedAt: testUpdateDate,
								Parent:     testParentWorkOrder1,
							}),
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testWorkOrder1RecipeID,
							ProcessName: testProcessA,
							ProcessType: "",
						},
					},
					Output: mock.Output{
						Error: errors.New(testInternalServerError),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, err := mock.New(tt.script)
			assert.NoErrorf(err, tt.name)

			s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.GetWorkOrderInformation(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("GetWorkOrderInformation() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, err := mock.New([]mock.Script{})
		assert.NoError(err)
		defer dm.Close()
		p := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetWorkOrderInformation(work_order.GetWorkOrderInformationParams{
			HTTPRequest: httpRequestWithHeader,
			WorkOrderID: testWorkOrder1,
		}, principal).(*work_order.GetWorkOrderInformationDefault)
		assert.True(ok)
		assert.Equal(work_order.NewGetWorkOrderInformationDefault(http.StatusForbidden), rep)
	}
}

func Test_parseCreateWorkOrdersRequest(t *testing.T) {
	assert := assert.New(t)
	var (
		recipeID     = "U-Z-223G2056-N-2"
		versionStage = "NORMAL_PRODUCTION"
		productID    = "223G2056"
		station      = "KU-P2510-BOM-402-1"
		processName  = "curing"
		processType  = "PRODUCE"
		processOID   = "processOID"
		batchSize    = decimal.NewFromFloat(79.21)
	)

	{
		script := []mock.Script{
			{ // 2
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 3
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: productID,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: recipeID,
								Product: mcom.Product{
									ID: productID,
								},
								Version: mcom.RecipeVersion{
									Stage: versionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  processOID,
											Name: processName,
											Type: processType,
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{station},
													BatchSize: &batchSize,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{ // 4
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 5
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: productID,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: recipeID,
								Product: mcom.Product{
									ID: productID,
								},
								Version: mcom.RecipeVersion{
									Stage: versionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  processOID,
											Name: processName,
											Type: processType,
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{station},
													BatchSize: &batchSize,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{ // 6
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 7
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 8
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 9
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 10
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: "",
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_PROCESS_NOT_FOUND,
					},
				},
			},
			{ // 11
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: "",
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_PROCESS_NOT_FOUND,
					},
				},
			},
			{ // 12
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 13
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 14
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 15
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 16
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 17
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    "bad",
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_PROCESS_NOT_FOUND,
					},
				},
			},
			{ // 18
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: "bad",
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_PROCESS_NOT_FOUND,
					},
				},
			},
			{ // 19
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: "bad",
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_PROCESS_NOT_FOUND,
					},
				},
			},
			{ // 20
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 21
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 22
				Name: mock.FuncGetProcessDefinition,
				Input: mock.Input{
					Request: mcom.GetProcessDefinitionRequest{
						RecipeID:    recipeID,
						ProcessName: processName,
						ProcessType: processType,
					},
				},
				Output: mock.Output{
					Response: mcom.GetProcessDefinitionReply{
						ProcessDefinition: mcom.ProcessDefinition{
							OID:  processOID,
							Name: testProcessName,
							Type: testProcessType,
							Output: mcom.OutputProduct{
								ID: productID,
							},
							Configs: []*mcom.RecipeProcessConfig{
								{
									Stations:  []string{station},
									BatchSize: &batchSize,
								},
							},
						},
					},
				},
			},
			{ // 23
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: productID,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: recipeID,
								Product: mcom.Product{
									ID: productID,
								},
								Version: mcom.RecipeVersion{
									Stage: versionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  processOID,
											Name: processName,
											Type: processType,
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{station},
													BatchSize: &batchSize,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{ // 24
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: productID,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: recipeID,
								Product: mcom.Product{
									ID: productID,
								},
								Version: mcom.RecipeVersion{
									Stage: versionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  processOID,
											Name: processName,
											Type: processType,
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{station},
													BatchSize: &batchSize,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		dm, _ := mock.New(script)
		rows := [][]string{
			/*  1 */ {"產品代號", "機台號", "生產version_stage", "配合表編號", "工序名稱", "工序類別", "計算方式(0.首數/1.預計產量)", "首數/預計產量", "生產批量", "預計生產日"},
			/*  2 */ {"223G2056", "KU-P2510-BOM-402-1", "NORMAL_PRODUCTION", "U-Z-223G2056-N-2", "curing", "PRODUCE", "1", "5", "", "2022-10-20"},
			/*  3 */ {"223G2056", "KU-P2510-BOM-402-1", "NORMAL_PRODUCTION", "", "curing", "PRODUCE", "1", "6", "", "2022-10-20"},
			/*  4 */ {"223G2056", "KU-P2510-BOM-402-1", "qq", "U-Z-223G2056-N-2", "curing", "PRODUCE", "0", "7", "", "2022-10-20"},
			/*  5 */ {"223G2056", "KU-P2510-BOM-402-1", "", "", "curing", "PRODUCE", "0", "8", "", "2022-10-20"},
			/*  6 */ {"223G2056", "KU-P2510-BOM-402-1", "NORMAL_PRODUCTION", "U-Z-223G2056-N-2", "curing", "PRODUCE", "0", "9", "2", "2022-10-20"},
			/*  7 */ {"", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "0", "10", "", "2022-10-20"},
			/*  8 */ {"223G2056", "", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "0", "11", "", "2022-10-20"},
			/*  9 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "0", "12", "", "2022-10-20"},
			/* 10 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "", "PRODUCE", "0", "13", "", "2022-10-20"},
			/* 11 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "", "0", "14", "", "2022-10-20"},
			/* 12 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "", "15", "", "2022-10-20"},
			/* 13 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "0", "", "", "2022-10-20"},
			/* 14 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "0", "16", "", ""},
			/* 15 */ {"bad", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "1", "17", "", "2022-10-20"},
			/* 16 */ {"223G2056", "bad", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "1", "5", "", "2022-10-20"},
			/* 17 */ {"223G2056", "KU-P2510-BOM-402-1", "", "bad", "curing", "PRODUCE", "1", "5", "", "2022-10-20"},
			/* 18 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "bad", "PRODUCE", "1", "5", "", "2022-10-20"},
			/* 19 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "bad", "1", "5", "", "2022-10-20"},
			/* 20 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "bad", "5", "", "2022-10-20"},
			/* 21 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "1", "bad", "", "2022-10-20"},
			/* 22 */ {"223G2056", "KU-P2510-BOM-402-1", "", "U-Z-223G2056-N-2", "curing", "PRODUCE", "1", "5", "", "2022/10/20"},
			/* 23 */ {"223G2056", "KU-P2510-BOM-402-1", "bad", "", "curing", "PRODUCE", "1", "5", "", "2022-10-20"},
			/* 24 */ {"223G2056", "", "", "", "", "", "", "", "", ""},
		}

		req, failData, err := parseCreateWorkOrdersRequest(context.Background(), dm, "", rows)
		assert.NoError(err)
		assert.Equal([]*work_order.CreateWorkOrdersFromFileOKBodyDataFailDataItems0{
			{
				Columns: []string{
					"B(機台號)",
					"C(生產version_stage)",
					"E(工序名稱)",
					"F(工序類別)",
					"I(生產批量)",
				},
				Index: 5,
			},
			{
				Columns: []string{
					"A(產品代號)",
				},
				Index: 7,
			},
			{
				Columns: []string{
					"B(機台號)",
					"I(生產批量)",
				},
				Index: 8,
			},
			{
				Columns: []string{
					"D(配合表編號)",
					"E(工序名稱)",
					"F(工序類別)",
					"I(生產批量)",
				},
				Index: 10,
			},
			{
				Columns: []string{
					"D(配合表編號)",
					"E(工序名稱)",
					"F(工序類別)",
					"I(生產批量)",
				},
				Index: 11,
			},
			{
				Columns: []string{
					"G(計算方式(0.首數/1.預計產量))",
					"H(首數/預計產量)",
				},
				Index: 12,
			},
			{
				Columns: []string{
					"H(首數/預計產量)",
				},
				Index: 13,
			},
			{
				Columns: []string{
					"J(預計生產日)",
				},
				Index: 14,
			},
			{
				Columns: []string{
					"A(產品代號)",
				},
				Index: 15,
			},
			{
				Columns: []string{
					"B(機台號)",
					"I(生產批量)",
				},
				Index: 16,
			},
			{
				Columns: []string{
					"D(配合表編號)",
					"E(工序名稱)",
					"F(工序類別)",
					"I(生產批量)",
				},
				Index: 17,
			},
			{
				Columns: []string{
					"D(配合表編號)",
					"E(工序名稱)",
					"F(工序類別)",
					"I(生產批量)",
				},
				Index: 18,
			},
			{
				Columns: []string{
					"D(配合表編號)",
					"E(工序名稱)",
					"F(工序類別)",
					"I(生產批量)",
				},
				Index: 19,
			},
			{
				Columns: []string{
					"G(計算方式(0.首數/1.預計產量))",
					"H(首數/預計產量)",
				},
				Index: 20,
			},
			{
				Columns: []string{
					"H(首數/預計產量)",
				},
				Index: 21,
			},
			{
				Columns: []string{
					"J(預計生產日)",
				},
				Index: 22,
			},
			{
				Columns: []string{
					"B(機台號)",
					"C(生產version_stage)",
					"E(工序名稱)",
					"F(工序類別)",
					"I(生產批量)",
				},
				Index: 23,
			},
			{
				Columns: []string{
					"B(機台號)",
					"C(生產version_stage)",
					"E(工序名稱)",
					"F(工序類別)",
					"G(計算方式(0.首數/1.預計產量))",
					"H(首數/預計產量)",
					"J(預計生產日)",
				},
				Index: 24,
			},
		}, failData)
		assert.Empty(req)
		assert.NoError(dm.Close())
	}
	{ // 未指定配合表編號, 但系統有兩筆配合表, 且各自配合表各有兩個 processes, 每個 process 的機台皆不同.
		script := []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: productID,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: recipeID,
								Product: mcom.Product{
									ID: productID,
								},
								Version: mcom.RecipeVersion{
									Stage: versionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  "processOID1",
											Name: "processName1",
											Type: "processType1",
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{"station1"},
													BatchSize: &batchSize,
												},
											},
										},
									},
									{
										Info: mcom.ProcessDefinition{
											OID:  "processOID2",
											Name: "processName2",
											Type: "processType2",
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{"station2"},
													BatchSize: &batchSize,
												},
											},
										},
									},
								},
							},
							{
								ID: recipeID,
								Product: mcom.Product{
									ID: productID,
								},
								Version: mcom.RecipeVersion{
									Stage: versionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  "processOID3",
											Name: "processName3",
											Type: "processType3",
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{"station3"},
													BatchSize: &batchSize,
												},
											},
										},
									},
									{
										Info: mcom.ProcessDefinition{
											OID:  processOID,
											Name: processName,
											Type: processType,
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{station},
													BatchSize: &batchSize,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		dm, _ := mock.New(script)
		rows := [][]string{
			/*  1 */ {"產品代號", "機台號", "生產version_stage", "配合表編號", "工序名稱", "工序類別", "計算方式(0.首數/1.預計產量)", "首數/預計產量", "生產批量", "預計生產日"},
			/*  2 */ {"223G2056", "KU-P2510-BOM-402-1", "NORMAL_PRODUCTION", "", "curing", "PRODUCE", "1", "5", "", "2022-10-20"},
		}
		date, _ := time.Parse("2006-01-02", "2022-10-20")
		req, failData, err := parseCreateWorkOrdersRequest(context.Background(), dm, "", rows)
		assert.NoError(err)
		assert.Len(failData, 0)
		assert.Equal(mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					RecipeID:        recipeID,
					ProcessOID:      processOID,
					ProcessName:     processName,
					ProcessType:     processType,
					Station:         station,
					BatchesQuantity: mcom.NewPlanQuantity(1, decimal.NewFromInt(5)),
					Date:            date,
				},
			},
		}, req)
		assert.NoError(dm.Close())
	}
	{ // 配合表未定義batch-size, 使用者有填寫batch-size.
		script := []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: productID,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: recipeID,
								Product: mcom.Product{
									ID: productID,
								},
								Version: mcom.RecipeVersion{
									Stage: versionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  processOID,
											Name: processName,
											Type: processType,
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{station},
													BatchSize: nil,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		dm, _ := mock.New(script)
		rows := [][]string{
			/*  1 */ {"產品代號", "機台號", "生產version_stage", "配合表編號", "工序名稱", "工序類別", "計算方式(0.首數/1.預計產量)", "首數/預計產量", "生產批量", "預計生產日"},
			/*  2 */ {"223G2056", "KU-P2510-BOM-402-1", "NORMAL_PRODUCTION", "", "curing", "PRODUCE", "1", "5", "4", "2022-10-20"},
		}
		date, _ := time.Parse("2006-01-02", "2022-10-20")
		req, failData, err := parseCreateWorkOrdersRequest(context.Background(), dm, "", rows)
		assert.NoError(err)
		assert.Len(failData, 0)
		assert.Equal(mcom.CreateWorkOrdersRequest{
			WorkOrders: []mcom.CreateWorkOrder{
				{
					RecipeID:        recipeID,
					ProcessOID:      processOID,
					ProcessName:     processName,
					ProcessType:     processType,
					Station:         station,
					BatchesQuantity: mcom.NewPlanQuantity(2, decimal.NewFromInt(5)),
					Date:            date,
				},
			},
		}, req)
		assert.NoError(dm.Close())
	}
	{ // 配合表未定義batch-size, 使用者未填寫batch-size.
		script := []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: productID,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: recipeID,
								Product: mcom.Product{
									ID: station,
								},
								Version: mcom.RecipeVersion{
									Stage: versionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											Name: processName,
											Type: processType,
											Output: mcom.OutputProduct{
												ID: productID,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{station},
													BatchSize: nil,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		dm, _ := mock.New(script)
		rows := [][]string{
			/*  1 */ {"產品代號", "機台號", "生產version_stage", "配合表編號", "工序名稱", "工序類別", "計算方式(0.首數/1.預計產量)", "首數/預計產量", "生產批量", "預計生產日"},
			/*  2 */ {"223G2056", "KU-P2510-BOM-402-1", "NORMAL_PRODUCTION", "", "curing", "PRODUCE", "1", "5", "", "2022-10-20"},
		}
		req, failData, err := parseCreateWorkOrdersRequest(context.Background(), dm, "", rows)
		assert.NoError(err)
		assert.Equal([]*work_order.CreateWorkOrdersFromFileOKBodyDataFailDataItems0{
			{
				Index: 2,
				Columns: []string{
					"I(生產批量)",
				},
			},
		}, failData)
		assert.Equal(mcom.CreateWorkOrdersRequest{}, req)
		assert.NoError(dm.Close())
	}
}

func Test_getLatestRecipe(t *testing.T) {
	assert := assert.New(t)
	var (
		testVersionStage = "VersionStage"
		testRecipeID     = "RecipeID"
	)
	{ // case success
		dm, _ := mock.New([]mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testWorkOrder1ProductA,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: testRecipeID,
								Product: mcom.Product{
									ID: testWorkOrder1ProductA,
								},
								Version: mcom.RecipeVersion{
									Stage: testVersionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											Name: testProcessName,
											Type: testProcessType,
											Output: mcom.OutputProduct{
												ID: testWorkOrder1ProductA,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations: []string{testStationA},
												},
											},
										},
									},
								},
							},
							{
								ID: "RecipeID2",
								Product: mcom.Product{
									ID: "ProductID",
								},
								Version: mcom.RecipeVersion{
									Stage: "VersionStage",
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations: []string{testStationA},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
		rep, err := getLatestRecipe(context.Background(), dm, Row([]string{testWorkOrder1ProductA, testStationA, testVersionStage, "", testProcessName, testProcessType}))
		assert.Equal(len(err), 0)
		assert.Equal(rep, mcom.GetRecipeReply{
			ID: testRecipeID,
			Product: mcom.Product{
				ID: testWorkOrder1ProductA,
			},
			Version: mcom.RecipeVersion{
				Stage: testVersionStage,
			},
			Processes: []*mcom.ProcessEntity{
				{
					Info: mcom.ProcessDefinition{
						Name: testProcessName,
						Type: testProcessType,
						Output: mcom.OutputProduct{
							ID: testWorkOrder1ProductA,
						},
						Configs: []*mcom.RecipeProcessConfig{
							{
								Stations: []string{testStationA},
							},
						},
					},
				},
			},
		})
		dm.Close()
	}
	{ // case not found matched latest recipe with versionStage
		dm, _ := mock.New([]mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testWorkOrder1ProductA,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: testRecipeID,
								Product: mcom.Product{
									ID: testWorkOrder1ProductA,
								},
								Version: mcom.RecipeVersion{
									Stage: "9999",
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations: []string{testStationA},
												},
											},
										},
									},
								},
							},
							{
								ID: "RecipeID2",
								Product: mcom.Product{
									ID: testWorkOrder1ProductA,
								},
								Version: mcom.RecipeVersion{
									Stage: "7777",
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations: []string{testStationA},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
		_, err := getLatestRecipe(context.Background(), dm, Row([]string{testWorkOrder1ProductA, "testStationB", testVersionStage, "", testProcessName, testProcessType}))
		assert.Equal(err, []int{colStation, colVersionStage, colProcessName, colProcessType})
		dm.Close()
	}
	{ // case not found matched latest recipe with processName & processType
		dm, _ := mock.New([]mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testWorkOrder1ProductA,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: testRecipeID,
								Product: mcom.Product{
									ID: testWorkOrder1ProductA,
								},
								Version: mcom.RecipeVersion{
									Stage: testVersionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations: []string{testStationA},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
		_, err := getLatestRecipe(context.Background(), dm, Row([]string{testWorkOrder1ProductA, "testStationB", testVersionStage, "", testProcessName, testProcessType}))
		assert.Equal(err, []int{colStation, colProcessName, colProcessType})
		dm.Close()
	}
	{ // case not found matched latest recipe with station
		dm, _ := mock.New([]mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testWorkOrder1ProductA,
					}.WithOrder(
						mcom.Order{
							Name:       "released_at",
							Descending: true,
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: testRecipeID,
								Product: mcom.Product{
									ID: testWorkOrder1ProductA,
								},
								Version: mcom.RecipeVersion{
									Stage: testVersionStage,
								},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											Name: testProcessName,
											Type: testProcessType,
											Output: mcom.OutputProduct{
												ID: testWorkOrder1ProductA,
											},
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations: []string{testStationA},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
		_, err := getLatestRecipe(context.Background(), dm, Row([]string{testWorkOrder1ProductA, "testStationB", testVersionStage, "", testProcessName, testProcessType}))
		assert.Equal(err, []int{colStation})
		dm.Close()
	}
}

func Test_parseBatchQuantity(t *testing.T) {
	assert := assert.New(t)

	{ // case success fixed
		rowFixed := Row{"0", "1", "2", "3", "4", "5", "0", "10", "8", "9"}
		batchSize := decimal.NewFromFloat(6)
		rep, badColumnsIndex := parseBatchQuantity(&batchSize, rowFixed)
		assert.Equal(len(badColumnsIndex), 0)
		assert.Equal(rep, mcom.NewFixedQuantity(uint(10), decimal.NewFromInt(60)))
	}
	{ // case success plan
		rowPlan := Row{"0", "1", "2", "3", "4", "5", "1", "1111.11", "8", "9"}
		batchSize := decimal.NewFromFloat(1.11)
		rep, badColumnsIndex := parseBatchQuantity(&batchSize, rowPlan)
		assert.Equal(len(badColumnsIndex), 0)
		assert.Equal(rep, mcom.NewPlanQuantity(uint(1001), decimal.NewFromFloat(1111.11)))
	}
	{ // case batchSize empty success
		rowFixed := Row{"0", "1", "2", "3", "4", "5", "0", "10", "8", "9"}
		rep, badColumnsIndex := parseBatchQuantity(nil, rowFixed)
		assert.Equal(len(badColumnsIndex), 0)
		assert.Equal(rep, mcom.NewFixedQuantity(uint(10), decimal.NewFromInt(80)))
	}
	{ // case batchSize empty error
		rowFixed := Row{"0", "1", "2", "3", "4", "5", "0", "10", "", "9"}
		_, badColumnsIndex := parseBatchQuantity(nil, rowFixed)
		assert.Equal(badColumnsIndex, []int{8})
	}
}

func TestWorkOrder_UpdateWorkOrder(t *testing.T) {
	assert := assert.New(t)
	httpRequestWithHeader := httptest.NewRequest("PUT", "/work-orders/"+testWorkOrder1, nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	badStatus := work_order.NewUpdateWorkOrderDefault(400)
	badStatus.Payload = &models.Error{
		Code:    int64(mcomErrors.Code_BAD_REQUEST),
		Details: "work order status not pending",
	}
	date := strfmt.Date(testSchedulingDate)
	type args struct {
		params    work_order.UpdateWorkOrderParams
		principal *models.Principal
	}
	tests := []struct {
		name    string
		args    args
		want    middleware.Responder
		scripts []mock.Script
	}{
		{ // case with quantities per batch
			name: "success update with quantities per batch",
			args: args{
				params: work_order.UpdateWorkOrderParams{
					HTTPRequest: httpRequestWithHeader,
					Body: &models.UpdateWorkOrder{
						BatchesQuantity: []string{"100", "100"},
						PlanDate:        &date,
						Recipe: &models.Recipe{
							ProcessOID:  testProcessOID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
							ID:          testWorkOrder1RecipeID,
						},
						Station: &testStationID,
					},
					ID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewUpdateWorkOrderOK(),
			scripts: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							Status: workorder.Status_PENDING,
						},
					},
				},
				{
					Name: mock.FuncUpdateWorkOrders,
					Input: mock.Input{
						Request: mcom.UpdateWorkOrdersRequest{
							Orders: []mcom.UpdateWorkOrder{
								{
									ID:              testWorkOrder1,
									Station:         testStationID,
									RecipeID:        testWorkOrder1RecipeID,
									ProcessOID:      testProcessOID,
									ProcessName:     testProcessName,
									ProcessType:     testProcessType,
									Date:            testSchedulingDate,
									BatchesQuantity: mcom.NewQuantityPerBatch([]decimal.Decimal{decimal.NewFromInt(100), decimal.NewFromInt(100)}),
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // case with plan quantity
			name: "success update with plan quantity",
			args: args{
				params: work_order.UpdateWorkOrderParams{
					HTTPRequest: httpRequestWithHeader,
					Body: &models.UpdateWorkOrder{
						BatchSize:    int64(mcomWorkOrder.BatchSize_PLAN_QUANTITY),
						BatchCount:   5,
						PlanQuantity: "500",
						PlanDate:     &date,
						Recipe: &models.Recipe{
							ProcessOID:  testProcessOID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
							ID:          testWorkOrder1RecipeID,
						},
						Station: &testStationID,
					},
					ID: testWorkOrder1,
				},
				principal: principal,
			},
			want: work_order.NewUpdateWorkOrderOK(),
			scripts: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							Status: workorder.Status_PENDING,
						},
					},
				},
				{
					Name: mock.FuncUpdateWorkOrders,
					Input: mock.Input{
						Request: mcom.UpdateWorkOrdersRequest{
							Orders: []mcom.UpdateWorkOrder{
								{
									ID:              testWorkOrder1,
									Station:         testStationID,
									RecipeID:        testWorkOrder1RecipeID,
									ProcessOID:      testProcessOID,
									ProcessName:     testProcessName,
									ProcessType:     testProcessType,
									Date:            testSchedulingDate,
									BatchesQuantity: mcom.NewPlanQuantity(5, decimal.NewFromInt(500)),
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // case with bad status
			name: "fail to update with bad status",
			args: args{
				params: work_order.UpdateWorkOrderParams{
					HTTPRequest: httpRequestWithHeader,
					Body: &models.UpdateWorkOrder{
						BatchSize:    int64(mcomWorkOrder.BatchSize_PLAN_QUANTITY),
						BatchCount:   5,
						PlanQuantity: "500",
						PlanDate:     &date,
						Recipe: &models.Recipe{
							ProcessOID:  testProcessOID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
							ID:          testWorkOrder1RecipeID,
						},
						Station: &testStationID,
					},
					ID: testWorkOrder1,
				},
				principal: principal,
			},
			want: badStatus,
			scripts: []mock.Script{
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrder1,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							Status: workorder.Status_ACTIVE,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, err := mock.New(tt.scripts)
			assert.NoError(err)
			s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.UpdateWorkOrder(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateWorkOrder() = %v, want %v", got, tt.want)
			}
			assert.NoError(dm.Close())
		})
	}
	{ // missing batch count
		dm, err := mock.New([]mock.Script{})
		assert.NoError(err)
		defer dm.Close()
		s := mustNewWorkorder(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep := s.UpdateWorkOrder(work_order.UpdateWorkOrderParams{
			HTTPRequest: httpRequestWithHeader,
			Body: &models.UpdateWorkOrder{
				BatchCount:      0,
				BatchSize:       int64(mcomWorkOrder.BatchSize_FIXED_QUANTITY),
				BatchesQuantity: []string{"20"},
				PlanDate:        &strfmt.Date{},
				PlanQuantity:    "20",
				Recipe: &models.Recipe{
					ID:          userID,
					ProcessName: testProcessName,
					ProcessOID:  testProcessOID,
					ProcessType: testProcessType,
				},
				Station: new(string),
			},
			ID: testWorkOrder1,
		}, principal)
		expected := work_order.NewUpdateWorkOrderDefault(400)
		expected.Payload = &models.Error{
			Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			Details: "missing batch count or quantity",
		}
		assert.Equal(expected, rep)
	}
}

func mustNewWorkorder(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool,
) service.WorkOrder {
	s := NewWorkOrder(dm, hasPermission, Config{})
	return s
}
