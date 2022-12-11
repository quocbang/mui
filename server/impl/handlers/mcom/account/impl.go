package account

import (
	"context"
	"fmt"
	"net/http"
	"time"

	apiErrors "github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"

	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/account"
)

const (
	// AuthorizationKey use for authentication of each secure services.
	AuthorizationKey = "x-mui-auth-key"

	invalidUserError  = "invalid user"
	tokenExpiredError = "token expired"
)

// Authorization definitions.
type Authorization struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
	tokenLifeTime time.Duration
}

// NewAuthorization returns Authorization service.
func NewAuthorization(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool, tokenLifeTime time.Duration) service.AccountAuthorization {
	return Authorization{
		dm:            dm,
		hasPermission: hasPermission,
		tokenLifeTime: tokenLifeTime,
	}
}

// Auth implementation.
func (a Authorization) Auth(token string) (*models.Principal, error) {
	tokenInfo, err := a.dm.GetTokenInfo(context.Background(), mcom.GetTokenInfoRequest{
		Token: token,
	})
	if err != nil {
		if e, ok := mcomErrors.As(err); ok {
			return nil, apiErrors.New(http.StatusUnauthorized, fmt.Sprintf("%v", mcomErrors.Error{
				Code:    e.Code,
				Details: e.Details,
			}))
		}
		return nil, err
	}

	if !tokenInfo.Valid {
		return nil, apiErrors.New(http.StatusUnauthorized, invalidUserError)
	}

	if tokenInfo.ExpiryTime.Before(time.Now().Local()) {
		return nil, apiErrors.New(http.StatusUnauthorized, tokenExpiredError)
	}

	return &models.Principal{
		ID:    tokenInfo.User,
		Roles: handlerUtils.ToModelsRoles(tokenInfo.Roles),
	}, nil
}

// Login handler implementation.
func (a Authorization) Login(params account.LoginParams) middleware.Responder {
	id := *params.Body.ID
	signInRequest := mcom.SignInRequest{
		Account:  id,
		Password: *params.Body.Password,
	}

	if *params.Body.LoginType == 1 {
		signInRequest.ADUser = true
	}

	var options []mcom.SignInOption
	if a.tokenLifeTime > 0 {
		options = append(options, mcom.WithTokenExpiredAfter(a.tokenLifeTime))
	}

	signInReply, err := a.dm.SignIn(params.HTTPRequest.Context(), signInRequest, options...)
	if err != nil {
		if e, ok := mcomErrors.As(err); ok {
			return account.NewLoginBadRequest().WithPayload(&models.Error{
				Code:    int64(e.Code),
				Details: e.Details,
			})
		}
		return account.NewLoginInternalServerError().WithPayload(&models.Error{
			Details: err.Error(),
		})
	}

	return account.NewLoginOK().WithPayload(&account.LoginOKBody{Data: &models.LoginResponse{
		Token:                 signInReply.Token,
		TokenExpiry:           strfmt.DateTime(signInReply.TokenExpiry),
		Roles:                 handlerUtils.ToModelsRoles(signInReply.Roles),
		AuthorizedDepartments: handlerUtils.ToDepartmentsModel(signInReply.Departments),
	}})
}

// Logout handler implementation.
func (a Authorization) Logout(params account.LogoutParams) middleware.Responder {
	token := params.HTTPRequest.Header.Get(AuthorizationKey)

	ctx := params.HTTPRequest.Context()

	if err := a.dm.SignOut(ctx, mcom.SignOutRequest{
		Token: token,
	}); err != nil {
		logger := zap.L()
		logFunc := logger.Error
		if _, ok := mcomErrors.As(err); ok {
			logFunc = logger.Warn
		}
		logFunc("logout failed..", zap.String("token", token), zap.Error(err))
	}

	return account.NewLogoutOK()
}

// ChangePassword implementations.
func (a Authorization) ChangePassword(params account.ChangePasswordParams, principal *models.Principal) middleware.Responder {
	if !a.hasPermission(kenda.FunctionOperationID_CHANGE_USER_PASSWORD, principal.Roles) {
		return account.NewChangePasswordDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	if err := a.dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
		UserID: principal.ID,
		ChangePassword: &struct {
			NewPassword string
			OldPassword string
		}{
			*params.Body.NewPassword,
			*params.Body.CurrentPassword,
		},
	}); err != nil {
		return utils.ParseError(ctx, account.NewChangePasswordDefault(0), err)
	}

	return account.NewChangePasswordOK()
}

// ListAuthorizedAccount implementations
func (a Authorization) ListAuthorizedAccount(params account.ListAuthorizedAccountParams, principal *models.Principal) middleware.Responder {
	if !a.hasPermission(kenda.FunctionOperationID_LIST_AUTHORIZED_ACCOUNT, principal.Roles) {
		return account.NewListAuthorizedAccountDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	list, err := a.dm.ListUserRoles(ctx, mcom.ListUserRolesRequest{
		DepartmentID: params.DepartmentOID,
	})
	if err != nil {
		return utils.ParseError(ctx, account.NewListAuthorizedAccountDefault(0), err)
	}

	data := make([]*models.AccountDataItems0, len(list.Users))
	for i, user := range list.Users {
		id := user.ID
		data[i] = &models.AccountDataItems0{
			EmployeeID: &id,
			Roles:      handlerUtils.ToModelsRoles(user.Roles),
		}
	}

	return account.NewListAuthorizedAccountOK().WithPayload(&account.ListAuthorizedAccountOKBody{
		Data: data,
	})
}

// ListUnauthorizedAccount implementations
func (a Authorization) ListUnauthorizedAccount(params account.ListUnauthorizedAccountParams, principal *models.Principal) middleware.Responder {
	if !a.hasPermission(kenda.FunctionOperationID_LIST_UNAUTHORIZED_ACCOUNT, principal.Roles) {
		return account.NewListUnauthorizedAccountDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	list, err := a.dm.ListUnauthorizedUsers(ctx, mcom.ListUnauthorizedUsersRequest{
		DepartmentID: params.DepartmentOID,
	}, mcom.ExcludeUsers([]string{principal.ID}))
	if err != nil {
		return utils.ParseError(ctx, account.NewListUnauthorizedAccountDefault(0), err)
	}

	data := make([]*account.ListUnauthorizedAccountOKBodyDataItems0, len(list))
	for i, user := range list {
		data[i] = &account.ListUnauthorizedAccountOKBodyDataItems0{
			EmployeeID: user,
		}
	}

	return account.NewListUnauthorizedAccountOK().WithPayload(&account.ListUnauthorizedAccountOKBody{
		Data: data,
	})
}

// GetRoleList implementations
func (a Authorization) GetRoleList(params account.GetRoleListParams, principal *models.Principal) middleware.Responder {
	if !a.hasPermission(kenda.FunctionOperationID_GET_ROLE_LIST, principal.Roles) {
		return account.NewGetRoleListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	list, err := a.dm.ListRoles(ctx)
	if err != nil {
		return utils.ParseError(ctx, account.NewGetRoleListDefault(0), err)
	}

	data := make([]*account.GetRoleListOKBodyDataItems0, len(list.Roles))
	for i, role := range list.Roles {
		data[i] = &account.GetRoleListOKBodyDataItems0{
			ID:   models.Role(role.Value),
			Name: role.Name,
		}
	}

	return account.NewGetRoleListOK().WithPayload(&account.GetRoleListOKBody{Data: data})
}

// Create implementations
func (a Authorization) CreateAccountAuthorization(params account.CreateAccountAuthorizationParams, principal *models.Principal) middleware.Responder {
	if !a.hasPermission(kenda.FunctionOperationID_CREATE_ACCOUNT, principal.Roles) {
		return account.NewCreateAccountAuthorizationDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	req := mcom.CreateAccountsRequest{
		mcom.CreateAccountRequest{
			ID:    *params.Body.EmployeeID,
			Roles: handlerUtils.FromModelsRoles(params.Body.Roles),
		}.WithDefaultPassword(),
	}
	if err := a.dm.CreateAccounts(ctx, req); err != nil {
		return utils.ParseError(ctx, account.NewCreateAccountAuthorizationDefault(0), err)
	}

	return account.NewCreateAccountAuthorizationOK()
}

// Update implementations
func (a Authorization) UpdateAccountAuthorization(params account.UpdateAccountAuthorizationParams, principal *models.Principal) middleware.Responder {
	if !a.hasPermission(kenda.FunctionOperationID_UPDATE_ACCOUNT, principal.Roles) {
		return account.NewUpdateAccountAuthorizationDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	var requestOptions []mcom.UpdateAccountOption
	if *params.Body.ResetPassword {
		requestOptions = append(requestOptions, mcom.ResetPassword())
	}

	req := mcom.UpdateAccountRequest{
		UserID: params.EmployeeID,
		Roles:  handlerUtils.FromModelsRoles(params.Body.Roles),
	}
	if err := a.dm.UpdateAccount(ctx, req, requestOptions...); err != nil {
		return utils.ParseError(ctx, account.NewUpdateAccountAuthorizationDefault(0), err)
	}

	return account.NewUpdateAccountAuthorizationOK()
}

// DeleteAccount implementations
func (a Authorization) DeleteAccount(params account.DeleteAccountParams, principal *models.Principal) middleware.Responder {
	if !a.hasPermission(kenda.FunctionOperationID_DELETE_ACCOUNT, principal.Roles) {
		return account.NewDeleteAccountDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	if err := a.dm.DeleteAccount(ctx, mcom.DeleteAccountRequest{ID: params.EmployeeID}); err != nil {
		return utils.ParseError(ctx, account.NewDeleteAccountDefault(0), err)
	}

	return account.NewDeleteAccountOK()
}
