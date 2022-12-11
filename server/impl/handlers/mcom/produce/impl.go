package produce

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/shopspring/decimal"

	"gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"
	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	utilsResources "gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/internal/printer"
	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils/barcodes"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	mesModels "gitlab.kenda.com.tw/kenda/mui/server/mes"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/produce"
)

const (
	workOrderBatchStarted = workorder.BatchStatus_BATCH_STARTED
)

type Config struct {
	Printers map[string]string
	FontPath string
	MesPath  string
}

// Produce definitions
type Produce struct {
	dm mcom.DataManager

	config Config

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

//NewProduce returns Produce service.
func NewProduce(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool,
	config Config) service.Produce {

	return Produce{
		dm:            dm,
		config:        config,
		hasPermission: hasPermission,
	}
}

// FeedCollect implements.
func (p Produce) FeedCollect(params produce.FeedCollectParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_FEED_COLLECT, principal.Roles) {
		return produce.NewFeedCollectDefault(http.StatusForbidden)
	}
	date := time.Time(params.Body.Collect.WorkDate)

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	// Feed steps
	// Check work order
	getWorkOrder, err := p.dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{
		ID: params.WorkOrderID,
	})
	if err != nil {
		return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
	}

	// Check batch
	batch, err := p.dm.GetBatch(ctx, mcom.GetBatchRequest{
		WorkOrder: getWorkOrder.ID,
		Number:    int16(params.Body.Feed.Batch),
	})
	if err != nil {
		e, ok := mcomErrors.As(err)
		if !ok {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
		}
		// if batch not exist, create batch
		if e.Code == mcomErrors.Code_BATCH_NOT_FOUND {
			if err := p.dm.CreateBatch(ctx, mcom.CreateBatchRequest{
				WorkOrder: params.WorkOrderID,
				Number:    int16(params.Body.Feed.Batch),
				Status:    workOrderBatchStarted,
			}); err != nil {
				return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
			}
		} else {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
		}

		batch = mcom.GetBatchReply{
			Info: mcom.BatchInfo{
				WorkOrder: params.WorkOrderID,
				Number:    int16(params.Body.Feed.Batch),
				Status:    int32(workOrderBatchStarted),
			},
		}
	}

	// Check batch status
	if batch.Info.Status != int32(workorder.BatchStatus_BATCH_PREPARING) &&
		batch.Info.Status != int32(workorder.BatchStatus_BATCH_STARTED) {
		return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), mcomErrors.Error{
			Details: fmt.Sprintf("bad status: status=%s", workorder.BatchStatus(batch.Info.Status)),
		})
	}

	// Feed
	_, err = p.dm.Feed(ctx, mcom.FeedRequest{
		Batch: mcom.BatchID{
			WorkOrder: params.WorkOrderID,
			Number:    int16(params.Body.Feed.Batch),
		},
		FeedContent: parseFeedResource(params.Body.Feed.Source),
	})
	if err != nil {
		return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
	}

	// closing the batch
	if err := p.dm.UpdateBatch(ctx, mcom.UpdateBatchRequest{
		WorkOrder: params.WorkOrderID,
		Number:    int16(params.Body.Feed.Batch),
		Status:    workorder.BatchStatus_BATCH_CLOSING,
	}); err != nil {
		return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
	}

	// Collect steps
	// Get station code
	getStation, err := p.dm.GetStation(ctx, mcom.GetStationRequest{
		ID: params.Body.StationID,
	})
	if err != nil {
		return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
	}
	stationCode := getStation.Information.Code
	if len(stationCode) == 1 {
		stationCode = "0" + stationCode
	}
	lotNumber := fmt.Sprintf("%d%s-%02d%02d", params.Body.Collect.Group, stationCode[0:2], date.Month(), date.Day())

	carrier := params.Body.Collect.CarrierResource
	if carrier != "" {
		// check if the carrier is existed
		_, err := p.dm.GetCarrier(ctx, mcom.GetCarrierRequest{
			ID: params.Body.Collect.CarrierResource,
		})
		if err != nil {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
		}
	}

	// Get resource oid
	now := time.Now()
	getLimitaryHour, err := p.dm.GetLimitaryHour(ctx, mcom.GetLimitaryHourRequest{ProductType: getWorkOrder.Product.Type})
	if err != nil {
		if e, ok := mcomErrors.As(err); !ok || e.Code != mcomErrors.Code_LIMITARY_HOUR_NOT_FOUND {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
		}
	}
	expiryTime := now.Add(time.Duration(getLimitaryHour.LimitaryHour.Max) * time.Hour)

	resources, err := p.dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{
			{
				ID:             getWorkOrder.Product.ID,
				Type:           getWorkOrder.Product.Type,
				Status:         utilsResources.MaterialStatus_AVAILABLE,
				Quantity:       decimal.NewFromFloat(params.Body.Collect.Quantity),
				Unit:           getWorkOrder.Unit,
				LotNumber:      lotNumber,
				ResourceID:     params.Body.Collect.ResourceID,
				CarrierID:      carrier,
				ProductionTime: now,
				ExpiryTime:     expiryTime,
			},
		},
	})
	if err != nil {
		return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
	}

	if carrier != "" {
		// replace the current resource in the carrier
		err = p.dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID:     carrier,
			Action: mcom.ClearResources{},
		})
		if err != nil {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
		}

		err = p.dm.UpdateCarrier(ctx, mcom.UpdateCarrierRequest{
			ID: carrier,
			Action: mcom.BindResources{
				ResourcesID: []string{resources[0].ID},
			},
		})
		if err != nil {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
		}
	}

	// Collect
	err = p.dm.CreateCollectRecord(ctx, mcom.CreateCollectRecordRequest{
		WorkOrder:   params.WorkOrderID,
		LotNumber:   lotNumber,
		Sequence:    int16(params.Body.Collect.Sequence),
		Quantity:    decimal.NewFromFloat(params.Body.Collect.Quantity),
		Station:     params.Body.StationID,
		ResourceOID: resources[0].OID,
	})
	if err != nil {
		return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
	}

	// Print
	if params.Body.Collect.Print {
		printData := printer.PrintData{
			StationID:      getWorkOrder.Station,
			NextStationID:  "",
			ProductID:      getWorkOrder.Product.ID,
			ProductionDate: now,
			ExpiryDate:     expiryTime,
			Quantity:       decimal.NewFromFloat(params.Body.Collect.Quantity),
			ResourceID:     params.Body.Collect.ResourceID,
		}

		// Read Config Printer
		pdf, err := printer.CreateResourcesPDF(ctx, models.MaterialResourceLabelFieldName{}, printData, barcodes.Code39{}, p.config.FontPath)
		if err != nil {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
		}

		if printer := p.config.Printers[getWorkOrder.Station]; printer == "" {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), mcomErrors.Error{
				Code:    mcomErrors.Code_STATION_PRINTER_NOT_DEFINED,
				Details: fmt.Sprintf("station %s no defined printer", getWorkOrder.Station),
			})
		}

		err = mcom.Print(ctx, p.config.Printers[getWorkOrder.Station], pdf)
		if err != nil {
			return utils.ParseError(ctx, produce.NewFeedCollectDefault(0), err)
		}
	}
	return produce.NewFeedCollectOK()
}

// MesFeed implements.
func (p Produce) MesFeed(params produce.MesFeedParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_MES_FEED, principal.Roles) {
		return produce.NewMesFeedDefault(http.StatusForbidden)
	}
	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	var (
		siteName = ""
	)
	mesFeedResponse := models.MesResponse{
		EnableForce: false,
		Success:     false,
		Error:       []*models.MesResponseErrorItems0{},
	}

	if p.config.MesPath == "" {
		return produce.NewMesFeedDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Details: "no mes path",
			})
	}

	// get station config
	config, err := p.dm.GetStationConfiguration(ctx, mcom.GetStationConfigurationRequest{
		StationID: params.StationID,
	})
	if err != nil {
		return utils.ParseError(ctx, produce.NewMesFeedDefault(0), err)
	}

	// get actionMode & accordingRecipe & siteName
	accordingRecipe := config.Feed.QuantitySource == stations.FeedQuantitySource_FROM_RECIPE
	actionMode := mesModels.CheckActionModeACTIONAUTO
	if params.Body.ForceFeed.Force {
		actionMode = mesModels.CheckActionModeACTIONFORCE
	}

	mesFeedRequest := mesModels.APIResourceFeedRequest{
		Check: &mesModels.CheckResourceAction{
			ActionMode: &actionMode,
			WorkOrder:  *params.Body.WorkOrderID,
			Batch:      *params.Body.Batch,
		},

		CloseBatch:      params.Body.CloseBatch,
		AccordingRecipe: accordingRecipe,
	}
	mesFeedResources, err := parseMesFeedResource(params.Body.Resource, accordingRecipe)
	if err != nil {
		return utils.ParseError(ctx, produce.NewMesFeedDefault(0), err)
	}
	mesFeedRequest.Feeds = mesFeedResources
	if config.SplitFeedAndCollect {
		siteName = config.Feed.OperatorSites[0].SiteID.Name
	}

	// send mes feed request to MES
	mesFeedPath := fmt.Sprintf("%s/mes/api/v2/resource/feed", p.config.MesPath)
	mesHeader := handlerUtils.MesHeader{
		UserID:  principal.ID,
		Station: params.StationID,
		Site:    siteName,
		TrackID: handlerUtils.GetContextValue(params.HTTPRequest, "rid"),
	}
	httpResponse, err := handlerUtils.SendMesPOSTRequest[*mesModels.APIResourceFeedReply](mesFeedRequest, mesHeader, mesFeedPath)
	if err != nil {
		return utils.ParseError(ctx, produce.NewMesFeedDefault(0), err)
	}

	mesFeedResponse.EnableForce = httpResponse.Enforceable
	if httpResponse.Results == nil || httpResponse.EnforceDone {
		mesFeedResponse.Success = true
	} else {
		mesFeedResponse.Error = parseMesFeedError(httpResponse.Results)
	}

	return produce.NewMesFeedOK().WithPayload(&produce.MesFeedOKBody{
		Data: &mesFeedResponse,
	})
}

// MesCollect implements.
func (p Produce) MesCollect(params produce.MesCollectParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_MES_COLLECT, principal.Roles) {
		return produce.NewMesCollectDefault(http.StatusForbidden)
	}
	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	mesCollectResponse := models.MesResponse{
		EnableForce: false,
		Success:     false,
		Error:       []*models.MesResponseErrorItems0{},
	}
	var (
		printResponse *produce.MesCollectOKBodyDataPrint
		siteName      = ""
	)

	if p.config.MesPath == "" {
		return produce.NewMesCollectDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Details: "no mes path",
		})
	}

	// get station config
	config, err := p.dm.GetStationConfiguration(ctx, mcom.GetStationConfigurationRequest{
		StationID: params.StationID,
	})
	if err != nil {
		return utils.ParseError(ctx, produce.NewMesCollectDefault(0), err)
	}

	// get actionMode & quantity
	actionMode := mesModels.CheckActionModeACTIONAUTO
	if params.Body.ForceCollect.Force {
		actionMode = mesModels.CheckActionModeACTIONFORCE
	}
	quantity, err := decimal.NewFromString(params.Body.Quantity)
	if err != nil {
		return utils.ParseError(ctx, produce.NewMesCollectDefault(0), err)
	}

	// get work order
	workOrder, err := p.dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{
		ID: *params.Body.WorkOrderID,
	})
	if err != nil {
		return utils.ParseError(ctx, produce.NewMesCollectDefault(0), err)
	}

	mesCollectRequest := mesModels.APICollectRequest{
		ActionMode: &actionMode,
		WorkOrder:  *params.Body.WorkOrderID,
		Carrier:    params.Body.CarrierResource,
		Sequence:   int32(*params.Body.Sequence),
		Quantity: &mesModels.V2commonsDecimal{
			Exp:   quantity.Exponent(),
			Value: quantity.Coefficient().String(),
		},
		Resource: &mesModels.ResourceID{
			ID: params.Body.ResourceID,
		},
		Feed: &mesModels.CollectRequestFeed{
			AccordingRecipe: true,
			Feeds:           []*mesModels.APIResourceFeed{},
		},
		LabelFields: []string{
			"manufacture_date",
			"expiry",
		},
	}

	if config.SplitFeedAndCollect {
		siteName = config.Collect.OperatorSites[0].SiteID.Name
	} else {
		mesFeedResources := make([]*mesModels.APIResourceFeed, len(params.Body.FeedResourceIDs))
		for i, data := range params.Body.FeedResourceIDs {
			mesFeedResources[i] = &mesModels.APIResourceFeed{
				Resource: &mesModels.APIResourceFeedResource{
					ID: data,
				},
			}
		}
		mesCollectRequest.Feed = &mesModels.CollectRequestFeed{
			Feeds:           mesFeedResources,
			AccordingRecipe: true,
		}
	}

	// send mes collect request to MES
	mesCollectPath := fmt.Sprintf("%s/mes/api/v2/resource/collect", p.config.MesPath)
	mesHeader := handlerUtils.MesHeader{
		UserID:  principal.ID,
		Station: params.StationID,
		Site:    siteName,
		TrackID: handlerUtils.GetContextValue(params.HTTPRequest, "rid"),
	}
	httpResponse, err := handlerUtils.SendMesPOSTRequest[*mesModels.APICollectReply](mesCollectRequest, mesHeader, mesCollectPath)
	if err != nil {
		return utils.ParseError(ctx, produce.NewMesFeedDefault(0), err)
	}

	mesCollectResponse.EnableForce = httpResponse.Enforceable
	if checkMesCollectError(httpResponse.Error) || httpResponse.EnforceDone {
		mesCollectResponse.Success = true
	} else {
		if *httpResponse.Error.Code == mesModels.ErrorCodeERRORINTERNAL {
			return produce.NewMesCollectDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: "mes internal error",
			})
		}
		mesCollectResponse.Error = []*models.MesResponseErrorItems0{
			{
				Code:    utils.ParseMesErrorCode(string(*httpResponse.Error.Code)),
				Details: httpResponse.Error.Details,
			},
		}
	}

	if mesCollectResponse.Success {
		// Print
		if params.Body.Print {
			// parse mesTime
			mesTime, err := parseMesTime([]string{"manufacture_date", "expiry"}, httpResponse.LabelFields)
			if err != nil {
				printResponse = &produce.MesCollectOKBodyDataPrint{
					Success: false,
					Error: &models.ErrorResponse{
						Code:    int64(mcomErrors.Code_FAILED_TO_PRINT_RESOURCE),
						Details: "mes time parse fail.",
					},
				}
			} else {
				printData := printer.PrintData{
					StationID:      params.StationID,
					NextStationID:  "",
					ProductID:      workOrder.Product.ID,
					ProductionDate: mesTime["manufacture_date"],
					ExpiryDate:     mesTime["expiry"],
					Quantity:       quantity,
					ResourceID:     params.Body.ResourceID,
				}

				// create pdf
				pdf, err := printer.CreateResourcesPDF(ctx, models.MaterialResourceLabelFieldName{}, printData, barcodes.Code39{}, p.config.FontPath)
				if err != nil {
					printResponse = &produce.MesCollectOKBodyDataPrint{
						Success: false,
						Error: &models.ErrorResponse{
							Code:    int64(mcomErrors.Code_FAILED_TO_PRINT_RESOURCE),
							Details: "create pdf fail.",
						},
					}
				} else {
					// check printer
					if printer := p.config.Printers[params.StationID]; printer == "" {
						printResponse = &produce.MesCollectOKBodyDataPrint{
							Success: false,
							Error: &models.ErrorResponse{
								Code:    int64(mcomErrors.Code_STATION_PRINTER_NOT_DEFINED),
								Details: fmt.Sprintf("station %s no defined printer", params.StationID),
							},
						}
					} else {
						// print pdf
						err = mcom.Print(ctx, p.config.Printers[params.StationID], pdf)
						if err != nil {
							printResponse = &produce.MesCollectOKBodyDataPrint{
								Success: false,
								Error: &models.ErrorResponse{
									Code:    int64(mcomErrors.Code_FAILED_TO_PRINT_RESOURCE),
									Details: "print fail.",
								},
							}
						} else {
							printResponse = &produce.MesCollectOKBodyDataPrint{
								Success: true,
							}
						}
					}
				}
			}
		}
	}
	return produce.NewMesCollectOK().WithPayload(&produce.MesCollectOKBody{
		Data: &produce.MesCollectOKBodyData{
			MesResponse: &mesCollectResponse,
			Print:       printResponse,
		},
	})
}

func parseFeedResource(dataIn []*produce.FeedCollectParamsBodyFeedSourceItems0) []mcom.FeedPerSite {
	dataOut := make([]mcom.FeedPerSite, len(dataIn))
	for i, data := range dataIn {
		dataOut[i] = mcom.FeedPerSiteType1{
			Site: mcomModels.UniqueSite{
				Station: data.SiteInfo.StationID,
				SiteID: mcomModels.SiteID{
					Name:  data.SiteInfo.SiteName,
					Index: int16(data.SiteInfo.SiteIndex),
				},
			},
			Quantity: decimal.NewFromFloat(data.Quantity),
		}
	}
	return dataOut
}

func parseMesFeedResource(dataIn []*models.FeedResource, accordingRecipe bool) ([]*mesModels.APIResourceFeed, error) {
	dataOut := make([]*mesModels.APIResourceFeed, len(dataIn))
	for i, data := range dataIn {
		dataOut[i] = &mesModels.APIResourceFeed{
			Resource: &mesModels.APIResourceFeedResource{
				ID: data.ID,
			},
		}
		if !accordingRecipe {
			quantity, err := decimal.NewFromString(data.Quantity)
			if err != nil {
				return dataOut, err
			}
			if quantity.LessThanOrEqual(decimal.Zero) {
				return dataOut, mcomErrors.Error{
					Code:    mcomErrors.Code_INVALID_NUMBER,
					Details: fmt.Sprintf("invalid_number = %s", data.Quantity),
				}
			}
			dataOut[i].Quantity = &mesModels.V2commonsDecimal{
				Exp:   quantity.Exponent(),
				Value: quantity.Coefficient().String(),
			}
		}
	}
	return dataOut, nil
}

func parseMesFeedError(mesResponse []*mesModels.V2apiResourceResult) []*models.MesResponseErrorItems0 {
	dataOut := make([]*models.MesResponseErrorItems0, len(mesResponse))
	for i, errorString := range mesResponse {
		dataOut[i] = &models.MesResponseErrorItems0{
			Code:    utils.ParseMesErrorCode(string(*errorString.Error.Code)),
			Details: errorString.Error.Details,
		}
	}

	return dataOut
}

func parseMesTime(mesLabels []string, mesLabelFields map[string]mesModels.CollectReplylabelField) (map[string]time.Time, error) {
	dataOut := map[string]time.Time{}
	for _, data := range mesLabels {
		timeNano, err := strconv.ParseInt(mesLabelFields[data].Time.Nano, 10, 64)
		if err != nil {
			return nil, err
		}
		dataOut[data] = time.Unix(0, timeNano)
	}

	return dataOut, nil
}

func checkMesCollectError(data *mesModels.CheckError) bool {
	empty := mesModels.CheckError{}
	if data == nil || data == &empty || *data.Code == mesModels.ErrorCodeERRORNONE {
		return true
	}
	return false
}
