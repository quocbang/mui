package warehouse

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/warehouse"
)

// Warehouse definitions.
type Warehouse struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewWarehouse returns Warehouse service.
func NewWarehouse(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.Warehouse {
	return Warehouse{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// GetWarehouseInfo implementations.
func (w Warehouse) GetWarehouseInfo(params warehouse.GetWarehouseInfoParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_GET_WAREHOUSE_INFO, principal.Roles) {
		return warehouse.NewGetWarehouseInfoDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	resourceWarehouse, err := w.dm.GetResourceWarehouse(ctx, mcom.GetResourceWarehouseRequest{
		ResourceID: params.ID,
	})
	if err != nil {
		return utils.ParseError(ctx, warehouse.NewGetWarehouseInfoDefault(0), err)
	}

	return warehouse.NewGetWarehouseInfoOK().WithPayload(&warehouse.GetWarehouseInfoOKBody{
		Data: &warehouse.GetWarehouseInfoOKBodyData{
			Location:    resourceWarehouse.Location,
			WarehouseID: resourceWarehouse.ID,
		},
	})
}

// WarehouseTransaction implementations.
func (w Warehouse) WarehouseTransaction(params warehouse.WarehouseTransactionParams, principal *models.Principal) middleware.Responder {
	if !w.hasPermission(kenda.FunctionOperationID_WAREHOUSE_TRANSACTION, principal.Roles) {
		return warehouse.NewWarehouseTransactionDefault(http.StatusForbidden)
	}
	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	if err := w.dm.WarehousingStock(ctx, mcom.WarehousingStockRequest{
		Warehouse: mcom.Warehouse{
			ID:       *params.Body.NewWarehouseID,
			Location: *params.Body.NewLocation,
		},
		ResourceIDs: []string{params.ID},
	}); err != nil {
		return utils.ParseError(ctx, warehouse.NewWarehouseTransactionDefault(0), err)

	}
	return warehouse.NewWarehouseTransactionOK()
}
