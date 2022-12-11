package utils

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-openapi/runtime/middleware"
	"github.com/shopspring/decimal"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations"
)

func TestToModelsRoles(t *testing.T) {
	type args struct {
		r []mcomRoles.Role
	}
	tests := []struct {
		name string
		args args
		want []models.Role
	}{
		{
			name: "success",
			args: args{
				r: []mcomRoles.Role{
					mcomRoles.Role_ADMINISTRATOR,
					mcomRoles.Role_LEADER,
				},
			},
			want: []models.Role{
				models.Role(mcomRoles.Role_ADMINISTRATOR),
				models.Role(mcomRoles.Role_LEADER),
			},
		},
		{
			name: "empty role",
			args: args{
				r: []mcomRoles.Role{},
			},
			want: []models.Role{},
		},
		{
			name: "nil role",
			args: args{
				r: nil,
			},
			want: []models.Role{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToModelsRoles(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToModelsRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDepartmentsModel(t *testing.T) {
	type args struct {
		d []mcom.Department
	}
	tests := []struct {
		name string
		args args
		want models.Departments
	}{
		{
			name: "success",
			args: args{
				d: []mcom.Department{
					{
						OID: "M2110xx",
						ID:  "M2110",
					},
				},
			},
			want: []*models.Department{
				{
					ID:  "M2110",
					OID: "M2110xx",
				},
			},
		},
		{
			name: "nil department",
			args: args{
				d: nil,
			},
			want: models.Departments{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToDepartmentsModel(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToDepartmentsModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSlices(t *testing.T) {
	type args struct {
		numbers []decimal.Decimal
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "good case",
			args: args{
				numbers: []decimal.Decimal{
					decimal.NewFromInt(10),
					decimal.NewFromInt(20),
				},
			},
			want: []string{"10", "20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlices(tt.args.numbers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToSlices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDecimals(t *testing.T) {
	type args struct {
		str []string
	}
	tests := []struct {
		name    string
		args    args
		want    []decimal.Decimal
		wantErr bool
	}{
		{
			name: "good case",
			args: args{
				str: []string{"10", "20"},
			},
			want: []decimal.Decimal{
				decimal.NewFromInt(10),
				decimal.NewFromInt(20),
			},
		},
		{
			name: "error case",
			args: args{
				str: []string{"T_T", "20"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToDecimals(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToDecimals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToDecimals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetServerStatus(t *testing.T) {
	httpRequest := httptest.NewRequest("GET", "/server/status", nil)
	type args struct {
		params operations.CheckServerStatusParams
	}
	tests := []struct {
		name string
		args args
		want middleware.Responder
	}{
		{
			name: "good case",
			args: args{
				params: operations.CheckServerStatusParams{
					HTTPRequest: httpRequest,
				},
			},
			want: operations.NewCheckServerStatusOK(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetServerStatus(tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServerStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromModelsRoles(t *testing.T) {
	type args struct {
		r []models.Role
	}
	tests := []struct {
		name string
		args args
		want []mcomRoles.Role
	}{
		{
			name: "success",
			args: args{
				r: []models.Role{
					models.Role(mcomRoles.Role_ADMINISTRATOR),
					models.Role(mcomRoles.Role_LEADER),
				},
			},
			want: []mcomRoles.Role{
				mcomRoles.Role_ADMINISTRATOR,
				mcomRoles.Role_LEADER,
			},
		},
		{
			name: "empty role",
			args: args{
				r: []models.Role{},
			},
			want: []mcomRoles.Role{},
		},
		{
			name: "nil role",
			args: args{
				r: nil,
			},
			want: []mcomRoles.Role{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromModelsRoles(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromModelsRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}
