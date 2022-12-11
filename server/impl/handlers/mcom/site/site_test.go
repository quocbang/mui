package site

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	mcomSites "gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/site"
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

	testResourceID      = "R0147852369"
	testQuantity        = "200"
	testQuantityDecimal = types.Decimal.NewFromInt32(200)
	testProductID       = "BAN"
	testProductGrade    = "B"

	testResource1ID              = "R1478523690"
	testProductType1             = "RUBBER"
	testResource1Quantity        = "20.5"
	testResource1QuantityDecimal = decimal.NewFromFloat(20.5)

	testResource2ID              = "R2356897410"
	testProductType2             = "RUBBER"
	testResource2Quantity        = "27"
	testResource2QuantityDecimal = decimal.NewFromInt(27)

	testQueueIndex uint16 = 7

	bindTypeContainerAdd        = models.BindType(bindtype.BindType_RESOURCE_BINDING_CONTAINER_ADD)
	bindTypeColQueueAdd         = models.BindType(bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD)
	bindTypeColQueueClear       = models.BindType(bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR)
	siteIndex64                 = int64(siteIndex)
	siteName                    = "LINE"
	siteIndex             int16 = 1
	station                     = "U-F270"
	testBrokenQuantity          = "T_T"
	testMismatchProductID       = "PROD_XX"
	testWarehouseID             = "A"
	testWarehouseLocation       = "11"
	testExpiryTime              = time.Date(10100, 2, 14, 0, 0, 0, 0, time.Local)
	testExpired                 = time.Date(2022, 2, 14, 0, 0, 0, 0, time.Local)

	testStationID = "STATION1"
	testSiteName  = "SITENAME"
)

func mustNewSite(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool,
) service.Site {
	s := NewSite(dm, hasPermission, Config{})
	return s
}

func TestSite_GetSiteMaterialList(t *testing.T) {
	assert := assert.New(t)

	httpRequest := httptest.NewRequest("GET", "/site/material/site-name/{siteName}/site-index/{siteIndex}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	scripts := []mock.Script{
		{
			Name: mock.FuncListSiteMaterials,
			Input: mock.Input{
				Request: mcom.ListSiteMaterialsRequest{
					Station: station,
					Site: mcomModels.SiteID{
						Name:  siteName,
						Index: siteIndex,
					},
				},
			},
			Output: mock.Output{
				Response: mcom.ListSiteMaterialsReply([]mcom.SiteMaterial{
					{
						ResourceID: testResourceID,
						ID:         testProductID,
						Grade:      testProductGrade,
						Quantity:   testQuantityDecimal,
					},
				}),
			},
		},
		{
			Name: mock.FuncListSiteMaterials,
			Input: mock.Input{
				Request: mcom.ListSiteMaterialsRequest{
					Station: station,
					Site: mcomModels.SiteID{
						Name: "CC",
					},
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
			Name: mock.FuncListSiteMaterials,
			Input: mock.Input{
				Request: mcom.ListSiteMaterialsRequest{
					Station: station,
					Site: mcomModels.SiteID{
						Name: "CC",
					},
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
		params    site.GetSiteMaterialListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get site bind material list success",
			args: args{
				params: site.GetSiteMaterialListParams{
					HTTPRequest: httpRequest,
					SiteIndex:   siteIndex64,
					SiteName:    siteName,
					Station:     station,
				},
				principal: principal,
			},
			want: site.NewGetSiteMaterialListOK().WithPayload(&site.GetSiteMaterialListOKBody{
				Data: []*models.BindMaterialData{
					{
						Grade:      models.Grade(testProductGrade),
						ProductID:  testProductID,
						Quantity:   testQuantity,
						ResourceID: testResourceID,
					},
				},
			}),
		},
		{
			name: "not found resource",
			args: args{
				params: site.GetSiteMaterialListParams{
					HTTPRequest: httpRequest,
					SiteName:    "CC",
					Station:     station,
				},
				principal: principal,
			},
			want: site.NewGetSiteMaterialListDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
					Details: "not found resource",
				}),
		},
		{
			name: "internal error",
			args: args{
				params: site.GetSiteMaterialListParams{
					HTTPRequest: httpRequest,
					SiteName:    "CC",
					Station:     station,
				},
				principal: principal,
			},
			want: site.NewGetSiteMaterialListDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Details: testInternalServerError,
				}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.GetSiteMaterialList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSiteMaterialList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.GetSiteMaterialList(site.GetSiteMaterialListParams{
			HTTPRequest: httpRequest,
			SiteName:    siteName,
			SiteIndex:   siteIndex64,
			Station:     station,
		}, principal).(*site.GetSiteMaterialListDefault)
		assert.True(ok)
		assert.Equal(site.NewGetSiteMaterialListDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestSite_AutoBindResource(t *testing.T) {
	httpRequestWithHeader := httptest.NewRequest("POST", "/site/resources/bind/auto", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	assert := assert.New(t)

	var (
		testToolResourceID = "TESTTOOLRESOURCEID"
		testToolID         = "TESTTOOLID"

		testWorkOrderID = "WorkOrder1"
		testProcessName = "ProcessName1"
		testProcessType = "ProcessType1"
		testRecipeID    = "Recipe1"
		testProductID2  = "OIL"
	)

	limitationHandler := func(productID string) error {
		if productID != testProductID {
			return mcomErrors.Error{
				Code:    mcomErrors.Code_PRODUCT_ID_MISMATCH,
				Details: "product id: " + productID,
			}
		}
		return nil
	}

	type args struct {
		params    site.AutoBindSiteResourcesParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{ // material bind to head success.
			name: "material bind to head success",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						WorkOrderID: testWorkOrderID,
						Station:     &station,
						BindType:    &bindTypeColQueueAdd,
						SiteIndex:   &siteIndex64,
						SiteName:    &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ProductType: testProductType1,
								ResourceID:  testResource1ID,
							},
							{
								Quantity:    testResource2Quantity,
								ProductType: testProductType2,
								ResourceID:  testResource2ID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_COLQUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Colqueue: &mcomModels.Colqueue{},
							},
						},
					},
				},
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrderID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							RecipeID: testRecipeID,
							Station:  station,
							Process: mcom.WorkOrderProcess{
								Name: testProcessName,
								Type: testProcessType,
							},
						},
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testRecipeID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								Name: testProcessName,
								Type: testProcessType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations: []string{station},
										Steps: []*mcom.RecipeProcessStep{
											{
												Materials: []*mcom.RecipeMaterial{
													{
														Name:             testProductID,
														RequiredRecipeID: testRecipeID,
													},
													{
														Name:             testProductID,
														RequiredRecipeID: testRecipeID,
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
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
				{
					Name: mock.FuncMaterialResourceBindV2,
					Input: mock.Input{
						Request: mcom.MaterialResourceBindRequestV2{
							Details: []mcom.MaterialBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resources: []mcom.BindMaterialResource{
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											Quantity:    &testResource1QuantityDecimal,
											ResourceID:  testResource1ID,
											ProductType: testProductType1,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource2ID,
											ProductType: testProductType2,
											Quantity:    &testResource2QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
									},
									Option: mcomModels.BindOption{
										Head: true,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // material bind to head success & force bind true.
			name: "material bind to head success & force bind true",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						WorkOrderID: testWorkOrderID,
						Station:     &station,
						BindType:    &bindTypeColQueueAdd,
						SiteIndex:   &siteIndex64,
						SiteName:    &siteName,
						Resources: []*models.BindResource{
							{
								ResourceID: testResource1ID,
							},
							{
								Quantity:    testResource2Quantity,
								ProductType: testProductType2,
								ResourceID:  testResource2ID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
						ForceBind: &site.AutoBindSiteResourcesParamsBodyForceBind{
							Force: true,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_COLQUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Colqueue: &mcomModels.Colqueue{},
							},
						},
					},
				},
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrderID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							RecipeID: testRecipeID,
							Station:  station,
							Process: mcom.WorkOrderProcess{
								Name: testProcessName,
								Type: testProcessType,
							},
						},
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testRecipeID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								Name: testProcessName,
								Type: testProcessType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations: []string{station},
										Steps: []*mcom.RecipeProcessStep{
											{
												Materials: []*mcom.RecipeMaterial{
													{
														Name:             testProductID,
														RequiredRecipeID: testRecipeID,
													},
													{
														Name:             testProductID,
														RequiredRecipeID: testRecipeID,
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
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: "",
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								nil,
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
				{
					Name: mock.FuncMaterialResourceBindV2,
					Input: mock.Input{
						Request: mcom.MaterialResourceBindRequestV2{
							Details: []mcom.MaterialBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resources: []mcom.BindMaterialResource{
										{
											ResourceID:  testResource1ID,
											ProductType: "",
										},
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource2ID,
											ProductType: testProductType2,
											Quantity:    &testResource2QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
									},
									Option: mcomModels.BindOption{
										Head: true,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // material clear success.
			name: "material clear success",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						WorkOrderID: testWorkOrderID,
						Station:     &station,
						BindType:    &bindTypeColQueueClear,
						SiteIndex:   &siteIndex64,
						SiteName:    &siteName,
						Resources:   []*models.BindResource{},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_QUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Queue: &mcomModels.Queue{},
							},
						},
					},
				},
				{
					Name: mock.FuncMaterialResourceBindV2,
					Input: mock.Input{
						Request: mcom.MaterialResourceBindRequestV2{
							Details: []mcom.MaterialBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resources: []mcom.BindMaterialResource{},
									Option:    mcomModels.BindOption{},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // material bind to head success without work order.
			name: "material bind to head success without work order",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeColQueueAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ProductType: testProductType1,
								ResourceID:  testResource1ID,
							},
							{
								Quantity:    testResource2Quantity,
								ProductType: testProductType2,
								ResourceID:  testResource2ID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_COLQUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Colqueue: &mcomModels.Colqueue{},
							},
						},
					},
				},
				{
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
				{
					Name: mock.FuncMaterialResourceBindV2,
					Input: mock.Input{
						Request: mcom.MaterialResourceBindRequestV2{
							Details: []mcom.MaterialBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resources: []mcom.BindMaterialResource{
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											Quantity:    &testResource1QuantityDecimal,
											ResourceID:  testResource1ID,
											ProductType: testProductType1,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource2ID,
											ProductType: testProductType2,
											Quantity:    &testResource2QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
									},
									Option: mcomModels.BindOption{
										Head: true,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // material bind to head success without work order& quantity empty.
			name: "material bind to head success without work order& quantity empty",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeColQueueAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								ProductType: testProductType1,
								ResourceID:  testResource1ID,
							},
							{
								Quantity:    testResource2Quantity,
								ProductType: testProductType2,
								ResourceID:  testResource2ID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_COLQUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Colqueue: &mcomModels.Colqueue{},
							},
						},
					},
				},
				{
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("2"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
				{
					Name: mock.FuncMaterialResourceBindV2,
					Input: mock.Input{
						Request: mcom.MaterialResourceBindRequestV2{
							Details: []mcom.MaterialBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resources: []mcom.BindMaterialResource{
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource1ID,
											ProductType: testProductType1,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource2ID,
											ProductType: testProductType2,
											Quantity:    &testResource2QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
									},
									Option: mcomModels.BindOption{
										Head: true,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // material bind to tail success.
			name: "material bind to tail success",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						WorkOrderID: testWorkOrderID,
						Station:     &station,
						BindType:    &bindTypeColQueueAdd,
						SiteIndex:   &siteIndex64,
						SiteName:    &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ProductType: testProductType1,
								ResourceID:  testResource1ID,
							},
							{
								Quantity:    testResource2Quantity,
								ProductType: testProductType2,
								ResourceID:  testResource2ID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Tail: true,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_COLQUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Colqueue: &mcomModels.Colqueue{},
							},
						},
					},
				},
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrderID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							RecipeID: testRecipeID,
							Station:  station,
							Process: mcom.WorkOrderProcess{
								Name: testProcessName,
								Type: testProcessType,
							},
						},
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testRecipeID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								Name: testProcessName,
								Type: testProcessType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations: []string{station},
										Steps: []*mcom.RecipeProcessStep{
											{
												Materials: []*mcom.RecipeMaterial{
													{
														Name:             testProductID,
														RequiredRecipeID: testRecipeID,
													},
													{
														Name:             testProductID,
														RequiredRecipeID: testRecipeID,
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
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
				{
					Name: mock.FuncMaterialResourceBindV2,
					Input: mock.Input{
						Request: mcom.MaterialResourceBindRequestV2{
							Details: []mcom.MaterialBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resources: []mcom.BindMaterialResource{
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource1ID,
											ProductType: testProductType1,
											Quantity:    &testResource1QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource2ID,
											ProductType: testProductType2,
											Quantity:    &testResource2QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
									},
									Option: mcomModels.BindOption{
										Tail: true,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // material bind to index success.
			name: "material bind to index success",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						WorkOrderID: testWorkOrderID,
						Station:     &station,
						BindType:    &bindTypeColQueueAdd,
						SiteIndex:   &siteIndex64,
						SiteName:    &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ProductType: testProductType1,
								ResourceID:  testResource1ID,
							},
							{
								Quantity:    testResource2Quantity,
								ProductType: testProductType2,
								ResourceID:  testResource2ID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Index: 7,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_COLQUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Colqueue: &mcomModels.Colqueue{},
							},
						},
					},
				},
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrderID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							RecipeID: testRecipeID,
							Station:  station,
							Process: mcom.WorkOrderProcess{
								Name: testProcessName,
								Type: testProcessType,
							},
						},
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testRecipeID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								Name: testProcessName,
								Type: testProcessType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations: []string{station},
										Steps: []*mcom.RecipeProcessStep{
											{
												Materials: []*mcom.RecipeMaterial{
													{
														Name:             testProductID,
														RequiredRecipeID: testRecipeID,
													},
													{
														Name:             testProductID,
														RequiredRecipeID: testRecipeID,
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
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
				{
					Name: mock.FuncMaterialResourceBindV2,
					Input: mock.Input{
						Request: mcom.MaterialResourceBindRequestV2{
							Details: []mcom.MaterialBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resources: []mcom.BindMaterialResource{
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource1ID,
											ProductType: testProductType1,
											Quantity:    &testResource1QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource2ID,
											ProductType: testProductType2,
											Quantity:    &testResource2QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
									},
									Option: mcomModels.BindOption{
										QueueIndex: &testQueueIndex,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // tool bind success without work order.
			name: "tool bind success without work order",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeColQueueAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ProductType: testToolID,
								ResourceID:  testToolResourceID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
						ResourceType: 1,
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_QUEUE,
								SubType:      mcomSites.SubType_TOOL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Queue: &mcomModels.Queue{},
							},
						},
					},
				},
				{
					Name: mock.FuncGetToolResource,
					Input: mock.Input{
						Request: mcom.GetToolResourceRequest{
							ResourceID: testToolResourceID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetToolResourceReply{
							ToolID: testToolID,
						},
					},
				},
				{
					Name: mock.FuncToolResourceBindV2,
					Input: mock.Input{
						Request: mcom.ToolResourceBindRequestV2{
							Details: []mcom.ToolBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resource: mcom.ToolResource{
										ResourceID: testToolResourceID,
										ToolID:     testToolID,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // tool bind success & tool id not found & force bind true.
			name: "tool bind success & tool id not found & force bind true",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeColQueueAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:   testResource1Quantity,
								ResourceID: testToolResourceID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
						ResourceType: 1,
						ForceBind: &site.AutoBindSiteResourcesParamsBodyForceBind{
							Force: true,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_QUEUE,
								SubType:      mcomSites.SubType_TOOL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Queue: &mcomModels.Queue{},
							},
						},
					},
				},
				{
					Name: mock.FuncGetToolResource,
					Input: mock.Input{
						Request: mcom.GetToolResourceRequest{
							ResourceID: testToolResourceID,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_RESOURCE_NOT_FOUND,
						},
					},
				},
				{
					Name: mock.FuncToolResourceBindV2,
					Input: mock.Input{
						Request: mcom.ToolResourceBindRequestV2{
							Details: []mcom.ToolBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resource: mcom.ToolResource{
										ResourceID: testToolResourceID,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // tool bind success & tool in recipe & force bind false.
			name: "tool bind success & tool in recipe & force bind false",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						WorkOrderID: testWorkOrderID,
						Station:     &station,
						BindType:    &bindTypeColQueueAdd,
						SiteIndex:   &siteIndex64,
						SiteName:    &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:   testResource1Quantity,
								ResourceID: testToolResourceID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
						ResourceType: 1,
						ForceBind: &site.AutoBindSiteResourcesParamsBodyForceBind{
							Force: false,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_QUEUE,
								SubType:      mcomSites.SubType_TOOL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Queue: &mcomModels.Queue{},
							},
						},
					},
				},
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrderID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							RecipeID: testRecipeID,
							Station:  station,
							Process: mcom.WorkOrderProcess{
								Name: testProcessName,
								Type: testProcessType,
							},
						},
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testRecipeID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								Name: testProcessName,
								Type: testProcessType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations: []string{station},
										Tools: []*mcom.RecipeTool{
											{
												ID:       testToolID,
												Required: true,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Name: mock.FuncGetToolResource,
					Input: mock.Input{
						Request: mcom.GetToolResourceRequest{
							ResourceID: testToolResourceID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetToolResourceReply{
							ToolID: testToolID,
						},
					},
				},
				{
					Name: mock.FuncToolResourceBindV2,
					Input: mock.Input{
						Request: mcom.ToolResourceBindRequestV2{
							Details: []mcom.ToolBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resource: mcom.ToolResource{
										ResourceID: testToolResourceID,
										ToolID:     testToolID,
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // tool clear success.
			name: "tool clear success",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeColQueueClear,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
						ResourceType: 1,
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesOK(),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_QUEUE,
								SubType:      mcomSites.SubType_TOOL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Container: &mcomModels.Container{},
							},
						},
					},
				},
				{
					Name: mock.FuncToolResourceBindV2,
					Input: mock.Input{
						Request: mcom.ToolResourceBindRequestV2{
							Details: []mcom.ToolBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resource: mcom.ToolResource{},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{ // internal server error.
			name: "internal server error",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeContainerAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ProductType: testProductType1,
								ResourceID:  testResource1ID,
							},
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Details: testInternalServerError,
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_CONTAINER,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Container: &mcomModels.Container{},
							},
						},
					},
				},
				{
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
				{
					Name: mock.FuncMaterialResourceBindV2,
					Input: mock.Input{
						Request: mcom.MaterialResourceBindRequestV2{
							Details: []mcom.MaterialBindRequestDetailV2{
								{
									Type: bindtype.BindType_RESOURCE_BINDING_CONTAINER_ADD,
									Site: mcomModels.UniqueSite{
										Station: station,
										SiteID: mcomModels.SiteID{
											Name:  siteName,
											Index: siteIndex,
										},
									},
									Resources: []mcom.BindMaterialResource{
										{
											Material: mcomModels.Material{
												ID:    testProductID,
												Grade: testProductGrade,
											},
											ResourceID:  testResource1ID,
											ProductType: testProductType1,
											Quantity:    &testResource1QuantityDecimal,
											Warehouse: mcom.Warehouse{
												ID:       testWarehouseID,
												Location: testWarehouseLocation,
											},
											ExpiryTime: types.ToTimeNano(testExpiryTime),
											Status:     resources.MaterialStatus_AVAILABLE,
										},
									},
									Option: mcomModels.BindOption{},
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
		{ // invalid number.
			name: "invalid number",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeContainerAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ResourceID:  testResource1ID,
								ProductType: testProductType1,
							},
							{
								Quantity:    testBrokenQuantity,
								ResourceID:  testResource2ID,
								ProductType: testProductType2,
							},
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code:    int64(mcomErrors.Code_INVALID_NUMBER),
					Details: "invalid_number=" + testBrokenQuantity,
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_CONTAINER,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Container: &mcomModels.Container{},
							},
						},
					},
				},
				{
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
			},
		},
		{ // mismatch limitation.
			name: "mismatch limitation",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeContainerAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ResourceID:  testResource1ID,
								ProductType: testProductType1,
							},
							{
								Quantity:    testResource2Quantity,
								ResourceID:  testResource2ID,
								ProductType: testProductType2,
							},
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code:    int64(mcomErrors.Code_PRODUCT_ID_MISMATCH),
					Details: "product id: " + testMismatchProductID,
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_CONTAINER,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Container: &mcomModels.Container{},
							},
						},
					},
				},
				{
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testMismatchProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
			},
		},
		{ // material resource quantity shortage.
			name: "material resource quantity shortage",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeContainerAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ResourceID:  testResource1ID,
								ProductType: testProductType1,
							},
							{
								Quantity:    testResource2Quantity,
								ResourceID:  testResource2ID,
								ProductType: testProductType2,
							},
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_MATERIAL_SHORTAGE),
					Details: "not enough resource quantity to bind, index=1, storage=10, demand=27",
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_CONTAINER,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Container: &mcomModels.Container{},
							},
						},
					},
				},
				{
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("10"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
			},
		},
		{ // material resource resource expired.
			name: "material resource resource expired",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeContainerAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ResourceID:  testResource1ID,
								ProductType: testProductType1,
							},
							{
								Quantity:    testResource2Quantity,
								ResourceID:  testResource2ID,
								ProductType: testProductType2,
							},
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_EXPIRED),
					Details: "resource expired, index=1",
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_CONTAINER,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Container: &mcomModels.Container{},
							},
						},
					},
				},
				{
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("100"),
										ExpiryTime: testExpired,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
			},
		},
		{ // material resource not found.
			name: "material resource not found",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeContainerAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ResourceID:  testResource1ID,
								ProductType: testProductType1,
							},
							{
								Quantity:    testResource2Quantity,
								ResourceID:  testResource2ID,
								ProductType: testProductType2,
							},
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
					Details: testResource2ID,
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_CONTAINER,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Container: &mcomModels.Container{},
							},
						},
					},
				},
				{
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								nil,
							},
						},
					},
				},
			},
		},
		{ // material resource not in recipe.
			name: "material resource not in recipe",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						WorkOrderID: testWorkOrderID,
						Station:     &station,
						BindType:    &bindTypeColQueueAdd,
						SiteIndex:   &siteIndex64,
						SiteName:    &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ProductType: testProductType1,
								ResourceID:  testResource1ID,
							},
							{
								Quantity:    testResource2Quantity,
								ProductType: testProductType2,
								ResourceID:  testResource2ID,
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code:    int64(mcomErrors.Code_RESOURCE_WORKORDER_RESOURCE_UNEXPECTED),
					Details: "material resource not in recipe",
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_COLQUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
							Content: mcomModels.SiteContent{
								Colqueue: &mcomModels.Colqueue{},
							},
						},
					},
				},
				{
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrderID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetWorkOrderReply{
							RecipeID: testRecipeID,
							Station:  station,
							Process: mcom.WorkOrderProcess{
								Name: testProcessName,
								Type: testProcessType,
							},
						},
					},
				},
				{
					Name: mock.FuncGetProcessDefinition,
					Input: mock.Input{
						Request: mcom.GetProcessDefinitionRequest{
							RecipeID:    testRecipeID,
							ProcessName: testProcessName,
							ProcessType: testProcessType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetProcessDefinitionReply{
							ProcessDefinition: mcom.ProcessDefinition{
								Name: testProcessName,
								Type: testProcessType,
								Configs: []*mcom.RecipeProcessConfig{
									{
										Stations: []string{station},
										Steps: []*mcom.RecipeProcessStep{
											{
												Materials: []*mcom.RecipeMaterial{
													{
														Name:             testProductID2,
														RequiredRecipeID: testRecipeID,
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
					Name: mock.FuncListMaterialResourceIdentities,
					Input: mock.Input{
						Request: mcom.ListMaterialResourceIdentitiesRequest{
							Details: []mcom.GetMaterialResourceIdentityRequest{
								{
									ResourceID:  testResource1ID,
									ProductType: testProductType1,
								},
								{
									ResourceID:  testResource2ID,
									ProductType: testProductType2,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.ListMaterialResourceIdentitiesReply{
							Replies: []*mcom.MaterialReply{
								{
									Material: mcom.Material{
										Type:       testProductType1,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource1ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
								{
									Material: mcom.Material{
										Type:       testProductType2,
										ID:         testProductID,
										Grade:      testProductGrade,
										Status:     resources.MaterialStatus_AVAILABLE,
										Quantity:   decimal.RequireFromString("200"),
										ExpiryTime: testExpiryTime,
										ResourceID: testResource2ID,
									},
									Warehouse: mcom.Warehouse{
										ID:       testWarehouseID,
										Location: testWarehouseLocation,
									},
								},
							},
						},
					},
				},
			},
		},
		{ // insufficient request.
			name: "insufficient request",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeContainerAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ResourceID:  testResource1ID,
								ProductType: testProductType1,
							},
							{
								Quantity:    testResource2Quantity,
								ResourceID:  testResource2ID,
								ProductType: testProductType2,
							},
						},
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
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
		{ // tool resource not found.
			name: "tool resource not found",
			args: args{
				params: site.AutoBindSiteResourcesParams{
					HTTPRequest: httpRequestWithHeader,
					Body: site.AutoBindSiteResourcesBody{
						Station:   &station,
						BindType:  &bindTypeColQueueAdd,
						SiteIndex: &siteIndex64,
						SiteName:  &siteName,
						Resources: []*models.BindResource{
							{
								Quantity:    testResource1Quantity,
								ProductType: testToolID,
								ResourceID:  "ERRORID",
							},
						},
						QueueOption: &site.AutoBindSiteResourcesParamsBodyQueueOption{
							Head: true,
						},
						ResourceType: 1,
					},
				},
				principal: principal,
			},
			want: site.NewAutoBindSiteResourcesDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code: int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: station,
							SiteName:  siteName,
							SiteIndex: siteIndex,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:              siteName,
							Index:             siteIndex,
							AdminDepartmentID: "xxx",
							Attributes: mcom.SiteAttributes{
								Type:         mcomSites.Type_COLQUEUE,
								SubType:      mcomSites.SubType_MATERIAL,
								LimitHandler: limitationHandler,
							},
						},
					},
				},
				{
					Name: mock.FuncGetToolResource,
					Input: mock.Input{
						Request: mcom.GetToolResourceRequest{
							ResourceID: "ERRORID",
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_RESOURCE_NOT_FOUND,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, err := mock.New(tt.script)
			assert.NoErrorf(err, tt.name)

			s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.AutoBindResource(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("AutoBindResource() = %v, want %v", got, tt.want)
			}
			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New(nil)
		s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.AutoBindResource(site.AutoBindSiteResourcesParams{
			HTTPRequest: httpRequestWithHeader,
			Body: site.AutoBindSiteResourcesBody{
				Station:   &station,
				BindType:  &bindTypeContainerAdd,
				SiteIndex: &siteIndex64,
				SiteName:  &siteName,
				Resources: []*models.BindResource{
					{
						Quantity:   testResource1Quantity,
						ResourceID: testResource1ID,
					},
				},
			},
		}, principal).(*site.AutoBindSiteResourcesDefault)
		assert.True(ok)
		assert.Equal(site.NewAutoBindSiteResourcesDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestSite_ListSubType(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name:  mock.FuncListSiteSubType,
			Input: mock.Input{},
			Output: mock.Output{
				Response: mcom.ListSiteSubTypeReply([]mcom.SiteSubType{
					{
						Name:  mcomSites.SubType_OPERATOR.String(),
						Value: mcomSites.SubType_OPERATOR,
					},
					{
						Name:  mcomSites.SubType_MATERIAL.String(),
						Value: mcomSites.SubType_MATERIAL,
					},
					{
						Name:  mcomSites.SubType_TOOL.String(),
						Value: mcomSites.SubType_TOOL,
					},
				}),
			},
		},
		{
			Name:  mock.FuncListSiteSubType,
			Input: mock.Input{},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "fake error message",
				},
			},
		},
		{
			Name:  mock.FuncListSiteSubType,
			Input: mock.Input{},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/site/sub-type-list", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    site.GetSiteSubTypeListParams
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
				params: site.GetSiteSubTypeListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: site.NewGetSiteSubTypeListOK().WithPayload(&site.GetSiteSubTypeListOKBody{
				Data: []*site.GetSiteSubTypeListOKBodyDataItems0{
					{
						Name: mcomSites.SubType_OPERATOR.String(),
						ID:   models.SiteSubType(mcomSites.SubType_OPERATOR),
					},
					{
						Name: mcomSites.SubType_MATERIAL.String(),
						ID:   models.SiteSubType(mcomSites.SubType_MATERIAL),
					},
					{
						Name: mcomSites.SubType_TOOL.String(),
						ID:   models.SiteSubType(mcomSites.SubType_TOOL),
					},
				},
			}),
		},
		{
			name: "bad request",
			args: args{
				params: site.GetSiteSubTypeListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: site.NewGetSiteSubTypeListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "fake error message",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: site.GetSiteSubTypeListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: site.NewGetSiteSubTypeListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.ListSubType(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListSubType() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.ListSubType(site.GetSiteSubTypeListParams{
			HTTPRequest: httpRequestWithHeader,
		}, principal).(*site.GetSiteSubTypeListDefault)
		assert.True(ok)
		assert.Equal(site.NewGetSiteSubTypeListDefault(http.StatusForbidden), rep)
	}
}

func TestSite_ListType(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name:  mock.FuncListSiteType,
			Input: mock.Input{},
			Output: mock.Output{
				Response: mcom.ListSiteTypeReply([]mcom.SiteType{
					{
						Name:  mcomSites.Type_SLOT.String(),
						Value: mcomSites.Type_SLOT,
					},
					{
						Name:  mcomSites.Type_CONTAINER.String(),
						Value: mcomSites.Type_CONTAINER,
					},
					{
						Name:  mcomSites.Type_COLLECTION.String(),
						Value: mcomSites.Type_COLLECTION,
					},
					{
						Name:  mcomSites.Type_QUEUE.String(),
						Value: mcomSites.Type_QUEUE,
					},
					{
						Name:  mcomSites.Type_COLQUEUE.String(),
						Value: mcomSites.Type_COLQUEUE,
					},
				}),
			},
		},
		{
			Name:  mock.FuncListSiteType,
			Input: mock.Input{},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "fake error message",
				},
			},
		},
		{
			Name:  mock.FuncListSiteType,
			Input: mock.Input{},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/site/type-list", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    site.GetSiteTypeListParams
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
				params: site.GetSiteTypeListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: site.NewGetSiteTypeListOK().WithPayload(&site.GetSiteTypeListOKBody{
				Data: []*site.GetSiteTypeListOKBodyDataItems0{
					{
						Name: mcomSites.Type_SLOT.String(),
						ID:   models.SiteType(mcomSites.Type_SLOT),
					},
					{
						Name: mcomSites.Type_CONTAINER.String(),
						ID:   models.SiteType(mcomSites.Type_CONTAINER),
					},
					{
						Name: mcomSites.Type_COLLECTION.String(),
						ID:   models.SiteType(mcomSites.Type_COLLECTION),
					},
					{
						Name: mcomSites.Type_QUEUE.String(),
						ID:   models.SiteType(mcomSites.Type_QUEUE),
					},
					{
						Name: mcomSites.Type_COLQUEUE.String(),
						ID:   models.SiteType(mcomSites.Type_COLQUEUE),
					},
				},
			}),
		},
		{
			name: "bad request",
			args: args{
				params: site.GetSiteTypeListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: site.NewGetSiteTypeListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "fake error message",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: site.GetSiteTypeListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: site.NewGetSiteTypeListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.ListType(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListType() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.ListType(site.GetSiteTypeListParams{
			HTTPRequest: httpRequestWithHeader,
		}, principal).(*site.GetSiteTypeListDefault)
		assert.True(ok)
		assert.Equal(site.NewGetSiteTypeListDefault(http.StatusForbidden), rep)
	}
}

func TestStation_GetStationOperator(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("POST", "/station/{stationID}/operator", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    site.GetStationOperatorParams
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
				params: site.GetStationOperatorParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationID,
					Body: site.GetStationOperatorBody{
						Site: &models.SiteInfo{
							StationID: testStationID,
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
				},
				principal: principal,
			},
			want: site.NewGetStationOperatorOK().WithPayload(&site.GetStationOperatorOKBody{
				Data: &site.GetStationOperatorOKBodyData{
					OperatorID: userID,
				},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: testStationID,
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:  testSiteName,
							Index: 0,
							Attributes: mcom.SiteAttributes{
								Type:    mcomSites.Type_SLOT,
								SubType: mcomSites.SubType_OPERATOR,
							},
							Content: mcomModels.SiteContent{
								Slot: &mcomModels.Slot{
									Operator: &mcomModels.OperatorSite{
										EmployeeID: userID,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "success without operator",
			args: args{
				params: site.GetStationOperatorParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationID,
					Body: site.GetStationOperatorBody{
						Site: &models.SiteInfo{
							StationID: testStationID,
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
				},
				principal: principal,
			},
			want: site.NewGetStationOperatorOK().WithPayload(&site.GetStationOperatorOKBody{
				Data: &site.GetStationOperatorOKBodyData{
					OperatorID: "",
				},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: testStationID,
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:  testSiteName,
							Index: 0,
							Attributes: mcom.SiteAttributes{
								Type:    mcomSites.Type_SLOT,
								SubType: mcomSites.SubType_OPERATOR,
							},
							Content: mcomModels.SiteContent{
								Slot: &mcomModels.Slot{
									Operator: &mcomModels.OperatorSite{
										EmployeeID: "",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "station site not found",
			args: args{
				params: site.GetStationOperatorParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   "not found",
					Body: site.GetStationOperatorBody{
						Site: &models.SiteInfo{
							StationID: "not found",
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
				},
				principal: principal,
			},
			want: site.NewGetStationOperatorOK().WithPayload(&site.GetStationOperatorOKBody{
				Data: &site.GetStationOperatorOKBodyData{
					OperatorID: "",
				},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: "not found",
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_STATION_SITE_NOT_FOUND,
						},
					},
				},
			},
		},
		{
			name: "station site sub type not mismatch",
			args: args{
				params: site.GetStationOperatorParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationID,
					Body: site.GetStationOperatorBody{
						Site: &models.SiteInfo{
							StationID: testStationID,
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
				},
				principal: principal,
			},
			want: site.NewGetStationOperatorDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_STATION_SITE_SUB_TYPE_MISMATCH),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: testStationID,
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
					Output: mock.Output{
						Response: mcom.GetSiteReply{
							Name:  testSiteName,
							Index: 0,
							Attributes: mcom.SiteAttributes{
								Type:    mcomSites.Type_SLOT,
								SubType: mcomSites.SubType_MATERIAL,
							},
							Content: mcomModels.SiteContent{
								Slot: &mcomModels.Slot{
									Material: &mcomModels.MaterialSite{},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "insufficient request",
			args: args{
				params: site.GetStationOperatorParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   "",
					Body: site.GetStationOperatorBody{
						Site: &models.SiteInfo{
							StationID: "",
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
				},
				principal: principal,
			},
			want: site.NewGetStationOperatorDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: "",
							SiteName:  testSiteName,
							SiteIndex: 0,
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
			name: "internal error",
			args: args{
				params: site.GetStationOperatorParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationID,
					Body: site.GetStationOperatorBody{
						Site: &models.SiteInfo{
							StationID: testStationID,
							SiteName:  testSiteName,
							SiteIndex: 0,
						},
					},
				},
				principal: principal,
			},
			want: site.NewGetStationOperatorDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetSite,
					Input: mock.Input{
						Request: mcom.GetSiteRequest{
							StationID: testStationID,
							SiteName:  testSiteName,
							SiteIndex: 0,
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

			s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.GetStationOperator(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("GetStationOperator() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}

	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := mustNewSite(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.GetStationOperator(site.GetStationOperatorParams{
			HTTPRequest: httpRequestWithHeader,
			StationID:   testStationID,
			Body: site.GetStationOperatorBody{
				Site: &models.SiteInfo{
					StationID: testStationID,
					SiteName:  testSiteName,
					SiteIndex: 0,
				},
			},
		}, principal).(*site.GetStationOperatorDefault)
		assert.True(ok)
		assert.Equal(site.NewGetStationOperatorDefault(http.StatusForbidden), rep)
	}
}
