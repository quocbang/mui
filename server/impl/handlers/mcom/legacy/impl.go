package legacy

import (
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/legacy"
)

// Legacy for Legacy structure.
type Legacy struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewLegacy returns Legacy.
func NewLegacy(dm mcom.DataManager, hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.Legacy {
	return Legacy{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// GetBarcodeInfo implementations.
func (p Legacy) GetBarcodeInfo(params legacy.GetBarcodeInfoParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_BARCODE_INFO, principal.Roles) {
		return legacy.NewGetBarcodeInfoDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	getMaterialData, err := p.dm.GetMaterial(ctx, mcom.GetMaterialRequest{
		MaterialID: params.ID,
	})
	if err != nil {
		return utils.ParseError(ctx, legacy.NewGetBarcodeInfoDefault(0), err)
	}
	return legacy.NewGetBarcodeInfoOK().WithPayload(
		&legacy.GetBarcodeInfoOKBody{
			Data: &legacy.GetBarcodeInfoOKBodyData{Material: &models.Material{
				Barcode:   getMaterialData.MaterialID,
				ExpiredAt: strfmt.Date(getMaterialData.ExpireDate),
				Inventory: getMaterialData.Quantity.String(),
				ProductID: getMaterialData.MaterialProductID,
				Status:    getMaterialData.Status,
			}},
		})
}

// UpdateBarcode implementation.
func (p Legacy) UpdateBarcode(params legacy.UpdateBarcodeParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_UPDATE_BARCODE, principal.Roles) {
		return legacy.NewUpdateBarcodeDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	err := p.dm.UpdateMaterial(ctx, mcom.UpdateMaterialRequest{
		MaterialID:       params.ID,
		ExtendedDuration: time.Duration(*params.Body.ExtendDays * 24 * 60 * 60 * 1000000000),
		User:             principal.ID,
		NewStatus:        *params.Body.NewStatus,
		Reason:           *params.Body.HoldReason,
		ProductCate:      *params.Body.ProductCate,
		ControlArea:      *params.Body.ControlArea,
	})
	if err != nil {
		return utils.ParseError(ctx, legacy.NewUpdateBarcodeDefault(0), err)
	}
	return legacy.NewUpdateBarcodeOK()
}

// GetUpdateBarcodeStatusList implementation.
func (p Legacy) GetUpdateBarcodeStatusList(params legacy.GetUpdateBarcodeStatusListParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_UPDATE_BARCODE_STATUS_LIST, principal.Roles) {
		return legacy.NewGetUpdateBarcodeStatusListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	statusList, err := p.dm.ListChangeableStatus(ctx, mcom.ListChangeableStatusRequest{
		MaterialID: params.ID,
	})
	if err != nil {
		return utils.ParseError(ctx, legacy.NewGetUpdateBarcodeStatusListDefault(0), err)
	}
	statusCodeList := make([]*models.CodeDescription, len(statusList.Codes))
	for i, code := range statusList.Codes {
		statusCodeList[i] = &models.CodeDescription{
			Code:        code.Code,
			Description: code.CodeDescription,
		}
	}
	return legacy.NewGetUpdateBarcodeStatusListOK().WithPayload(&legacy.GetUpdateBarcodeStatusListOKBody{Data: statusCodeList})
}

// GetExtendDays implementation.
func (p Legacy) GetExtendDays(params legacy.GetExtendDaysParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_EXTEND_DAYS, principal.Roles) {
		return legacy.NewGetExtendDaysDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	duration, err := p.dm.GetMaterialExtendDate(ctx, mcom.GetMaterialExtendDateRequest{
		MaterialID: params.ID,
	})
	if err != nil {
		return utils.ParseError(ctx, legacy.NewGetExtendDaysDefault(0), err)
	}
	return legacy.NewGetExtendDaysOK().WithPayload(&legacy.GetExtendDaysOKBody{
		Data: &legacy.GetExtendDaysOKBodyData{
			ExtendDay: int64(time.Duration(duration).Hours() / 24),
		},
	})
}

// GetControlAreaList implementation.
func (p Legacy) GetControlAreaList(params legacy.GetControlAreaListParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_CONTROL_AREA_LIST, principal.Roles) {
		return legacy.NewGetControlAreaListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	controlArea, err := p.dm.ListControlAreas(ctx)
	if err != nil {
		return utils.ParseError(ctx, legacy.NewGetControlAreaListDefault(0), err)
	}
	controlAreaList := make([]*models.CodeDescription, len(controlArea.Codes))
	for i, code := range controlArea.Codes {
		controlAreaList[i] = &models.CodeDescription{
			Code:        code.Code,
			Description: code.CodeDescription,
		}
	}
	return legacy.NewGetControlAreaListOK().WithPayload(&legacy.GetControlAreaListOKBody{Data: controlAreaList})
}

// GetHoldReasonList implementation.
func (p Legacy) GetHoldReasonList(params legacy.GetHoldReasonListParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_HOLD_REASON_LIST, principal.Roles) {
		return legacy.NewGetHoldReasonListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	controlReason, err := p.dm.ListControlReasons(ctx)
	if err != nil {
		return utils.ParseError(ctx, legacy.NewGetHoldReasonListDefault(0), err)
	}
	holdReasonList := make([]*models.CodeDescription, len(controlReason.Codes))
	for i, code := range controlReason.Codes {
		holdReasonList[i] = &models.CodeDescription{
			Code:        code.Code,
			Description: code.CodeDescription,
		}
	}
	return legacy.NewGetHoldReasonListOK().WithPayload(&legacy.GetHoldReasonListOKBody{Data: holdReasonList})
}
