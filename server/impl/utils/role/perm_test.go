package role

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
)

var parsedPermission = map[string][]string{
	"GET_SERVER_STATUS": {
		"ADMINISTRATOR",
		"LEADER",
		"PLANNER",
		"SCHEDULER",
		"INSPECTOR",
		"QUALITY_CONTROLLER",
		"OPERATOR",
		"BEARER",
	},
	"GET_BARCODE_INFO": {
		"ADMINISTRATOR",
		"LEADER",
		"INSPECTOR",
		"QUALITY_CONTROLLER",
	},
	"UPDATE_BARCODE": {
		"ADMINISTRATOR",
		"LEADER",
		"INSPECTOR",
		"QUALITY_CONTROLLER",
	},
	"GET_UPDATE_BARCODE_STATUS_LIST": {
		"ADMINISTRATOR",
		"LEADER",
		"INSPECTOR",
		"QUALITY_CONTROLLER",
	},
	"GET_EXTEND_DAYS": {
		"ADMINISTRATOR",
		"LEADER",
		"INSPECTOR",
		"QUALITY_CONTROLLER",
	},
	"GET_CONTROL_AREA_LIST": {
		"ADMINISTRATOR",
		"LEADER",
		"INSPECTOR",
		"QUALITY_CONTROLLER",
	},
	"GET_HOLD_REASON_LIST": {
		"ADMINISTRATOR",
		"LEADER",
		"INSPECTOR",
		"QUALITY_CONTROLLER",
	},
}

func TestHasPermission(t *testing.T) {
	assert.NoError(t, InitPermission(parsedPermission))
	defer ClearPermission()
	type args struct {
		id    kenda.FunctionOperationID
		roles []models.Role
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "permitted",
			args: args{
				id: kenda.FunctionOperationID_GET_SERVER_STATUS,
				roles: []models.Role{
					models.Role(mcomRoles.Role_ADMINISTRATOR),
				},
			},
			want: true,
		},
		{
			name: "not permitted",
			args: args{
				id: kenda.FunctionOperationID_GET_CONTROL_AREA_LIST,
				roles: []models.Role{
					models.Role(mcomRoles.Role_BEARER),
					models.Role(mcomRoles.Role_PLANNER),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPermission(tt.args.id, tt.args.roles)
			if got != tt.want {
				t.Errorf("HasPermission() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitPermission(t *testing.T) {
	assert := assert.New(t)
	{ // success
		data := map[string][]string{
			"GET_SERVER_STATUS": {
				"ADMINISTRATOR",
				"LEADER",
			},
			"GET_BARCODE_INFO": {
				"ADMINISTRATOR",
				"INSPECTOR",
			},
		}
		assert.NoError(InitPermission(data))
		assert.Equal(funcRoleList{
			kenda.FunctionOperationID_GET_SERVER_STATUS: {
				mcomRoles.Role_ADMINISTRATOR: struct{}{},
				mcomRoles.Role_LEADER:        struct{}{},
			},
			kenda.FunctionOperationID_GET_BARCODE_INFO: {
				mcomRoles.Role_ADMINISTRATOR: struct{}{},
				mcomRoles.Role_INSPECTOR:     struct{}{},
			},
		}, permissionList)
		permissionList = nil
	}
	{ // repeated roles, it will be counted as one
		data := map[string][]string{
			"GET_SERVER_STATUS": {
				"ADMINISTRATOR",
				"ADMINISTRATOR",
			},
		}
		assert.NoError(InitPermission(data))
		assert.Equal(funcRoleList{
			kenda.FunctionOperationID_GET_SERVER_STATUS: {
				mcomRoles.Role_ADMINISTRATOR: struct{}{},
			},
		}, permissionList)
		permissionList = nil
	}
	{ // function not in the list
		data := map[string][]string{
			"GET_SERVER_STATUS": {},
			"GET_BARCODE_INFO": {
				"ADMINISTRATOR",
			},
			"GET_TAX_DATA": {
				"ADMINISTRATOR",
			},
		}
		assert.EqualError(InitPermission(data), "function GET_TAX_DATA was not in the list")
		permissionList = nil
	}
	{ // not existed role
		data := map[string][]string{
			"GET_SERVER_STATUS": {
				"ACTOR",
			},
			"GET_BARCODE_INFO": {
				"ADMINISTRATOR",
			},
		}
		assert.EqualError(InitPermission(data), "not existed role: ACTOR")
		permissionList = nil
	}
	{ // registered permission
		data := map[string][]string{
			"GET_SERVER_STATUS": {
				"ADMINISTRATOR",
				"LEADER",
			},
			"GET_BARCODE_INFO": {
				"ADMINISTRATOR",
				"INSPECTOR",
			},
		}
		assert.NoError(InitPermission(data))
		assert.EqualError(InitPermission(data), "role permissions have been initialized")
		permissionList = nil
	}
}

func TestClearPermission(t *testing.T) {
	ClearPermission()
	assert.Equal(t, funcRoleList(nil), permissionList)
}

func Test_rolesToMap(t *testing.T) {
	type args struct {
		roles []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[mcomRoles.Role]struct{}
		wantErr bool
	}{
		{
			name: "roles map success",
			args: args{
				roles: []string{
					"ADMINISTRATOR",
					"LEADER",
				},
			},
			want: map[mcomRoles.Role]struct{}{
				mcomRoles.Role_ADMINISTRATOR: struct{}{},
				mcomRoles.Role_LEADER:        struct{}{},
			},
			wantErr: false,
		},
		{
			name: "role not existed",
			args: args{
				roles: []string{
					"ADMINISTRATOR",
					"LEADER",
					"ACTOR",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rolesToMap(tt.args.roles)
			if (err != nil) != tt.wantErr {
				t.Errorf("rolesToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rolesToMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}
