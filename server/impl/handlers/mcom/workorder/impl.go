package workorder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/shopspring/decimal"
	excelize "github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"
	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	mcomWorkOrder "gitlab.kenda.com.tw/kenda/mcom/utils/workorder"

	"gitlab.kenda.com.tw/kenda/mui/server/configs"
	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	mesageModels "gitlab.kenda.com.tw/kenda/mui/server/models"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/work_order"
)

const (
	startWorkOrder = int64(0)
	closeWorkOrder = int64(1)

	colProductID            = 0
	colStation              = 1
	colVersionStage         = 2
	colRecipeID             = 3
	colProcessName          = 4
	colProcessType          = 5
	colExcelBatchSize       = 6
	colExcelBatchSizeData   = 7
	colExcelRecipeBatchSize = 8
	colDate                 = 9
)

type Config struct {
	StationFunctionConfig map[string]configs.FunctionAPIPath
}

// workorder definitions.
type WorkOrder struct {
	dm            mcom.DataManager
	config        Config
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

type Row []string

func (r Row) Column(i int) string {
	if len(r) <= i {
		return ""
	}
	return r[i]
}

// get excel column index
func excelColIndex(i int) string {
	return string(rune('A' + i))
}

// parse excel bad column
func excelBadColumns(header Row, colIndex []int) []string {
	dataOut := make([]string, len(colIndex))
	for i, index := range colIndex {
		dataOut[i] = fmt.Sprintf("%s(%s)", excelColIndex(index), header.Column(index))
	}
	return dataOut
}

// NewWorkOrder returns WorkOrder service.
func NewWorkOrder(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool,
	config Config) service.WorkOrder {
	return WorkOrder{
		dm:            dm,
		hasPermission: hasPermission,
		config:        config,
	}
}

// Create implements CreateStationScheduling.
func (w WorkOrder) CreateStationScheduling(params work_order.CreateStationSchedulingParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_CREATE_STATION_SCHEDULING, principal.Roles) {
		return work_order.NewCreateStationSchedulingDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	createList := make([]mcom.CreateWorkOrder, len(params.Body))
	for i, body := range params.Body {
		batchesQuantity, err := handlerUtils.ToDecimals(body.BatchesQuantity)
		if err != nil {
			return work_order.NewCreateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: fmt.Sprintf("invalid_numbers=%v", body.BatchesQuantity),
			})
		}
		createList[i] = mcom.CreateWorkOrder{
			ProcessOID:   body.Recipe.ProcessOID,
			RecipeID:     body.Recipe.ID,
			ProcessName:  body.Recipe.ProcessName,
			ProcessType:  body.Recipe.ProcessType,
			DepartmentID: *body.DepartmentOID,
			Station:      *body.Station,
			Date:         time.Time(*body.PlanDate),
			Parent:       body.ParentID,
		}
		switch body.BatchSize {
		case int64(mcomWorkOrder.BatchSize_PER_BATCH_QUANTITIES):
			if len(batchesQuantity) == 0 {
				return work_order.NewCreateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
					Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				})
			}
			createList[i].BatchesQuantity = mcom.NewQuantityPerBatch(batchesQuantity)

		case int64(mcomWorkOrder.BatchSize_FIXED_QUANTITY),
			int64(mcomWorkOrder.BatchSize_PLAN_QUANTITY):
			if body.BatchCount == 0 {
				return work_order.NewCreateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
					Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				})
			}

			planQuantity, err := decimal.NewFromString(body.PlanQuantity)
			if err != nil {
				return work_order.NewCreateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
					Code:    int64(mcomErrors.Code_INVALID_NUMBER),
					Details: "invalid_number=" + body.PlanQuantity,
				})
			}

			if body.BatchSize == int64(mcomWorkOrder.BatchSize_FIXED_QUANTITY) {
				createList[i].BatchesQuantity = mcom.NewFixedQuantity(uint(body.BatchCount), planQuantity)
			} else {
				createList[i].BatchesQuantity = mcom.NewPlanQuantity(uint(body.BatchCount), planQuantity)
			}

		default:
			return work_order.NewCreateStationSchedulingDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: fmt.Sprintf("no implementation with %d of BatchSize", body.BatchSize),
			})
		}
	}

	workorderIDs, err := w.dm.CreateWorkOrders(ctx,
		mcom.CreateWorkOrdersRequest{
			WorkOrders: createList,
		})
	if err != nil {
		return utils.ParseError(ctx, work_order.NewCreateStationSchedulingDefault(0), err)
	}

	return work_order.NewCreateStationSchedulingOK().WithPayload(&work_order.CreateStationSchedulingOKBody{Data: workorderIDs.IDs})
}

// Update implements UpdateStationScheduling.
func (w WorkOrder) UpdateStationScheduling(params work_order.UpdateStationSchedulingParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_UPDATE_STATION_SCHEDULING, principal.Roles) {
		return work_order.NewUpdateStationSchedulingDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	updateList := make([]mcom.UpdateWorkOrder, len(params.Body))
	for i, body := range params.Body {
		if *body.ForceToAbort {
			updateList[i] = mcom.UpdateWorkOrder{
				ID:     *body.ID,
				Status: workorder.Status_SKIPPED,
			}
			continue
		}
		updateList[i] = mcom.UpdateWorkOrder{
			ID:       *body.ID,
			Sequence: int32(*body.Sequence),
		}
	}

	if err := w.dm.UpdateWorkOrders(ctx, mcom.UpdateWorkOrdersRequest{Orders: updateList}); err != nil {
		return utils.ParseError(ctx, work_order.NewUpdateStationSchedulingDefault(0), err)
	}
	return work_order.NewUpdateStationSchedulingOK()
}

// GetStationScheduling implements.
func (w WorkOrder) GetStationScheduling(params work_order.GetStationSchedulingParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_GET_STATION_SCHEDULING, principal.Roles) {
		return work_order.NewGetStationSchedulingDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := w.dm.ListWorkOrdersByDuration(ctx, mcom.ListWorkOrdersByDurationRequest{
		Since: time.Time(params.Date),
		Until: time.Time(params.Date),
		// The function is to find the work order of the day, so it needs "Until" to limit the search range.
		Station: params.Station,
	}.WithOrder(
		mcom.Order{
			Name:       "reserved_date",
			Descending: false,
		},
		mcom.Order{
			Name:       "reserved_sequence",
			Descending: false,
		},
	))
	if err != nil {
		return utils.ParseError(ctx, work_order.NewGetStationSchedulingDefault(0), err)
	}
	data := make(models.WorkOrders, len(list.Contents))
	for i, wo := range list.Contents {
		data[i] = &models.WorkOrder{
			ID:            wo.ID,
			BatchSize:     int64(wo.BatchQuantityType),
			DepartmentOID: wo.DepartmentID,
			PlanDate:      strfmt.Date(wo.Date),
			ProductID:     wo.Product.ID,
			Recipe: &models.Recipe{
				ProcessName: wo.Process.Name,
				ProcessOID:  wo.Process.OID,
				ProcessType: wo.Process.Type,
				ID:          wo.RecipeID,
			},
			Sequence: int64(wo.Sequence),
			Station:  wo.Station,
			Status:   models.WorkOrderStatus(wo.Status),
			UpdateAt: strfmt.DateTime(wo.UpdatedAt),
			UpdateBy: wo.UpdatedBy,
			ParentID: wo.Parent,
		}

		batchQuantityDetails, err := handlerUtils.ParseBatchQuantityDetails(wo.BatchQuantityDetails)
		if err != nil {
			return work_order.NewGetStationSchedulingDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Details: err.Error(),
				},
			)
		}

		if len(batchQuantityDetails.PerBatchQuantity) != 0 {
			data[i].BatchesQuantity = handlerUtils.ToSlices(batchQuantityDetails.PerBatchQuantity)
		} else {
			data[i].BatchCount = int64(batchQuantityDetails.BatchCount)
			data[i].PlanQuantity = batchQuantityDetails.PlanQuantity.String()
		}
	}
	return work_order.NewGetStationSchedulingOK().WithPayload(&work_order.GetStationSchedulingOKBody{Data: data})
}

// ListWorkOrders implements.
func (w WorkOrder) ListWorkOrders(params work_order.ListWorkOrdersParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_LIST_WORK_ORDERS, principal.Roles) {
		return work_order.NewListWorkOrdersDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := w.dm.ListWorkOrdersByDuration(ctx, mcom.ListWorkOrdersByDurationRequest{
		Since:   time.Time(params.WorkDate).AddDate(0, 0, -9),
		Station: params.StationID,
	}.WithOrder(
		mcom.Order{
			Name:       "reserved_date",
			Descending: false,
		},
		mcom.Order{
			Name:       "reserved_sequence",
			Descending: false,
		},
	))
	if err != nil {
		return utils.ParseError(ctx, work_order.NewListWorkOrdersDefault(0), err)
	}

	data := []*work_order.ListWorkOrdersOKBodyDataItems0{}

	for _, wo := range list.Contents {
		if wo.Status == workorder.Status_PENDING || wo.Status == workorder.Status_ACTIVE || wo.Status == workorder.Status_CLOSING {
			batchQuantityDetails, err := handlerUtils.ParseBatchQuantityDetails(wo.BatchQuantityDetails)
			if err != nil {
				return work_order.NewListWorkOrdersDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{
						Details: err.Error(),
					},
				)
			}

			data = append(data, &work_order.ListWorkOrdersOKBodyDataItems0{
				WorkOrderID:     wo.ID,
				RecipeID:        wo.RecipeID,
				ProductID:       wo.Product.ID,
				ProductType:     wo.Product.Type,
				WorkOrderStatus: models.WorkOrderStatus(wo.Status),
				PlanQuantity:    batchQuantityDetails.PlanQuantity.String(),
				Date:            strfmt.Date(wo.Date),
			})
		}
	}
	return work_order.NewListWorkOrdersOK().WithPayload(&work_order.ListWorkOrdersOKBody{Data: data})
}

// ListWorkOrdersRate implements.
func (w WorkOrder) ListWorkOrdersRate(params work_order.ListWorkOrdersRateParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_LIST_WORK_ORDERS_RATE, principal.Roles) {
		return work_order.NewListWorkOrdersRateDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	pageRequest := mcom.PaginationRequest{}
	if params.Page != nil && params.Limit != nil {
		pageRequest = mcom.PaginationRequest{
			PageCount:      uint(*params.Page),
			ObjectsPerPage: uint(*params.Limit),
		}
	}

	orderRequest := parseOrderRequest(params.Body.OrderRequest, getWorkOrderListInfoByTypeDefaultOrderFunc)

	list, err := w.dm.ListWorkOrdersByDuration(ctx, mcom.ListWorkOrdersByDurationRequest{
		Since:        time.Time(params.WorkStartDate),
		Until:        time.Time(params.WorkEndDate),
		DepartmentID: params.DepartmentID,
	}.WithPagination(pageRequest).
		WithOrder(orderRequest...))
	if err != nil {
		return utils.ParseError(ctx, work_order.NewListWorkOrdersRateDefault(0), err)
	}

	data := make([]*models.WorkOrderRateData, len(list.Contents))
	for j, wo := range list.Contents {
		batchQuantityDetails, err := handlerUtils.ParseBatchQuantityDetails(wo.BatchQuantityDetails)
		if err != nil {
			return work_order.NewListWorkOrdersRateDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Details: err.Error(),
				},
			)
		}
		var ratio float64 = 0
		if wo.CollectedQuantity.LessThanOrEqual(batchQuantityDetails.PlanQuantity) {
			ratio = (wo.CollectedQuantity.InexactFloat64() / batchQuantityDetails.PlanQuantity.InexactFloat64()) * 100
		} else {
			ratio = 100 - (((wo.CollectedQuantity.InexactFloat64() - batchQuantityDetails.PlanQuantity.InexactFloat64()) / batchQuantityDetails.PlanQuantity.InexactFloat64()) * 100)
		}

		var productionTime string = ""
		var productionEndTime string = ""
		if wo.Status == workorder.Status_ACTIVE || wo.Status == workorder.Status_CLOSING || wo.Status == workorder.Status_CLOSED {
			if wo.CurrentBatch > 0 {
				batch, err := w.dm.GetBatch(ctx, mcom.GetBatchRequest{
					WorkOrder: wo.ID,
					Number:    int16(1),
				})
				if err != nil {
					return utils.ParseError(ctx, work_order.NewListWorkOrdersRateDefault(0), err)
				}
				if batch.Info.Status == int32(workorder.BatchStatus_BATCH_STARTED) || batch.Info.Status == int32(workorder.BatchStatus_BATCH_CLOSING) || batch.Info.Status == int32(workorder.BatchStatus_BATCH_CLOSED) {
					if len(batch.Info.Records) != 0 {
						productionTime = batch.Info.Records[0].Time.Format("2006-01-02")
					}
				}
			}
			if wo.Status == workorder.Status_CLOSED {
				productionEndTime = wo.UpdatedAt.Format("2006-01-02")
			}
		}
		data[j] = &models.WorkOrderRateData{
			DepartmentID:      wo.DepartmentID,
			WorkOrderID:       wo.ID,
			ProductID:         wo.Product.ID,
			Station:           wo.Station,
			PlanQuantity:      batchQuantityDetails.PlanQuantity.String(),
			CollectedQuantity: wo.CollectedQuantity.String(),
			Ratio:             fmt.Sprintf("%.2f%%", ratio),
			ProductionTime:    &productionTime,
			ProductionEndTime: &productionEndTime,
			UpdateBy:          wo.UpdatedBy,
			CreatedBy:         wo.InsertedBy,
			RecipeID:          wo.RecipeID,
		}
	}
	return work_order.NewListWorkOrdersRateOK().WithPayload(&work_order.ListWorkOrdersRateOKBody{
		Data: &work_order.ListWorkOrdersRateOKBodyData{
			Items: data,
			Total: list.AmountOfData},
	})
}

func parseOrderRequest(dataIn []*work_order.ListWorkOrdersRateParamsBodyOrderRequestItems0, defaultOrderFunc func() []mcom.Order) []mcom.Order {
	length := len(dataIn)
	if length == 0 {
		return defaultOrderFunc()
	}
	dataOut := make([]mcom.Order, length)
	for i, d := range dataIn {
		dataOut[i] = mcom.Order{
			Name:       d.OrderName,
			Descending: d.Descending,
		}
	}
	return dataOut
}

func getWorkOrderListInfoByTypeDefaultOrderFunc() []mcom.Order {
	return []mcom.Order{
		{
			Name:       "reserved_date",
			Descending: false,
		},
		{
			Name:       "reserved_sequence",
			Descending: false,
		},
	}
}

// ChangeWorkOrderStatus implements.
func (w WorkOrder) ChangeWorkOrderStatus(params work_order.ChangeWorkOrderStatusParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_CHANGE_WORK_ORDER_STATUS, principal.Roles) {
		return work_order.NewChangeWorkOrderStatusDefault(http.StatusForbidden)
	}

	var (
		status = params.Body.Type
	)
	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	getWorkOrder, err := w.dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{
		ID: params.WorkOrderID,
	})
	if err != nil {
		return utils.ParseError(ctx, work_order.NewChangeWorkOrderStatusDefault(0), err)
	}

	switch status {
	case startWorkOrder:
		if getWorkOrder.Status == workorder.Status_PENDING {
			status = int64(workorder.Status_ACTIVE)
		} else {
			return utils.ParseError(ctx, work_order.NewChangeWorkOrderStatusDefault(0), mcomErrors.Error{
				Code:    mcomErrors.Code_BAD_REQUEST,
				Details: "work order status not pending",
			})
		}
	case closeWorkOrder:
		if getWorkOrder.Status == workorder.Status_ACTIVE || getWorkOrder.Status == workorder.Status_CLOSING {
			status = int64(workorder.Status_CLOSED)
		} else {
			return utils.ParseError(ctx, work_order.NewChangeWorkOrderStatusDefault(0), mcomErrors.Error{
				Code:    mcomErrors.Code_BAD_REQUEST,
				Details: "work order status not active or closing",
			})
		}
	default:
		return utils.ParseError(ctx, work_order.NewChangeWorkOrderStatusDefault(0), mcomErrors.Error{
			Code:    mcomErrors.Code_BAD_REQUEST,
			Details: "status not within the specified range",
		})
	}

	abnormality := mcomWorkOrder.Abnormality(params.Body.Remark + 1)
	if err := w.dm.UpdateWorkOrders(ctx, mcom.UpdateWorkOrdersRequest{
		Orders: []mcom.UpdateWorkOrder{
			{
				ID:          params.WorkOrderID,
				Status:      workorder.Status(status),
				Abnormality: abnormality,
			},
		},
	}); err != nil {
		return utils.ParseError(ctx, work_order.NewChangeWorkOrderStatusDefault(0), err)
	}

	if status == int64(workorder.Status_CLOSED) {
		if apiConfig := w.config.StationFunctionConfig[getWorkOrder.Station]; apiConfig.ClosedWorkOrderAPIPath != "" {
			closedWorkOrderRequest := mesageModels.NotifyWorkOrderClosedRequestBody{
				WorkOrderID: params.WorkOrderID,
			}

			// send closedWorkOrder request to MES agent
			if err := handlerUtils.SendMesAgePOSTRequest(closedWorkOrderRequest, apiConfig.ClosedWorkOrderAPIPath); err != nil {
				commonsCtx.Logger(ctx).Error("failed to send the request to MES Agent", zap.Error(err))
			}
		}
	}

	return work_order.NewChangeWorkOrderStatusOK()
}

// GetWorkOrderInformation implements.
func (w WorkOrder) GetWorkOrderInformation(params work_order.GetWorkOrderInformationParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_GET_WORK_ORDER_INFORMATION, principal.Roles) {
		return work_order.NewGetWorkOrderInformationDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	// planQuantity & sequence
	getWorkOrder, err := w.dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{
		ID: params.WorkOrderID,
	})
	if err != nil {
		return utils.ParseError(ctx, work_order.NewGetWorkOrderInformationDefault(0), err)
	}

	batchQuantityDetails, err := handlerUtils.ParseBatchQuantityDetails(getWorkOrder.BatchQuantityDetails)
	if err != nil {
		return work_order.NewGetWorkOrderInformationDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Details: err.Error(),
			},
		)
	}

	// materials & tools & useValue(max min mid)
	getRecipe, err := w.dm.GetProcessDefinition(ctx, mcom.GetProcessDefinitionRequest{
		RecipeID:    getWorkOrder.RecipeID,
		ProcessName: getWorkOrder.Process.Name,
		ProcessType: getWorkOrder.Process.Type,
	})
	if err != nil {
		return utils.ParseError(ctx, work_order.NewGetWorkOrderInformationDefault(0), err)
	}

	var (
		loadWorkOrderRequest = mesageModels.NotifyWorkOrderStartRequestBody{
			OperatorID: principal.ID,
			WorkOrder: &mesageModels.NotifyWorkOrderStartRequestBodyWorkOrder{
				ID:                params.WorkOrderID,
				CollectedSequence: int64(getWorkOrder.CollectedSequence),
				CurrentBatch:      int64(getWorkOrder.CurrentBatch),
				PlanBatchCount:    batchQuantityDetails.BatchCount,
				PlanQuantity: &mesageModels.Decimal{
					Value: batchQuantityDetails.PlanQuantity.CoefficientInt64(),
					Exp:   int64(batchQuantityDetails.PlanQuantity.Exponent()),
				},
				ProductID: getWorkOrder.Product.ID,
			},
			Recipe: &mesageModels.NotifyWorkOrderStartRequestBodyRecipe{
				ID: getWorkOrder.RecipeID,
			},
		}

		step = []*mcom.RecipeProcessStep{}
	)

	tools := []*work_order.GetWorkOrderInformationOKBodyDataRecipeToolsItems0{}
	materials := []*work_order.GetWorkOrderInformationOKBodyDataRecipeMaterialsItems0{}
	for _, data := range getRecipe.Configs {
		for _, configStation := range data.Stations {
			if configStation == getWorkOrder.Station {
				// mesage commonControls
				loadWorkOrderRequest.Recipe.CommonControls = append(loadWorkOrderRequest.Recipe.CommonControls,
					parseRecipeParameter(data.CommonControls)...)

				// commonProperties
				loadWorkOrderRequest.Recipe.CommonProperties = append(loadWorkOrderRequest.Recipe.CommonProperties,
					parseRecipeParameter(data.CommonProperties)...)

				// tool
				informationTool, mesageTool := parseToolsList(data.Tools)
				tools = append(tools, informationTool...)
				loadWorkOrderRequest.Recipe.Tools = append(loadWorkOrderRequest.Recipe.Tools, mesageTool...)

				// material
				informationMaterial, mesageMaterial, err := parseMaterialsList(data.Steps)
				if err != nil {
					return utils.ParseError(ctx, work_order.NewGetWorkOrderInformationDefault(0), err)
				}
				materials = append(materials, informationMaterial...)
				loadWorkOrderRequest.Recipe.ProcessSteps = append(loadWorkOrderRequest.Recipe.ProcessSteps, mesageMaterial...)
				step = data.Steps

				loadWorkOrderRequest.Recipe.BatchSize = &mesageModels.Decimal{
					Value: data.BatchSize.CoefficientInt64(),
					Exp:   int64(data.BatchSize.Exponent()),
				}
				break
			}
		}
	}

	if apiConfig := w.config.StationFunctionConfig[getWorkOrder.Station]; apiConfig.LoadWorkOrderAPIPath != "" {

		//station site material check
		getStationSite, err := w.dm.GetStation(ctx, mcom.GetStationRequest{
			ID: getWorkOrder.Station,
		})
		if err != nil {
			return work_order.NewGetWorkOrderInformationDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: err.Error(),
			})
		}

		for _, data := range getStationSite.Sites {
			if data.Information.SubType == sites.SubType_MATERIAL {
				tempSite := &mesageModels.NotifyWorkOrderStartRequestBodySitesItems0{
					Site: &mesageModels.Site{
						Station: data.Information.Station,
						Name:    data.Information.SiteID.Name,
						Index:   int64(data.Information.SiteID.Index),
					},
					CurrentState: &mesageModels.SiteBindingState{
						NotOK: false,
					},
				}

				switch data.Information.Type {
				case sites.Type_SLOT:
					m := data.Content.Slot
					if m.Material != nil {
						tempSite.CurrentState.NotOK = handlerUtils.SiteMaterialNotOKCheck(mcomModels.BoundResource(*m), step)
						tempSite.CurrentState.Resources = append(tempSite.CurrentState.Resources, &mesageModels.SiteBindingStateResourcesItems0{
							ID: m.Material.ResourceID,
						})
					} else {
						tempSite.CurrentState.NotOK = true
					}

				case sites.Type_CONTAINER, sites.Type_COLLECTION:
					var (
						content []mcomModels.BoundResource
					)
					if data.Information.Type == sites.Type_CONTAINER {
						content = *data.Content.Container
					} else {
						content = *data.Content.Collection
					}
					if len(content) != 0 {
						for _, m := range content {
							if !tempSite.CurrentState.NotOK {
								tempSite.CurrentState.NotOK = handlerUtils.SiteMaterialNotOKCheck(mcomModels.BoundResource(m), step)
							}
							tempSite.CurrentState.Resources = append(tempSite.CurrentState.Resources,
								&mesageModels.SiteBindingStateResourcesItems0{
									ID: m.Material.ResourceID,
								})
						}
					} else {
						tempSite.CurrentState.NotOK = true
					}

				case sites.Type_QUEUE:
					content := data.Content.Queue
					if len(*content) != 0 {
						for _, m := range *content {
							if !tempSite.CurrentState.NotOK {
								tempSite.CurrentState.NotOK = handlerUtils.SiteMaterialNotOKCheck(mcomModels.BoundResource(m), step)
							}
							tempSite.CurrentState.Resources = append(tempSite.CurrentState.Resources,
								&mesageModels.SiteBindingStateResourcesItems0{
									ID: m.Material.ResourceID,
								})
						}
					} else {
						tempSite.CurrentState.NotOK = true
					}

				case sites.Type_COLQUEUE:
					content := data.Content.Colqueue

					if len(*content) != 0 {
						for _, m := range *content {
							for _, n := range m {
								if !tempSite.CurrentState.NotOK {
									tempSite.CurrentState.NotOK = handlerUtils.SiteMaterialNotOKCheck(mcomModels.BoundResource(n), step)
								}
								tempSite.CurrentState.Resources = append(tempSite.CurrentState.Resources,
									&mesageModels.SiteBindingStateResourcesItems0{
										ID: n.Material.ResourceID,
									})
							}
						}
					} else {
						tempSite.CurrentState.NotOK = true
					}
				}
				loadWorkOrderRequest.Sites = append(loadWorkOrderRequest.Sites, tempSite)
			}
		}

		// send loadWorkOrder request to MES agent
		if err := handlerUtils.SendMesAgePOSTRequest(loadWorkOrderRequest, apiConfig.LoadWorkOrderAPIPath); err != nil {
			commonsCtx.Logger(ctx).Error("failed to send the request to MES Agent", zap.Error(err))
		}
	}

	return work_order.NewGetWorkOrderInformationOK().WithPayload(&work_order.GetWorkOrderInformationOKBody{
		Data: &work_order.GetWorkOrderInformationOKBodyData{
			WorkOrderID:     getWorkOrder.ID,
			ProductID:       getWorkOrder.Product.ID,
			ProductType:     getWorkOrder.Product.Type,
			RecipeID:        getWorkOrder.RecipeID,
			Date:            strfmt.Date(getWorkOrder.Date),
			WorkOrderStatus: int64(getWorkOrder.Status),
			CollectSequence: int64(getWorkOrder.CollectedSequence),
			PlanQuantity:    batchQuantityDetails.PlanQuantity.String(),
			CurrentBatch:    int64(getWorkOrder.CurrentBatch),
			CurrentQuantity: getWorkOrder.CollectedQuantity.InexactFloat64(),
			Recipe: &work_order.GetWorkOrderInformationOKBodyDataRecipe{
				Materials: materials,
				Tools:     tools,
			},
		},
	})
}

func (w WorkOrder) CreateWorkOrdersFromFile(params work_order.CreateWorkOrdersFromFileParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_CREATE_WORK_ORDERS_FROM_FILE, principal.Roles) {
		return work_order.NewCreateWorkOrdersFromFileDefault(http.StatusForbidden)
	}
	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	rows, err := getExcelRows(params.UploadFile)
	if err != nil {
		return utils.ParseError(ctx, work_order.NewCreateWorkOrdersFromFileDefault(0), err)
	}
	params.UploadFile.Close()

	createRequest, failData, err := parseCreateWorkOrdersRequest(ctx, w.dm, params.Department, rows)
	if err != nil {
		return utils.ParseError(ctx, work_order.NewCreateWorkOrdersFromFileDefault(0), err)
	}
	if len(failData) > 0 {
		return work_order.NewCreateWorkOrdersFromFileOK().WithPayload(&work_order.CreateWorkOrdersFromFileOKBody{
			Data: &work_order.CreateWorkOrdersFromFileOKBodyData{
				FailData: failData,
			},
		})
	}

	_, err = w.dm.CreateWorkOrders(ctx, createRequest)
	if err != nil {
		return utils.ParseError(ctx, work_order.NewCreateWorkOrdersFromFileDefault(0), err)
	}

	return work_order.NewCreateWorkOrdersFromFileOK().WithPayload(&work_order.CreateWorkOrdersFromFileOKBody{
		Data: &work_order.CreateWorkOrdersFromFileOKBodyData{
			FailData: []*work_order.CreateWorkOrdersFromFileOKBodyDataFailDataItems0{},
		},
	})
}

func getExcelRows(file io.Reader) ([][]string, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// get sheet name list
	sheetList := f.GetSheetList()

	// if empty sheet
	if len(sheetList) == 0 {
		return nil, mcomErrors.Error{
			Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
			Details: "no sheet",
		}
	}

	return f.GetRows(sheetList[0])
}

func parseCreateWorkOrdersRequest(ctx context.Context, dm mcom.DataManager, department string, rows [][]string) (mcom.CreateWorkOrdersRequest, []*work_order.CreateWorkOrdersFromFileOKBodyDataFailDataItems0, error) {
	/*
		excel column name
		0 row[colProductID]				productID
		1 row[colStation]				stationID
		2 row[colVersionStage]			versionStage
		3 row[colRecipeID]				recipeID
		4 row[colProcessName]			processName
		5 row[colProcessType]			processType
		6 row[colExcelBatchSize]		batchSize 0:FixedQuantity;1:PlanQuantity
		7 row[colExcelBatchSizeMode]	batch/planQuantity
		8 row[colExcelRecipeBatchSize]	recipe batchSize
		9 row[colDate]					date
	*/

	// if file empty
	if len(rows) <= 1 {
		return mcom.CreateWorkOrdersRequest{}, nil, mcomErrors.Error{
			Code:    mcomErrors.Code_INSUFFICIENT_REQUEST,
			Details: "file empty",
		}
	}

	header := Row(rows[0])
	rows = rows[1:]

	var failData []*work_order.CreateWorkOrdersFromFileOKBodyDataFailDataItems0

	createRequest := []mcom.CreateWorkOrder{}

	for i, r := range rows {
		var failColumns []int
		var (
			batchSize *decimal.Decimal
			row       Row = r
		)

		tempCreateRequest := mcom.CreateWorkOrder{
			DepartmentID: department,
			Station:      row.Column(colStation),
			RecipeID:     row.Column(colRecipeID),
			ProcessName:  row.Column(colProcessName),
			ProcessType:  row.Column(colProcessType),
		}

		// check recipeID
		if tempCreateRequest.RecipeID == "" {
			// check productID
			if row.Column(colProductID) != "" {
				latestRecipe, err := getLatestRecipe(ctx, dm, Row(r))

				// if err length not 0
				if len(err) != 0 {
					failColumns = append(failColumns, err...)
				} else {
					tempCreateRequest.RecipeID = latestRecipe.ID

					// find the process with station
					for _, i := range latestRecipe.Processes {
						var find bool
						// get batchSize
						batchSize, find = findBatchSize(i.Info, row)
						if find {
							tempCreateRequest.ProcessOID = i.Info.OID
							break
						}
					}
				}
			} else {
				failColumns = append(failColumns, colProductID)
			}
		} else {
			// get processID & recipe batchSize
			getProcessDefinition, err := dm.GetProcessDefinition(ctx, mcom.GetProcessDefinitionRequest{
				RecipeID:    tempCreateRequest.RecipeID,
				ProcessName: tempCreateRequest.ProcessName,
				ProcessType: tempCreateRequest.ProcessType,
			})

			// if err equal Code_PROCESS_NOT_FOUND
			if err != nil {
				if e, ok := mcomErrors.As(err); ok {
					if e.Code == mcomErrors.Code_PROCESS_NOT_FOUND || e.Code == mcomErrors.Code_INSUFFICIENT_REQUEST {
						failColumns = append(failColumns, colRecipeID, colProcessName, colProcessType)
					} else {
						return mcom.CreateWorkOrdersRequest{}, nil, err
					}
				}
			} else {
				// check output ID
				if getProcessDefinition.Output.ID != row.Column(colProductID) {
					failColumns = append(failColumns, colProductID)
				}
				//check station
				if !recipeStationCheck(getProcessDefinition.ProcessDefinition, tempCreateRequest.Station) {
					failColumns = append(failColumns, colStation)
				}

				tempCreateRequest.ProcessOID = getProcessDefinition.OID
				// get batchSize
				batchSize, _ = findBatchSize(getProcessDefinition.ProcessDefinition, row)
			}
		}

		// batchQuantity
		batchQuantity, badColumnsIndex := parseBatchQuantity(batchSize, row)
		// check error index is nil
		if len(badColumnsIndex) != 0 {
			failColumns = append(failColumns, badColumnsIndex...)
		} else {
			tempCreateRequest.BatchesQuantity = batchQuantity
		}

		// parse date
		date, err := time.Parse("2006-01-02", row.Column(colDate))
		if err != nil {
			failColumns = append(failColumns, colDate)
		} else {
			tempCreateRequest.Date = date
		}

		createRequest = append(createRequest, tempCreateRequest)

		if len(failColumns) != 0 {
			failData = append(failData, &work_order.CreateWorkOrdersFromFileOKBodyDataFailDataItems0{
				Index:   int64(i + 2),
				Columns: excelBadColumns(header, failColumns),
			})
		}
	}

	if len(failData) > 0 {
		return mcom.CreateWorkOrdersRequest{}, failData, nil
	}

	return mcom.CreateWorkOrdersRequest{
		WorkOrders: createRequest,
	}, nil, nil
}

func getLatestRecipe(ctx context.Context, dm mcom.DataManager, dataIn Row) (mcom.GetRecipeReply, []int) {
	var (
		dataOut     []mcom.GetRecipeReply
		failColumns []int
	)
	getRecipeList, err := dm.ListRecipesByProduct(ctx, mcom.ListRecipesByProductRequest{
		ProductID: dataIn.Column(colProductID),
	}.WithOrder(mcom.Order{
		Name:       "released_at",
		Descending: true,
	}))
	if err != nil {
		failColumns = append(failColumns, colProductID)
	}

	var recipeListVersionStage []mcom.GetRecipeReply

	// if versionStage not empty
	if dataIn.Column(colVersionStage) != "" {
		// find match recipe with version stage
		for _, recipe := range getRecipeList.Recipes {
			if recipe.Version.Stage == dataIn.Column(colVersionStage) {
				recipeListVersionStage = append(recipeListVersionStage, recipe)
			}
		}
	}

	var recipeListProcess []mcom.GetRecipeReply
	if dataIn.Column(colVersionStage) == "" || len(recipeListVersionStage) == 0 {
		failColumns = append(failColumns, colVersionStage)
	} else {
		// find match recipe with processName & processType
		for _, recipe := range recipeListVersionStage {
			for _, process := range recipe.Processes {
				if processCheck(process.Info, dataIn) {
					recipeListProcess = append(recipeListProcess, recipe)
				}
			}
		}
	}

	if len(recipeListProcess) == 0 {
		failColumns = append(failColumns, colProcessName, colProcessType)
	} else {
		// find match recipe with stationID
		for _, recipe := range recipeListProcess {
			for _, process := range recipe.Processes {
				if processCheck(process.Info, dataIn) {
					if recipeStationCheck(process.Info, dataIn.Column(colStation)) {
						dataOut = append(dataOut, recipe)
					}
				}
			}
		}
	}

	if len(dataOut) == 0 {
		failColumns = append(failColumns, colStation)
	}

	if len(failColumns) != 0 {
		sort.Ints(failColumns)
		return mcom.GetRecipeReply{}, failColumns
	}

	// from recipeList find latest recipeID
	// if descending is true
	// find match first is latest
	return dataOut[0], nil
}

// find match process & productID
func processCheck(processInfo mcom.ProcessDefinition, dataIn Row) bool {
	if processInfo.Name == dataIn.Column(colProcessName) && processInfo.Type == dataIn.Column(colProcessType) &&
		processInfo.Output.ID == dataIn.Column(colProductID) {
		return true
	}
	return false
}

// parseBatchQuantity if the length of `badColumnsIndex` is 0, it means there is no
// bad column, otherwise it does and `batchQuantity` is nil.
func parseBatchQuantity(batchSize *decimal.Decimal, row Row) (batchQuantity mcom.BatchQuantity, badColumnsIndex []int) {
	var (
		batchesQuantity mcom.BatchQuantity
		batch           int
		planQuantity    decimal.Decimal
	)

	if batchSize == nil || batchSize.Equals(decimal.Zero) {
		if row.Column(colExcelRecipeBatchSize) != "" {
			size, err := decimal.NewFromString(row.Column(colExcelRecipeBatchSize))
			if err == nil {
				batchSize = &size
			}
		}
	}

	// batchQuantity
	switch row.Column(colExcelBatchSize) {
	// 0:FixedQuantity
	case "0":
		batch, err := strconv.Atoi(row.Column(colExcelBatchSizeData))
		if err != nil || batch == 0 {
			badColumnsIndex = append(badColumnsIndex, colExcelBatchSizeData)
			break
		}
		//
		if batchSize == nil {
			badColumnsIndex = append(badColumnsIndex, colExcelRecipeBatchSize)
			break
		}

		planQuantity = decimal.NewFromInt(int64(batch)).Mul(*batchSize)
		batchesQuantity = mcom.NewFixedQuantity(uint(batch), planQuantity)
	// 1:PlanQuantity
	case "1":
		planQuantity, err := decimal.NewFromString(row.Column(colExcelBatchSizeData))
		if err != nil || planQuantity.Equals(decimal.Zero) {
			badColumnsIndex = append(badColumnsIndex, colExcelBatchSizeData)
			break
		}
		if batchSize == nil {
			badColumnsIndex = append(badColumnsIndex, colExcelRecipeBatchSize)
			break
		}

		batch = int(planQuantity.Div(*batchSize).Ceil().IntPart())
		batchesQuantity = mcom.NewPlanQuantity(uint(batch), planQuantity)
	default:
		badColumnsIndex = append(badColumnsIndex, colExcelBatchSize, colExcelBatchSizeData)
	}

	if len(badColumnsIndex) > 0 {
		return nil, badColumnsIndex
	}
	return batchesQuantity, nil
}

func recipeStationCheck(process mcom.ProcessDefinition, station string) bool {
	for _, recipeConfig := range process.Configs {
		for _, recipeConfigStation := range recipeConfig.Stations {
			if recipeConfigStation == station {
				return true
			}
		}
	}

	return false
}

func findBatchSize(process mcom.ProcessDefinition, row Row) (*decimal.Decimal, bool) {
	for _, recipeConfig := range process.Configs {
		for _, recipeConfigStation := range recipeConfig.Stations {
			if recipeConfigStation == row.Column(colStation) {
				return recipeConfig.BatchSize, true
			}
		}
	}

	return nil, false
}

// UpdateWorkOrder implements.
func (w WorkOrder) UpdateWorkOrder(params work_order.UpdateWorkOrderParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_UPDATE_WORK_ORDER, principal.Roles) {
		return work_order.NewUpdateWorkOrderDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	batchesQuantity, err := parseBatchQuantities(*params.Body)
	if err != nil {
		return work_order.NewUpdateWorkOrderDefault(http.StatusBadRequest).WithPayload(err)
	}

	getWorkOrderRep, e := w.dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{
		ID: params.ID,
	})
	if e != nil {
		return utils.ParseError(ctx, work_order.NewUpdateWorkOrderDefault(0), e)
	}
	if getWorkOrderRep.Status != workorder.Status_PENDING {
		return utils.ParseError(ctx, work_order.NewUpdateWorkOrderDefault(0), mcomErrors.Error{
			Code:    mcomErrors.Code_BAD_REQUEST,
			Details: "work order status not pending",
		})
	}

	updateWorkOrder := mcom.UpdateWorkOrdersRequest{
		Orders: []mcom.UpdateWorkOrder{
			{
				ID:              params.ID,
				Station:         *params.Body.Station,
				RecipeID:        params.Body.Recipe.ID,
				ProcessOID:      params.Body.Recipe.ProcessOID,
				ProcessName:     params.Body.Recipe.ProcessName,
				ProcessType:     params.Body.Recipe.ProcessType,
				Date:            time.Time(*params.Body.PlanDate),
				BatchesQuantity: batchesQuantity,
			},
		},
	}

	if err := w.dm.UpdateWorkOrders(ctx, updateWorkOrder); err != nil {
		return utils.ParseError(ctx, work_order.NewUpdateWorkOrderDefault(0), err)
	}
	return work_order.NewUpdateWorkOrderOK()
}

func parseBatchQuantities(req models.UpdateWorkOrder) (mcom.BatchQuantity, *models.Error) {
	switch req.BatchSize {
	case int64(mcomWorkOrder.BatchSize_PER_BATCH_QUANTITIES):
		batchesQuantity, err := handlerUtils.ToDecimals(req.BatchesQuantity)
		if err != nil {
			return nil, &models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: fmt.Sprintf("invalid_numbers=%v", req.BatchesQuantity),
			}
		}
		return mcom.NewQuantityPerBatch(batchesQuantity), nil
	case int64(mcomWorkOrder.BatchSize_FIXED_QUANTITY):
		qty, err := decimal.NewFromString(req.PlanQuantity)
		if err != nil {
			return nil, &models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: fmt.Sprintf("invalid_numbers=%v", req.BatchesQuantity),
			}
		}
		if req.BatchCount <= 0 || qty.LessThanOrEqual(decimal.Zero) {
			return nil, &models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "missing batch count or quantity",
			}
		}
		return mcom.NewFixedQuantity(uint(req.BatchCount), qty), nil
	case int64(mcomWorkOrder.BatchSize_PLAN_QUANTITY):
		qty, err := decimal.NewFromString(req.PlanQuantity)
		if err != nil {
			return nil, &models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: fmt.Sprintf("invalid_numbers=%v", req.BatchesQuantity),
			}
		}
		if req.BatchCount <= 0 || qty.LessThanOrEqual(decimal.Zero) {
			return nil, &models.Error{
				Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
				Details: "missing batch count or quantity",
			}
		}
		return mcom.NewPlanQuantity(uint(req.BatchCount), qty), nil
	default:
		return nil, &models.Error{Code: int64(mcomErrors.Code_BAD_REQUEST), Details: "invalid enum value of batch size"}
	}
}

func parseRecipeParameter(dataIn []*mcom.RecipeProperty) mesageModels.RecipeParameter {
	dataOut := mesageModels.RecipeParameter{}
	for _, data := range dataIn {
		tempCommonRecipeProperty := mesageModels.RecipeParameterItems0{
			Name:  data.Name,
			Value: &mesageModels.RecipeParameterItems0Value{},
		}

		if data.Param.High != nil {
			tempCommonRecipeProperty.Value.Max = &mesageModels.Decimal{
				Value: data.Param.High.CoefficientInt64(),
				Exp:   int64(data.Param.High.Exponent()),
			}
		}
		if data.Param.Mid != nil {
			tempCommonRecipeProperty.Value.Mid = &mesageModels.Decimal{
				Value: data.Param.Mid.CoefficientInt64(),
				Exp:   int64(data.Param.Mid.Exponent()),
			}
		}
		if data.Param.Low != nil {
			tempCommonRecipeProperty.Value.Min = &mesageModels.Decimal{
				Value: data.Param.Low.CoefficientInt64(),
				Exp:   int64(data.Param.Low.Exponent()),
			}
		}

		dataOut = append(dataOut, &tempCommonRecipeProperty)

	}
	return dataOut
}

func parseToolsList(dataIn []*mcom.RecipeTool) ([]*work_order.GetWorkOrderInformationOKBodyDataRecipeToolsItems0,
	[]*mesageModels.NotifyWorkOrderStartRequestBodyRecipeToolsItems0) {
	dataOut1 := make([]*work_order.GetWorkOrderInformationOKBodyDataRecipeToolsItems0, len(dataIn))
	dataOut2 := make([]*mesageModels.NotifyWorkOrderStartRequestBodyRecipeToolsItems0, len(dataIn))
	for i, data := range dataIn {
		dataOut1[i] = &work_order.GetWorkOrderInformationOKBodyDataRecipeToolsItems0{
			ID:        data.ID,
			Necessity: data.Required,
		}

		dataOut2[i] = &mesageModels.NotifyWorkOrderStartRequestBodyRecipeToolsItems0{
			ID: data.ID,
		}
	}
	return dataOut1, dataOut2
}

func parseMaterialsList(dataIn []*mcom.RecipeProcessStep) ([]*work_order.GetWorkOrderInformationOKBodyDataRecipeMaterialsItems0,
	[]*mesageModels.NotifyWorkOrderStartRequestBodyRecipeProcessStepsItems0, error) {
	materialList := make(map[string]int)
	dataOut1 :=
		[]*work_order.GetWorkOrderInformationOKBodyDataRecipeMaterialsItems0{}
	dataOut2 :=
		[]*mesageModels.NotifyWorkOrderStartRequestBodyRecipeProcessStepsItems0{}

	for _, data := range dataIn {
		// mesage control
		tempRecipeMaterial := mesageModels.NotifyWorkOrderStartRequestBodyRecipeProcessStepsItems0{
			Controls:  make(mesageModels.RecipeParameter, len(data.Controls)),
			Materials: []*mesageModels.NotifyWorkOrderStartRequestBodyRecipeProcessStepsItems0MaterialsItems0{},
		}
		for i, control := range data.Controls {
			tempControl := mesageModels.RecipeParameterItems0{
				Name:  control.Name,
				Value: &mesageModels.RecipeParameterItems0Value{},
			}

			if control.Param.High != nil {
				tempControl.Value.Max = &mesageModels.Decimal{
					Value: control.Param.High.CoefficientInt64(),
					Exp:   int64(control.Param.High.Exponent()),
				}
			}
			if control.Param.Mid != nil {
				tempControl.Value.Mid = &mesageModels.Decimal{
					Value: control.Param.Mid.CoefficientInt64(),
					Exp:   int64(control.Param.Mid.Exponent()),
				}
			}
			if control.Param.Low != nil {
				tempControl.Value.Min = &mesageModels.Decimal{
					Value: control.Param.Low.CoefficientInt64(),
					Exp:   int64(control.Param.Low.Exponent()),
				}
			}

			tempRecipeMaterial.Controls[i] = &tempControl
		}

		for _, materialData := range data.Materials {
			if materialData.Value.Mid == nil {
				return []*work_order.GetWorkOrderInformationOKBodyDataRecipeMaterialsItems0{},
					[]*mesageModels.NotifyWorkOrderStartRequestBodyRecipeProcessStepsItems0{},
					mcomErrors.Error{
						Details: "some step standard value is nil",
					}
			}

			index := findIndex(materialData.Name, materialList)
			if index == -1 {
				materialList[materialData.Name] = len(materialList)
				dataOut1 = append(dataOut1, &work_order.GetWorkOrderInformationOKBodyDataRecipeMaterialsItems0{
					ID:            materialData.Name,
					SiteName:      materialData.Site,
					StandardValue: materialData.Value.Mid.InexactFloat64(),
				})
			} else {
				dataOut1[index].StandardValue += materialData.Value.Mid.InexactFloat64()
			}

			// mesage material
			tempMaterial := &mesageModels.NotifyWorkOrderStartRequestBodyRecipeProcessStepsItems0MaterialsItems0{
				Grade:    materialData.Grade,
				ID:       materialData.Name,
				Quantity: &mesageModels.NotifyWorkOrderStartRequestBodyRecipeProcessStepsItems0MaterialsItems0Quantity{},
			}

			if materialData.Value.High != nil {
				tempMaterial.Quantity.Max = &mesageModels.Decimal{
					Value: materialData.Value.High.CoefficientInt64(),
					Exp:   int64(materialData.Value.High.Exponent()),
				}
			}
			if materialData.Value.Mid != nil {
				tempMaterial.Quantity.Mid = &mesageModels.Decimal{
					Value: materialData.Value.Mid.CoefficientInt64(),
					Exp:   int64(materialData.Value.Mid.Exponent()),
				}
			}
			if materialData.Value.Low != nil {
				tempMaterial.Quantity.Min = &mesageModels.Decimal{
					Value: materialData.Value.Low.CoefficientInt64(),
					Exp:   int64(materialData.Value.Low.Exponent()),
				}
			}

			tempRecipeMaterial.Materials = append(tempRecipeMaterial.Materials, tempMaterial)
		}
		dataOut2 = append(dataOut2, &tempRecipeMaterial)
	}
	return dataOut1, dataOut2, nil
}

func findIndex(target string, dataIn map[string]int) int {
	if dataOut, ok := dataIn[target]; ok {
		return dataOut
	}
	return -1
}
