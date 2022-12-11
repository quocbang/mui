package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"

	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
)

var (
	errorCode = map[string]int64{
		"Error_ERROR_BAD_REQUEST":      int64(mcomErrors.Code_BAD_REQUEST),
		"ERROR_RESOURCE_NOT_FOUND":     int64(mcomErrors.Code_RESOURCE_NOT_FOUND),
		"ERROR_RESOURCE_EXPIRED":       int64(mcomErrors.Code_RESOURCE_EXPIRED),
		"ERROR_RESOURCE_ON_HOLD":       int64(mcomErrors.Code_RESOURCE_UNAVAILABLE),
		"ERROR_WORKORDER_NOT_FOUND":    int64(mcomErrors.Code_WORKORDER_NOT_FOUND),
		"ERROR_WORKORDER_BAD_BATCH":    int64(mcomErrors.Code_WORKORDER_BAD_BATCH),
		"ERROR_BATCH_NOT_FOUND":        int64(mcomErrors.Code_BATCH_NOT_FOUND),
		"ERROR_US_MISMATCH":            int64(mcomErrors.Code_USER_STATION_MISMATCH),
		"ERROR_SR_SITE_NOT_FOUND":      int64(mcomErrors.Code_STATION_SITE_NOT_FOUND),
		"ERROR_SW_MISMATCH":            int64(mcomErrors.Code_STATION_WORKORDER_MISMATCH),
		"ERROR_RW_QUANTITY_BELOW_MIN":  int64(mcomErrors.Code_RESOURCE_WORKORDER_QUANTITY_BELOW_MIN),
		"ERROR_RW_QUANTITY_ABOVE_MAX":  int64(mcomErrors.Code_RESOURCE_WORKORDER_QUANTITY_ABOVE_MAX),
		"ERROR_RW_QUANTITY_SHORTAGE":   int64(mcomErrors.Code_RESOURCE_MATERIAL_SHORTAGE),
		"ERROR_RW_BAD_GRADE":           int64(mcomErrors.Code_RESOURCE_WORKORDER_BAD_GRADE),
		"ERROR_RW_RESOURCE_UNEXPECTED": int64(mcomErrors.Code_RESOURCE_WORKORDER_RESOURCE_UNEXPECTED),
		"ERROR_RW_RESOURCE_MISSING":    int64(mcomErrors.Code_RESOURCE_WORKORDER_RESOURCE_MISSING),
		"ERROR_RESOURCE_EXISTED":       int64(mcomErrors.Code_RESOURCE_EXISTED),
		"ERROR_CARRIER_NOT_FOUND":      int64(mcomErrors.Code_CARRIER_NOT_FOUND),
		"ERROR_CARRIER_IN_USE":         int64(mcomErrors.Code_CARRIER_IN_USE),
		"ERROR_BATCH_NOT_READY":        int64(mcomErrors.Code_BATCH_NOT_READY),
		"ERROR_WORKORDER_BAD_STATUS":   int64(mcomErrors.Code_WORKORDER_BAD_STATUS),
		"ERROR_RECORD_EXISTED":         int64(mcomErrors.Code_RECORD_ALREADY_EXISTS),
		"ERROR_RECORD_NOT_FOUND":       int64(mcomErrors.Code_RECORD_NOT_FOUND),
	}
)

// ParseError parse default error.
func ParseError(ctx context.Context, d defaultError, err error) middleware.Responder {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		d.SetStatusCode(http.StatusRequestTimeout)
		d.SetPayload(&models.Error{
			Details: err.Error(),
		})
		return d
	}
	if e, ok := mcomErrors.As(err); ok {
		d.SetStatusCode(http.StatusBadRequest)
		d.SetPayload(&models.Error{
			Code:    int64(e.Code),
			Details: e.Details,
		})
	} else {
		d.SetStatusCode(http.StatusInternalServerError)
		d.SetPayload(&models.Error{
			Details: err.Error(),
		})
	}

	return d
}

type defaultError interface {
	middleware.Responder

	SetStatusCode(int)
	SetPayload(*models.Error)
}

func ParseMesErrorCode(mesCode string) int64 {
	return errorCode[mesCode]
}
