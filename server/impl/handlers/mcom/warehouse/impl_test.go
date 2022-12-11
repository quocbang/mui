package warehouse

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/warehouse"
)

const (
	userID = "tester"

	testResourceID = "R0147852369"

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

	testWarehouse = "WX"
	testLocation  = "001"
)

func TestWarehouse_GetWarehouseInfo(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("GET", "/warehouse/resource/{ID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	{ // normal case
		scripts := []mock.Script{
			{
				Name: mock.FuncGetResourceWarehouse,
				Input: mock.Input{
					Request: mcom.GetResourceWarehouseRequest{
						ResourceID: testResourceID,
					},
				},
				Output: mock.Output{
					Response: mcom.GetResourceWarehouseReply{
						ID:       testWarehouse,
						Location: testLocation,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		w := NewWarehouse(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := w.GetWarehouseInfo(warehouse.GetWarehouseInfoParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testResourceID,
		}, principal).(*warehouse.GetWarehouseInfoOK)
		if assert.True(ok) {
			assert.Equal(warehouse.NewGetWarehouseInfoOK().WithPayload(&warehouse.GetWarehouseInfoOKBody{
				Data: &warehouse.GetWarehouseInfoOKBodyData{
					Location:    testLocation,
					WarehouseID: testWarehouse,
				},
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // user error case
		scripts := []mock.Script{
			{
				Name: mock.FuncGetResourceWarehouse,
				Input: mock.Input{
					Request: mcom.GetResourceWarehouseRequest{
						ResourceID: testResourceID,
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_RESOURCE_NOT_FOUND,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		w := NewWarehouse(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := w.GetWarehouseInfo(warehouse.GetWarehouseInfoParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testResourceID,
		}, principal).(*warehouse.GetWarehouseInfoDefault)
		if assert.True(ok) {
			assert.Equal(warehouse.NewGetWarehouseInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // internal error
		scripts := []mock.Script{
			{
				Name: mock.FuncGetResourceWarehouse,
				Input: mock.Input{
					Request: mcom.GetResourceWarehouseRequest{
						ResourceID: testResourceID,
					},
				},
				Output: mock.Output{
					Error: errors.New(testInternalServerError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		w := NewWarehouse(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := w.GetWarehouseInfo(warehouse.GetWarehouseInfoParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testResourceID,
		}, principal).(*warehouse.GetWarehouseInfoDefault)
		if assert.True(ok) {
			assert.Equal(warehouse.NewGetWarehouseInfoDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		w := NewWarehouse(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := w.GetWarehouseInfo(warehouse.GetWarehouseInfoParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testResourceID,
		}, principal).(*warehouse.GetWarehouseInfoDefault)
		assert.True(ok)
		assert.Equal(warehouse.NewGetWarehouseInfoDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}

func TestWarehouse_WarehouseTransaction(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("PUT", "/warehouse/resource/{ID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	{ // normal case
		scripts := []mock.Script{
			{
				Name: mock.FuncWarehousingStock,
				Input: mock.Input{
					Request: mcom.WarehousingStockRequest{
						Warehouse: mcom.Warehouse{
							ID:       testWarehouse,
							Location: testLocation,
						},
						ResourceIDs: []string{testResourceID},
					},
				},
				Output: mock.Output{},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		w := NewWarehouse(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := w.WarehouseTransaction(warehouse.WarehouseTransactionParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testResourceID,
			Body: warehouse.WarehouseTransactionBody{
				NewLocation:    &testLocation,
				NewWarehouseID: &testWarehouse,
			},
		}, principal).(*warehouse.WarehouseTransactionOK)
		if assert.True(ok) {
			assert.Equal(warehouse.NewWarehouseTransactionOK(), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // user error case
		scripts := []mock.Script{
			{
				Name: mock.FuncWarehousingStock,
				Input: mock.Input{
					Request: mcom.WarehousingStockRequest{
						Warehouse: mcom.Warehouse{
							ID:       testWarehouse,
							Location: testLocation,
						},
						ResourceIDs: []string{testResourceID},
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_RESOURCE_NOT_FOUND,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		w := NewWarehouse(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := w.WarehouseTransaction(warehouse.WarehouseTransactionParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testResourceID,
			Body: warehouse.WarehouseTransactionBody{
				NewLocation:    &testLocation,
				NewWarehouseID: &testWarehouse,
			},
		}, principal).(*warehouse.WarehouseTransactionDefault)
		if assert.True(ok) {
			assert.Equal(warehouse.NewWarehouseTransactionDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // internal error
		scripts := []mock.Script{
			{
				Name: mock.FuncWarehousingStock,
				Input: mock.Input{
					Request: mcom.WarehousingStockRequest{
						Warehouse: mcom.Warehouse{
							ID:       testWarehouse,
							Location: testLocation,
						},
						ResourceIDs: []string{testResourceID},
					},
				},
				Output: mock.Output{
					Error: errors.New(testInternalServerError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		w := NewWarehouse(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		rep, ok := w.WarehouseTransaction(warehouse.WarehouseTransactionParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testResourceID,
			Body: warehouse.WarehouseTransactionBody{
				NewLocation:    &testLocation,
				NewWarehouseID: &testWarehouse,
			},
		}, principal).(*warehouse.WarehouseTransactionDefault)
		if assert.True(ok) {
			assert.Equal(warehouse.NewWarehouseTransactionDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), rep)
		}
		assert.NoError(dm.Close())
	}
	{ // forbidden access
		dm, err := mock.New(nil)
		assert.NoError(err)
		w := NewWarehouse(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := w.WarehouseTransaction(warehouse.WarehouseTransactionParams{
			HTTPRequest: httpRequestWithHeader,
			ID:          testResourceID,
			Body: warehouse.WarehouseTransactionBody{
				NewLocation:    &testLocation,
				NewWarehouseID: &testWarehouse,
			},
		}, principal).(*warehouse.WarehouseTransactionDefault)
		assert.True(ok)
		assert.Equal(warehouse.NewWarehouseTransactionDefault(http.StatusForbidden), rep)
		assert.NoError(dm.Close())
	}
}
