package resource

import (
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

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/resource"
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

	testProductType     = "RUBBER"
	testProductID       = "BAN"
	testProductGrade    = "B"
	testLotNumber       = "xxx-0120"
	testQuantity        = "11"
	testQuantityDecimal = decimal.NewFromInt(11)
	testUnit            = "sku"

	testWarehouse = "WX"
	testLocation  = "001"

	productDate = strfmt.DateTime(testProductionDate)
	expiredDate = strfmt.DateTime(testExpirationDate)

	testProductionDate = time.Date(2021, 8, 9, 0, 0, 0, 0, time.Local)
	testExpirationDate = time.Date(2021, 8, 11, 0, 0, 0, 0, time.Local)

	resourceID           = "R0147852369"
	unexpectedResourceID = "X1424788124"

	testBrokenQuantity = "T_T"

	testTime = time.Unix(0, 1646803251)

	testSiteName1 = "TESTSITENAME1"

	testToolID = "ToolID"

	testInspections = mcomModels.Inspections{
		{ID: 1,
			Remark: "001"},
		{ID: 2,
			Remark: "002"},
	}
)

func intToInt64(dataIn []int) []int64 {
	dataOut := make([]int64, len(dataIn))
	for i := 0; i < len(dataIn); i++ {
		dataOut[i] = int64(dataIn[i])
	}
	return dataOut
}

func TestResource_AddMaterial(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("POST", "/resource/material/stock", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	{ // normal case
		scripts := []mock.Script{
			{
				Name: mock.FuncCreateMaterialResources,
				Input: mock.Input{
					Request: mcom.CreateMaterialResourcesRequest{
						Materials: []mcom.CreateMaterialResourcesRequestDetail{
							{
								Type:           testProductType,
								ID:             testProductID,
								Grade:          testProductGrade,
								Status:         resources.MaterialStatus_INSPECTION,
								Quantity:       testQuantityDecimal,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ProductionTime: testProductionDate,
								ExpiryTime:     testExpirationDate,
							},
						},
					},
					Options: []interface{}{
						mcom.WithStockIn(mcom.Warehouse{ID: testWarehouse, Location: testLocation}),
					},
				},
				Output: mock.Output{
					Response: mcom.CreateMaterialResourcesReply{
						{
							ID: resourceID,
						},
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.AddMaterial(resource.AddMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.AddMaterialBody{
				Resource: &models.ProductInfo{
					ExpiryTime:     &expiredDate,
					Grade:          &testProductGrade,
					LotNumber:      &testLotNumber,
					ProductID:      &testProductID,
					ProductType:    &testProductType,
					ProductionTime: &productDate,
					Quantity:       &testQuantity,
					Unit:           &testUnit,
				},
				Warehouse: &models.Warehouse{
					ID:       &testWarehouse,
					Location: &testLocation,
				},
			},
		}, principal).(*resource.AddMaterialOK)
		if assert.True(ok) {
			assert.Equal(resource.NewAddMaterialOK().WithPayload(&resource.AddMaterialOKBody{
				Data: &resource.AddMaterialOKBodyData{
					ResourceID: resourceID,
				},
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // invalid number
		dm, err := mock.New([]mock.Script{})
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.AddMaterial(resource.AddMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.AddMaterialBody{
				Resource: &models.ProductInfo{
					ExpiryTime:     &expiredDate,
					Grade:          &testProductGrade,
					LotNumber:      &testLotNumber,
					ProductID:      &testProductID,
					ProductType:    &testProductType,
					ProductionTime: &productDate,
					Quantity:       &testBrokenQuantity,
					Unit:           &testUnit,
				},
				Warehouse: &models.Warehouse{
					ID:       &testWarehouse,
					Location: &testLocation,
				},
			},
		}, principal).(*resource.AddMaterialDefault)
		if assert.True(ok) {
			assert.Equal(resource.NewAddMaterialDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: "invalid_number=" + testBrokenQuantity,
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // bad request on create material
		scripts := []mock.Script{
			{
				Name: mock.FuncCreateMaterialResources,
				Input: mock.Input{
					Request: mcom.CreateMaterialResourcesRequest{
						Materials: []mcom.CreateMaterialResourcesRequestDetail{
							{
								Type:           testProductType,
								ID:             testProductID,
								Grade:          testProductGrade,
								Status:         resources.MaterialStatus_INSPECTION,
								Quantity:       testQuantityDecimal,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ProductionTime: testProductionDate,
								ExpiryTime:     testExpirationDate,
							},
						},
					},
					Options: []interface{}{
						mcom.WithStockIn(mcom.Warehouse{ID: testWarehouse, Location: testLocation}),
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_RESOURCE_EXISTED,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.AddMaterial(resource.AddMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.AddMaterialBody{
				Resource: &models.ProductInfo{
					ExpiryTime:     &expiredDate,
					Grade:          &testProductGrade,
					LotNumber:      &testLotNumber,
					ProductID:      &testProductID,
					ProductType:    &testProductType,
					ProductionTime: &productDate,
					Quantity:       &testQuantity,
					Unit:           &testUnit,
				},
				Warehouse: &models.Warehouse{
					ID:       &testWarehouse,
					Location: &testLocation,
				},
			},
		}, principal).(*resource.AddMaterialDefault)
		if assert.True(ok) {
			assert.Equal(resource.NewAddMaterialDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_RESOURCE_EXISTED),
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // internal error on create material
		scripts := []mock.Script{
			{
				Name: mock.FuncCreateMaterialResources,
				Input: mock.Input{
					Request: mcom.CreateMaterialResourcesRequest{
						Materials: []mcom.CreateMaterialResourcesRequestDetail{
							{
								Type:           testProductType,
								ID:             testProductID,
								Grade:          testProductGrade,
								Status:         resources.MaterialStatus_INSPECTION,
								Quantity:       testQuantityDecimal,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ProductionTime: testProductionDate,
								ExpiryTime:     testExpirationDate,
							},
						},
					},
					Options: []interface{}{
						mcom.WithStockIn(mcom.Warehouse{ID: testWarehouse, Location: testLocation}),
					},
				},
				Output: mock.Output{
					Error: errors.New(testInternalServerError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.AddMaterial(resource.AddMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.AddMaterialBody{
				Resource: &models.ProductInfo{
					ExpiryTime:     &expiredDate,
					Grade:          &testProductGrade,
					LotNumber:      &testLotNumber,
					ProductID:      &testProductID,
					ProductType:    &testProductType,
					ProductionTime: &productDate,
					Quantity:       &testQuantity,
					Unit:           &testUnit,
				},
				Warehouse: &models.Warehouse{
					ID:       &testWarehouse,
					Location: &testLocation,
				},
			},
		}, principal).(*resource.AddMaterialDefault)
		if assert.True(ok) {
			assert.Equal(resource.NewAddMaterialDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // unexpected create material results
		response := mcom.CreateMaterialResourcesReply{
			{
				ID: resourceID,
			},
			{
				ID: unexpectedResourceID,
			},
		}
		scripts := []mock.Script{
			{
				Name: mock.FuncCreateMaterialResources,
				Input: mock.Input{
					Request: mcom.CreateMaterialResourcesRequest{
						Materials: []mcom.CreateMaterialResourcesRequestDetail{
							{
								Type:           testProductType,
								ID:             testProductID,
								Grade:          testProductGrade,
								Status:         resources.MaterialStatus_INSPECTION,
								Quantity:       testQuantityDecimal,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ProductionTime: testProductionDate,
								ExpiryTime:     testExpirationDate,
							},
						},
					},
					Options: []interface{}{
						mcom.WithStockIn(mcom.Warehouse{ID: testWarehouse, Location: testLocation}),
					},
				},
				Output: mock.Output{
					Response: response,
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.AddMaterial(resource.AddMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.AddMaterialBody{
				Resource: &models.ProductInfo{
					ExpiryTime:     &expiredDate,
					Grade:          &testProductGrade,
					LotNumber:      &testLotNumber,
					ProductID:      &testProductID,
					ProductType:    &testProductType,
					ProductionTime: &productDate,
					Quantity:       &testQuantity,
					Unit:           &testUnit,
				},
				Warehouse: &models.Warehouse{
					ID:       &testWarehouse,
					Location: &testLocation,
				},
			},
		}, principal).(*resource.AddMaterialDefault)
		if assert.True(ok) {
			assert.Equal(resource.NewAddMaterialDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: fmt.Sprintf("unexpected add material result=%v", response),
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.AddMaterial(resource.AddMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.AddMaterialBody{
				Resource: &models.ProductInfo{
					ExpiryTime:     &expiredDate,
					Grade:          &testProductGrade,
					LotNumber:      &testLotNumber,
					ProductID:      &testProductID,
					ProductType:    &testProductType,
					ProductionTime: &productDate,
					Quantity:       &testQuantity,
					Unit:           &testUnit,
				},
				Warehouse: &models.Warehouse{
					ID:       &testWarehouse,
					Location: &testLocation,
				},
			},
		}, principal).(*resource.AddMaterialDefault)
		assert.True(ok)
		assert.Equal(resource.NewAddMaterialDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestResource_SplitMaterial(t *testing.T) {
	assert := assert.New(t)
	var (
		testSplitQuantity = 1.234567
		errorQuantity     = 0.000000001

		testInspectionIDs = []int{4, 11, 20, 22}
		testRemark        = "split test"

		errorResourceID  = "XXXXXXXXXXXXXXX"
		errorProductType = ""
	)
	httpRequestWithHeader := httptest.NewRequest("POST", "/resource/material/split", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	{ // normal case OK
		scripts := []mock.Script{
			{
				Name: mock.FuncSplitMaterialResource,
				Input: mock.Input{
					Request: mcom.SplitMaterialResourceRequest{
						ResourceID:    testProductID,
						ProductType:   testProductType,
						Quantity:      decimal.NewFromFloat(testSplitQuantity),
						InspectionIDs: testInspectionIDs,
						Remark:        testRemark,
					},
				},
				Output: mock.Output{
					Response: mcom.SplitMaterialResourceReply{
						NewResourceID: resourceID,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.SplitMaterial(resource.SplitMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.SplitMaterialBody{
				ResourceID:    testProductID,
				ProductType:   testProductType,
				SplitQuantity: testSplitQuantity,
				Inspections:   intToInt64(testInspectionIDs),
				Remark:        testRemark,
			},
		}, principal).(*resource.SplitMaterialOK)
		if assert.True(ok) {
			assert.Equal(resource.NewSplitMaterialOK().WithPayload(&resource.SplitMaterialOKBody{
				Data: &resource.SplitMaterialOKBodyData{
					ResourceID: resourceID,
				},
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // INVALID NUMBER
		scripts := []mock.Script{
			{
				Name: mock.FuncSplitMaterialResource,
				Input: mock.Input{
					Request: mcom.SplitMaterialResourceRequest{
						ResourceID:    testProductID,
						ProductType:   testProductType,
						Quantity:      decimal.NewFromFloat(errorQuantity),
						InspectionIDs: testInspectionIDs,
						Remark:        testRemark,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_INVALID_NUMBER,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.SplitMaterial(resource.SplitMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.SplitMaterialBody{
				ResourceID:    testProductID,
				ProductType:   testProductType,
				SplitQuantity: errorQuantity,
				Inspections:   intToInt64(testInspectionIDs),
				Remark:        testRemark,
			},
		}, principal).(*resource.SplitMaterialDefault)
		if assert.True(ok) {
			assert.Equal(resource.NewSplitMaterialDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INVALID_NUMBER),
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // RESOURCE NOT FOUND
		scripts := []mock.Script{
			{
				Name: mock.FuncSplitMaterialResource,
				Input: mock.Input{
					Request: mcom.SplitMaterialResourceRequest{
						ResourceID:    errorResourceID,
						ProductType:   testProductType,
						Quantity:      decimal.NewFromFloat(testSplitQuantity),
						InspectionIDs: testInspectionIDs,
						Remark:        testRemark,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_RESOURCE_NOT_FOUND,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.SplitMaterial(resource.SplitMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.SplitMaterialBody{
				ResourceID:    errorResourceID,
				ProductType:   testProductType,
				SplitQuantity: testSplitQuantity,
				Inspections:   intToInt64(testInspectionIDs),
				Remark:        testRemark,
			},
		}, principal).(*resource.SplitMaterialDefault)
		if assert.True(ok) {
			assert.Equal(resource.NewSplitMaterialDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // INNSUFFICIENT REQUEST
		scripts := []mock.Script{
			{
				Name: mock.FuncSplitMaterialResource,
				Input: mock.Input{
					Request: mcom.SplitMaterialResourceRequest{
						ResourceID:    testProductID,
						ProductType:   errorProductType,
						Quantity:      decimal.NewFromFloat(testSplitQuantity),
						InspectionIDs: testInspectionIDs,
						Remark:        testRemark,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_INSUFFICIENT_REQUEST,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.SplitMaterial(resource.SplitMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.SplitMaterialBody{
				ResourceID:    testProductID,
				ProductType:   errorProductType,
				SplitQuantity: testSplitQuantity,
				Inspections:   intToInt64(testInspectionIDs),
				Remark:        testRemark,
			},
		}, principal).(*resource.SplitMaterialDefault)
		if assert.True(ok) {
			assert.Equal(resource.NewSplitMaterialDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // internal error
		scripts := []mock.Script{
			{
				Name: mock.FuncSplitMaterialResource,
				Input: mock.Input{
					Request: mcom.SplitMaterialResourceRequest{
						ResourceID:    testProductID,
						ProductType:   testProductType,
						Quantity:      decimal.NewFromFloat(testSplitQuantity),
						InspectionIDs: testInspectionIDs,
						Remark:        testRemark,
					},
				},
				Output: mock.Output{
					Error: errors.New(testInternalServerError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := r.SplitMaterial(resource.SplitMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.SplitMaterialBody{
				ResourceID:    testProductID,
				ProductType:   testProductType,
				SplitQuantity: testSplitQuantity,
				Inspections:   intToInt64(testInspectionIDs),
				Remark:        testRemark,
			},
		}, principal).(*resource.SplitMaterialDefault)
		if assert.True(ok) {
			assert.Equal(resource.NewSplitMaterialDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.SplitMaterial(resource.SplitMaterialParams{
			HTTPRequest: httpRequestWithHeader,
			Body: resource.SplitMaterialBody{
				ResourceID:    testProductID,
				ProductType:   testProductType,
				SplitQuantity: testSplitQuantity,
				Inspections:   intToInt64(testInspectionIDs),
				Remark:        testRemark,
			},
		}, principal).(*resource.SplitMaterialDefault)
		assert.True(ok)
		assert.Equal(resource.NewSplitMaterialDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestResource_GetMaterialResourceInfo(t *testing.T) {
	assert := assert.New(t)

	httpRequest := httptest.NewRequest("GET", "/resource/material/info/resource-id/{ID}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	scripts := []mock.Script{
		{
			Name: mock.FuncGetMaterialResource,
			Input: mock.Input{
				Request: mcom.GetMaterialResourceRequest{
					ResourceID: resourceID,
				},
			},
			Output: mock.Output{
				Response: mcom.GetMaterialResourceReply{
					{
						Material: mcom.Material{
							Type:           testProductType,
							ID:             "68000",
							Grade:          "B",
							Status:         resources.MaterialStatus_AVAILABLE,
							Quantity:       decimal.RequireFromString("100"),
							Unit:           testUnit,
							LotNumber:      testLotNumber,
							ProductionTime: testTime,
							ExpiryTime:     testExpirationDate,
							ResourceID:     resourceID,
							UpdatedAt:      types.ToTimeNano(testTime),
							UpdatedBy:      userID,
							CreatedAt:      types.ToTimeNano(testTime),
							CreatedBy:      userID,
							MinDosage:      decimal.NewFromInt32(30),
							Inspections:    testInspections,
							CarrierID:      "",
							Remark:         "",
						},
						Warehouse: mcom.Warehouse{
							ID:       testWarehouse,
							Location: testLocation,
						},
					},
				},
			},
		},
		{
			Name: mock.FuncGetMaterialResource,
			Input: mock.Input{
				Request: mcom.GetMaterialResourceRequest{
					ResourceID: unexpectedResourceID,
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
					Details: "not found resource",
				},
			},
		},
		{
			Name: mock.FuncGetMaterialResource,
			Input: mock.Input{
				Request: mcom.GetMaterialResourceRequest{
					ResourceID: unexpectedResourceID,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}
	dm, err := mock.New(scripts)
	assert.NoError(err)

	type args struct {
		params    resource.GetMaterialResourceInfoParams
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
				params: resource.GetMaterialResourceInfoParams{
					HTTPRequest: httpRequest,
					ID:          resourceID,
				},
				principal: principal,
			},
			want: resource.NewGetMaterialResourceInfoOK().WithPayload(
				&resource.GetMaterialResourceInfoOKBody{
					Data: models.ResourceMaterials{
						&models.ResourceMaterial{
							ID:            "68000",
							CarrierID:     "",
							Remark:        "",
							CreatedAt:     strfmt.DateTime(testTime),
							CreatedBy:     userID,
							ExpiredDate:   strfmt.DateTime(testExpirationDate),
							Grade:         "B",
							Inspections:   inspectionStructType(testInspections),
							MinimumDosage: "30",
							ProductType:   testProductType,
							Quantity:      "100",
							ResourceID:    resourceID,
							Status:        models.MaterialStatus(1),
							Unit:          testUnit,
							UpdatedAt:     strfmt.DateTime(testTime),
							UpdatedBy:     userID,
							Warehouse: &models.Warehouse{
								ID:       &testWarehouse,
								Location: &testLocation,
							},
						},
					},
				}),
		},
		{
			name: "not found resource",
			args: args{
				params: resource.GetMaterialResourceInfoParams{
					HTTPRequest: httpRequest,
					ID:          unexpectedResourceID,
				},
				principal: principal,
			},
			want: resource.NewGetMaterialResourceInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
				Details: "not found resource",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: resource.GetMaterialResourceInfoParams{
					HTTPRequest: httpRequest,
					ID:          unexpectedResourceID,
				},
				principal: principal,
			},
			want: resource.NewGetMaterialResourceInfoDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := r.GetMaterialResourceInfo(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMaterialResourceInfo() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.GetMaterialResourceInfo(resource.GetMaterialResourceInfoParams{
			HTTPRequest: httpRequest,
			ID:          resourceID,
		}, principal).(*resource.GetMaterialResourceInfoDefault)
		assert.True(ok)
		assert.Equal(resource.NewGetMaterialResourceInfoDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestResource_ListMaterialStatus(t *testing.T) {
	assert := assert.New(t)

	httpRequest := httptest.NewRequest("GET", "/resource/material/status", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	scripts := []mock.Script{
		{
			Name:  mock.FuncListMaterialResourceStatus,
			Input: mock.Input{},
			Output: mock.Output{
				Response: mcom.ListMaterialResourceStatusReply{
					resources.MaterialStatus_MATERIAL_STATUS_UNSPECIFIED.String(),
					resources.MaterialStatus_AVAILABLE.String(),
					resources.MaterialStatus_HOLD.String(),
					resources.MaterialStatus_INSPECTION.String(),
					resources.MaterialStatus_MOUNTED.String(),
					resources.MaterialStatus_UNAVAILABLE.String(),
				},
			},
		},
		{
			Name:  mock.FuncListMaterialResourceStatus,
			Input: mock.Input{},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "fake error message",
				},
			},
		},
		{
			Name:  mock.FuncListMaterialResourceStatus,
			Input: mock.Input{},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}
	dm, err := mock.New(scripts)
	assert.NoError(err)

	type args struct {
		params    resource.ListMaterialStatusParams
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
				params: resource.ListMaterialStatusParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: resource.NewListMaterialStatusOK().WithPayload(&resource.ListMaterialStatusOKBody{
				Data: []*resource.ListMaterialStatusOKBodyDataItems0{
					{
						Name: resources.MaterialStatus_MATERIAL_STATUS_UNSPECIFIED.String(),
						ID:   models.MaterialStatus(resources.MaterialStatus_MATERIAL_STATUS_UNSPECIFIED),
					},
					{
						Name: resources.MaterialStatus_AVAILABLE.String(),
						ID:   models.MaterialStatus(resources.MaterialStatus_AVAILABLE),
					},
					{
						Name: resources.MaterialStatus_HOLD.String(),
						ID:   models.MaterialStatus(resources.MaterialStatus_HOLD),
					},
					{
						Name: resources.MaterialStatus_INSPECTION.String(),
						ID:   models.MaterialStatus(resources.MaterialStatus_INSPECTION),
					},
					{
						Name: resources.MaterialStatus_MOUNTED.String(),
						ID:   models.MaterialStatus(resources.MaterialStatus_MOUNTED),
					},
					{
						Name: resources.MaterialStatus_UNAVAILABLE.String(),
						ID:   models.MaterialStatus(resources.MaterialStatus_UNAVAILABLE),
					},
				},
			}),
		},
		{
			name: "bad request",
			args: args{
				params: resource.ListMaterialStatusParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: resource.NewListMaterialStatusDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "fake error message",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: resource.ListMaterialStatusParams{
					HTTPRequest: httpRequest,
				},
				principal: principal,
			},
			want: resource.NewListMaterialStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := r.ListMaterialStatus(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListMaterialStatus() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.ListMaterialStatus(resource.ListMaterialStatusParams{
			HTTPRequest: httpRequest,
		}, principal).(*resource.ListMaterialStatusDefault)
		assert.True(ok)
		assert.Equal(resource.NewListMaterialStatusDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestResource_GetToolID(t *testing.T) {
	var (
		testResourceToolID = "ResourceToolID"
		testBindingSite    = mcomModels.UniqueSite{
			SiteID: mcomModels.SiteID{
				Name:  testSiteName1,
				Index: 1,
			},
			Station: "testStation1",
		}
		testTime             = time.Unix(0, 1646803251)
		unexpectedResourceID = "X1424788124"
	)

	assert := assert.New(t)

	httpRequest := httptest.NewRequest("GET", "/production-flow/tool-resource/{toolResourceID}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    resource.GetToolIDParams
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
				params: resource.GetToolIDParams{
					HTTPRequest:    httpRequest,
					ToolResourceID: testResourceToolID,
				},
				principal: principal,
			},
			want: resource.NewGetToolIDOK().WithPayload(
				&resource.GetToolIDOKBody{
					Data: &resource.GetToolIDOKBodyData{
						ToolID: testToolID,
					},
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetToolResource,
					Input: mock.Input{
						Request: mcom.GetToolResourceRequest{
							ResourceID: testResourceToolID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetToolResourceReply{
							ToolID:      testToolID,
							BindingSite: testBindingSite,
							CreatedBy:   userID,
							CreatedAt:   types.ToTimeNano(testTime),
						},
					},
				},
			},
		},
		{
			name: "not found resource",
			args: args{
				params: resource.GetToolIDParams{
					HTTPRequest:    httpRequest,
					ToolResourceID: unexpectedResourceID,
				},
				principal: principal,
			},
			want: resource.NewGetToolIDDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
				Details: "not found resource",
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetToolResource,
					Input: mock.Input{
						Request: mcom.GetToolResourceRequest{
							ResourceID: unexpectedResourceID,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
							Details: "not found resource",
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: resource.GetToolIDParams{
					HTTPRequest:    httpRequest,
					ToolResourceID: unexpectedResourceID,
				},
				principal: principal,
			},
			want: resource.NewGetToolIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetToolResource,
					Input: mock.Input{
						Request: mcom.GetToolResourceRequest{
							ResourceID: unexpectedResourceID,
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

			s := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.GetToolID(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("GetToolID() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := mustNewResource(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.GetToolID(resource.GetToolIDParams{
			HTTPRequest:    httpRequest,
			ToolResourceID: testResourceToolID,
		}, principal).(*resource.GetToolIDDefault)
		assert.True(ok)
		assert.Equal(resource.NewGetToolIDDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func mustNewResource(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool,
) service.Resource {
	s := NewResource(dm, hasPermission, Config{FontPath: "fake-path"})
	return s
}
