package plan

import (
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/shopspring/decimal"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/plan"
)

// Plan definitions.
type Plan struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewPlan returns Plan service.
func NewPlan(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.Plan {
	return Plan{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// List implements GetPlanList.
func (p Plan) GetPlanList(params plan.GetPlanListParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_PLAN_LIST, principal.Roles) {
		return plan.NewGetPlanListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := p.dm.ListProductPlans(ctx, mcom.ListProductPlansRequest{
		Date:         time.Time(params.Date),
		DepartmentID: params.DepartmentOID,
		ProductType:  params.ProductType,
	})
	if err != nil {
		return utils.ParseError(ctx, plan.NewGetPlanListDefault(0), err)
	}
	data := make([]*models.PlanData, len(list.ProductPlans))
	for i, productPlan := range list.ProductPlans {
		data[i] = &models.PlanData{
			ProductID:         productPlan.ProductID,
			DayQuantity:       productPlan.Quantity.Daily.String(),
			WeekQuantity:      productPlan.Quantity.Week.String(),
			StockQuantity:     productPlan.Quantity.Stock.String(),
			ScheduledQuantity: productPlan.Quantity.Reserved.String(),
		}
	}
	return plan.NewGetPlanListOK().WithPayload(&plan.GetPlanListOKBody{Data: data})
}

// Create implements AddPlan.
func (p Plan) AddPlan(params plan.AddPlanParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_ADD_PLAN, principal.Roles) {
		return plan.NewAddPlanDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	dayQuantity, err := decimal.NewFromString(*params.Body.DayQuantity)
	if err != nil {
		return plan.NewAddPlanDefault(http.StatusBadRequest).WithPayload(&models.Error{
			Code:    int64(mcomErrors.Code_INVALID_NUMBER),
			Details: "invalid_number=" + *params.Body.DayQuantity,
		})
	}
	if err := p.dm.CreateProductPlan(ctx, mcom.CreateProductionPlanRequest{
		Date: time.Time(*params.Body.Date),
		Product: mcom.Product{
			ID:   *params.Body.ProductID,
			Type: *params.Body.ProductType,
		},
		DepartmentID: *params.Body.DepartmentOID,
		Quantity:     dayQuantity,
	}); err != nil {
		return utils.ParseError(ctx, plan.NewAddPlanDefault(0), err)
	}

	return plan.NewAddPlanOK()
}
