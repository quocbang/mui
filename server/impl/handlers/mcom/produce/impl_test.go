package produce

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/produce"
)

const (
	userID = "tester"

	testInternalServerError = "internal error"
	remarkNone              = 0
)

var (
	principal = &models.Principal{
		ID: userID,
		Roles: []models.Role{
			models.Role(mcomRoles.Role_ADMINISTRATOR),
			models.Role(mcomRoles.Role_LEADER),
		},
	}
	testStationA   = "STATION-A"
	testWorkOrder1 = "WORKORDERID001"
	testSiteName1  = "TESTSITENAME1"

	testWorkOrder1ProductA    = "PRODUCT-A"
	testWorkOrder1ProductType = "RUBBER"

	testSchedulingDate  = time.Date(2021, 8, 9, 0, 0, 0, 0, time.Local)
	testSequence        = 99
	testQuantity        = decimal.Decimal(decimal.NewFromFloat(79.21))
	testResourceID      = "TESTRESOURCEID"
	testWorkOrderError1 = "WORKORDERERRORID"
	testStationErrorID  = "TESTSTATIONERRORNAME1"
	testSiteErrorName1  = "TESTSITEERRORNAME1"
)

func TestProduce_FeedCollect(t *testing.T) {
	var (
		testBatch             = 19
		testCarrierResourceID = "CARRIERRESOURCEID"
		testCarrierID         = "CARRIERID"
		testUnit              = "UNIT"
		testResourceOID       = "RESOURCEOID"
		date                  = time.Time(testSchedulingDate)
		testLotNumber         = fmt.Sprintf("%1s%s-%02d%02d", "1", "99", date.Month(), date.Day())
		testLimitaryHourMin   = 0
		testLimitaryHourMax   = 144
	)

	timeNow := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return timeNow
	})
	defer monkey.UnpatchAll()
	expiryTime := timeNow.Add(time.Duration(testLimitaryHourMax) * time.Hour)
	assert := assert.New(t)

	httpRequest := httptest.NewRequest("POST", "/production-flow/feed-collect/work-order/{workOrderID}", nil)
	httpRequest.Header.Set(account.AuthorizationKey, "token-for-tester")

	type args struct {
		params    produce.FeedCollectParams
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
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectOK(),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workOrderBatchStarted),
							},
						},
					},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteName1,
											Index: 0,
										},
										Station: testStationA,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{
							FeedRecordID: "my-feed-id",
						},
					},
				},
				{
					Name: mock.FuncUpdateBatch,
					Input: mock.Input{
						Request: mcom.UpdateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workorder.BatchStatus_BATCH_CLOSING,
						},
					},
					Output: mock.Output{
						Error: nil,
					},
				},
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationReply{
							Information: mcom.StationInformation{
								Code: "99",
							},
						},
					},
				},
				{
					Name: mock.FuncGetCarrier,
					Input: mock.Input{
						Request: mcom.GetCarrierRequest{
							ID: testCarrierResourceID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetCarrierReply{
							ID: testCarrierID,
						},
					},
				},
				{
					Name: mock.FuncGetLimitaryHour,
					Input: mock.Input{
						Request: mcom.GetLimitaryHourRequest{
							ProductType: testWorkOrder1ProductType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetLimitaryHourReply{
							LimitaryHour: mcom.LimitaryHourParameter{
								Min: int32(testLimitaryHourMin),
								Max: int32(testLimitaryHourMax),
							},
						},
					},
				},
				{
					Name: mock.FuncCreateMaterialResources,
					Input: mock.Input{
						Request: mcom.CreateMaterialResourcesRequest{
							Materials: []mcom.CreateMaterialResourcesRequestDetail{{
								Type:           testWorkOrder1ProductType,
								ID:             testWorkOrder1ProductA,
								Status:         resources.MaterialStatus_AVAILABLE,
								Quantity:       testQuantity,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ResourceID:     testResourceID,
								CarrierID:      testCarrierResourceID,
								ProductionTime: timeNow,
								ExpiryTime:     expiryTime,
							}},
						},
					},
					Output: mock.Output{
						Response: mcom.CreateMaterialResourcesReply{
							{
								ID:  testResourceID,
								OID: testResourceOID,
							},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID:     testCarrierResourceID,
							Action: mcom.ClearResources{},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID: testCarrierResourceID,
							Action: mcom.BindResources{
								ResourcesID: []string{testResourceID},
							},
						},
					},
				},
				{
					Name: mock.FuncCreateCollectRecord,
					Input: mock.Input{
						Request: mcom.CreateCollectRecordRequest{
							Sequence:    int16(testSequence),
							LotNumber:   testLotNumber,
							WorkOrder:   testWorkOrder1,
							Station:     testStationA,
							Quantity:    testQuantity,
							ResourceOID: testResourceOID,
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "not batch, create batch, success",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectOK(),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_BATCH_NOT_FOUND,
						},
					},
				},
				{
					Name: mock.FuncCreateBatch,
					Input: mock.Input{
						Request: mcom.CreateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workOrderBatchStarted,
						},
					},
					Output: mock.Output{},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteName1,
											Index: 0,
										},
										Station: testStationA,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{
							FeedRecordID: "my-feed-id",
						},
					},
				},
				{
					Name: mock.FuncUpdateBatch,
					Input: mock.Input{
						Request: mcom.UpdateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workorder.BatchStatus_BATCH_CLOSING,
						},
					},
					Output: mock.Output{
						Error: nil,
					},
				},
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationReply{
							Information: mcom.StationInformation{
								Code: "99",
							},
						},
					},
				},
				{
					Name: mock.FuncGetCarrier,
					Input: mock.Input{
						Request: mcom.GetCarrierRequest{
							ID: testCarrierResourceID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetCarrierReply{
							ID: testCarrierID,
						},
					},
				},
				{
					Name: mock.FuncGetLimitaryHour,
					Input: mock.Input{
						Request: mcom.GetLimitaryHourRequest{
							ProductType: testWorkOrder1ProductType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetLimitaryHourReply{
							LimitaryHour: mcom.LimitaryHourParameter{
								Min: int32(testLimitaryHourMin),
								Max: int32(testLimitaryHourMax),
							},
						},
					},
				},
				{
					Name: mock.FuncCreateMaterialResources,
					Input: mock.Input{
						Request: mcom.CreateMaterialResourcesRequest{
							Materials: []mcom.CreateMaterialResourcesRequestDetail{{
								Type:           testWorkOrder1ProductType,
								ID:             testWorkOrder1ProductA,
								Status:         resources.MaterialStatus_AVAILABLE,
								Quantity:       testQuantity,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ResourceID:     testResourceID,
								CarrierID:      testCarrierResourceID,
								ProductionTime: timeNow,
								ExpiryTime:     expiryTime,
							}},
						},
					},
					Output: mock.Output{
						Response: mcom.CreateMaterialResourcesReply{
							{
								ID:  testResourceID,
								OID: testResourceOID,
							},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID:     testCarrierResourceID,
							Action: mcom.ClearResources{},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID: testCarrierResourceID,
							Action: mcom.BindResources{
								ResourcesID: []string{testResourceID},
							},
						},
					},
				},
				{
					Name: mock.FuncCreateCollectRecord,
					Input: mock.Input{
						Request: mcom.CreateCollectRecordRequest{
							Sequence:    int16(testSequence),
							LotNumber:   testLotNumber,
							WorkOrder:   testWorkOrder1,
							Station:     testStationA,
							Quantity:    testQuantity,
							ResourceOID: testResourceOID,
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "not found work order",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrderError1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
				Details: "work order not found",
			}),
			script: []mock.Script{
				{ // fail of work order not found
					Name: mock.FuncGetWorkOrder,
					Input: mock.Input{
						Request: mcom.GetWorkOrderRequest{
							ID: testWorkOrderError1,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code:    mcomErrors.Code_WORKORDER_NOT_FOUND,
							Details: "work order not found",
						}},
				},
			},
		},
		{
			name: "batch already exists",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_BATCH_ALREADY_EXISTS),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_BATCH_NOT_FOUND,
						},
					},
				},
				{
					Name: mock.FuncCreateBatch,
					Input: mock.Input{
						Request: mcom.CreateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workOrderBatchStarted,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_BATCH_ALREADY_EXISTS,
						},
					},
				},
			},
		},
		{
			name: "batch status not preparing or started",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Details: fmt.Sprintf("bad status: status=%s", workorder.BatchStatus_BATCH_CANCELLED),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workorder.BatchStatus_BATCH_CANCELLED),
							},
						},
					},
				},
			},
		},
		{
			name: "not found station site",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationErrorID,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationErrorID,
										SiteName:  testSiteErrorName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:      1,
							WorkDate:   strfmt.Date(testSchedulingDate),
							ResourceID: testResourceID,
							Sequence:   int64(testSequence),
							Quantity:   testQuantity.InexactFloat64(),
							Print:      false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_STATION_SITE_NOT_FOUND),
				Details: "station site not found",
			}),
			script: []mock.Script{
				{ // fail feed of station site not found
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workOrderBatchStarted),
							},
						},
					},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteErrorName1,
											Index: 0,
										},
										Station: testStationErrorID,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{},
						Error: mcomErrors.Error{
							Code:    mcomErrors.Code_STATION_SITE_NOT_FOUND,
							Details: "station site not found",
						}},
				},
			},
		},
		{
			name: "station not found",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_STATION_NOT_FOUND),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workOrderBatchStarted),
							},
						},
					},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteName1,
											Index: 0,
										},
										Station: testStationA,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{
							FeedRecordID: "my-feed-id",
						},
					},
				},
				{
					Name: mock.FuncUpdateBatch,
					Input: mock.Input{
						Request: mcom.UpdateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workorder.BatchStatus_BATCH_CLOSING,
						},
					},
					Output: mock.Output{
						Error: nil,
					},
				},
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
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
			name: "carrier not found",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_CARRIER_NOT_FOUND),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workOrderBatchStarted),
							},
						},
					},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteName1,
											Index: 0,
										},
										Station: testStationA,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{
							FeedRecordID: "my-feed-id",
						},
					},
				},
				{
					Name: mock.FuncUpdateBatch,
					Input: mock.Input{
						Request: mcom.UpdateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workorder.BatchStatus_BATCH_CLOSING,
						},
					},
					Output: mock.Output{
						Error: nil,
					},
				},
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationReply{
							Information: mcom.StationInformation{
								Code: "99",
							},
						},
					},
				},
				{
					Name: mock.FuncGetCarrier,
					Input: mock.Input{
						Request: mcom.GetCarrierRequest{
							ID: testCarrierResourceID,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_CARRIER_NOT_FOUND,
						},
					},
				},
			},
		},
		{
			name: "resource existed",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_RESOURCE_EXISTED),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workOrderBatchStarted),
							},
						},
					},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteName1,
											Index: 0,
										},
										Station: testStationA,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{
							FeedRecordID: "my-feed-id",
						},
					},
				},
				{
					Name: mock.FuncUpdateBatch,
					Input: mock.Input{
						Request: mcom.UpdateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workorder.BatchStatus_BATCH_CLOSING,
						},
					},
					Output: mock.Output{
						Error: nil,
					},
				},
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationReply{
							Information: mcom.StationInformation{
								Code: "99",
							},
						},
					},
				},
				{
					Name: mock.FuncGetCarrier,
					Input: mock.Input{
						Request: mcom.GetCarrierRequest{
							ID: testCarrierResourceID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetCarrierReply{
							ID: testCarrierID,
						},
					},
				},
				{
					Name: mock.FuncGetLimitaryHour,
					Input: mock.Input{
						Request: mcom.GetLimitaryHourRequest{
							ProductType: testWorkOrder1ProductType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetLimitaryHourReply{
							LimitaryHour: mcom.LimitaryHourParameter{
								Min: int32(testLimitaryHourMin),
								Max: int32(testLimitaryHourMax),
							},
						},
					},
				},
				{
					Name: mock.FuncCreateMaterialResources,
					Input: mock.Input{
						Request: mcom.CreateMaterialResourcesRequest{
							Materials: []mcom.CreateMaterialResourcesRequestDetail{{
								Type:           testWorkOrder1ProductType,
								ID:             testWorkOrder1ProductA,
								Status:         resources.MaterialStatus_AVAILABLE,
								Quantity:       testQuantity,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ResourceID:     testResourceID,
								CarrierID:      testCarrierResourceID,
								ProductionTime: timeNow,
								ExpiryTime:     expiryTime,
							}},
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_RESOURCE_EXISTED,
						},
					},
				},
			},
		},
		{
			name: "not assign resource id",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      "",
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectOK(),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workOrderBatchStarted),
							},
						},
					},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteName1,
											Index: 0,
										},
										Station: testStationA,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{
							FeedRecordID: "my-feed-id",
						},
					},
				},
				{
					Name: mock.FuncUpdateBatch,
					Input: mock.Input{
						Request: mcom.UpdateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workorder.BatchStatus_BATCH_CLOSING,
						},
					},
					Output: mock.Output{
						Error: nil,
					},
				},
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationReply{
							Information: mcom.StationInformation{
								Code: "99",
							},
						},
					},
				},
				{
					Name: mock.FuncGetCarrier,
					Input: mock.Input{
						Request: mcom.GetCarrierRequest{
							ID: testCarrierResourceID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetCarrierReply{
							ID: testCarrierID,
						},
					},
				},
				{
					Name: mock.FuncGetLimitaryHour,
					Input: mock.Input{
						Request: mcom.GetLimitaryHourRequest{
							ProductType: testWorkOrder1ProductType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetLimitaryHourReply{
							LimitaryHour: mcom.LimitaryHourParameter{
								Min: int32(testLimitaryHourMin),
								Max: int32(testLimitaryHourMax),
							},
						},
					},
				},
				{
					Name: mock.FuncCreateMaterialResources,
					Input: mock.Input{
						Request: mcom.CreateMaterialResourcesRequest{
							Materials: []mcom.CreateMaterialResourcesRequestDetail{{
								Type:           testWorkOrder1ProductType,
								ID:             testWorkOrder1ProductA,
								Status:         resources.MaterialStatus_AVAILABLE,
								Quantity:       testQuantity,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ResourceID:     "",
								CarrierID:      testCarrierResourceID,
								ProductionTime: timeNow,
								ExpiryTime:     expiryTime,
							}},
						},
					},
					Output: mock.Output{
						Response: mcom.CreateMaterialResourcesReply{
							{
								ID:  testResourceID,
								OID: testResourceOID,
							},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID:     testCarrierResourceID,
							Action: mcom.ClearResources{},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID: testCarrierResourceID,
							Action: mcom.BindResources{
								ResourcesID: []string{testResourceID},
							},
						},
					},
				},
				{
					Name: mock.FuncCreateCollectRecord,
					Input: mock.Input{
						Request: mcom.CreateCollectRecordRequest{
							Sequence:    int16(testSequence),
							LotNumber:   testLotNumber,
							WorkOrder:   testWorkOrder1,
							Station:     testStationA,
							Quantity:    testQuantity,
							ResourceOID: testResourceOID,
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "record already exists",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_RECORD_ALREADY_EXISTS),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workOrderBatchStarted),
							},
						},
					},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteName1,
											Index: 0,
										},
										Station: testStationA,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{
							FeedRecordID: "my-feed-id",
						},
					},
				},
				{
					Name: mock.FuncUpdateBatch,
					Input: mock.Input{
						Request: mcom.UpdateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workorder.BatchStatus_BATCH_CLOSING,
						},
					},
					Output: mock.Output{
						Error: nil,
					},
				},
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationReply{
							Information: mcom.StationInformation{
								Code: "99",
							},
						},
					},
				},
				{
					Name: mock.FuncGetCarrier,
					Input: mock.Input{
						Request: mcom.GetCarrierRequest{
							ID: testCarrierResourceID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetCarrierReply{
							ID: testCarrierID,
						},
					},
				},
				{
					Name: mock.FuncGetLimitaryHour,
					Input: mock.Input{
						Request: mcom.GetLimitaryHourRequest{
							ProductType: testWorkOrder1ProductType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetLimitaryHourReply{
							LimitaryHour: mcom.LimitaryHourParameter{
								Min: int32(testLimitaryHourMin),
								Max: int32(testLimitaryHourMax),
							},
						},
					},
				},
				{
					Name: mock.FuncCreateMaterialResources,
					Input: mock.Input{
						Request: mcom.CreateMaterialResourcesRequest{
							Materials: []mcom.CreateMaterialResourcesRequestDetail{{
								Type:           testWorkOrder1ProductType,
								ID:             testWorkOrder1ProductA,
								Status:         resources.MaterialStatus_AVAILABLE,
								Quantity:       testQuantity,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ResourceID:     testResourceID,
								CarrierID:      testCarrierResourceID,
								ProductionTime: timeNow,
								ExpiryTime:     expiryTime,
							}},
						},
					},
					Output: mock.Output{
						Response: mcom.CreateMaterialResourcesReply{
							{
								ID:  testResourceID,
								OID: testResourceOID,
							},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID:     testCarrierResourceID,
							Action: mcom.ClearResources{},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID: testCarrierResourceID,
							Action: mcom.BindResources{
								ResourcesID: []string{testResourceID},
							},
						},
					},
				},
				{
					Name: mock.FuncCreateCollectRecord,
					Input: mock.Input{
						Request: mcom.CreateCollectRecordRequest{
							Sequence:    int16(testSequence),
							LotNumber:   testLotNumber,
							WorkOrder:   testWorkOrder1,
							Station:     testStationA,
							Quantity:    testQuantity,
							ResourceOID: testResourceOID,
						},
					},
					Output: mock.Output{
						Error: mcomErrors.Error{
							Code: mcomErrors.Code_RECORD_ALREADY_EXISTS,
						},
					},
				},
			},
		},
		{
			name: "limitary hour not found, success",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectOK(),
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
							Unit:   testUnit,
							Status: workorder.Status_ACTIVE,
						},
					},
				},
				{
					Name: mock.FuncGetBatch,
					Input: mock.Input{
						Request: mcom.GetBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
						},
					},
					Output: mock.Output{
						Response: mcom.GetBatchReply{
							Info: mcom.BatchInfo{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
								Status:    int32(workOrderBatchStarted),
							},
						},
					},
				},
				{
					Name: mock.FuncFeed,
					Input: mock.Input{
						Request: mcom.FeedRequest{
							Batch: mcom.BatchID{
								WorkOrder: testWorkOrder1,
								Number:    int16(testBatch),
							},
							FeedContent: []mcom.FeedPerSite{
								mcom.FeedPerSiteType1{
									Site: mcomModels.UniqueSite{
										SiteID: mcomModels.SiteID{
											Name:  testSiteName1,
											Index: 0,
										},
										Station: testStationA,
									},
									Quantity: testQuantity,
								},
							},
						},
					},
					Output: mock.Output{
						Response: mcom.FeedReply{
							FeedRecordID: "my-feed-id",
						},
					},
				},
				{
					Name: mock.FuncUpdateBatch,
					Input: mock.Input{
						Request: mcom.UpdateBatchRequest{
							WorkOrder: testWorkOrder1,
							Number:    int16(testBatch),
							Status:    workorder.BatchStatus_BATCH_CLOSING,
						},
					},
					Output: mock.Output{
						Error: nil,
					},
				},
				{
					Name: mock.FuncGetStation,
					Input: mock.Input{
						Request: mcom.GetStationRequest{
							ID: testStationA,
						},
					},
					Output: mock.Output{
						Response: mcom.GetStationReply{
							Information: mcom.StationInformation{
								Code: "99",
							},
						},
					},
				},
				{
					Name: mock.FuncGetCarrier,
					Input: mock.Input{
						Request: mcom.GetCarrierRequest{
							ID: testCarrierResourceID,
						},
					},
					Output: mock.Output{
						Response: mcom.GetCarrierReply{
							ID: testCarrierID,
						},
					},
				},
				{
					Name: mock.FuncGetLimitaryHour,
					Input: mock.Input{
						Request: mcom.GetLimitaryHourRequest{
							ProductType: testWorkOrder1ProductType,
						},
					},
					Output: mock.Output{
						Response: mcom.GetLimitaryHourReply{},
						Error:    mcomErrors.Error{Code: mcomErrors.Code_LIMITARY_HOUR_NOT_FOUND},
					},
				},
				{
					Name: mock.FuncCreateMaterialResources,
					Input: mock.Input{
						Request: mcom.CreateMaterialResourcesRequest{
							Materials: []mcom.CreateMaterialResourcesRequestDetail{{
								Type:           testWorkOrder1ProductType,
								ID:             testWorkOrder1ProductA,
								Status:         resources.MaterialStatus_AVAILABLE,
								Quantity:       testQuantity,
								Unit:           testUnit,
								LotNumber:      testLotNumber,
								ResourceID:     testResourceID,
								CarrierID:      testCarrierResourceID,
								ProductionTime: timeNow,
								ExpiryTime:     timeNow,
							}},
						},
					},
					Output: mock.Output{
						Response: mcom.CreateMaterialResourcesReply{
							{
								ID:  testResourceID,
								OID: testResourceOID,
							},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID:     testCarrierResourceID,
							Action: mcom.ClearResources{},
						},
					},
				},
				{
					Name: mock.FuncUpdateCarrier,
					Input: mock.Input{
						Request: mcom.UpdateCarrierRequest{
							ID: testCarrierResourceID,
							Action: mcom.BindResources{
								ResourcesID: []string{testResourceID},
							},
						},
					},
				},
				{
					Name: mock.FuncCreateCollectRecord,
					Input: mock.Input{
						Request: mcom.CreateCollectRecordRequest{
							Sequence:    int16(testSequence),
							LotNumber:   testLotNumber,
							WorkOrder:   testWorkOrder1,
							Station:     testStationA,
							Quantity:    testQuantity,
							ResourceOID: testResourceOID,
						},
					},
					Output: mock.Output{},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: produce.FeedCollectParams{
					HTTPRequest: httpRequest,
					WorkOrderID: testWorkOrder1,
					Body: produce.FeedCollectBody{
						StationID: testStationA,
						Feed: &produce.FeedCollectParamsBodyFeed{
							Batch: int64(testBatch),
							Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
								{
									SiteInfo: &models.SiteInfo{
										StationID: testStationA,
										SiteName:  testSiteName1,
										SiteIndex: 0,
									},
									Quantity: testQuantity.InexactFloat64(),
								},
							},
						},
						Collect: &produce.FeedCollectParamsBodyCollect{
							Group:           1,
							WorkDate:        strfmt.Date(testSchedulingDate),
							ResourceID:      testResourceID,
							CarrierResource: testCarrierResourceID,
							Sequence:        int64(testSequence),
							Quantity:        testQuantity.InexactFloat64(),
							Print:           false,
						},
					},
				},
				principal: principal,
			},
			want: produce.NewFeedCollectDefault(http.StatusInternalServerError).WithPayload(&models.Error{
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

			s := mustNewProduce(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.FeedCollect(tt.args.params, tt.args.principal); !assert.Equal(got, tt.want) {
				t.Errorf("FeedCollect() = %v, want %v", got, tt.want)
			}

			assert.NoErrorf(dm.Close(), tt.name)
		})
	}

	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := mustNewProduce(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := r.FeedCollect(produce.FeedCollectParams{
			HTTPRequest: httpRequest,
			WorkOrderID: testWorkOrder1,
			Body: produce.FeedCollectBody{
				StationID: testStationA,
				Feed: &produce.FeedCollectParamsBodyFeed{
					Batch: int64(testBatch),
					Source: []*produce.FeedCollectParamsBodyFeedSourceItems0{
						{
							SiteInfo: &models.SiteInfo{
								StationID: testStationA,
								SiteName:  testSiteName1,
								SiteIndex: 0,
							},
							Quantity: testQuantity.InexactFloat64(),
						},
					},
				},
				Collect: &produce.FeedCollectParamsBodyCollect{
					Group:      1,
					WorkDate:   strfmt.Date(testSchedulingDate),
					ResourceID: testResourceID,
					Sequence:   int64(testSequence),
					Quantity:   testQuantity.InexactFloat64(),
				},
			},
		}, principal).(*produce.FeedCollectDefault)
		assert.True(ok)
		assert.Equal(produce.NewFeedCollectDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func mustNewProduce(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool,
) service.Produce {
	s := NewProduce(dm, hasPermission, Config{FontPath: "fake-path"})
	return s
}
