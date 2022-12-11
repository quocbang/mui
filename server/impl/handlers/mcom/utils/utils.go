package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/shopspring/decimal"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	mcomWorkOrder "gitlab.kenda.com.tw/kenda/mcom/utils/workorder"
	mesModels "gitlab.kenda.com.tw/kenda/mui/server/mes"

	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations"
)

type UtilsString string

type MesHeader struct {
	UserID, Station, Site, TrackID string
}
type batchDetails struct {
	PerBatchQuantity []decimal.Decimal
	BatchCount       int64
	PlanQuantity     decimal.Decimal
}

// GetServerStatus to make sure the server is running.
func GetServerStatus(_ operations.CheckServerStatusParams) middleware.Responder {
	return operations.NewCheckServerStatusOK()
}

// ToModelsRoles format multiple mcom Role Type to models' Role.
func ToModelsRoles(r []mcomRoles.Role) []models.Role {
	roles := make([]models.Role, len(r))
	for i, v := range r {
		roles[i] = models.Role(v)
	}
	return roles
}

// FromModelsRoles format multiple models' Role Type to mcom Role.
func FromModelsRoles(r []models.Role) []mcomRoles.Role {
	roles := make([]mcomRoles.Role, len(r))
	for i, v := range r {
		roles[i] = mcomRoles.Role(v)
	}
	return roles
}

// ToDepartmentsModel converts list of mcom.Department to models.Departments
func ToDepartmentsModel(d []mcom.Department) models.Departments {
	departments := make([]*models.Department, len(d))
	for i, dep := range d {
		departments[i] = &models.Department{
			OID: dep.OID,
			ID:  dep.ID,
		}
	}
	return departments
}

// ToSlices convert list of decimal.Decimal into slice of string.
func ToSlices(numbers []decimal.Decimal) []string {
	s := make([]string, len(numbers))
	for i, number := range numbers {
		s[i] = number.String()
	}
	return s
}

// ToDecimals convert slice of string to list of decimal.Decimal.
func ToDecimals(str []string) ([]decimal.Decimal, error) {
	decimals := make([]decimal.Decimal, len(str))
	for i, s := range str {
		d, err := decimal.NewFromString(s)
		if err != nil {
			return nil, err
		}
		decimals[i] = d
	}
	return decimals, nil
}

// check material status & expiry time & in recipe
// return true represent status not available or expired or not in recipe
func SiteMaterialNotOKCheck(target mcomModels.BoundResource, recipe []*mcom.RecipeProcessStep) bool {
	for _, m := range recipe {
		for _, n := range m.Materials {
			if n.Name == target.Material.Material.ID {
				if target.Material.Status != resources.MaterialStatus_AVAILABLE {
					return true
				}
				if target.Material.ExpiryTime.Time().Before(time.Now()) {
					return true
				}
				return false
			}
		}
	}
	return true
}

func SendMesAgePOSTRequest(requestBody interface{}, url string) error {
	requestJson, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestJson))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func SendMesPOSTRequest[T *mesModels.APIResourceFeedReply | *mesModels.APICollectReply](requestBody interface{}, mesHeader MesHeader, url string) (T, error) {
	requestJson, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", mesHeader.UserID)
	req.Header.Set("station", mesHeader.Station)
	req.Header.Set("site", mesHeader.Site)
	req.Header.Set("pid", mesHeader.TrackID)

	httpResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if httpResponse.StatusCode/100 != 2 {
		return nil, errors.New("mes error")
	}

	defer httpResponse.Body.Close()
	dataOut := new(T)
	json.NewDecoder(httpResponse.Body).Decode(dataOut)

	return *dataOut, nil
}

func NewBoolean(data bool) *bool {
	dataOut := data
	return &dataOut
}

func SetContextValue(dataIn *http.Request, key string, value string) *http.Request {
	return dataIn.WithContext(context.WithValue(dataIn.Context(), UtilsString(key), value))
}

func GetContextValue(dataIn *http.Request, key string) string {
	return fmt.Sprint(dataIn.Context().Value(UtilsString(key)))
}

func ParseBatchQuantityDetails(dataIn mcomModels.BatchQuantityDetails) (batchDetails, error) {
	var dataOut batchDetails

	switch dataIn.BatchQuantityType {
	case mcomWorkOrder.BatchSize_PER_BATCH_QUANTITIES:
		dataOut.PerBatchQuantity = dataIn.QuantityForBatches
		dataOut.BatchCount = int64(len(dataIn.QuantityForBatches))
		dataOut.PlanQuantity = sum(dataIn.QuantityForBatches)

	case mcomWorkOrder.BatchSize_FIXED_QUANTITY:
		dataOut.BatchCount = int64(dataIn.FixedQuantity.BatchCount)
		dataOut.PlanQuantity = dataIn.FixedQuantity.PlanQuantity

	case mcomWorkOrder.BatchSize_PLAN_QUANTITY:
		dataOut.BatchCount = int64(dataIn.PlanQuantity.BatchCount)
		dataOut.PlanQuantity = dataIn.PlanQuantity.PlanQuantity

	default:
		return dataOut, fmt.Errorf(fmt.Sprintf("no implementation with %d of BatchSize", dataIn.BatchQuantityType))
	}

	return dataOut, nil
}

func sum(dataIn []decimal.Decimal) decimal.Decimal {
	var dataOut = decimal.Zero
	for _, data := range dataIn {
		dataOut = dataOut.Add(data)
	}
	return dataOut
}
