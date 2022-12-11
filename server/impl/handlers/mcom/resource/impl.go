package resource

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/shopspring/decimal"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	resources "gitlab.kenda.com.tw/kenda/mcom/utils/resources"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/internal/printer"
	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils/barcodes"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/resource"
)

type Config struct {
	Printers map[string]string
	FontPath string
}

// Resource definitions
type Resource struct {
	dm mcom.DataManager

	config Config

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewResource returns Resource service.
func NewResource(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool,
	config Config) service.Resource {
	return Resource{
		dm:            dm,
		config:        config,
		hasPermission: hasPermission,
	}
}
func int64ToInt(dataIn []int64) []int {
	dataOut := make([]int, len(dataIn))
	for i := 0; i < len(dataIn); i++ {
		dataOut[i] = int(dataIn[i])
	}
	return dataOut
}

// AddMaterial implementations.
func (r Resource) AddMaterial(params resource.AddMaterialParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_ADD_MATERIAL, principal.Roles) {
		return resource.NewAddMaterialDefault(http.StatusForbidden)
	}

	quantity, err := decimal.NewFromString(*params.Body.Resource.Quantity)
	if err != nil {
		return resource.NewAddMaterialDefault(http.StatusBadRequest).WithPayload(&models.Error{
			Code:    int64(mcomErrors.Code_INVALID_NUMBER),
			Details: "invalid_number=" + *params.Body.Resource.Quantity,
		})
	}

	req := mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{
			{
				Type:           *params.Body.Resource.ProductType,
				ID:             *params.Body.Resource.ProductID,
				Grade:          *params.Body.Resource.Grade,
				Status:         resources.MaterialStatus_INSPECTION,
				Quantity:       quantity,
				Unit:           *params.Body.Resource.Unit,
				LotNumber:      *params.Body.Resource.LotNumber,
				ProductionTime: time.Time(*params.Body.Resource.ProductionTime),
				ExpiryTime:     time.Time(*params.Body.Resource.ExpiryTime),
			},
		},
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	resourceIDs, err := r.dm.CreateMaterialResources(ctx, req,
		mcom.WithStockIn(mcom.Warehouse{
			ID:       *params.Body.Warehouse.ID,
			Location: *params.Body.Warehouse.Location,
		}))
	if err != nil {
		return utils.ParseError(ctx, resource.NewAddMaterialDefault(0), err)
	}

	if len(resourceIDs) != 1 {
		return resource.NewAddMaterialDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Details: fmt.Sprintf("unexpected add material result=%v", resourceIDs),
		})
	}

	return resource.NewAddMaterialOK().WithPayload(
		&resource.AddMaterialOKBody{Data: &resource.AddMaterialOKBodyData{ResourceID: resourceIDs[0].ID}})
}

// SplitMaterial implementations.
func (r Resource) SplitMaterial(params resource.SplitMaterialParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_SPLIT_MATERIAL, principal.Roles) {
		return resource.NewSplitMaterialDefault(http.StatusForbidden)
	}

	req := mcom.SplitMaterialResourceRequest{
		ResourceID:    params.Body.ResourceID,
		ProductType:   params.Body.ProductType,
		Quantity:      decimal.NewFromFloat(params.Body.SplitQuantity),
		InspectionIDs: int64ToInt(params.Body.Inspections),
		Remark:        params.Body.Remark,
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	resourceID, err := r.dm.SplitMaterialResource(ctx, req)
	if err != nil {
		return utils.ParseError(ctx, resource.NewSplitMaterialDefault(0), err)
	}

	return resource.NewSplitMaterialOK().WithPayload(
		&resource.SplitMaterialOKBody{Data: &resource.SplitMaterialOKBodyData{
			ResourceID: resourceID.NewResourceID}})
}

// GetMaterialResourceInfo implementation.
func (r Resource) GetMaterialResourceInfo(params resource.GetMaterialResourceInfoParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_GET_MATERIAL_RESOURCE_INFO, principal.Roles) {
		return resource.NewGetMaterialResourceInfoDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	reply, err := r.dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{
		ResourceID: params.ID,
	})
	if err != nil {
		return utils.ParseError(ctx, resource.NewGetMaterialResourceInfoDefault(0), err)
	}

	return resource.NewGetMaterialResourceInfoOK().WithPayload(&resource.GetMaterialResourceInfoOKBody{
		Data: parseResourceMaterials(reply),
	})
}

// ListMaterialStatus implementation.
func (r Resource) ListMaterialStatus(params resource.ListMaterialStatusParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_LIST_MATERIAL_STATUS, principal.Roles) {
		return resource.NewListMaterialStatusDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := r.dm.ListMaterialResourceStatus(ctx)
	if err != nil {
		return utils.ParseError(ctx, resource.NewListMaterialStatusDefault(0), err)
	}

	data := make([]*resource.ListMaterialStatusOKBodyDataItems0, len(list))
	for i, materialStatus := range list {
		data[i] = &resource.ListMaterialStatusOKBodyDataItems0{
			ID:   models.MaterialStatus(resources.MaterialStatus_value[materialStatus]),
			Name: materialStatus,
		}
	}

	return resource.NewListMaterialStatusOK().WithPayload(&resource.ListMaterialStatusOKBody{Data: data})
}

// DownloadMaterialResource implementation.
func (r Resource) DownloadMaterialResource(params resource.DownloadMaterialResourceParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_DOWNLOAD_MATERIAL_RESOURCE, principal.Roles) {
		return resource.NewGetMaterialResourceInfoDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	reply, err := r.dm.GetMaterialResourceIdentity(ctx, mcom.GetMaterialResourceIdentityRequest{
		ResourceID:  params.Body.ResourceID,
		ProductType: params.Body.ProductType,
	})
	if err != nil {
		return utils.ParseError(ctx, resource.NewGetMaterialResourceInfoDefault(0), err)
	}

	printData := printer.PrintData{
		StationID:      reply.Material.Station,
		NextStationID:  "",
		ProductID:      reply.Material.ID,
		ProductionDate: reply.Material.ProductionTime,
		ExpiryDate:     reply.Material.ExpiryTime,
		Quantity:       reply.Material.Quantity,
		ResourceID:     reply.Material.ResourceID,
	}

	f, err := printer.CreateResourcesPDF(ctx, models.MaterialResourceLabelFieldName{}, printData, barcodes.Code39{}, r.config.FontPath)
	if err != nil {
		return resource.NewDownloadMaterialResourceDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Details: err.Error(),
		})
	}
	return resource.NewDownloadMaterialResourceOK().WithPayload(f)
}

// GetToolID implements.
func (r Resource) GetToolID(params resource.GetToolIDParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_GET_TOOL_ID, principal.Roles) {
		return resource.NewGetToolIDDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := r.dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
		ResourceID: params.ToolResourceID,
	})
	if err != nil {
		return utils.ParseError(ctx, resource.NewGetToolIDDefault(0), err)
	}

	return resource.NewGetToolIDOK().WithPayload(&resource.GetToolIDOKBody{
		Data: &resource.GetToolIDOKBodyData{
			ToolID: list.ToolID,
		},
	})
}

// PrintMaterialResource implements.
func (r Resource) PrintMaterialResource(params resource.PrintMaterialResourceParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_PRINT_MATERIAL_RESOURCE, principal.Roles) {
		return resource.NewPrintMaterialResourceDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	// Read Config Printer
	getWorkOrder, err := r.dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{
		ID: params.Body.WorkOrderID,
	})
	if err != nil {
		return utils.ParseError(ctx, resource.NewPrintMaterialResourceDefault(0), err)
	}

	getCollect, err := r.dm.GetCollectRecord(ctx, mcom.GetCollectRecordRequest{
		WorkOrder: params.Body.WorkOrderID,
		Sequence:  int16(params.Body.Sequence),
	})
	if err != nil {
		return utils.ParseError(ctx, resource.NewPrintMaterialResourceDefault(0), err)
	}

	getResource, err := r.dm.GetMaterialResourceIdentity(ctx, mcom.GetMaterialResourceIdentityRequest{
		ResourceID:  getCollect.ResourceID,
		ProductType: getWorkOrder.Product.Type,
	})
	if err != nil {
		return utils.ParseError(ctx, resource.NewPrintMaterialResourceDefault(0), err)
	}

	printData := printer.PrintData{
		StationID:      getWorkOrder.Station,
		NextStationID:  "",
		ProductID:      getCollect.ProductID,
		ProductionDate: getResource.Material.ProductionTime,
		ExpiryDate:     getResource.Material.ExpiryTime,
		Quantity:       getCollect.Quantity,
		ResourceID:     getCollect.ResourceID,
	}

	pdf, err := printer.CreateResourcesPDF(ctx, models.MaterialResourceLabelFieldName{}, printData, barcodes.Code39{}, r.config.FontPath)
	if err != nil {
		return utils.ParseError(ctx, resource.NewPrintMaterialResourceDefault(0), err)
	}

	if printer := r.config.Printers[getWorkOrder.Station]; printer == "" {
		return utils.ParseError(ctx, resource.NewPrintMaterialResourceDefault(0), mcomErrors.Error{
			Code:    mcomErrors.Code_STATION_PRINTER_NOT_DEFINED,
			Details: fmt.Sprintf("station %s no defined printer", getWorkOrder.Station),
		})
	}

	err = mcom.Print(ctx, r.config.Printers[getWorkOrder.Station], pdf)
	if err != nil {
		return utils.ParseError(ctx, resource.NewPrintMaterialResourceDefault(0), err)
	}

	return resource.NewPrintMaterialResourceOK()
}

func (r Resource) DownloadPreMaterialResource(params resource.DownloadPreMaterialResourceParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_DOWNLOAD_PRE_MATERIAL_RESOURCE, principal.Roles) {
		return resource.NewDownloadPreMaterialResourceDefault(http.StatusForbidden)
	}
	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	now := time.Now()
	getWorkOrder, err := r.dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{
		ID: params.WorkOrderID,
	})
	if err != nil {
		return utils.ParseError(ctx, resource.NewDownloadPreMaterialResourceDefault(0), err)
	}

	getLimitaryHour, err := r.dm.GetLimitaryHour(ctx, mcom.GetLimitaryHourRequest{ProductType: getWorkOrder.Product.Type})
	if err != nil {
		if e, ok := mcomErrors.As(err); !ok || e.Code != mcomErrors.Code_LIMITARY_HOUR_NOT_FOUND {
			return utils.ParseError(ctx, resource.NewDownloadPreMaterialResourceDefault(0), err)
		}
	}
	expiryTime := now.Add(time.Duration(getLimitaryHour.LimitaryHour.Max) * time.Hour)

	batchQuantityDetails, err := handlerUtils.ParseBatchQuantityDetails(getWorkOrder.BatchQuantityDetails)
	if err != nil {
		return resource.NewDownloadPreMaterialResourceDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Details: err.Error(),
			},
		)
	}
	resources, err := r.dm.CreateMaterialResources(ctx, mcom.CreateMaterialResourcesRequest{
		Materials: []mcom.CreateMaterialResourcesRequestDetail{
			{
				Type:            getWorkOrder.Product.Type,
				ID:              getWorkOrder.Product.ID,
				Grade:           "",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.Zero,
				PlannedQuantity: batchQuantityDetails.PlanQuantity,
				Station:         getWorkOrder.Station,
				Unit:            getWorkOrder.Unit,
				LotNumber:       "",
				ProductionTime:  now,
				ExpiryTime:      expiryTime,
				ResourceID:      "",
				MinDosage:       decimal.Decimal{},
				Inspections:     []mcomModels.Inspection{},
				Remark:          "",
				CarrierID:       "",
			},
		},
	})
	if err != nil {
		return utils.ParseError(ctx, resource.NewDownloadPreMaterialResourceDefault(0), err)
	}

	//Print
	printData := printer.PrintData{
		StationID:      getWorkOrder.Station,
		NextStationID:  "",
		ProductID:      getWorkOrder.Product.ID,
		ProductionDate: now,
		ExpiryDate:     expiryTime,
		Quantity:       batchQuantityDetails.PlanQuantity,
		ResourceID:     resources[0].ID,
	}

	f, err := printer.CreateResourcesPDF(ctx, *params.Body.FieldName, printData, barcodes.Code39{}, r.config.FontPath)
	if err != nil {
		return resource.NewDownloadPreMaterialResourceDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Details: err.Error(),
		})
	}
	return resource.NewDownloadPreMaterialResourceOK().WithPayload(f)
}

func parseResourceMaterials(replies []mcom.MaterialReply) models.ResourceMaterials {
	resources := make(models.ResourceMaterials, len(replies))
	for i, resourceReply := range replies {
		warehouseID, warehouseLocation := resourceReply.Warehouse.ID, resourceReply.Warehouse.Location
		resources[i] = &models.ResourceMaterial{
			ID:            resourceReply.Material.ID,
			CarrierID:     resourceReply.Material.CarrierID,
			CreatedAt:     strfmt.DateTime(resourceReply.Material.CreatedAt.Time()),
			CreatedBy:     resourceReply.Material.CreatedBy,
			ExpiredDate:   strfmt.DateTime(resourceReply.Material.ExpiryTime),
			Grade:         models.Grade(resourceReply.Material.Grade),
			Inspections:   inspectionStructType(resourceReply.Material.Inspections),
			MinimumDosage: resourceReply.Material.MinDosage.String(),
			Remark:        resourceReply.Material.Remark,
			ProductType:   resourceReply.Material.Type,
			Quantity:      resourceReply.Material.Quantity.String(),
			Unit:          resourceReply.Material.Unit,
			ResourceID:    resourceReply.Material.ResourceID,
			Status:        models.MaterialStatus(resourceReply.Material.Status),
			UpdatedAt:     strfmt.DateTime(resourceReply.Material.UpdatedAt.Time()),
			UpdatedBy:     resourceReply.Material.UpdatedBy,
			Warehouse: &models.Warehouse{
				ID:       &warehouseID,
				Location: &warehouseLocation,
			},
		}
	}
	return resources
}

func inspectionStructType(dataIn mcomModels.Inspections) []*models.ResourceMaterialInspectionsItems0 {

	dataOut := make([]*models.ResourceMaterialInspectionsItems0, len(dataIn))
	for i, d := range dataIn {
		dataOut[i] = &models.ResourceMaterialInspectionsItems0{
			ID:     int64(d.ID),
			Remark: d.Remark,
		}
	}
	return dataOut
}
