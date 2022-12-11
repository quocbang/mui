package product

import (
	"context"
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
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/product"
)

const (
	userID = "tester"

	testProductTypeA = "TYPE-A"
	testProductTypeB = "TYPE-B"
	testProductTypeC = "TYPE-C"
	testProductTypeD = "TYPE-D"
	testProductTypeE = "TYPE-E"

	testProduct1Parent    = "PRODUCT-A"
	testProduct1Children1 = "PRODUCT-B"
	testProduct1Children2 = "PRODUCT-C"
	testProduct1Children3 = "PRODUCT-D"
	testProduct2Parent    = "PRODUCT-W"
	testProduct2Children1 = "PRODUCT-X"
	testProduct2Children2 = "PRODUCT-Y"
	testProduct2Children3 = "PRODUCT-Z"

	testMaterialAID           = "MATERIAL-A"
	testMaterialAGrade        = "X"
	testMaterialAType         = "NATURAL_RUBBER"
	testMaterialARecipeID     = "CUTA"
	testMaterialAProcessAOID  = "MATAPROCESS001"
	testMaterialAProcessA     = "MATAPROCESS-A"
	testMaterialAProcessAType = "PROCESS"

	testMaterialBID           = "MATERIAL-B"
	testMaterialBGrade        = ""
	testMaterialBType         = "COMPOUND_INGREDIENTS"
	testMaterialBRecipeID     = "CMP001"
	testMaterialBProcessAOID  = "MATBPROCESS001"
	testMaterialBProcessA     = "MATBPROCESS-A"
	testMaterialBProcessAType = "PROCESS"

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

	testDepartmentOID  = "M2100"
	onlyLastProcess    = true
	testProductType    = "RUBBER"
	testProductID      = "BAN"
	testLotNumber      = "xxx-0120"
	testUnit           = "sku"
	testWarehouse      = "WX"
	testLocation       = "001"
	testWarehouse2     = "AX"
	testLocation2      = "002"
	testExpirationDate = time.Date(2021, 8, 11, 0, 0, 0, 0, time.Local)
	resourceID         = "R0147852369"
	testTime           = time.Unix(0, 1646803251)
	testMaterialStatus = int64(resources.MaterialStatus_AVAILABLE)
	testStartDate      = strfmt.DateTime(testTime)
	testInspections    = mcomModels.Inspections{
		{ID: 1,
			Remark: "001"},
		{ID: 2,
			Remark: "002"},
	}
)

func TestProduct_GetProductTypeByDepartmentList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListProductTypes,
			Input: mock.Input{
				Request: mcom.ListProductTypesRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Response: mcom.ListProductTypesReply{
					testProductTypeA,
					testProductTypeB,
					testProductTypeC,
					testProductTypeD,
					testProductTypeE,
				},
			},
		},
		{
			Name: mock.FuncListProductTypes,
			Input: mock.Input{
				Request: mcom.ListProductTypesRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/product/active-product-types/department-oid/{departmentOID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    product.GetProductTypeByDepartmentListParams
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
				params: product.GetProductTypeByDepartmentListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: product.NewGetProductTypeByDepartmentListOK().WithPayload(&product.GetProductTypeByDepartmentListOKBody{
				Data: []*models.ProductsItems0{
					{
						Type: testProductTypeA,
					},
					{
						Type: testProductTypeB,
					},
					{
						Type: testProductTypeC,
					},
					{
						Type: testProductTypeD,
					},
					{
						Type: testProductTypeE,
					},
				},
			}),
		},
		{
			name: "internal error",
			args: args{
				params: product.GetProductTypeByDepartmentListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: product.NewGetProductTypeByDepartmentListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProduct(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetProductTypeByDepartmentList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProductTypeList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewProduct(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetProductTypeByDepartmentList(product.GetProductTypeByDepartmentListParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentOID: testDepartmentOID,
		}, principal).(*product.GetProductTypeByDepartmentListDefault)
		assert.True(ok)
		assert.Equal(product.NewGetProductTypeByDepartmentListDefault(http.StatusForbidden), rep)
	}
}
func TestProduct_GetProductTypeList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListProductTypes,
			Input: mock.Input{
				Request: mcom.ListProductTypesRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Response: mcom.ListProductTypesReply{
					testProductTypeA,
					testProductTypeB,
					testProductTypeC,
					testProductTypeD,
					testProductTypeE,
				},
			},
		},
		{
			Name: mock.FuncListProductTypes,
			Input: mock.Input{
				Request: mcom.ListProductTypesRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/product/active-product-types", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    product.GetProductTypeListParams
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
				params: product.GetProductTypeListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOid: &testDepartmentOID,
				},
				principal: principal,
			},
			want: product.NewGetProductTypeListOK().WithPayload(&product.GetProductTypeListOKBody{
				Data: []*models.ProductsItems0{
					{
						Type: testProductTypeA,
					},
					{
						Type: testProductTypeB,
					},
					{
						Type: testProductTypeC,
					},
					{
						Type: testProductTypeD,
					},
					{
						Type: testProductTypeE,
					},
				},
			}),
		},
		{
			name: "internal error",
			args: args{
				params: product.GetProductTypeListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOid: &testDepartmentOID,
				},
				principal: principal,
			},
			want: product.NewGetProductTypeListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProduct(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetProductTypeList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProductTypeList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewProduct(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetProductTypeList(product.GetProductTypeListParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentOid: &testDepartmentOID,
		}, principal).(*product.GetProductTypeListDefault)
		assert.True(ok)
		assert.Equal(product.NewGetProductTypeListDefault(http.StatusForbidden), rep)
	}
}

func TestProduct_GetProductList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListProductIDs,
			Input: mock.Input{
				Request: mcom.ListProductIDsRequest{
					Type: testProductTypeA,
				},
			},
			Output: mock.Output{
				Response: mcom.ListProductIDsReply{
					testProduct1Parent,
					testProduct1Children1,
					testProduct2Children2,
					testProduct1Children3,
					testProduct2Parent,
					testProduct2Children1,
					testProduct2Children2,
					testProduct2Children3,
				},
			},
		},
		{
			Name: mock.FuncListProductIDs,
			Input: mock.Input{
				Request: mcom.ListProductIDsRequest{
					Type:          testProductTypeA,
					IsLastProcess: true,
				},
			},
			Output: mock.Output{
				Response: mcom.ListProductIDsReply{
					testProduct1Parent,
					testProduct2Parent,
				},
			},
		},
		{
			Name: mock.FuncListProductIDs,
			Input: mock.Input{
				Request: mcom.ListProductIDsRequest{
					Type: testProductTypeA,
				},
			},
			Output: mock.Output{
				Error: &mcomErrors.Error{
					Code: mcomErrors.Code_PRODUCT_ID_NOT_FOUND, // for test case, in fact it won't return this code
				},
			},
		},
		{
			Name: mock.FuncListProductIDs,
			Input: mock.Input{
				Request: mcom.ListProductIDsRequest{
					Type: testProductTypeA,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/product/active-products/product-type/{productType}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    product.GetProductListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "getting all process successfully",
			args: args{
				params: product.GetProductListParams{
					HTTPRequest: httpRequestWithHeader,
					ProductType: testProductTypeA,
				},
				principal: principal,
			},
			want: product.NewGetProductListOK().WithPayload(&product.GetProductListOKBody{
				Data: []string{
					testProduct1Parent,
					testProduct1Children1,
					testProduct2Children2,
					testProduct1Children3,
					testProduct2Parent,
					testProduct2Children1,
					testProduct2Children2,
					testProduct2Children3,
				},
			}),
		},
		{
			name: "getting only last process successfully",
			args: args{
				params: product.GetProductListParams{
					HTTPRequest:   httpRequestWithHeader,
					ProductType:   testProductTypeA,
					IsLastProcess: &onlyLastProcess,
				},
				principal: principal,
			},
			want: product.NewGetProductListOK().WithPayload(&product.GetProductListOKBody{
				Data: []string{
					testProduct1Parent,
					testProduct2Parent,
				},
			}),
		},
		{
			name: "not found",
			args: args{
				params: product.GetProductListParams{
					HTTPRequest: httpRequestWithHeader,
					ProductType: testProductTypeA,
				},
				principal: principal,
			},
			want: product.NewGetProductListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_PRODUCT_ID_NOT_FOUND),
			}),
		},
		{
			name: "internal error",
			args: args{
				params: product.GetProductListParams{
					HTTPRequest: httpRequestWithHeader,
					ProductType: testProductTypeA,
				},
				principal: principal,
			},
			want: product.NewGetProductListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProduct(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetProductList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProductList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewProduct(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetProductList(product.GetProductListParams{
			HTTPRequest: httpRequestWithHeader,
			ProductType: testProductTypeA,
		}, principal).(*product.GetProductListDefault)
		assert.True(ok)
		assert.Equal(product.NewGetProductListDefault(http.StatusForbidden), rep)
	}
}

func TestProduct_GetMaterialResourceInfoByType(t *testing.T) {
	assert := assert.New(t)

	httpRequest := httptest.NewRequest("GET", "/resource/material/info/product-type/{productType}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	var (
		testPage        = int64(1)
		testLimit       = int64(20)
		testOrderName   = "id"
		testDescending  = true
		testErrorPage   = int64(0)
		testErrorLimit  = int64(0)
		testPageRequest = mcom.PaginationRequest{
			PageCount:      uint(testPage),
			ObjectsPerPage: uint(testLimit),
		}
		testOrderRequest = mcom.Order{
			Name:       testOrderName,
			Descending: testDescending,
		}
	)
	scripts := []mock.Script{
		{
			Name: mock.FuncListMaterialResources,
			Input: mock.Input{
				Request: mcom.ListMaterialResourcesRequest{
					ProductType: testProductType,
				}.WithPagination(testPageRequest).WithOrder(testOrderRequest),
			},
			Output: mock.Output{
				Response: mcom.ListMaterialResourcesReply{
					Resources: []mcom.MaterialReply{
						{
							Material: mcom.Material{
								Type:           testProductType,
								ID:             "68000",
								Grade:          "B",
								Status:         resources.MaterialStatus_AVAILABLE,
								Quantity:       decimal.RequireFromString("100"),
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ProductionTime: time.Time{},
								ExpiryTime:     testExpirationDate,
								ResourceID:     resourceID,
								UpdatedAt:      0,
								UpdatedBy:      "",
								CreatedAt:      0,
								CreatedBy:      "",
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
								MinDosage:      decimal.NewFromInt32(10),
								Inspections:    testInspections,
								CarrierID:      "",
								Remark:         "",
							},
							Warehouse: mcom.Warehouse{
								ID:       testWarehouse2,
								Location: testLocation2,
							},
						},
					},
					PaginationReply: mcom.PaginationReply{
						AmountOfData: 10,
					},
				},
			},
		},
		{
			Name: mock.FuncListMaterialResources,
			Input: mock.Input{
				Request: mcom.ListMaterialResourcesRequest{
					ProductType: testProductType,
					ProductID:   testProductID,
					Status:      resources.MaterialStatus_AVAILABLE,
					CreatedAt:   types.ToTimeNano(testTime),
				}.WithPagination(testPageRequest).WithOrder(testOrderRequest),
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_PRODUCT_ID_NOT_FOUND,
					Details: "not found product id",
				},
			},
		},
		{
			Name: mock.FuncListMaterialResources,
			Input: mock.Input{
				Request: mcom.ListMaterialResourcesRequest{
					ProductType: testProductType,
				}.WithPagination(mcom.PaginationRequest{
					PageCount:      uint(testErrorPage),
					ObjectsPerPage: uint(testErrorLimit),
				}).WithOrder(testOrderRequest),
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}
	dm, err := mock.New(scripts)
	assert.NoError(err)

	type args struct {
		params    product.GetMaterialResourceInfoByTypeParams
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
				params: product.GetMaterialResourceInfoByTypeParams{
					HTTPRequest: httpRequest,
					ProductType: testProductType,
					Page:        &testPage,
					Limit:       &testLimit,
					Body: product.GetMaterialResourceInfoByTypeBody{
						OrderRequest: []*product.GetMaterialResourceInfoByTypeParamsBodyOrderRequestItems0{{
							OrderName:  testOrderName,
							Descending: testDescending,
						}},
					},
				},
				principal: principal,
			},
			want: product.NewGetMaterialResourceInfoByTypeOK().WithPayload(
				&product.GetMaterialResourceInfoByTypeOKBody{
					Data: &product.GetMaterialResourceInfoByTypeOKBodyData{
						Items: models.ResourceMaterials{
							&models.ResourceMaterial{
								ID:            "68000",
								CarrierID:     "",
								Remark:        "",
								CreatedAt:     strfmt.DateTime(time.Unix(0, 0)),
								CreatedBy:     "",
								ExpiredDate:   strfmt.DateTime(testExpirationDate),
								Grade:         "B",
								Inspections:   inspectionStructType(testInspections),
								MinimumDosage: "30",
								ProductType:   testProductType,
								Quantity:      "100",
								Unit:          testUnit,
								ResourceID:    resourceID,
								Status:        models.MaterialStatus(1),
								UpdatedAt:     strfmt.DateTime(time.Unix(0, 0)),
								UpdatedBy:     "",
								Warehouse: &models.Warehouse{
									ID:       &testWarehouse,
									Location: &testLocation,
								},
							},
							&models.ResourceMaterial{
								ID:            "68000",
								CarrierID:     "",
								Remark:        "",
								CreatedAt:     strfmt.DateTime(testTime),
								CreatedBy:     userID,
								ExpiredDate:   strfmt.DateTime(testExpirationDate),
								Grade:         "B",
								Inspections:   inspectionStructType(testInspections),
								MinimumDosage: "10",
								ProductType:   testProductType,
								Quantity:      "100",
								Unit:          testUnit,
								ResourceID:    resourceID,
								Status:        models.MaterialStatus(1),
								UpdatedAt:     strfmt.DateTime(testTime),
								UpdatedBy:     userID,
								Warehouse: &models.Warehouse{
									ID:       &testWarehouse2,
									Location: &testLocation2,
								},
							},
						},
						Total: 10,
					},
				}),
		},
		{
			name: "not found product id",
			args: args{
				params: product.GetMaterialResourceInfoByTypeParams{
					HTTPRequest: httpRequest,
					ProductType: testProductType,
					ProductID:   &testProductID,
					Status:      &testMaterialStatus,
					StartDate:   &testStartDate,
					Page:        &testPage,
					Limit:       &testLimit,
					Body: product.GetMaterialResourceInfoByTypeBody{
						OrderRequest: []*product.GetMaterialResourceInfoByTypeParamsBodyOrderRequestItems0{
							{
								OrderName:  testOrderName,
								Descending: testDescending,
							},
						},
					},
				},
				principal: principal,
			},
			want: product.NewGetMaterialResourceInfoByTypeDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_PRODUCT_ID_NOT_FOUND),
				Details: "not found product id",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: product.GetMaterialResourceInfoByTypeParams{
					HTTPRequest: httpRequest,
					ProductType: testProductType,
					Page:        &testErrorPage,
					Limit:       &testErrorLimit,
					Body: product.GetMaterialResourceInfoByTypeBody{
						OrderRequest: []*product.GetMaterialResourceInfoByTypeParamsBodyOrderRequestItems0{
							{
								OrderName:  testOrderName,
								Descending: testDescending,
							},
						},
					},
				},
				principal: principal,
			},
			want: product.NewGetMaterialResourceInfoByTypeDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProduct(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetMaterialResourceInfoByType(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("GetMaterialResourceInfoByType() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := NewProduct(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.GetMaterialResourceInfoByType(product.GetMaterialResourceInfoByTypeParams{
			HTTPRequest: httpRequest,
			ProductType: testProductType,
			Page:        &testPage,
			Limit:       &testLimit,
		}, principal).(*product.GetMaterialResourceInfoByTypeDefault)
		assert.True(ok)
		assert.Equal(product.NewGetMaterialResourceInfoByTypeDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestProduct_substituteBuilder_Build(t *testing.T) {
	assert := assert.New(t)

	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListMultipleSubstitutions,
			Input: mock.Input{
				Request: mcom.ListMultipleSubstitutionsRequest{
					ProductIDs: []mcomModels.ProductID{
						{
							ID:    testMaterialBID,
							Grade: testMaterialBGrade,
						},
					},
				},
			},
			Output: mock.Output{
				Response: mcom.ListMultipleSubstitutionsReply{
					Reply: map[mcomModels.ProductID]mcom.ListSubstitutionsReply{
						{
							ID:    testMaterialBID,
							Grade: testMaterialBGrade,
						}: {
							Substitutions: []mcomModels.Substitution{
								{
									ID:    testMaterialAID,
									Grade: testMaterialAGrade,
								},
							},
						},
					},
				},
			},
		},
		{
			Name: mock.FuncListMultipleSubstitutions,
			Input: mock.Input{
				Request: mcom.ListMultipleSubstitutionsRequest{
					ProductIDs: []mcomModels.ProductID{
						{
							ID:    testMaterialBID,
							Grade: testMaterialBGrade,
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

	defer dm.Close()

	tests := []struct {
		name    string
		want    map[mcomModels.ProductID][]string
		wantErr bool
	}{
		{
			name: "success",
			want: map[mcomModels.ProductID][]string{
				{
					ID:    testMaterialBID,
					Grade: testMaterialBGrade,
				}: {testMaterialAID + testMaterialAGrade},
			},
			wantErr: false,
		},
		{
			name:    "internal error",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := newSubstituteBuilder(context.Background(), dm, []mcomModels.ProductID{
				{
					ID:    testMaterialBID,
					Grade: testMaterialBGrade,
				},
			})
			got, err := sb.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Build() got = %v, want %v", got, tt.want)
			}
		})
	}
	{ // skip empty material name
		sb := newSubstituteBuilder(context.Background(), dm, nil)
		list, err := sb.Build()
		assert.NoError(err)
		assert.Len(list, 0)
	}
}
