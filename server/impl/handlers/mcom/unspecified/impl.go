package unspecified

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/unspecified"
)

// Unspecified definitions.
type Unspecified struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewUnspecified returns Unspecified service.
func NewUnspecified(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.Unspecified {
	return Unspecified{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// ListDepartmentIDs implementation.
func (u Unspecified) ListDepartmentIDs(params unspecified.ListDepartmentIDsParams, principal *models.Principal) middleware.Responder {
	if !u.hasPermission(kenda.FunctionOperationID_LIST_DEPARTMENT_IDS, principal.Roles) {
		return unspecified.NewListDepartmentIDsDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := u.dm.ListAllDepartment(ctx)
	if err != nil {
		return utils.ParseError(ctx, unspecified.NewListDepartmentIDsDefault(0), err)
	}

	data := make([]*unspecified.ListDepartmentIDsOKBodyDataItems0, len(list.IDs))
	for i, department := range list.IDs {
		data[i] = &unspecified.ListDepartmentIDsOKBodyDataItems0{
			DepartmentID: department,
		}
	}
	return unspecified.NewListDepartmentIDsOK().WithPayload(&unspecified.ListDepartmentIDsOKBody{Data: data})
}
