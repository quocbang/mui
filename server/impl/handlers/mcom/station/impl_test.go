package station

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
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	mcomSites "gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/station"
)

const (
	userID = "tester"

	testStationB = "Station-B"
	testStationC = "Station-C"
	testStationD = "Station-D"
	testStationE = "Station-E"
	testStationF = "Station-F"
	testStationG = "Station-G"

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
	testStationA      = "Station-A"
	testDepartmentOID = "ABCDEFGHIJKLMOPQRSTUVWXYZ123456789"

	testInsertDate = time.Date(2021, 8, 9, 16, 45, 23, 500, time.Local)
	testUpdateDate = time.Date(2021, 8, 9, 16, 45, 23, 500, time.Local)

	testEmpty       = ""
	testCodeA11     = "A11"
	testDescription = "A"
	testIdle        = models.StationState(stations.State_IDLE)

	testQuantity1Decimal = decimal.NewFromInt(174)
	testQuantity2Decimal = decimal.NewFromInt(188)

	testSiteName1 = "TESTSITENAME1"

	testSchedulingDate = time.Date(2021, 8, 9, 0, 0, 0, 0, time.Local)
	testStationErrorID = "TESTSTATIONERRORNAME1"

	testStationID = "STATION1"
)

func TestStation_GetStationList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListStations,
			Input: mock.Input{
				Request: mcom.ListStationsRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Response: mcom.ListStationsReply{
					Stations: []mcom.Station{
						{
							ID: testStationA,
							// TODO: 之後補上相關資訊
						},
						{
							ID: testStationB,
						},
						{
							ID: testStationC,
						},
						{
							ID: testStationD,
						},
						{
							ID: testStationE,
						},
						{
							ID: testStationF,
						},
						{
							ID: testStationG,
						},
					},
				},
			},
		},
		{
			Name: mock.FuncListStations,
			Input: mock.Input{
				Request: mcom.ListStationsRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_STATION_NOT_FOUND, // for test case, in fact it won't return this code
				},
			},
		},
		{
			Name: mock.FuncListStations,
			Input: mock.Input{
				Request: mcom.ListStationsRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/station-list/department-oid/{departmentOID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.GetStationListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get station list success",
			args: args{
				params: station.GetStationListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: station.NewGetStationListOK().WithPayload(&station.GetStationListOKBody{
				Data: []*station.GetStationListOKBodyDataItems0{
					{
						ID: testStationA,
					},
					{
						ID: testStationB,
					},
					{
						ID: testStationC,
					},
					{
						ID: testStationD,
					},
					{
						ID: testStationE,
					},
					{
						ID: testStationF,
					},
					{
						ID: testStationG,
					},
				},
			}),
		},
		{
			name: "not found data",
			args: args{
				params: station.GetStationListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: station.NewGetStationListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_STATION_NOT_FOUND), // for test case, in fact it won't return this code
			}),
		},
		{
			name: "internal error",
			args: args{
				params: station.GetStationListParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: station.NewGetStationListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.GetStationList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStationList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.GetStationList(station.GetStationListParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentOID: testDepartmentOID,
		}, principal).(*station.GetStationListDefault)
		assert.True(ok)
		assert.Equal(station.NewGetStationListDefault(http.StatusForbidden), rep)
	}
}

func TestStation_ListStationInfo(t *testing.T) {
	assert := assert.New(t)

	var (
		testTotal       = 10
		testPage        = int64(1)
		testLimit       = int64(20)
		testOrderName   = "id"
		testDescending  = false
		testPageRequest = mcom.PaginationRequest{
			PageCount:      uint(testPage),
			ObjectsPerPage: uint(testLimit),
		}
		testOrderRequest = []mcom.Order{
			{
				Name:       testOrderName,
				Descending: testDescending,
			},
		}
		testOrder = []*station.ListStationInfoParamsBodyOrderRequestItems0{
			{
				OrderName:  testOrderName,
				Descending: testDescending,
			}}
	)

	var (
		container = mcomModels.Container([]mcomModels.BoundResource{
			{
				Material: &mcomModels.MaterialSite{
					Material: mcomModels.Material{
						ID:    "MAT11",
						Grade: "B",
					},
					Quantity:   &testQuantity1Decimal,
					ResourceID: "R0147852369X",
				},
			},
			{
				Material: &mcomModels.MaterialSite{
					Material: mcomModels.Material{
						ID:    "MAT11",
						Grade: "B",
					},
					Quantity:   nil,
					ResourceID: "R0147852369A",
				},
			},
		})

		slot = mcomModels.Slot(mcomModels.BoundResource{
			Operator: &mcomModels.OperatorSite{
				EmployeeID: userID,
				Group:      7,
				WorkDate:   testInsertDate,
			},
		})

		collection = mcomModels.Collection([]mcomModels.BoundResource{
			{
				Tool: &mcomModels.ToolSite{
					ResourceID:    "R01234567890",
					InstalledTime: testInsertDate,
				},
			},
			{
				Tool: &mcomModels.ToolSite{
					ResourceID:    "R01234567891",
					InstalledTime: testInsertDate,
				},
			},
		})

		queue = mcomModels.Queue([]mcomModels.Slot{
			{
				Material: &mcomModels.MaterialSite{
					Material: mcomModels.Material{
						ID:    "MAT11",
						Grade: "B",
					},
					Quantity:   &testQuantity1Decimal,
					ResourceID: "R0147852369X",
				},
			},
			{
				Material: &mcomModels.MaterialSite{
					Material: mcomModels.Material{
						ID:    "MAT11",
						Grade: "B",
					},
					Quantity:   &testQuantity2Decimal,
					ResourceID: "R0147852369A",
				},
			},
		})

		colQueue = mcomModels.Colqueue([]mcomModels.Collection{
			{
				{
					Material: &mcomModels.MaterialSite{
						Material: mcomModels.Material{
							ID:    "MAT11",
							Grade: "B",
						},
						Quantity:   &testQuantity1Decimal,
						ResourceID: "R0147852369X",
					},
				},
				{
					Material: &mcomModels.MaterialSite{
						Material: mcomModels.Material{
							ID:    "MAT11",
							Grade: "B",
						},
						Quantity:   &testQuantity2Decimal,
						ResourceID: "R0147852369A",
					},
				},
			},
			{
				{
					Material: &mcomModels.MaterialSite{
						Material: mcomModels.Material{
							ID:    "MAT12",
							Grade: "B",
						},
						Quantity:   &testQuantity1Decimal,
						ResourceID: "R0147852369B",
					},
				},
				{
					Material: &mcomModels.MaterialSite{
						Material: mcomModels.Material{
							ID:    "MAT12",
							Grade: "B",
						},
						Quantity:   &testQuantity2Decimal,
						ResourceID: "R0147852369C",
					},
				},
			},
		})

		stationData1 = []*models.StationData{
			{
				ID:          testStationA,
				Code:        &testCodeA11,
				Description: &testDescription,
				InsertedAt:  strfmt.DateTime(testInsertDate),
				InsertedBy:  userID,
				Sites: []*models.Site{
					{
						Content: &models.SiteContent{
							Container: []*models.BoundResource{
								{
									MaterialSite: &models.BoundResourceMaterialSite{
										ID:         "MAT11",
										Grade:      "B",
										Quantity:   "174",
										ResourceID: "R0147852369X",
									},
								},
								{
									MaterialSite: &models.BoundResourceMaterialSite{
										ID:         "MAT11",
										Grade:      "B",
										Quantity:   "",
										ResourceID: "R0147852369A",
									},
								},
							},
						},
						Index:   1,
						Name:    "A",
						SubType: models.SiteSubType(mcomSites.SubType_MATERIAL),
						Type:    models.SiteType(mcomSites.Type_CONTAINER),
					},
				},
				State:    &testIdle,
				UpdateAt: strfmt.DateTime(testUpdateDate),
				UpdateBy: userID,
			},
		}

		stationData2 = []*models.StationData{
			{
				ID:          testStationA,
				Code:        &testCodeA11,
				Description: &testDescription,
				InsertedAt:  strfmt.DateTime(testInsertDate),
				InsertedBy:  userID,
				Sites: []*models.Site{
					{
						Content: &models.SiteContent{
							Slot: &models.BoundResource{
								OperatorSite: &models.BoundResourceOperatorSite{
									EmployeeID: userID,
									Group:      7,
									WorkDate:   strfmt.DateTime(testInsertDate),
								},
							},
						},
						Index:   1,
						Name:    "A",
						SubType: models.SiteSubType(mcomSites.SubType_OPERATOR),
						Type:    models.SiteType(mcomSites.Type_SLOT),
					},
				},
				State:    &testIdle,
				UpdateAt: strfmt.DateTime(testUpdateDate),
				UpdateBy: userID,
			},
		}

		stationData3 = []*models.StationData{
			{
				ID:          testStationA,
				Code:        &testCodeA11,
				Description: &testDescription,
				InsertedAt:  strfmt.DateTime(testInsertDate),
				InsertedBy:  userID,
				Sites: []*models.Site{
					{
						Content: &models.SiteContent{
							Collection: []*models.BoundResource{
								{
									ToolSite: &models.BoundResourceToolSite{
										InstalledTime: strfmt.DateTime(testInsertDate),
										ResourceID:    "R01234567890",
									},
								},
								{
									ToolSite: &models.BoundResourceToolSite{
										InstalledTime: strfmt.DateTime(testInsertDate),
										ResourceID:    "R01234567891",
									},
								},
							},
						},
						Index:   1,
						Name:    "A",
						SubType: models.SiteSubType(mcomSites.SubType_TOOL),
						Type:    models.SiteType(mcomSites.Type_COLLECTION),
					},
				},
				State:    &testIdle,
				UpdateAt: strfmt.DateTime(testUpdateDate),
				UpdateBy: userID,
			},
		}

		stationData4 = []*models.StationData{
			{
				ID:          testStationA,
				Code:        &testCodeA11,
				Description: &testDescription,
				InsertedAt:  strfmt.DateTime(testInsertDate),
				InsertedBy:  userID,
				Sites: []*models.Site{
					{
						Content: &models.SiteContent{
							Queue: []*models.BoundResource{
								{
									MaterialSite: &models.BoundResourceMaterialSite{
										ID:         "MAT11",
										Grade:      "B",
										Quantity:   "174",
										ResourceID: "R0147852369X",
									},
								},
								{
									MaterialSite: &models.BoundResourceMaterialSite{
										ID:         "MAT11",
										Grade:      "B",
										Quantity:   "188",
										ResourceID: "R0147852369A",
									},
								},
							},
						},
						Index:   1,
						Name:    "A",
						SubType: models.SiteSubType(mcomSites.SubType_MATERIAL),
						Type:    models.SiteType(mcomSites.Type_QUEUE),
					},
				},
				State:    &testIdle,
				UpdateAt: strfmt.DateTime(testUpdateDate),
				UpdateBy: userID,
			},
		}

		stationData5 = []*models.StationData{
			{
				ID:          testStationA,
				Code:        &testCodeA11,
				Description: &testDescription,
				InsertedAt:  strfmt.DateTime(testInsertDate),
				InsertedBy:  userID,
				Sites: []*models.Site{
					{
						Content: &models.SiteContent{
							Colqueue: []models.CollectionContent{
								{
									{
										MaterialSite: &models.BoundResourceMaterialSite{
											ID:         "MAT11",
											Grade:      "B",
											Quantity:   "174",
											ResourceID: "R0147852369X",
										},
									},
									{
										MaterialSite: &models.BoundResourceMaterialSite{
											ID:         "MAT11",
											Grade:      "B",
											Quantity:   "188",
											ResourceID: "R0147852369A",
										},
									},
								},
								{
									{
										MaterialSite: &models.BoundResourceMaterialSite{
											ID:         "MAT12",
											Grade:      "B",
											Quantity:   "174",
											ResourceID: "R0147852369B",
										},
									},
									{
										MaterialSite: &models.BoundResourceMaterialSite{
											ID:         "MAT12",
											Grade:      "B",
											Quantity:   "188",
											ResourceID: "R0147852369C",
										},
									},
								},
							},
						},
						Index:   1,
						Name:    "A",
						SubType: models.SiteSubType(mcomSites.SubType_MATERIAL),
						Type:    models.SiteType(mcomSites.Type_COLQUEUE),
					},
				},
				State:    &testIdle,
				UpdateAt: strfmt.DateTime(testUpdateDate),
				UpdateBy: userID,
			},
		}
	)

	httpRequestWithHeader := httptest.NewRequest("GET", "/station/maintenance/department-oid/{departmentOID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.ListStationInfoParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success[container-material]",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: stationData1,
					Total: int64(testTotal)},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_CONTAINER,
												SubType: mcomSites.SubType_MATERIAL,
											},
											Content: mcomModels.SiteContent{
												Container: &container,
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[container-material] empty content",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: []*models.StationData{
						{
							ID:          testStationA,
							Code:        &testCodeA11,
							Description: &testDescription,
							InsertedAt:  strfmt.DateTime(testInsertDate),
							InsertedBy:  userID,
							Sites: []*models.Site{
								{
									Content: &models.SiteContent{},
									Index:   1,
									Name:    "A",
									SubType: models.SiteSubType(mcomSites.SubType_MATERIAL),
									Type:    models.SiteType(mcomSites.Type_CONTAINER),
								},
							},
							State:    &testIdle,
							UpdateAt: strfmt.DateTime(testUpdateDate),
							UpdateBy: userID,
						},
					},
					Total: int64(testTotal),
				},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_CONTAINER,
												SubType: mcomSites.SubType_MATERIAL,
											},
											Content: mcomModels.SiteContent{
												Container: &mcomModels.Container{},
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[slot-operator]",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: stationData2,
					Total: int64(testTotal)},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_SLOT,
												SubType: mcomSites.SubType_OPERATOR,
											},
											Content: mcomModels.SiteContent{
												Slot: &slot,
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[slot-operator] empty content",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: []*models.StationData{
						{
							ID:          testStationA,
							Code:        &testCodeA11,
							Description: &testDescription,
							InsertedAt:  strfmt.DateTime(testInsertDate),
							InsertedBy:  userID,
							Sites: []*models.Site{
								{
									Content: &models.SiteContent{},
									Index:   1,
									Name:    "A",
									SubType: models.SiteSubType(mcomSites.SubType_OPERATOR),
									Type:    models.SiteType(mcomSites.Type_SLOT),
								},
							},
							State:    &testIdle,
							UpdateAt: strfmt.DateTime(testUpdateDate),
							UpdateBy: userID,
						},
					},
					Total: int64(testTotal),
				},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_SLOT,
												SubType: mcomSites.SubType_OPERATOR,
											},
											Content: mcomModels.SiteContent{
												Slot: &mcomModels.Slot{},
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[collection-tool]",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: stationData3,
					Total: int64(testTotal)},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_COLLECTION,
												SubType: mcomSites.SubType_TOOL,
											},
											Content: mcomModels.SiteContent{
												Collection: &collection,
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[collection-tool] empty content",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: []*models.StationData{
						{
							ID:          testStationA,
							Code:        &testCodeA11,
							Description: &testDescription,
							InsertedAt:  strfmt.DateTime(testInsertDate),
							InsertedBy:  userID,
							Sites: []*models.Site{
								{
									Content: &models.SiteContent{},
									Index:   1,
									Name:    "A",
									SubType: models.SiteSubType(mcomSites.SubType_TOOL),
									Type:    models.SiteType(mcomSites.Type_COLLECTION),
								},
							},
							State:    &testIdle,
							UpdateAt: strfmt.DateTime(testUpdateDate),
							UpdateBy: userID,
						},
					},
					Total: int64(testTotal)},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_COLLECTION,
												SubType: mcomSites.SubType_TOOL,
											},
											Content: mcomModels.SiteContent{
												Collection: &mcomModels.Collection{},
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[queue-material]",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: stationData4,
					Total: int64(testTotal)},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_QUEUE,
												SubType: mcomSites.SubType_MATERIAL,
											},
											Content: mcomModels.SiteContent{
												Queue: &queue,
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[queue-material] empty content",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: []*models.StationData{
						{
							ID:          testStationA,
							Code:        &testCodeA11,
							Description: &testDescription,
							InsertedAt:  strfmt.DateTime(testInsertDate),
							InsertedBy:  userID,
							Sites: []*models.Site{
								{
									Content: &models.SiteContent{},
									Index:   1,
									Name:    "A",
									SubType: models.SiteSubType(mcomSites.SubType_MATERIAL),
									Type:    models.SiteType(mcomSites.Type_QUEUE),
								},
							},
							State:    &testIdle,
							UpdateAt: strfmt.DateTime(testUpdateDate),
							UpdateBy: userID,
						},
					},
					Total: int64(testTotal)},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_QUEUE,
												SubType: mcomSites.SubType_MATERIAL,
											},
											Content: mcomModels.SiteContent{
												Queue: &mcomModels.Queue{},
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[collection-queue-material]",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: stationData5,
					Total: int64(testTotal)},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_COLQUEUE,
												SubType: mcomSites.SubType_MATERIAL,
											},
											Content: mcomModels.SiteContent{
												Colqueue: &colQueue,
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "success[collection-queue-material] empty content",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
				Data: &station.ListStationInfoOKBodyData{
					Items: []*models.StationData{
						{
							ID:          testStationA,
							Code:        &testCodeA11,
							Description: &testDescription,
							InsertedAt:  strfmt.DateTime(testInsertDate),
							InsertedBy:  userID,
							Sites: []*models.Site{
								{
									Content: &models.SiteContent{},
									Index:   1,
									Name:    "A",
									SubType: models.SiteSubType(mcomSites.SubType_MATERIAL),
									Type:    models.SiteType(mcomSites.Type_COLQUEUE),
								},
							},
							State:    &testIdle,
							UpdateAt: strfmt.DateTime(testUpdateDate),
							UpdateBy: userID,
						},
					},
					Total: int64(testTotal)},
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Response: mcom.ListStationsReply{
							Stations: []mcom.Station{
								{
									ID:                testStationA,
									AdminDepartmentID: testDepartmentOID,
									Sites: []mcom.ListStationSite{
										{
											Information: mcom.ListStationSitesInformation{
												UniqueSite: mcomModels.UniqueSite{
													SiteID: mcomModels.SiteID{
														Name:  "A",
														Index: 1,
													},
													Station: testStationA,
												},
												Type:    mcomSites.Type_COLQUEUE,
												SubType: mcomSites.SubType_MATERIAL,
											},
											Content: mcomModels.SiteContent{
												Colqueue: &mcomModels.Colqueue{},
											},
										},
									},
									State: stations.State_IDLE,
									Information: mcom.StationInformation{
										Code:        testCodeA11,
										Description: testDescription,
									},
									UpdatedBy:  userID,
									UpdatedAt:  testUpdateDate,
									InsertedBy: userID,
									InsertedAt: testInsertDate,
								},
							},
							PaginationReply: mcom.PaginationReply{
								AmountOfData: int64(testTotal),
							},
						},
					},
				},
			},
		},
		{
			name: "bad request",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: "",
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "empty request",
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: ""}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
							Details: "empty request",
						},
					},
				},
			},
		},
		{
			name: "internal server error",
			args: args{
				params: station.ListStationInfoParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
					Limit:         &testLimit,
					Page:          &testPage,
					Body: station.ListStationInfoBody{
						OrderRequest: testOrder,
					},
				},
				principal: principal,
			},
			want: station.NewListStationInfoDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncListStations,
					Input: mock.Input{
						Request: mcom.ListStationsRequest{DepartmentID: testDepartmentOID}.
							WithPagination(testPageRequest).WithOrder(testOrderRequest...),
					},
					Output: mock.Output{
						Error: errors.New(testInternalServerError),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		dm, err := mock.New(tt.script)
		assert.NoErrorf(err, tt.name)
		t.Run(tt.name, func(t *testing.T) {
			m := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := m.ListStationInfo(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		m := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := m.ListStationInfo(station.ListStationInfoParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentOID: testDepartmentOID,
		}, principal).(*station.ListStationInfoDefault)
		assert.True(ok)
		assert.Equal(station.NewListStationInfoDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestStation_CreateStation(t *testing.T) {
	assert := assert.New(t)

	createScripts := []mock.Script{
		{
			Name: mock.FuncCreateStation,
			Input: mock.Input{
				Request: mcom.CreateStationRequest{
					ID:           testStationA,
					DepartmentID: testDepartmentOID,
					State:        stations.State_SHUTDOWN,
					Sites: []mcom.SiteInformation{
						{
							Name:    "AS",
							Index:   1,
							Type:    mcomSites.Type_SLOT,
							SubType: mcomSites.SubType_OPERATOR,
						},
					},
					Information: mcom.StationInformation{
						Code:        testCodeA11,
						Description: testDescription,
					},
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncCreateStation,
			Input: mock.Input{
				Request: mcom.CreateStationRequest{
					ID:           "",
					DepartmentID: testDepartmentOID,
					State:        stations.State_SHUTDOWN,
					Sites: []mcom.SiteInformation{
						{
							Name:    "AS",
							Index:   1,
							Type:    mcomSites.Type_SLOT,
							SubType: mcomSites.SubType_OPERATOR,
						},
					},
					Information: mcom.StationInformation{
						Code:        testCodeA11,
						Description: testDescription,
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
			Name: mock.FuncCreateStation,
			Input: mock.Input{
				Request: mcom.CreateStationRequest{
					ID:           testStationA,
					DepartmentID: testDepartmentOID,
					State:        stations.State_SHUTDOWN,
					Sites: []mcom.SiteInformation{
						{
							Name:    "AS",
							Index:   1,
							Type:    mcomSites.Type_SLOT,
							SubType: mcomSites.SubType_OPERATOR,
						},
					},
					Information: mcom.StationInformation{
						Code:        testCodeA11,
						Description: testDescription,
					},
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}
	dm, err := mock.New(createScripts)
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("POST", "/station/maintenance", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.CreateStationParams
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
				params: station.CreateStationParams{
					HTTPRequest: httpRequestWithHeader,
					Body: station.CreateStationBody{
						ID:            &testStationA,
						Code:          &testCodeA11,
						DepartmentOID: &testDepartmentOID,
						Description:   &testDescription,
						Sites: []*models.Site{
							{
								ActionMode: models.SiteActionMode(1),
								Content: &models.SiteContent{
									Slot: &models.BoundResource{
										OperatorSite: &models.BoundResourceOperatorSite{
											EmployeeID: userID,
											Group:      11,
											WorkDate:   strfmt.DateTime(testInsertDate),
										},
									},
								},
								Index:   1,
								Name:    "AS",
								SubType: models.SiteSubType(mcomSites.SubType_OPERATOR),
								Type:    models.SiteType(mcomSites.Type_SLOT),
							},
						},
					},
				},
				principal: principal,
			},
			want: station.NewCreateStationOK(),
		},
		{
			name: "bad site action mode",
			args: args{
				params: station.CreateStationParams{
					HTTPRequest: httpRequestWithHeader,
					Body: station.CreateStationBody{
						ID:            &testEmpty,
						Code:          &testCodeA11,
						DepartmentOID: &testDepartmentOID,
						Description:   &testDescription,
						Sites: []*models.Site{
							{
								ActionMode: models.SiteActionMode(0),
								Content: &models.SiteContent{
									Slot: &models.BoundResource{
										OperatorSite: &models.BoundResourceOperatorSite{
											EmployeeID: userID,
											ExpiryTime: strfmt.DateTime{},
											Group:      11,
											WorkDate:   strfmt.DateTime(testInsertDate),
										},
									},
								},
								Index:   1,
								Name:    "AS",
								SubType: models.SiteSubType(mcomSites.SubType_OPERATOR),
								Type:    models.SiteType(mcomSites.Type_SLOT),
							},
						},
					},
				},
				principal: principal,
			},
			want: station.NewCreateStationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Details: "wrong site action mode",
			}),
		},
		{
			name: "bad request",
			args: args{
				params: station.CreateStationParams{
					HTTPRequest: httpRequestWithHeader,
					Body: station.CreateStationBody{
						ID:            &testEmpty,
						Code:          &testCodeA11,
						DepartmentOID: &testDepartmentOID,
						Description:   &testDescription,
						Sites: []*models.Site{
							{
								ActionMode: models.SiteActionMode(1),
								Content: &models.SiteContent{
									Slot: &models.BoundResource{
										OperatorSite: &models.BoundResourceOperatorSite{
											EmployeeID: userID,
											ExpiryTime: strfmt.DateTime{},
											Group:      11,
											WorkDate:   strfmt.DateTime(testInsertDate),
										},
									},
								},
								Index:   1,
								Name:    "AS",
								SubType: models.SiteSubType(mcomSites.SubType_OPERATOR),
								Type:    models.SiteType(mcomSites.Type_SLOT),
							},
						},
					},
				},
				principal: principal,
			},
			want: station.NewCreateStationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "insufficient request",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: station.CreateStationParams{
					HTTPRequest: httpRequestWithHeader,
					Body: station.CreateStationBody{
						ID:            &testStationA,
						Code:          &testCodeA11,
						DepartmentOID: &testDepartmentOID,
						Description:   &testDescription,
						Sites: []*models.Site{
							{
								ActionMode: models.SiteActionMode(1),
								Content: &models.SiteContent{
									Slot: &models.BoundResource{
										OperatorSite: &models.BoundResourceOperatorSite{
											EmployeeID: userID,
											ExpiryTime: strfmt.DateTime{},
											Group:      11,
											WorkDate:   strfmt.DateTime(testInsertDate),
										},
									},
								},
								Index:   1,
								Name:    "AS",
								SubType: models.SiteSubType(mcomSites.SubType_OPERATOR),
								Type:    models.SiteType(mcomSites.Type_SLOT),
							},
						},
					},
				},
				principal: principal,
			},
			want: station.NewCreateStationDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := m.CreateStation(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		m := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := m.CreateStation(station.CreateStationParams{
			HTTPRequest: httpRequestWithHeader,
			Body:        station.CreateStationBody{},
		}, principal).(*station.CreateStationDefault)
		assert.True(ok)
		assert.Equal(station.NewCreateStationDefault(http.StatusForbidden), rep)
	}
}

func TestStation_UpdateStationInfo(t *testing.T) {
	assert := assert.New(t)

	updateScripts := []mock.Script{
		{
			Name: mock.FuncUpdateStation,
			Input: mock.Input{
				Request: mcom.UpdateStationRequest{
					ID: testStationA,
					Sites: []mcom.UpdateStationSite{
						{
							ActionMode: mcomSites.ActionType_REMOVE,
							Information: mcom.SiteInformation{
								Name:    "SA",
								Index:   11,
								Type:    mcomSites.Type_SLOT,
								SubType: mcomSites.SubType_OPERATOR,
							},
						},
					},
					State: stations.State_IDLE,
					Information: mcom.StationInformation{
						Code:        testCodeA11,
						Description: testDescription,
					},
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncUpdateStation,
			Input: mock.Input{
				Request: mcom.UpdateStationRequest{
					ID:    "",
					Sites: []mcom.UpdateStationSite{},
					State: stations.State_IDLE,
					Information: mcom.StationInformation{
						Code:        testCodeA11,
						Description: testDescription,
					},
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "empty id",
				},
			},
		},
		{
			Name: mock.FuncUpdateStation,
			Input: mock.Input{
				Request: mcom.UpdateStationRequest{
					ID:    testStationA,
					Sites: []mcom.UpdateStationSite{},
					State: stations.State_IDLE,
					Information: mcom.StationInformation{
						Code:        testCodeA11,
						Description: testDescription,
					},
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}
	dm, err := mock.New(updateScripts)
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("PATCH", "/station/maintenance/{ID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.UpdateStationInfoParams
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
				params: station.UpdateStationInfoParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testStationA,
					Body: &models.StationData{
						Code:        &testCodeA11,
						Description: &testDescription,
						State:       &testIdle,
						Sites: []*models.Site{
							{
								ActionMode: models.SiteActionMode(2),
								Index:      11,
								Name:       "SA",
								SubType:    models.SiteSubType(mcomSites.SubType_OPERATOR),
								Type:       models.SiteType(mcomSites.Type_SLOT),
							},
						},
					},
				},
				principal: principal,
			},
			want: station.NewUpdateStationInfoOK(),
		},
		{
			name: "bad read site action mode",
			args: args{
				params: station.UpdateStationInfoParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testStationA,
					Body: &models.StationData{
						Code:        &testCodeA11,
						Description: &testDescription,
						State:       &testIdle,
						Sites: []*models.Site{
							{
								ActionMode: models.SiteActionMode(0),
								Index:      11,
								Name:       "SA",
								SubType:    models.SiteSubType(mcomSites.SubType_OPERATOR),
								Type:       models.SiteType(mcomSites.Type_SLOT),
							},
						},
					},
				},
				principal: principal,
			},
			want: station.NewUpdateStationInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Details: "not allow READ site on update station",
			}),
		},
		{
			name: "bad request",
			args: args{
				params: station.UpdateStationInfoParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          "",
					Body: &models.StationData{
						Code:        &testCodeA11,
						Description: &testDescription,
						State:       &testIdle,
						Sites:       []*models.Site{},
					},
				},
				principal: principal,
			},
			want: station.NewUpdateStationInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "empty id",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: station.UpdateStationInfoParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testStationA,
					Body: &models.StationData{
						Code:        &testCodeA11,
						Description: &testDescription,
						State:       &testIdle,
						Sites:       []*models.Site{},
					},
				},
				principal: principal,
			},
			want: station.NewUpdateStationInfoDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := m.UpdateStationInfo(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		m := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := m.UpdateStationInfo(station.UpdateStationInfoParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testStationA,
		}, principal).(*station.UpdateStationInfoDefault)
		assert.True(ok)
		assert.Equal(station.NewUpdateStationInfoDefault(http.StatusForbidden), rep)
	}
}

func TestStation_DeleteStation(t *testing.T) {
	assert := assert.New(t)

	deleteScripts := []mock.Script{
		{
			Name: mock.FuncDeleteStation,
			Input: mock.Input{
				Request: mcom.DeleteStationRequest{
					StationID: testStationA,
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncDeleteStation,
			Input: mock.Input{
				Request: mcom.DeleteStationRequest{
					StationID: "",
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
					Details: "empty id",
				},
			},
		},
		{
			Name: mock.FuncDeleteStation,
			Input: mock.Input{
				Request: mcom.DeleteStationRequest{
					StationID: testStationA,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}
	dm, err := mock.New(deleteScripts)
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("DELETE", "/station/maintenance/{ID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.DeleteStationParams
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
				params: station.DeleteStationParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testStationA,
				},
				principal: principal,
			},
			want: station.NewDeleteStationOK(),
		},
		{
			name: "bad request",
			args: args{
				params: station.DeleteStationParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          "",
				},
				principal: principal,
			},
			want: station.NewDeleteStationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "empty id",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: station.DeleteStationParams{
					HTTPRequest: httpRequestWithHeader,
					ID:          testStationA,
				},
				principal: principal,
			},
			want: station.NewDeleteStationDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := m.DeleteStation(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Delete() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		m := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := m.DeleteStation(station.DeleteStationParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testStationA,
		}, principal).(*station.DeleteStationDefault)
		assert.True(ok)
		assert.Equal(station.NewDeleteStationDefault(http.StatusForbidden), rep)
	}
}

func TestStation_GetStationStateList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name:  mock.FuncListStationState,
			Input: mock.Input{},
			Output: mock.Output{
				Response: mcom.ListStationStateReply{
					{
						Value: stations.State_SHUTDOWN,
						Name:  "SHUTDOWN",
					},
					{
						Value: stations.State_IDLE,
						Name:  "IDLE",
					},
					{
						Value: stations.State_RUNNING,
						Name:  "RUNNING",
					},
					{
						Value: stations.State_MAINTENANCE,
						Name:  "MAINTENANCE",
					},
					{
						Value: stations.State_REPAIRING,
						Name:  "REPAIRING",
					},
					{
						Value: stations.State_MALFUNCTION,
						Name:  "MALFUNCTION",
					},
					{
						Value: stations.State_DISPOSAL,
						Name:  "DISPOSAL",
					},
				},
			},
		},
		{
			Name:  mock.FuncListStationState,
			Input: mock.Input{},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_STATION_NOT_FOUND, // for test case, in fact it won't return this code
				},
			},
		},
		{
			Name:  mock.FuncListStationState,
			Input: mock.Input{},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/station/state", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.GetStationStateListParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "get station state list success",
			args: args{
				params: station.GetStationStateListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: station.NewGetStationStateListOK().WithPayload(&station.GetStationStateListOKBody{
				Data: []*station.GetStationStateListOKBodyDataItems0{
					{
						ID:   models.StationState(stations.State_SHUTDOWN),
						Name: "SHUTDOWN",
					},
					{
						ID:   models.StationState(stations.State_IDLE),
						Name: "IDLE",
					},
					{
						ID:   models.StationState(stations.State_RUNNING),
						Name: "RUNNING",
					},
					{
						ID:   models.StationState(stations.State_MAINTENANCE),
						Name: "MAINTENANCE",
					},
					{
						ID:   models.StationState(stations.State_REPAIRING),
						Name: "REPAIRING",
					},
					{
						ID:   models.StationState(stations.State_MALFUNCTION),
						Name: "MALFUNCTION",
					},
					{
						ID:   models.StationState(stations.State_DISPOSAL),
						Name: "DISPOSAL",
					},
				},
			}),
		},
		{
			name: "not found data",
			args: args{
				params: station.GetStationStateListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: station.NewGetStationStateListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_STATION_NOT_FOUND), // for test case, in fact it won't return this code
			}),
		},
		{
			name: "internal error",
			args: args{
				params: station.GetStationStateListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: station.NewGetStationStateListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.GetStationStateList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStationStateList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.GetStationStateList(station.GetStationStateListParams{
			HTTPRequest: httpRequestWithHeader,
		}, principal).(*station.GetStationStateListDefault)
		assert.True(ok)
		assert.Equal(station.NewGetStationStateListDefault(http.StatusForbidden), rep)
	}
}

func TestStation_ListStations(t *testing.T) {
	assert := assert.New(t)

	var (
		testStationList        = []string{testStationA, testStationErrorID}
		testDepartmentErrorOID = "NOTEXIST"
	)

	httpRequestWithHeader := httptest.NewRequest("GET", "/production-flow/station/department-oid/{departmentOID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.ListStationsParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success all station",
			args: args{
				params: station.ListStationsParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: nil,
				},
				principal: principal,
			},
			want: station.NewListStationsOK().WithPayload(&station.ListStationsOKBody{
				Data: []*station.ListStationsOKBodyDataItems0{
					{
						StationID: testStationA,
					},
					{
						StationID: testStationErrorID,
					}},
			}),
			script: []mock.Script{
				{ // success all station
					Name: mock.FuncListStationIDs,
					Input: mock.Input{
						Request: mcom.ListStationIDsRequest{},
					},
					Output: mock.Output{
						Response: mcom.ListStationIDsReply{
							Stations: testStationList,
						},
					},
				},
			},
		},
		{
			name: "success",
			args: args{
				params: station.ListStationsParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: &testDepartmentOID,
				},
				principal: principal,
			},
			want: station.NewListStationsOK().WithPayload(&station.ListStationsOKBody{
				Data: []*station.ListStationsOKBodyDataItems0{
					{
						StationID: testStationA,
					},
					{
						StationID: testStationErrorID,
					}},
			}),
			script: []mock.Script{
				{ // success
					Name: mock.FuncListStationIDs,
					Input: mock.Input{
						Request: mcom.ListStationIDsRequest{
							DepartmentID: testDepartmentOID,
						},
					},
					Output: mock.Output{
						Response: mcom.ListStationIDsReply{
							Stations: testStationList,
						},
					},
				},
			},
		},
		{
			name: "insufficient request",
			args: args{
				params: station.ListStationsParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: &testDepartmentErrorOID,
				},
				principal: principal,
			},
			want: station.NewListStationsDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
			script: []mock.Script{
				{ // insufficient request
					Name: mock.FuncListStationIDs,
					Input: mock.Input{
						Request: mcom.ListStationIDsRequest{
							DepartmentID: testDepartmentErrorOID,
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
				params: station.ListStationsParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: nil,
				},
				principal: principal,
			},
			want: station.NewListStationsDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{ // internal error
					Name: mock.FuncListStationIDs,
					Input: mock.Input{
						Request: mcom.ListStationIDsRequest{},
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

			s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.ListStations(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("ListStations() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New(nil)
		s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.ListStations(station.ListStationsParams{
			HTTPRequest: httpRequestWithHeader,
		}, principal).(*station.ListStationsDefault)
		assert.True(ok)
		assert.Equal(station.NewListStationsDefault(http.StatusForbidden), rep)
	}
}

func TestStation_ListStationSites(t *testing.T) {
	assert := assert.New(t)

	httpRequest := httptest.NewRequest("GET", "/production-flow/site/station/{stationID}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.ListStationSitesParams
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
				params: station.ListStationSitesParams{
					HTTPRequest: httpRequest,
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: station.NewListStationSitesOK().WithPayload(
				&station.ListStationSitesOKBody{
					Data: []*station.ListStationSitesOKBodyDataItems0{
						{
							Site: &models.SiteInfo{
								StationID: testStationA,
								SiteName:  testSiteName1,
								SiteIndex: 1,
							},
							Type:    mcomSites.Type_CONTAINER.String(),
							SubType: int64(mcomSites.SubType_MATERIAL),
						},
					},
				}),
			script: []mock.Script{
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationReply(
							mcom.Station{
								Sites: []mcom.ListStationSite{
									{
										Information: mcom.ListStationSitesInformation{
											UniqueSite: mcomModels.UniqueSite{
												SiteID: mcomModels.SiteID{
													Name:  testSiteName1,
													Index: 1,
												},
												Station: testStationA,
											},
											Type:    mcomSites.Type_CONTAINER,
											SubType: mcomSites.SubType_MATERIAL,
										},
									},
								},
							},
						),
					},
				},
			},
		},
		{
			name: "station not found",
			args: args{
				params: station.ListStationSitesParams{
					HTTPRequest: httpRequest,
					StationID:   testStationErrorID,
				},
				principal: principal,
			},
			want: station.NewListStationSitesDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_STATION_NOT_FOUND),
				Details: "station not found",
			}),
			script: []mock.Script{
				{ // station not found
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationErrorID,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code:    mcomErrors.Code_STATION_NOT_FOUND,
							Details: "station not found",
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: station.ListStationSitesParams{
					HTTPRequest: httpRequest,
					StationID:   testStationA,
				},
				principal: principal,
			},
			want: station.NewListStationSitesDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{ // internal error
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
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

			s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.ListStationSites(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("ListStationSites() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.ListStationSites(station.ListStationSitesParams{
			HTTPRequest: httpRequest,
			StationID:   testStationA,
		}, principal).(*station.ListStationSitesDefault)
		assert.True(ok)
		assert.Equal(station.NewListStationSitesDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestStation_StationForceSignIn(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("POST", "/station/{stationID}/sign-in", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.StationForceSignInParams
		principal *models.Principal
	}
	tests := []struct {
		name   string
		args   args
		want   middleware.Responder
		script []mock.Script
	}{
		{
			name: "success separate",
			args: args{
				params: station.StationForceSignInParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationID,
					Body: station.StationForceSignInBody{
						SiteName: testSiteName1,
						Group:    1,
						WorkDate: strfmt.Date(testSchedulingDate),
					},
				},
				principal: principal,
			},
			want: station.NewStationForceSignInOK(),
			script: []mock.Script{
				{
					Name: mock.FuncSignInStation,
					Input: mock.Input{
						Request: mcom.SignInStationRequest{
							Station: testStationID,
							Site: mcomModels.SiteID{
								Name:  testSiteName1,
								Index: 0,
							},
							Group:    1,
							WorkDate: time.Time(testSchedulingDate),
						},
						Options: []interface{}{mcom.ForceSignIn(), mcom.CreateSiteIfNotExists()},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "success together",
			args: args{
				params: station.StationForceSignInParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationID,
					Body: station.StationForceSignInBody{
						SiteName: "",
						Group:    1,
						WorkDate: strfmt.Date(testSchedulingDate),
					},
				},
				principal: principal,
			},
			want: station.NewStationForceSignInOK(),
			script: []mock.Script{
				{
					Name: mock.FuncSignInStation,
					Input: mock.Input{
						Request: mcom.SignInStationRequest{
							Station: testStationID,
							Site: mcomModels.SiteID{
								Name:  "",
								Index: 0,
							},
							Group:    1,
							WorkDate: time.Time(testSchedulingDate),
						},
						Options: []interface{}{mcom.ForceSignIn(), mcom.CreateSiteIfNotExists()},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "station not found",
			args: args{
				params: station.StationForceSignInParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   "XXX",
					Body: station.StationForceSignInBody{
						SiteName: testSiteName1,
						Group:    1,
						WorkDate: strfmt.Date(testSchedulingDate),
					},
				},
				principal: principal,
			},
			want: station.NewStationForceSignInDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_STATION_NOT_FOUND),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncSignInStation,
					Input: mock.Input{
						Request: mcom.SignInStationRequest{
							Station: "XXX",
							Site: mcomModels.SiteID{
								Name:  testSiteName1,
								Index: 0,
							},
							Group:    1,
							WorkDate: time.Time(testSchedulingDate),
						},
						Options: []interface{}{mcom.ForceSignIn(), mcom.CreateSiteIfNotExists()},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_STATION_NOT_FOUND,
						},
					},
				},
			},
		},
		{
			name: "insufficient request",
			args: args{
				params: station.StationForceSignInParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationID,
					Body: station.StationForceSignInBody{
						Group:    1,
						WorkDate: strfmt.Date(testSchedulingDate),
					},
				},
				principal: principal,
			},
			want: station.NewStationForceSignInDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncSignInStation,
					Input: mock.Input{
						Request: mcom.SignInStationRequest{
							Station: testStationID,
							Site: mcomModels.SiteID{
								Name:  "",
								Index: 0,
							},
							Group:    1,
							WorkDate: time.Time(testSchedulingDate),
						},
						Options: []interface{}{mcom.ForceSignIn(), mcom.CreateSiteIfNotExists()},
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
				params: station.StationForceSignInParams{
					HTTPRequest: httpRequestWithHeader,
					StationID:   testStationID,
					Body: station.StationForceSignInBody{
						Group:    1,
						WorkDate: strfmt.Date(testSchedulingDate),
					},
				},
				principal: principal,
			},
			want: station.NewStationForceSignInDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncSignInStation,
					Input: mock.Input{
						Request: mcom.SignInStationRequest{
							Station: testStationID,
							Site: mcomModels.SiteID{
								Name:  "",
								Index: 0,
							},
							Group:    1,
							WorkDate: time.Time(testSchedulingDate),
						},
						Options: []interface{}{mcom.ForceSignIn(), mcom.CreateSiteIfNotExists()},
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

			s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.StationForceSignIn(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("StationForceSignIn() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}

	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.StationForceSignIn(station.StationForceSignInParams{
			HTTPRequest: httpRequestWithHeader,
			StationID:   testStationID,
			Body: station.StationForceSignInBody{
				Group:    1,
				WorkDate: strfmt.Date(testSchedulingDate),
			},
		}, principal).(*station.StationForceSignInDefault)
		assert.True(ok)
		assert.Equal(station.NewStationForceSignInDefault(http.StatusForbidden), rep)
	}
}

func TestStation_StationSignOut(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("POST", "/stations/sign-out", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    station.StationSignOutParams
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
				params: station.StationSignOutParams{
					HTTPRequest: httpRequestWithHeader,
					Body: station.StationSignOutBody{
						StationSites: []*models.SiteInfo{
							{
								StationID: testStationID,
								SiteName:  "",
							},
							{
								StationID: testStationA,
								SiteName:  testSiteName1,
							},
						},
					},
				},
				principal: principal,
			},
			want: station.NewStationSignOutOK(),
			script: []mock.Script{
				{
					Name: mock.FuncSignOutStations,
					Input: mock.Input{
						Request: mcom.SignOutStationsRequest{
							Sites: []mcomModels.UniqueSite{
								{
									Station: testStationID,
									SiteID: mcomModels.SiteID{
										Name:  "",
										Index: 0,
									},
								},
								{
									Station: testStationA,
									SiteID: mcomModels.SiteID{
										Name:  testSiteName1,
										Index: 0,
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
				params: station.StationSignOutParams{
					HTTPRequest: httpRequestWithHeader,
					Body: station.StationSignOutBody{
						StationSites: []*models.SiteInfo{},
					},
				},
				principal: principal,
			},
			want: station.NewStationSignOutDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
			script: []mock.Script{
				{
					Name: mock.FuncSignOutStations,
					Input: mock.Input{
						Request: mcom.SignOutStationsRequest{
							Sites: []mcomModels.UniqueSite{},
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
				params: station.StationSignOutParams{
					HTTPRequest: httpRequestWithHeader,
					Body: station.StationSignOutBody{
						StationSites: []*models.SiteInfo{
							{
								StationID: testStationID,
								SiteName:  "",
							},
						},
					},
				},
				principal: principal,
			},
			want: station.NewStationSignOutDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name: mock.FuncSignOutStations,
					Input: mock.Input{
						Request: mcom.SignOutStationsRequest{
							Sites: []mcomModels.UniqueSite{
								{
									Station: testStationID,
									SiteID: mcomModels.SiteID{
										Name:  "",
										Index: 0,
									},
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

			s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.StationSignOut(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("StationSignOut() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}

	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := NewStation(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.StationSignOut(station.StationSignOutParams{
			HTTPRequest: httpRequestWithHeader,
			Body: station.StationSignOutBody{
				StationSites: []*models.SiteInfo{
					{
						StationID: testStationID,
						SiteName:  "",
					},
				},
			},
		}, principal).(*station.StationSignOutDefault)
		assert.True(ok)
		assert.Equal(station.NewStationSignOutDefault(http.StatusForbidden), rep)
	}
}
