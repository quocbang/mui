package ui

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"

	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/ui"
)

// UI definitions
type UI struct {
	dm            mcom.DataManager
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewUI returns Production service.
func NewUI(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.UI {
	return UI{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// GetProductGroupList implementation.
func (u UI) GetProductGroupList(params ui.GetProductGroupListParams, principal *models.Principal) middleware.Responder {
	if !u.hasPermission(kenda.FunctionOperationID_GET_PRODUCT_GROUP_LIST, principal.Roles) {
		return ui.NewGetProductGroupListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	reply, err := u.dm.ListProductGroups(ctx, mcom.ListProductGroupsRequest{
		DepartmentID: params.DepartmentOID,
		Type:         params.ProductType,
	})
	if err != nil {
		return utils.ParseError(ctx, ui.NewGetProductGroupListDefault(0), err)
	}
	data := make([]*ui.GetProductGroupListOKBodyDataItems0, len(reply.Products))
	for i, group := range reply.Products {
		data[i] = &ui.GetProductGroupListOKBodyDataItems0{
			Children: group.Children,
			Parent:   group.ID,
		}
	}
	return ui.NewGetProductGroupListOK().WithPayload(&ui.GetProductGroupListOKBody{Data: data})
}

// SetStationConfig implements.
func (u UI) SetStationConfig(params ui.SetStationConfigParams, principal *models.Principal) middleware.Responder {
	if !u.hasPermission(kenda.FunctionOperationID_SET_STATION_CONFIG, principal.Roles) {
		return ui.NewSetStationConfigDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	quantitySource := stations.CollectQuantitySource_FROM_STATION_CONFIGS
	defaultCollectQuantity := decimal.Zero

	// collect quantity
	switch params.Body.StationConfig.Collect.Quantity.Type {
	case int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS):
		quantity, err := decimal.NewFromString(params.Body.StationConfig.Collect.Quantity.Value)
		if err != nil {
			return utils.ParseError(ctx, ui.NewSetStationConfigDefault(0), err)
		}
		defaultCollectQuantity = quantity
	case int64(stations.CollectQuantitySource_FROM_STATION_PARAMS):
		quantitySource = stations.CollectQuantitySource_FROM_STATION_PARAMS
	default:
		return ui.NewSetStationConfigDefault(http.StatusBadRequest).WithPayload(
			&models.Error{
				Code:    int64(mcomErrors.Code_INVALID_NUMBER),
				Details: "collect quantity type is invalid number.",
			})
	}
	setStationConfigRequest := mcom.SetStationConfigurationRequest{}
	setStationConfigRequest.StationID = params.StationID
	setStationConfigRequest.SplitFeedAndCollect = params.Body.StationConfig.SeparateMode
	setStationConfigRequest.Feed = mcom.StationFeedConfigs{
		ProductTypes:         params.Body.StationConfig.Feed.ProductType,
		NeedMaterialResource: *params.Body.StationConfig.Feed.MaterialResource,
		QuantitySource:       stations.FeedQuantitySource_FROM_RECIPE,
	}

	setStationConfigRequest.Collect = mcom.StationCollectConfigs{
		NeedCollectResource: *params.Body.StationConfig.Collect.Resource,
		NeedCarrierResource: *params.Body.StationConfig.Collect.CarrierResource,
		QuantitySource:      quantitySource,
		DefaultQuantity:     defaultCollectQuantity,
	}

	// separate feed & collect
	if setStationConfigRequest.SplitFeedAndCollect {
		setStationConfigRequest.Feed.QuantitySource = stations.FeedQuantitySource(params.Body.StationConfig.Feed.StandardQuantity)

		// feed & collect operator site
		if len(params.Body.StationConfig.Feed.OperatorSites) == 0 || len(params.Body.StationConfig.Collect.OperatorSites) == 0 {
			return ui.NewSetStationConfigDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Code:    int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
					Details: "operator site is empty.",
				})
		} else {
			setStationConfigRequest.Feed.OperatorSites = parseMcomSite(params.Body.StationConfig.Feed.OperatorSites)
			setStationConfigRequest.Collect.OperatorSites = parseMcomSite(params.Body.StationConfig.Collect.OperatorSites)
		}
	}
	if err := u.dm.SetStationConfiguration(ctx, setStationConfigRequest); err != nil {
		return utils.ParseError(ctx, ui.NewSetStationConfigDefault(0), err)
	}

	return ui.NewSetStationConfigOK()
}

// GetStationConfig implements.
func (u UI) GetStationConfig(params ui.GetStationConfigParams, principal *models.Principal) middleware.Responder {
	if !u.hasPermission(kenda.FunctionOperationID_GET_STATION_CONFIG, principal.Roles) {
		return ui.NewGetStationConfigDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	config, err := u.dm.GetStationConfiguration(ctx, mcom.GetStationConfigurationRequest{
		StationID: params.StationID,
	})
	if err != nil {
		return utils.ParseError(ctx, ui.NewGetStationConfigDefault(0), err)
	}

	stationConfig := models.StationConfig{
		SeparateMode: false,
		Feed: &models.StationConfigFeed{
			MaterialResource: handlerUtils.NewBoolean(true),
			ProductType:      []string{},
			OperatorSites:    defaultOperatorSite(),
			StandardQuantity: int64(stations.FeedQuantitySource_FROM_RECIPE),
		},
		Collect: &models.StationConfigCollect{
			Resource:        handlerUtils.NewBoolean(true),
			CarrierResource: handlerUtils.NewBoolean(true),
			Quantity: &models.StationConfigCollectQuantity{
				Type: int64(stations.CollectQuantitySource_FROM_STATION_CONFIGS),
			},
			OperatorSites: defaultOperatorSite(),
		},
	}

	quantity := config.Collect.DefaultQuantity.String()
	if hasStationConfig(config) {
		stationConfig.SeparateMode = config.SplitFeedAndCollect
		// Feed
		stationConfig.Feed = &models.StationConfigFeed{
			MaterialResource: &config.Feed.NeedMaterialResource,
			ProductType:      config.Feed.ProductTypes,
			OperatorSites:    defaultOperatorSite(),
			StandardQuantity: int64(stations.FeedQuantitySource_FROM_RECIPE),
		}
		// Collect
		stationConfig.Collect = &models.StationConfigCollect{
			Resource:        &config.Collect.NeedCollectResource,
			CarrierResource: &config.Collect.NeedCarrierResource,
			Quantity:        &models.StationConfigCollectQuantity{},
			OperatorSites:   defaultOperatorSite(),
		}

		// spilt feed & collect
		if config.SplitFeedAndCollect {
			// feed quantity
			stationConfig.Feed.StandardQuantity = int64(config.Feed.QuantitySource)

			// feed & collect operatorSites
			stationConfig.Feed.OperatorSites = parseSwaggerSite(config.Feed.OperatorSites)
			stationConfig.Collect.OperatorSites = parseSwaggerSite(config.Collect.OperatorSites)
		}

		// collect quantity
		if config.Collect.QuantitySource == stations.CollectQuantitySource_FROM_STATION_PARAMS {
			stationConfig.Collect.Quantity.Type = int64(stations.CollectQuantitySource_FROM_STATION_PARAMS)
		}
	}

	stationConfig.Collect.Quantity.Value = quantity

	return ui.NewGetStationConfigOK().WithPayload(&ui.GetStationConfigOKBody{
		Data: &ui.GetStationConfigOKBodyData{
			StationConfig: &stationConfig,
		},
	})
}

func parseMcomSite(dataIn []*models.SiteInfo) []mcomModels.UniqueSite {
	dataOut := make([]mcomModels.UniqueSite, len(dataIn))
	for i, site := range dataIn {
		dataOut[i] = mcomModels.UniqueSite{
			Station: site.StationID,
			SiteID: mcomModels.SiteID{
				Name:  site.SiteName,
				Index: 0,
			},
		}
	}
	return dataOut
}

func parseSwaggerSite(dataIn []mcomModels.UniqueSite) []*models.SiteInfo {
	dataOut := make([]*models.SiteInfo, len(dataIn))
	for i, site := range dataIn {
		dataOut[i] = &models.SiteInfo{
			StationID: site.Station,
			SiteName:  site.SiteID.Name,
			SiteIndex: int64(site.SiteID.Index),
		}
	}
	return dataOut
}

func hasStationConfig(config mcom.GetStationConfigurationReply) bool {
	zero := mcom.GetStationConfigurationReply{}
	return !cmp.Equal(zero, config)
}

func defaultOperatorSite() []*models.SiteInfo {
	return []*models.SiteInfo{
		{
			StationID: "",
			SiteName:  "",
			SiteIndex: 0,
		},
	}
}
