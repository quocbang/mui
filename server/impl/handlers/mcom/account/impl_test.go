package account

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	apiErrors "github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	authorization "gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/account"
)

const (
	tokenFor = "token-for-"

	brokenError             = "broken user"
	invalidToken            = "invalid token"
	logoutFailed            = "logout failed"
	internalError           = "internal server error"
	notFoundToken           = "token not found"
	alreadyLogout           = "already logout"
	userNotFound            = "user not found"
	missingPassword         = "missing password"
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

	testEmpty           = ""
	testUsernameDan     = "dan"
	testUsernameSpencer = "spencer"

	testDepartmentOID = "M2110"

	testDanRoles = []mcomRoles.Role{
		mcomRoles.Role_INSPECTOR,
		mcomRoles.Role_QUALITY_CONTROLLER,
	}

	testSpencerRoles = []mcomRoles.Role{
		mcomRoles.Role_SCHEDULER,
		mcomRoles.Role_OPERATOR,
	}

	falseReset = false
	trueReset  = true

	loginType  = models.LoginType(0)
	userID     = "tester"
	badUserID  = "xxx"
	brokenUser = "broken"
	password   = "p4s5w0rd"
	empty      = ""

	tokenExpiry = time.Now().Add(8 * time.Hour)
	expiredDate = time.Date(1999, 07, 30, 0, 0, 0, 0, time.Local)

	departments = []mcom.Department{
		{
			OID: "M2110xx",
			ID:  "M2110",
		},
	}
	roles = []mcomRoles.Role{
		mcomRoles.Role_ADMINISTRATOR,
		mcomRoles.Role_LEADER,
	}
)

func TestAuthorization(t *testing.T) {
	assert := assert.New(t)

	httpRequest := httptest.NewRequest(http.MethodPost, "/login", nil)

	{ // broken user
		scripts := []mock.Script{
			{ // login internal error
				Name: mock.FuncSignIn,
				Input: mock.Input{
					Request: mcom.SignInRequest{
						Account:  brokenUser,
						Password: password,
					},
				},
				Output: mock.Output{
					Error: errors.New(brokenError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		params := authorization.LoginParams{
			HTTPRequest: httpRequest,
			Body: &models.LoginRequest{
				ID:        &brokenUser,
				Password:  &password,
				LoginType: &loginType,
			},
		}
		r, ok := u.Login(params).(*authorization.LoginInternalServerError)
		assert.True(ok)
		assert.Equal(brokenError, r.Payload.Details)
		assert.NoError(dm.Close())
	}
	{ // Auth: invalid token.
		scripts := []mock.Script{
			{ // invalid token
				Name: mock.FuncGetTokenInfo,
				Input: mock.Input{
					Request: mcom.GetTokenInfoRequest{
						Token: tokenFor + brokenUser,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code:    mcomErrors.Code_USER_UNKNOWN_TOKEN,
						Details: invalidToken,
					},
				},
			},
		}

		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		_, err = u.Auth(tokenFor + brokenUser)
		assert.Equal(apiErrors.New(http.StatusUnauthorized, fmt.Sprintf("%v", mcomErrors.Error{
			Code:    mcomErrors.Code_USER_UNKNOWN_TOKEN,
			Details: invalidToken,
		})), err)
		assert.NoError(dm.Close())
	}
	{ // Logout failed before login, but also allowed logout
		scripts := []mock.Script{
			{ // logout failed
				Name: mock.FuncSignOut,
				Input: mock.Input{
					Request: mcom.SignOutRequest{
						Token: tokenFor + brokenUser,
					},
				},
				Output: mock.Output{
					Error: errors.New(logoutFailed),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		logoutRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
		logoutRequest.Header.Set(AuthorizationKey, tokenFor+brokenUser)
		params := authorization.LogoutParams{HTTPRequest: logoutRequest}
		_, ok := u.Logout(params).(*authorization.LogoutOK)
		assert.True(ok)
		assert.NoError(dm.Close())
	}
	{ // login success with options
		scripts := []mock.Script{
			{ // login success with options
				Name: mock.FuncSignIn,
				Input: mock.Input{
					Request: mcom.SignInRequest{
						Account:  userID,
						Password: password,
					},
					Options: []interface{}{
						mcom.WithTokenExpiredAfter(8 * time.Hour),
					},
				},
				Output: mock.Output{
					Response: mcom.SignInReply{
						Token:       tokenFor + userID,
						TokenExpiry: tokenExpiry,
						Departments: departments,
						Roles:       roles,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 8*60*60*time.Second)

		params := authorization.LoginParams{
			HTTPRequest: httpRequest,
			Body: &models.LoginRequest{
				ID:        &userID,
				Password:  &password,
				LoginType: &loginType,
			},
		}
		r, ok := u.Login(params).(*authorization.LoginOK)
		if assert.True(ok) {
			assert.Equal(&models.LoginResponse{
				AuthorizedDepartments: handlerUtils.ToDepartmentsModel(departments),
				Roles:                 handlerUtils.ToModelsRoles(roles),
				Token:                 tokenFor + userID,
				TokenExpiry:           strfmt.DateTime(tokenExpiry),
			}, r.Payload.Data)
		}
		assert.NoError(dm.Close())
	}
	{ // login success with AD
		scripts := []mock.Script{
			{
				Name: mock.FuncSignIn,
				Input: mock.Input{
					Request: mcom.SignInRequest{
						Account:  userID,
						Password: password,
						ADUser:   true,
					},
				},
				Output: mock.Output{
					Response: mcom.SignInReply{
						Token:       tokenFor + userID,
						Departments: departments,
						Roles:       roles,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		}, 0)

		windowsLoginType := models.LoginType(1)

		params := authorization.LoginParams{
			HTTPRequest: httpRequest,
			Body: &models.LoginRequest{
				ID:        &userID,
				Password:  &password,
				LoginType: &windowsLoginType,
			},
		}
		r, ok := u.Login(params).(*authorization.LoginOK)
		if assert.True(ok) {
			assert.Equal(&models.LoginResponse{
				AuthorizedDepartments: handlerUtils.ToDepartmentsModel(departments),
				Roles:                 handlerUtils.ToModelsRoles(roles),
				Token:                 tokenFor + userID,
			}, r.Payload.Data)
		}
		assert.NoError(dm.Close())
	}
	{ // Auth success.
		scripts := []mock.Script{
			{ // getTokenInfo success
				Name: mock.FuncGetTokenInfo,
				Input: mock.Input{
					Request: mcom.GetTokenInfoRequest{
						Token: tokenFor + userID,
					},
				},
				Output: mock.Output{
					Response: mcom.GetTokenInfoReply{
						User:        userID,
						Valid:       true,
						ExpiryTime:  tokenExpiry,
						CreatedTime: time.Now(),
						Roles:       roles,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		_, err = u.Auth(tokenFor + userID)
		assert.NoError(err)
		assert.NoError(dm.Close())
	}
	{ // Auth internal error.
		scripts := []mock.Script{
			{ // getTokenInfo internal error
				Name: mock.FuncGetTokenInfo,
				Input: mock.Input{
					Request: mcom.GetTokenInfoRequest{
						Token: tokenFor + userID,
					},
				},
				Output: mock.Output{
					Error: errors.New(internalError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		_, err = u.Auth(tokenFor + userID)
		assert.EqualError(err, internalError)
		assert.NoError(dm.Close())
	}
	{ // Auth bad request.
		scripts := []mock.Script{
			{ // token not found
				Name: mock.FuncGetTokenInfo,
				Input: mock.Input{
					Request: mcom.GetTokenInfoRequest{
						Token: tokenFor + userID,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code:    mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
						Details: notFoundToken,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		_, err = u.Auth(tokenFor + userID)
		assert.Equal(apiErrors.New(http.StatusUnauthorized, fmt.Sprintf("%v", mcomErrors.Error{
			Code:    mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
			Details: notFoundToken,
		})), err)
		assert.NoError(dm.Close())
	}
	{ // Auth user invalid.
		scripts := []mock.Script{
			{ // invalid user
				Name: mock.FuncGetTokenInfo,
				Input: mock.Input{
					Request: mcom.GetTokenInfoRequest{
						Token: tokenFor + userID,
					},
				},
				Output: mock.Output{
					Response: mcom.GetTokenInfoReply{
						User:        userID,
						Valid:       false,
						ExpiryTime:  tokenExpiry,
						CreatedTime: time.Now(),
						Roles:       roles,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		_, err = u.Auth(tokenFor + userID)
		assert.EqualError(err, invalidUserError)
		assert.NoError(dm.Close())
	}
	{ // Auth expired token.
		scripts := []mock.Script{
			{ // expired token
				Name: mock.FuncGetTokenInfo,
				Input: mock.Input{
					Request: mcom.GetTokenInfoRequest{
						Token: tokenFor + userID,
					},
				},
				Output: mock.Output{
					Response: mcom.GetTokenInfoReply{
						User:        userID,
						Valid:       true,
						ExpiryTime:  expiredDate,
						CreatedTime: time.Now(),
						Roles:       roles,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		_, err = u.Auth(tokenFor + userID)
		assert.EqualError(err, tokenExpiredError)
		assert.NoError(dm.Close())
	}
	{ // logout success.
		scripts := []mock.Script{
			{ // logout success
				Name: mock.FuncSignOut,
				Input: mock.Input{
					Request: mcom.SignOutRequest{
						Token: tokenFor + userID,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		logoutRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
		logoutRequest.Header.Set(AuthorizationKey, tokenFor+userID)
		params := authorization.LogoutParams{HTTPRequest: logoutRequest}
		_, ok := u.Logout(params).(*authorization.LogoutOK)
		assert.True(ok)
		assert.NoError(dm.Close())
	}
	{ // logout error. (mes error code), but also allowed logout
		scripts := []mock.Script{
			{ // already logout
				Name: mock.FuncSignOut,
				Input: mock.Input{
					Request: mcom.SignOutRequest{
						Token: tokenFor + userID,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code:    mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
						Details: alreadyLogout,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		logoutRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
		logoutRequest.Header.Set(AuthorizationKey, tokenFor+userID)
		params := authorization.LogoutParams{HTTPRequest: logoutRequest}
		_, ok := u.Logout(params).(*authorization.LogoutOK)
		assert.True(ok)
		assert.NoError(dm.Close())
	}
	{ // logout error. (internal server error), but also allowed logout
		scripts := []mock.Script{
			{ // logout internal server error
				Name: mock.FuncSignOut,
				Input: mock.Input{
					Request: mcom.SignOutRequest{
						Token: tokenFor + userID,
					},
				},
				Output: mock.Output{
					Error: errors.New(internalError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		logoutRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
		logoutRequest.Header.Set(AuthorizationKey, tokenFor+userID)
		params := authorization.LogoutParams{HTTPRequest: logoutRequest}
		_, ok := u.Logout(params).(*authorization.LogoutOK)
		assert.True(ok)
		assert.NoError(dm.Close())
	}
	{ // user not exist.
		scripts := []mock.Script{
			{ // user not exist
				Name: mock.FuncSignIn,
				Input: mock.Input{
					Request: mcom.SignInRequest{
						Account:  badUserID,
						Password: password,
					},
				},
				Output: mock.Output{
					Error: &mcomErrors.Error{
						Code:    mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
						Details: userNotFound,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		params := authorization.LoginParams{
			HTTPRequest: httpRequest,
			Body: &models.LoginRequest{
				ID:        &badUserID,
				Password:  &password,
				LoginType: &loginType,
			},
		}
		r, ok := u.Login(params).(*authorization.LoginBadRequest)
		assert.True(ok)
		assert.Equal(authorization.NewLoginBadRequest().WithPayload(&models.Error{
			Code:    int64(mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD),
			Details: userNotFound,
		}), r)
		assert.NoError(dm.Close())
	}
	{ // login missing password
		scripts := []mock.Script{
			{ // missing password
				Name: mock.FuncSignIn,
				Input: mock.Input{
					Request: mcom.SignInRequest{
						Account:  userID,
						Password: empty,
					},
				},
				Output: mock.Output{
					Error: &mcomErrors.Error{
						Code:    mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
						Details: missingPassword,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		password := ""
		params := authorization.LoginParams{
			HTTPRequest: httpRequest,
			Body: &models.LoginRequest{
				ID:        &userID,
				Password:  &password,
				LoginType: &loginType,
			},
		}

		r, ok := u.Login(params).(*authorization.LoginBadRequest)
		assert.True(ok)
		assert.Equal(authorization.NewLoginBadRequest().WithPayload(&models.Error{
			Code:    int64(mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD),
			Details: missingPassword,
		}), r)
		assert.NoError(dm.Close())
	}
	{
		scripts := []mock.Script{
			{ // missing user
				Name: mock.FuncSignIn,
				Input: mock.Input{
					Request: mcom.SignInRequest{
						Account:  empty,
						Password: password,
					},
				},
				Output: mock.Output{
					Error: &mcomErrors.Error{
						Code:    mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
						Details: userNotFound,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)

		// missing user
		userID, password = "", "p4s5w0rd"
		params := authorization.LoginParams{
			HTTPRequest: httpRequest,
			Body: &models.LoginRequest{
				ID:        &userID,
				Password:  &password,
				LoginType: &loginType,
			},
		}
		r, ok := u.Login(params).(*authorization.LoginBadRequest)
		assert.True(ok)
		assert.Equal(authorization.NewLoginBadRequest().WithPayload(&models.Error{
			Code:    int64(mcomErrors.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD),
			Details: userNotFound,
		}), r)
	}
}

var (
	oldPassword = "YOUAREGREAT"
	newPassword = "YOUDIDIT"
)

func TestAuthorization_ChangePassword(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("PUT", "/user/change-password", nil)
	httpRequestWithHeader.Header.Set(AuthorizationKey, "token-for-tester")

	scripts := []mock.Script{
		{
			Name: mock.FuncUpdateAccount,
			Input: mock.Input{
				Request: mcom.UpdateAccountRequest{
					UserID: principal.ID,
					ChangePassword: &struct {
						NewPassword string
						OldPassword string
					}{
						NewPassword: newPassword,
						OldPassword: oldPassword,
					},
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncUpdateAccount,
			Input: mock.Input{
				Request: mcom.UpdateAccountRequest{
					UserID: principal.ID,
					ChangePassword: &struct {
						NewPassword string
						OldPassword string
					}{
						NewPassword: newPassword,
						OldPassword: oldPassword,
					},
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_ACCOUNT_BAD_OLD_PASSWORD,
					Details: "bad password",
				},
			},
		},
		{
			Name: mock.FuncUpdateAccount,
			Input: mock.Input{
				Request: mcom.UpdateAccountRequest{
					UserID: principal.ID,
					ChangePassword: &struct {
						NewPassword string
						OldPassword string
					}{
						NewPassword: newPassword,
						OldPassword: oldPassword,
					},
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	}
	dm, err := mock.New(scripts)
	assert.NoError(err)

	type args struct {
		params    authorization.ChangePasswordParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "change password success",
			args: args{
				params: authorization.ChangePasswordParams{
					HTTPRequest: httpRequestWithHeader,
					Body: authorization.ChangePasswordBody{
						CurrentPassword: &oldPassword,
						NewPassword:     &newPassword,
					},
				},
				principal: principal,
			},
			want: authorization.NewChangePasswordOK(),
		},
		{
			name: "wrong old password",
			args: args{
				params: authorization.ChangePasswordParams{
					HTTPRequest: httpRequestWithHeader,
					Body: authorization.ChangePasswordBody{
						CurrentPassword: &oldPassword,
						NewPassword:     &newPassword,
					},
				},
				principal: principal,
			},
			want: authorization.NewChangePasswordDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_ACCOUNT_BAD_OLD_PASSWORD),
				Details: "bad password",
			}),
		},
		{
			name: "internal error",
			args: args{
				params: authorization.ChangePasswordParams{
					HTTPRequest: httpRequestWithHeader,
					Body: authorization.ChangePasswordBody{
						CurrentPassword: &oldPassword,
						NewPassword:     &newPassword,
					},
				},
				principal: principal,
			},
			want: authorization.NewChangePasswordDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			}, 0)
			if got := u.ChangePassword(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChangePassword() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())

	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		r := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)
		rep, ok := r.ChangePassword(authorization.ChangePasswordParams{
			HTTPRequest: httpRequestWithHeader,
			Body: authorization.ChangePasswordBody{
				CurrentPassword: &oldPassword,
				NewPassword:     &newPassword,
			},
		}, principal).(*authorization.ChangePasswordDefault)
		assert.True(ok)
		assert.Equal(authorization.NewChangePasswordDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestAuthorization_ListAuthorizedAccount(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListUserRoles,
			Input: mock.Input{
				Request: mcom.ListUserRolesRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Response: mcom.ListUserRolesReply{
					Users: []mcom.UserRoles{
						{
							ID:    testUsernameDan,
							Roles: testDanRoles,
						},
						{
							ID:    testUsernameSpencer,
							Roles: testSpencerRoles,
						},
					},
				},
			},
		},
		{
			Name: mock.FuncListUserRoles,
			Input: mock.Input{
				Request: mcom.ListUserRolesRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_DEPARTMENT_NOT_FOUND,
				},
			},
		},
		{
			Name: mock.FuncListUserRoles,
			Input: mock.Input{
				Request: mcom.ListUserRolesRequest{
					DepartmentID: testDepartmentOID,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/account/authorized/department-oid/{departmentOID}", nil)
	httpRequestWithHeader.Header.Set(AuthorizationKey, "token-for-tester")

	type args struct {
		params    authorization.ListAuthorizedAccountParams
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
				params: authorization.ListAuthorizedAccountParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: authorization.NewListAuthorizedAccountOK().WithPayload(&authorization.ListAuthorizedAccountOKBody{Data: []*models.AccountDataItems0{
				{
					EmployeeID: &testUsernameDan,
					Roles:      handlerUtils.ToModelsRoles(testDanRoles),
				},
				{
					EmployeeID: &testUsernameSpencer,
					Roles:      handlerUtils.ToModelsRoles(testSpencerRoles),
				},
			}}),
		},
		{
			name: "not found department oid",
			args: args{
				params: authorization.ListAuthorizedAccountParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: authorization.NewListAuthorizedAccountDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_DEPARTMENT_NOT_FOUND),
			}),
		},
		{
			name: "internal error",
			args: args{
				params: authorization.ListAuthorizedAccountParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: authorization.NewListAuthorizedAccountDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			}, 0)
			if got := a.ListAuthorizedAccount(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListAuthorizedAccount() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)
		rep, ok := a.ListAuthorizedAccount(authorization.ListAuthorizedAccountParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentOID: testDepartmentOID,
		}, principal).(*authorization.ListAuthorizedAccountDefault)
		assert.True(ok)
		assert.Equal(authorization.NewListAuthorizedAccountDefault(http.StatusForbidden), rep)
	}
}

func TestAuthorization_ListUnauthorizedAccount(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("GET", "/account/unauthorized/department-oid/{departmentOID}", nil)
	httpRequestWithHeader.Header.Set(AuthorizationKey, "token-for-tester")

	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncListUnauthorizedUsers,
			Input: mock.Input{
				Request: mcom.ListUnauthorizedUsersRequest{
					DepartmentID: testDepartmentOID,
				},
				Options: []interface{}{
					mcom.ExcludeUsers([]string{principal.ID}),
				},
			},
			Output: mock.Output{
				Response: mcom.ListUnauthorizedUsersReply([]string{testUsernameDan, testUsernameSpencer}),
			},
		},
		{
			Name: mock.FuncListUnauthorizedUsers,
			Input: mock.Input{
				Request: mcom.ListUnauthorizedUsersRequest{
					DepartmentID: testDepartmentOID,
				},
				Options: []interface{}{
					mcom.ExcludeUsers([]string{principal.ID}),
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_DEPARTMENT_NOT_FOUND,
				},
			},
		},
		{
			Name: mock.FuncListUnauthorizedUsers,
			Input: mock.Input{
				Request: mcom.ListUnauthorizedUsersRequest{
					DepartmentID: testDepartmentOID,
				},
				Options: []interface{}{
					mcom.ExcludeUsers([]string{principal.ID}),
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	type args struct {
		params    authorization.ListUnauthorizedAccountParams
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
				params: authorization.ListUnauthorizedAccountParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: authorization.NewListUnauthorizedAccountOK().WithPayload(&authorization.ListUnauthorizedAccountOKBody{
				Data: []*authorization.ListUnauthorizedAccountOKBodyDataItems0{
					{
						EmployeeID: testUsernameDan,
					},
					{
						EmployeeID: testUsernameSpencer,
					},
				}}),
		},
		{
			name: "not found department oid",
			args: args{
				params: authorization.ListUnauthorizedAccountParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: authorization.NewListUnauthorizedAccountDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_DEPARTMENT_NOT_FOUND),
			}),
		},
		{
			name: "internal error",
			args: args{
				params: authorization.ListUnauthorizedAccountParams{
					HTTPRequest:   httpRequestWithHeader,
					DepartmentOID: testDepartmentOID,
				},
				principal: principal,
			},
			want: authorization.NewListUnauthorizedAccountDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			}, 0)
			if got := a.ListUnauthorizedAccount(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListUnauthorizedAccount() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)
		rep, ok := a.ListUnauthorizedAccount(authorization.ListUnauthorizedAccountParams{
			HTTPRequest:   httpRequestWithHeader,
			DepartmentOID: testDepartmentOID,
		}, principal).(*authorization.ListUnauthorizedAccountDefault)
		assert.True(ok)
		assert.Equal(authorization.NewListUnauthorizedAccountDefault(http.StatusForbidden), rep)
	}
}

func TestAuthorization_GetRoleList(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name:  mock.FuncListRoles,
			Input: mock.Input{},
			Output: mock.Output{
				Response: mcom.ListRolesReply{
					Roles: []mcom.Role{
						{
							Name:  mcomRoles.Role_PLANNER.String(),
							Value: mcomRoles.Role_PLANNER,
						},
						{
							Name:  mcomRoles.Role_SCHEDULER.String(),
							Value: mcomRoles.Role_SCHEDULER,
						},
						{
							Name:  mcomRoles.Role_INSPECTOR.String(),
							Value: mcomRoles.Role_INSPECTOR,
						},
						{
							Name:  mcomRoles.Role_QUALITY_CONTROLLER.String(),
							Value: mcomRoles.Role_QUALITY_CONTROLLER,
						},
						{
							Name:  mcomRoles.Role_OPERATOR.String(),
							Value: mcomRoles.Role_OPERATOR,
						},
						{
							Name:  mcomRoles.Role_BEARER.String(),
							Value: mcomRoles.Role_BEARER,
						},
					},
				},
			},
		},
		{
			Name:  mock.FuncListRoles,
			Input: mock.Input{},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code:    mcomErrors.Code_NONE,
					Details: "fake error message",
				},
			},
		},
		{
			Name:  mock.FuncListRoles,
			Input: mock.Input{},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("GET", "/account/role-list", nil)
	httpRequestWithHeader.Header.Set(AuthorizationKey, "token-for-tester")

	type args struct {
		params    authorization.GetRoleListParams
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
				params: authorization.GetRoleListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: authorization.NewGetRoleListOK().WithPayload(&authorization.GetRoleListOKBody{
				Data: []*authorization.GetRoleListOKBodyDataItems0{
					{
						ID:   models.Role(mcomRoles.Role_PLANNER),
						Name: mcomRoles.Role_PLANNER.String(),
					},
					{
						ID:   models.Role(mcomRoles.Role_SCHEDULER),
						Name: mcomRoles.Role_SCHEDULER.String(),
					},
					{
						ID:   models.Role(mcomRoles.Role_INSPECTOR),
						Name: mcomRoles.Role_INSPECTOR.String(),
					},
					{
						ID:   models.Role(mcomRoles.Role_QUALITY_CONTROLLER),
						Name: mcomRoles.Role_QUALITY_CONTROLLER.String(),
					},
					{
						ID:   models.Role(mcomRoles.Role_OPERATOR),
						Name: mcomRoles.Role_OPERATOR.String(),
					},
					{
						ID:   models.Role(mcomRoles.Role_BEARER),
						Name: mcomRoles.Role_BEARER.String(),
					},
				},
			}),
		},
		{
			name: "bad request",
			args: args{
				params: authorization.GetRoleListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: authorization.NewGetRoleListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_NONE),
				Details: "fake error message",
			}),
		},
		{
			name: "bad request",
			args: args{
				params: authorization.GetRoleListParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: authorization.NewGetRoleListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			}, 0)
			if got := a.GetRoleList(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRoleList() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)
		rep, ok := a.GetRoleList(authorization.GetRoleListParams{
			HTTPRequest: httpRequestWithHeader,
		}, principal).(*authorization.GetRoleListDefault)
		assert.True(ok)
		assert.Equal(authorization.NewGetRoleListDefault(http.StatusForbidden), rep)
	}
}

func TestAuthorization_Create(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncCreateAccounts,
			Input: mock.Input{
				Request: mcom.CreateAccountsRequest{
					mcom.CreateAccountRequest{
						ID:    testUsernameDan,
						Roles: testDanRoles,
					}.WithDefaultPassword(),
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncCreateAccounts,
			Input: mock.Input{
				Request: mcom.CreateAccountsRequest{
					mcom.CreateAccountRequest{
						ID:    "",
						Roles: testSpencerRoles,
					}.WithDefaultPassword(),
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_INSUFFICIENT_REQUEST,
				},
			},
		},
		{
			Name: mock.FuncCreateAccounts,
			Input: mock.Input{
				Request: mcom.CreateAccountsRequest{
					mcom.CreateAccountRequest{
						ID:    testUsernameSpencer,
						Roles: testSpencerRoles,
					}.WithDefaultPassword(),
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("POST", "/account/authorization", nil)
	httpRequestWithHeader.Header.Set(AuthorizationKey, "token-for-tester")

	type args struct {
		params    authorization.CreateAccountAuthorizationParams
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
				params: authorization.CreateAccountAuthorizationParams{
					HTTPRequest: httpRequestWithHeader,
					Body: authorization.CreateAccountAuthorizationBody{
						EmployeeID: &testUsernameDan,
						Roles:      handlerUtils.ToModelsRoles(testDanRoles),
					},
				},
				principal: principal,
			},
			want: authorization.NewCreateAccountAuthorizationOK(),
		},
		{
			name: "insufficient request",
			args: args{
				params: authorization.CreateAccountAuthorizationParams{
					HTTPRequest: httpRequestWithHeader,
					Body: authorization.CreateAccountAuthorizationBody{
						EmployeeID: &testEmpty,
						Roles:      handlerUtils.ToModelsRoles(testSpencerRoles),
					},
				},
				principal: principal,
			},
			want: authorization.NewCreateAccountAuthorizationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
		},
		{
			name: "internal error",
			args: args{
				params: authorization.CreateAccountAuthorizationParams{
					HTTPRequest: httpRequestWithHeader,
					Body: authorization.CreateAccountAuthorizationBody{
						EmployeeID: &testUsernameSpencer,
						Roles:      handlerUtils.ToModelsRoles(testSpencerRoles),
					},
				},
				principal: principal,
			},
			want: authorization.NewCreateAccountAuthorizationDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			}, 0)
			if got := a.CreateAccountAuthorization(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)
		rep, ok := a.CreateAccountAuthorization(authorization.CreateAccountAuthorizationParams{
			HTTPRequest: httpRequestWithHeader,
			Body: authorization.CreateAccountAuthorizationBody{
				EmployeeID: &testUsernameSpencer,
				Roles:      handlerUtils.ToModelsRoles(testSpencerRoles),
			},
		}, principal).(*authorization.CreateAccountAuthorizationDefault)
		assert.True(ok)
		assert.Equal(authorization.NewCreateAccountAuthorizationDefault(http.StatusForbidden), rep)
	}
}

func TestAuthorization_UpdateAccount(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncUpdateAccount,
			Input: mock.Input{
				Request: mcom.UpdateAccountRequest{
					UserID: testUsernameDan,
					Roles:  testDanRoles,
				},
			},
			Output: mock.Output{},
		},
		{ // with reset password
			Name: mock.FuncUpdateAccount,
			Input: mock.Input{
				Request: mcom.UpdateAccountRequest{
					UserID: testUsernameDan,
					Roles:  testDanRoles,
				},
				Options: []interface{}{mcom.ResetPassword()},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncUpdateAccount,
			Input: mock.Input{
				Request: mcom.UpdateAccountRequest{
					UserID: "",
					Roles:  testSpencerRoles,
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_INSUFFICIENT_REQUEST,
				},
			},
		},
		{
			Name: mock.FuncUpdateAccount,
			Input: mock.Input{
				Request: mcom.UpdateAccountRequest{
					UserID: testUsernameSpencer,
					Roles:  testSpencerRoles,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("PUT", "/account/authorization/{employeeID}", nil)
	httpRequestWithHeader.Header.Set(AuthorizationKey, "token-for-tester")

	type args struct {
		params    authorization.UpdateAccountAuthorizationParams
		principal *models.Principal
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "update roles successfully",
			args: args{
				params: authorization.UpdateAccountAuthorizationParams{
					HTTPRequest: httpRequestWithHeader,
					EmployeeID:  testUsernameDan,
					Body: authorization.UpdateAccountAuthorizationBody{
						Roles:         handlerUtils.ToModelsRoles(testDanRoles),
						ResetPassword: &falseReset,
					},
				},
				principal: principal,
			},
			want: authorization.NewUpdateAccountAuthorizationOK(),
		},
		{
			name: "reset password successfully",
			args: args{
				params: authorization.UpdateAccountAuthorizationParams{
					HTTPRequest: httpRequestWithHeader,
					EmployeeID:  testUsernameDan,
					Body: authorization.UpdateAccountAuthorizationBody{
						Roles:         handlerUtils.ToModelsRoles(testDanRoles),
						ResetPassword: &trueReset,
					},
				},
				principal: principal,
			},
			want: authorization.NewUpdateAccountAuthorizationOK(),
		},
		{
			name: "insufficient request",
			args: args{
				params: authorization.UpdateAccountAuthorizationParams{
					HTTPRequest: httpRequestWithHeader,
					EmployeeID:  testEmpty,
					Body: authorization.UpdateAccountAuthorizationBody{
						Roles:         handlerUtils.ToModelsRoles(testSpencerRoles),
						ResetPassword: &falseReset,
					},
				},
				principal: principal,
			},
			want: authorization.NewUpdateAccountAuthorizationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
		},
		{
			name: "internal error",
			args: args{
				params: authorization.UpdateAccountAuthorizationParams{
					HTTPRequest: httpRequestWithHeader,
					EmployeeID:  testUsernameSpencer,
					Body: authorization.UpdateAccountAuthorizationBody{
						Roles:         handlerUtils.ToModelsRoles(testSpencerRoles),
						ResetPassword: &falseReset,
					},
				},
				principal: principal,
			},
			want: authorization.NewUpdateAccountAuthorizationDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			}, 0)
			if got := a.UpdateAccountAuthorization(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)
		rep, ok := a.UpdateAccountAuthorization(authorization.UpdateAccountAuthorizationParams{
			HTTPRequest: httpRequestWithHeader,
			EmployeeID:  testUsernameSpencer,
			Body: authorization.UpdateAccountAuthorizationBody{
				Roles:         handlerUtils.ToModelsRoles(testSpencerRoles),
				ResetPassword: &falseReset,
			},
		}, principal).(*authorization.UpdateAccountAuthorizationDefault)
		assert.True(ok)
		assert.Equal(authorization.NewUpdateAccountAuthorizationDefault(http.StatusForbidden), rep)
	}
}

func TestAuthorization_DeleteAccount(t *testing.T) {
	assert := assert.New(t)
	dm, err := mock.New([]mock.Script{
		{
			Name: mock.FuncDeleteAccount,
			Input: mock.Input{
				Request: mcom.DeleteAccountRequest{
					ID: testUsernameDan,
				},
			},
			Output: mock.Output{},
		},
		{
			Name: mock.FuncDeleteAccount,
			Input: mock.Input{
				Request: mcom.DeleteAccountRequest{
					ID: "",
				},
			},
			Output: mock.Output{
				Error: mcomErrors.Error{
					Code: mcomErrors.Code_INSUFFICIENT_REQUEST,
				},
			},
		},
		{
			Name: mock.FuncDeleteAccount,
			Input: mock.Input{
				Request: mcom.DeleteAccountRequest{
					ID: testUsernameSpencer,
				},
			},
			Output: mock.Output{
				Error: errors.New(testInternalServerError),
			},
		},
	})
	assert.NoError(err)

	httpRequestWithHeader := httptest.NewRequest("DELETE", "/account/authorization/{employeeID}", nil)
	httpRequestWithHeader.Header.Set(AuthorizationKey, "token-for-tester")

	type args struct {
		params    authorization.DeleteAccountParams
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
				params: authorization.DeleteAccountParams{
					HTTPRequest: httpRequestWithHeader,
					EmployeeID:  testUsernameDan,
				},
				principal: principal,
			},
			want: authorization.NewDeleteAccountOK(),
		},
		{
			name: "insufficient request",
			args: args{
				params: authorization.DeleteAccountParams{
					HTTPRequest: httpRequestWithHeader,
					EmployeeID:  testEmpty,
				},
				principal: principal,
			},
			want: authorization.NewDeleteAccountDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}),
		},
		{
			name: "internal error",
			args: args{
				params: authorization.DeleteAccountParams{
					HTTPRequest: httpRequestWithHeader,
					EmployeeID:  testUsernameSpencer,
				},
				principal: principal,
			},
			want: authorization.NewDeleteAccountDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			}, 0)
			if got := a.DeleteAccount(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteAccount() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.NoError(dm.Close())
	{ // forbidden access
		a := NewAuthorization(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		}, 0)
		rep, ok := a.DeleteAccount(authorization.DeleteAccountParams{
			HTTPRequest: httpRequestWithHeader,
			EmployeeID:  testUsernameSpencer,
		}, principal).(*authorization.DeleteAccountDefault)
		assert.True(ok)
		assert.Equal(authorization.NewDeleteAccountDefault(http.StatusForbidden), rep)
	}
}
