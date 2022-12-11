package ui

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-openapi/runtime/middleware"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/ui"
)

const (
	userID = "tester"

	testInternalServerError = "internal error"
	remarkNone              = 0

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
)

var (
	principal = &models.Principal{
		ID: userID,
		Roles: []models.Role{
			models.Role(mcomRoles.Role_ADMINISTRATOR),
			models.Role(mcomRoles.Role_LEADER),
		},
	}

	testDepartmentOID = "M2100"

	testStationA              = "STATION-A"
	testWorkOrder1ProductType = "RUBBER"
	testQuantity              = decimal.Decimal(decimal.NewFromFloat(79.21))
)

func TestUI_GetProductGroupList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListProductGroups,
			Input: mock.Input{
				Request: mcom.ListProductGroupsRequest{
					DepartmentID: testDepartmentOID,
					Type:         testProductTypeA,
				},
			},
			Output: mock.Output{
				Response: mcom.ListProductGroupsReply{
					Products: []mcom.ProductGroup{
						{
							ID: testProduct1Parent,
							Children: []string{
								testProduct1Children1,
								testProduct1Children2,
								testProduct1Children3,
							},
						},
						{
							ID: testProduct2Parent,
							Children: []string{
								testProduct2Children1,
								testProduct2Children2,
								testProduct2Children3,
							},
						},
					},
				},
			},
		},
		{
			Name: mock.FuncListProductGroups,
			Input: mock.Input{
				Request: mcom.ListProductGroupsRequest{
					DepartmentID: testDepartmentOID,
					Type:         testProductTypeA,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/product/groups/department-oid/{departmentOID}/product-type/{productType}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    ui.GetProductGroupListParams
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
				params: ui.GetProductGroupListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					ProductType:   testProductTypeA,
				},
				principal: principal,
			},
			want: ui.NewGetProductGroupListOK().WithPayload(&ui.GetProductGroupListOKBody{
				Data: []*ui.GetProductGroupListOKBodyDataItems0{
					{
						Parent: testProduct1Parent,
						Children: []string{
							testProduct1Children1,
							testProduct1Children2,
							testProduct1Children3,
						},
					},
					{
						Parent: testProduct2Parent,
						Children: []string{
							testProduct2Children1,
							testProduct2Children2,
							testProduct2Children3,
						},
					},
				},
			}),
		},
		{
			name: "internal error",
			args: args{
				params: ui.GetProductGroupListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					ProductType:   testProductTypeA,
				},
				principal: principal,
			},
			want: ui.NewGetProductGroupListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewUI(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := p.GetProductGroupList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProductGroupList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		p := NewUI(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetProductGroupList(ui.GetProductGroupListParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentOID: testDepartmentOID,
			ProductType:   testProductTypeA,
		}, principal).(*ui.GetProductGroupListDefault)
		assert.True(ok)
		assert.Equal(ui.NewGetProductGroupListDefault(http.StatusForbidden), rep)
	}
}

func TestUI_SetStationConfig(t *testing.T) {

	httpRequestWithHeader := httptest.NewRequest("POST", "/production-flow/config/station/{stationID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	assert := assert.New(t)

	type args struct {
		params    ui.SetStationConfigParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success feed & collect together",
			args: args{
				params: ui.SetStationConfigParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationA,
					Body: ui.SetStationConfigBody{
						StationConfig: &models.StationConfig{
							SeparateMode: false,
							Feed: &models.StationConfigFeed{
								ProductType:      []string{testWorkOrder1ProductType},
								MaterialResource: handlerUtils.NewBoolean(true),
							},
							Collect: &models.StationConfigCollect{
								Resource:        handlerUtils.NewBoolean(true),
								CarrierResource: handlerUtils.NewBoolean(true),
								Quantity: &models.StationConfigCollectQuantity{
									Type:  int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
									Value: testQuantity.String(),
								},
							},
						},
					},
				},
				principal: principal,
			},
			want: ui.NewSetStationConfigOK(),
			script: []mock.Script{
				{
					Name: mock.FuncSetStationConfiguration,
					Input: mock.Input{
						Request: mcom.SetStationConfigurationRequest{
							StationID:           testStationA,
							SplitFeedAndCollect: false,
							Feed: mcom.StationFeedConfigs{
								ProductTypes:         []string{testWorkOrder1ProductType},
								NeedMaterialResource: true,
								QuantitySource:       stations.FeedQuantitySource_FROM_RECIPE,
							},
							Collect: mcom.StationCollectConfigs{
								NeedCollectResource: true,
								NeedCarrierResource: true,
								QuantitySource:      stations.CollectQuantitySource_FROM_STATION_CONFIGS,
								DefaultQuantity:     testQuantity,
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "success feed & collect separate",
			args: args{
				params: ui.SetStationConfigParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationA,
					Body: ui.SetStationConfigBody{
						StationConfig: &models.StationConfig{
							SeparateMode: true,
							Feed: &models.StationConfigFeed{
								ProductType:      []string{testWorkOrder1ProductType},
								MaterialResource: handlerUtils.NewBoolean(true),
								StandardQuantity: int64(stations.FeedQuantitySource_FROM_RECIPE),
								OperatorSites: []*models.SiteInfo{
									{
										StationID: testStationA,
										SiteName:  "Feed",
										SiteIndex: 0,
									},
								},
							},
							Collect: &models.StationConfigCollect{
								Resource:        handlerUtils.NewBoolean(true),
								CarrierResource: handlerUtils.NewBoolean(true),
								OperatorSites: []*models.SiteInfo{
									{
										StationID: testStationA,
										SiteName:  "Collect",
										SiteIndex: 0,
									},
								},
								Quantity: &models.StationConfigCollectQuantity{
									Type:  int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
									Value: testQuantity.String(),
								},
							},
						},
					},
				},
				principal: principal,
			},
			want: ui.NewSetStationConfigOK(),
			script: []mock.Script{
				{
					Name: mock.FuncSetStationConfiguration,
					Input: mock.Input{
						Request: mcom.SetStationConfigurationRequest{
							StationID:           testStationA,
							SplitFeedAndCollect: true,
							Feed: mcom.StationFeedConfigs{
								ProductTypes:         []string{testWorkOrder1ProductType},
								NeedMaterialResource: true,
								QuantitySource:       stations.FeedQuantitySource_FROM_RECIPE,
								OperatorSites: []mcomModels.UniqueSite{
									{
										Station: testStationA,
										SiteID: mcomModels.SiteID{
											Name:  "Feed",
											Index: 0,
										},
									},
								},
							},

							Collect: mcom.StationCollectConfigs{
								NeedCarrierResource: true,
								NeedCollectResource: true,
								QuantitySource:      stations.CollectQuantitySource_FROM_STATION_CONFIGS,
								DefaultQuantity:     decimal.NewFromFloat(79.21),
								OperatorSites: []mcomModels.UniqueSite{
									{
										Station: testStationA,
										SiteID: mcomModels.SiteID{
											Name:  "Collect",
											Index: 0,
										},
									},
								},
							},
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "insufficient request",
			args: args{
				params: ui.SetStationConfigParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationA,
					Body: ui.SetStationConfigBody{
						StationConfig: &models.StationConfig{
							SeparateMode: false,
							Feed: &models.StationConfigFeed{
								ProductType:      []string{},
								MaterialResource: handlerUtils.NewBoolean(true),
							},
							Collect: &models.StationConfigCollect{
								Resource: handlerUtils.NewBoolean(true),

								CarrierResource: handlerUtils.NewBoolean(true),
								Quantity: &models.StationConfigCollectQuantity{
									Type:  int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
									Value: testQuantity.String(),
								},
							},
						},
					},
				},
				principal: principal,
			},
			want: ui.NewSetStationConfigDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "insufficient request",
			}),
			script: []mock.Script{
				{
					Name: mock.FuncSetStationConfiguration,
					Input: mock.Input{
						Request: mcom.SetStationConfigurationRequest{
							StationID: testStationA,
							Feed: mcom.StationFeedConfigs{
								ProductTypes:         []string{},
								NeedMaterialResource: true,
								QuantitySource:       stations.FeedQuantitySource_FROM_RECIPE,
							},
							Collect: mcom.StationCollectConfigs{
								NeedCarrierResource: true,
								NeedCollectResource: true,
								QuantitySource:      stations.CollectQuantitySource_FROM_STATION_CONFIGS,
								DefaultQuantity:     testQuantity,
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
			},
		},
		{
			name: "internal error",
			args: args{
				params: ui.SetStationConfigParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationA,
					Body: ui.SetStationConfigBody{
						StationConfig: &models.StationConfig{
							SeparateMode: false,
							Feed: &models.StationConfigFeed{
								ProductType:      []string{testWorkOrder1ProductType},
								MaterialResource: handlerUtils.NewBoolean(true),
							},
							Collect: &models.StationConfigCollect{
								Resource:        handlerUtils.NewBoolean(true),
								CarrierResource: handlerUtils.NewBoolean(true),
								Quantity: &models.StationConfigCollectQuantity{
									Type:  int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
									Value: testQuantity.String(),
								},
							},
						},
					},
				},
				principal: principal,
			},
			want: ui.NewSetStationConfigDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncSetStationConfiguration,
					Input: mock.Input{
						Request: mcom.SetStationConfigurationRequest{
							StationID: testStationA,
							Feed: mcom.StationFeedConfigs{
								ProductTypes:         []string{testWorkOrder1ProductType},
								NeedMaterialResource: true,
								QuantitySource:       stations.FeedQuantitySource_FROM_RECIPE,
							},
							Collect: mcom.StationCollectConfigs{
								NeedCarrierResource: true,
								NeedCollectResource: true,
								QuantitySource:      stations.CollectQuantitySource_FROM_STATION_CONFIGS,
								DefaultQuantity:     testQuantity,
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

			s := NewUI(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.SetStationConfig(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("SetStationConfig() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New(nil)
		s := NewUI(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.SetStationConfig(ui.SetStationConfigParams{
			HTTPRequest: httpRequestWithHeader,
			StationID:   testStationA,
			Body: ui.SetStationConfigBody{
				StationConfig: &models.StationConfig{
					SeparateMode: false,
					Feed: &models.StationConfigFeed{
						ProductType:      []string{testWorkOrder1ProductType},
						MaterialResource: handlerUtils.NewBoolean(true),
					},
					Collect: &models.StationConfigCollect{
						Resource:        handlerUtils.NewBoolean(true),
						CarrierResource: handlerUtils.NewBoolean(true),
						Quantity: &models.StationConfigCollectQuantity{
							Type:  int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
							Value: testQuantity.String(),
						},
					},
				},
			},
		}, principal).(*ui.SetStationConfigDefault)
		assert.True(ok)
		assert.Equal(ui.NewSetStationConfigDefault(http.StatusForbidden), rep)
	}
}

func TestUI_GetStationConfig(t *testing.T) {
	assert := assert.New(t)

	httpRequest := httptest.NewRequest("GET", "/production-flow/config/station/{stationID}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    ui.GetStationConfigParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success feed & collect together",
			args: args{
				params: ui.GetStationConfigParams{
					HTTPRequest: httpRequest,
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: ui.NewGetStationConfigOK().WithPayload(
				&ui.GetStationConfigOKBody{
					Data: &ui.GetStationConfigOKBodyData{
						StationConfig: &models.StationConfig{
							SeparateMode: false,
							Feed: &models.StationConfigFeed{
								MaterialResource: handlerUtils.NewBoolean(true),
								ProductType:      []string{testWorkOrder1ProductType},
								OperatorSites:    defaultOperatorSite(),
								StandardQuantity: int64(stations.FeedQuantitySource_FROM_RECIPE),
							},
							Collect: &models.StationConfigCollect{
								CarrierResource: handlerUtils.NewBoolean(true),
								Resource:        handlerUtils.NewBoolean(false),
								Quantity: &models.StationConfigCollectQuantity{
									Type:  int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
									Value: testQuantity.String(),
								},
								OperatorSites: defaultOperatorSite(),
							},
						},
					},
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetStationConfiguration,
					Input: mock.Input{
						Request: mcom.GetStationConfigurationRequest{
							StationID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationConfigurationReply{
							SplitFeedAndCollect: false,
							Feed: mcom.StationFeedConfigs{
								NeedMaterialResource: true,
								ProductTypes:         []string{testWorkOrder1ProductType},
								QuantitySource:       stations.FeedQuantitySource_FROM_RECIPE,
							},
							Collect: mcom.StationCollectConfigs{
								NeedCarrierResource: true,
								NeedCollectResource: false,
								QuantitySource:      stations.CollectQuantitySource_FROM_STATION_CONFIGS,
								DefaultQuantity:     testQuantity,
							},
						},
					},
				},
			},
		},
		{
			name: "success feed & collect separate",
			args: args{
				params: ui.GetStationConfigParams{
					HTTPRequest: httpRequest,
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: ui.NewGetStationConfigOK().WithPayload(
				&ui.GetStationConfigOKBody{
					Data: &ui.GetStationConfigOKBodyData{
						StationConfig: &models.StationConfig{
							SeparateMode: true,
							Feed: &models.StationConfigFeed{
								MaterialResource: handlerUtils.NewBoolean(true),
								ProductType:      []string{testWorkOrder1ProductType},
								StandardQuantity: int64(stations.FeedQuantitySource_USER_DEFINITION),
								OperatorSites: []*models.SiteInfo{
									{
										StationID: testStationA,
										SiteName:  "Feed",
										SiteIndex: 0,
									},
								},
							},
							Collect: &models.StationConfigCollect{
								Resource:        handlerUtils.NewBoolean(false),
								CarrierResource: handlerUtils.NewBoolean(true),
								Quantity: &models.StationConfigCollectQuantity{
									Type:  int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
									Value: testQuantity.String(),
								},
								OperatorSites: []*models.SiteInfo{
									{
										StationID: testStationA,
										SiteName:  "Collect",
										SiteIndex: 0,
									},
								},
							},
						},
					},
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetStationConfiguration,
					Input: mock.Input{
						Request: mcom.GetStationConfigurationRequest{
							StationID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationConfigurationReply{
							SplitFeedAndCollect: true,
							Feed: mcom.StationFeedConfigs{
								NeedMaterialResource: true,
								ProductTypes:         []string{testWorkOrder1ProductType},
								QuantitySource:       stations.FeedQuantitySource_USER_DEFINITION,
								OperatorSites: []mcomModels.UniqueSite{
									{
										Station: testStationA,
										SiteID: mcomModels.SiteID{
											Name:  "Feed",
											Index: 0,
										},
									},
								},
							},
							Collect: mcom.StationCollectConfigs{
								NeedCarrierResource: true,
								NeedCollectResource: false,
								QuantitySource:      stations.CollectQuantitySource_FROM_STATION_CONFIGS,
								DefaultQuantity:     testQuantity,
								OperatorSites: []mcomModels.UniqueSite{
									{
										Station: testStationA,
										SiteID: mcomModels.SiteID{
											Name:  "Collect",
											Index: 0,
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
			name: "not found config, get empty config",
			args: args{
				params: ui.GetStationConfigParams{
					HTTPRequest: httpRequest,
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: ui.NewGetStationConfigOK().WithPayload(
				&ui.GetStationConfigOKBody{
					Data: &ui.GetStationConfigOKBodyData{
						StationConfig: &models.StationConfig{
							SeparateMode: false,
							Feed: &models.StationConfigFeed{
								MaterialResource: handlerUtils.NewBoolean(true),
								ProductType:      []string{},
								OperatorSites:    defaultOperatorSite(),
								StandardQuantity: int64(stations.FeedQuantitySource_FROM_RECIPE),
							},
							Collect: &models.StationConfigCollect{
								Resource:        handlerUtils.NewBoolean(true),
								CarrierResource: handlerUtils.NewBoolean(true),
								Quantity: &models.StationConfigCollectQuantity{
									Type:  int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
									Value: "0",
								},
								OperatorSites: defaultOperatorSite(),
							},
						},
					},
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetStationConfiguration,
					Input: mock.Input{
						Request: mcom.GetStationConfigurationRequest{
							StationID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationConfigurationReply{},
					},
				},
			},
		},
		{
			name: "insufficient request",
			args: args{
				params: ui.GetStationConfigParams{
					HTTPRequest: httpRequest,
					StationID:   "",
				},
				principal: principal,
			},
			want: ui.NewGetStationConfigDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "insufficient request",
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetStationConfiguration,
					Input: mock.Input{
						Request: mcom.GetStationConfigurationRequest{
							StationID: "",
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
							Details: "insufficient request",
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: ui.GetStationConfigParams{
					HTTPRequest: httpRequest,
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: ui.NewGetStationConfigDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncGetStationConfiguration,
					Input: mock.Input{
						Request: mcom.GetStationConfigurationRequest{
							StationID: testStationA,
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

			s := NewUI(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.GetStationConfig(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("GetStationConfig() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := NewUI(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.GetStationConfig(ui.GetStationConfigParams{
			HTTPRequest: httpRequest,
			StationID:   testStationA,
		}, principal).(*ui.GetStationConfigDefault)
		assert.True(ok)
		assert.Equal(ui.NewGetStationConfigDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}
