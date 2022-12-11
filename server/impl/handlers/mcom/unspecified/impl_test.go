package unspecified

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-openapi/runtime/middleware"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/unspecified"
)

const (
	userID = "tester"

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
)

func TestUnspecified_ListDepartmentIDs(t *testing.T) {

	httpRequestWithHeader := httptest.NewRequest("GET", "/departments", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	assert := assert.New(t)

	type args struct {
		params    unspecified.ListDepartmentIDsParams
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
				params: unspecified.ListDepartmentIDsParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: unspecified.NewListDepartmentIDsOK().WithPayload(&unspecified.ListDepartmentIDsOKBody{
				Data: []*unspecified.ListDepartmentIDsOKBodyDataItems0{
					{
						DepartmentID: "M2100",
					},
					{
						DepartmentID: "M2110",
					},
					{
						DepartmentID: "M2120",
					},
				},
			}),
			script: []mock.Script{
				{
					Name:  mock.FuncListAllDepartment,
					Input: mock.Input{},
					Output: mock.Output{
						Response: mcom.ListAllDepartmentReply{
							IDs: []string{"M2100", "M2110", "M2120"},
						},
					},
				},
			},
		},
		{
			name: "internal error",
			args: args{
				params: unspecified.ListDepartmentIDsParams{
					HTTPRequest: httpRequestWithHeader,
				},
				principal: principal,
			},
			want: unspecified.NewListDepartmentIDsDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}),
			script: []mock.Script{
				{
					Name:  mock.FuncListAllDepartment,
					Input: mock.Input{},
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
			s := NewUnspecified(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
				return true
			})
			if got := s.ListDepartmentIDs(tt.args.params, tt.args.principal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListDepartmentIDs() = %v, want %v", got, tt.want)
			}
			assert.NoErrorf(dm.Close(), tt.name)
		})
	}
	{ // forbidden access
		dm, _ := mock.New([]mock.Script{})
		s := NewUnspecified(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := s.ListDepartmentIDs(unspecified.ListDepartmentIDsParams{
			HTTPRequest: httpRequestWithHeader,
		}, principal).(*unspecified.ListDepartmentIDsDefault)
		assert.True(ok)
		assert.Equal(unspecified.NewListDepartmentIDsDefault(http.StatusForbidden), rep)
	}
}
