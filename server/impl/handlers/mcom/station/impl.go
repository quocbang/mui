package station

import (
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	mcomSites "gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/station"
)

var toActionType = map[models.SiteActionMode]mcomSites.ActionType{
	models.SiteActionMode(1): mcomSites.ActionType_ADD,
	models.SiteActionMode(2): mcomSites.ActionType_REMOVE,
}

// Station definitions.
type Station struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewStation returns Station service.
func NewStation(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.Station {
	return Station{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// GetStationList implementation.
func (s Station) GetStationList(params station.GetStationListParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_GET_STATION_LIST, principal.Roles) {
		return station.NewGetStationListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := s.dm.ListStations(ctx, mcom.ListStationsRequest{
		DepartmentID: params.DepartmentOID,
	})
	if err != nil {
		return utils.ParseError(ctx, station.NewGetStationListDefault(0), err)
	}
	data := make([]*station.GetStationListOKBodyDataItems0, len(list.Stations))
	for i, s := range list.Stations {
		data[i] = &station.GetStationListOKBodyDataItems0{ID: s.ID}
	}
	return station.NewGetStationListOK().WithPayload(&station.GetStationListOKBody{Data: data})
}

// List implementation
func (s Station) ListStationInfo(params station.ListStationInfoParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_LIST_STATION_INFO, principal.Roles) {
		return station.NewListStationInfoDefault(http.StatusForbidden)
	}

	pageRequest := mcom.PaginationRequest{}
	if params.Page != nil && params.Limit != nil {
		pageRequest = mcom.PaginationRequest{
			PageCount:      uint(*params.Page),
			ObjectsPerPage: uint(*params.Limit),
		}
	}

	orderRequest := parseOrderRequest(params.Body.OrderRequest, getStationMaintenanceListInfoByTypeDefaultOrderFunc)

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	list, err := s.dm.ListStations(ctx, mcom.ListStationsRequest{DepartmentID: params.DepartmentOID}.WithPagination(pageRequest).WithOrder(orderRequest...))
	if err != nil {
		return utils.ParseError(ctx, station.NewListStationInfoDefault(0), err)
	}

	data := make([]*models.StationData, len(list.Stations))
	for i, station := range list.Stations {
		sites := make([]*models.Site, len(station.Sites))
		for j, site := range station.Sites {
			content := &models.SiteContent{}
			switch site.Information.Type {
			case mcomSites.Type_SLOT:
				slot := mcomModels.BoundResource(*site.Content.Slot)
				content.Slot = parseBoundResources(site.Information.SubType, slot)[0]
			case mcomSites.Type_CONTAINER:
				container := []mcomModels.BoundResource(*site.Content.Container)
				content.Container = parseBoundResources(site.Information.SubType, container...)
			case mcomSites.Type_COLLECTION:
				collection := []mcomModels.BoundResource(*site.Content.Collection)
				content.Collection = parseBoundResources(site.Information.SubType, collection...)
			case mcomSites.Type_QUEUE:
				slots := []mcomModels.Slot(*site.Content.Queue)
				queue := make([]mcomModels.BoundResource, len(slots))
				for k, slot := range slots {
					queue[k] = mcomModels.BoundResource(slot)
				}
				content.Queue = parseBoundResources(site.Information.SubType, queue...)
			case mcomSites.Type_COLQUEUE:
				collections := []mcomModels.Collection(*site.Content.Colqueue)
				if len(collections) > 0 {
					colQueue := make([]models.CollectionContent, len(collections))
					for k, collection := range collections {
						colQueue[k] = parseBoundResources(site.Information.SubType, []mcomModels.BoundResource(collection)...)
					}
					content.Colqueue = colQueue
				}
			}
			sites[j] = &models.Site{
				Content: content,
				Index:   int64(site.Information.UniqueSite.SiteID.Index),
				Name:    site.Information.UniqueSite.SiteID.Name,
				SubType: models.SiteSubType(site.Information.SubType),
				Type:    models.SiteType(site.Information.Type),
			}
		}
		state := models.StationState(station.State)
		stationCode, stationDescription := station.Information.Code, station.Information.Description
		data[i] = &models.StationData{
			ID:          station.ID,
			Code:        &stationCode,
			Description: &stationDescription,
			InsertedAt:  strfmt.DateTime(station.InsertedAt),
			InsertedBy:  station.InsertedBy,
			Sites:       sites,
			State:       &state,
			UpdateAt:    strfmt.DateTime(station.UpdatedAt),
			UpdateBy:    station.UpdatedBy,
		}
	}

	return station.NewListStationInfoOK().WithPayload(&station.ListStationInfoOKBody{
		Data: &station.ListStationInfoOKBodyData{
			Items: data,
			Total: list.AmountOfData},
	})
}

// Create implementation
func (s Station) CreateStation(params station.CreateStationParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_CREATE_STATION, principal.Roles) {
		return station.NewCreateStationDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	sites := make([]mcom.SiteInformation, len(params.Body.Sites))
	for i, site := range params.Body.Sites {
		if site.ActionMode != models.SiteActionMode(1) {
			return station.NewCreateStationDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Details: "wrong site action mode",
			})
		}
		sites[i] = mcom.SiteInformation{
			Name:    site.Name,
			Index:   int(site.Index),
			Type:    mcomSites.Type(site.Type),
			SubType: mcomSites.SubType(site.SubType),
		}
	}
	req := mcom.CreateStationRequest{
		ID:           *params.Body.ID,
		DepartmentID: *params.Body.DepartmentOID,
		Sites:        sites,
		State:        stations.State_SHUTDOWN,
		Information: mcom.StationInformation{
			Code:        *params.Body.Code,
			Description: *params.Body.Description,
		},
	}
	if err := s.dm.CreateStation(ctx, req); err != nil {
		return utils.ParseError(ctx, station.NewCreateStationDefault(0), err)
	}

	return station.NewCreateStationOK()
}

// Update implementation
func (s Station) UpdateStationInfo(params station.UpdateStationInfoParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_UPDATE_STATION_INFO, principal.Roles) {
		return station.NewUpdateStationInfoDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	sites := make([]mcom.UpdateStationSite, len(params.Body.Sites))
	for i, site := range params.Body.Sites {
		actionType, ok := toActionType[site.ActionMode]
		if !ok {
			return station.NewUpdateStationInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Details: "not allow READ site on update station",
			})
		}
		sites[i] = mcom.UpdateStationSite{
			// minus one is for Action Mode READ
			ActionMode: actionType,
			Information: mcom.SiteInformation{
				Name:    site.Name,
				Index:   int(site.Index),
				Type:    mcomSites.Type(site.Type),
				SubType: mcomSites.SubType(site.SubType),
			},
		}
	}
	req := mcom.UpdateStationRequest{
		ID:    params.ID,
		Sites: sites,
		State: stations.State(*params.Body.State),
		Information: mcom.StationInformation{
			Code:        *params.Body.Code,
			Description: *params.Body.Description,
		},
	}
	if err := s.dm.UpdateStation(ctx, req); err != nil {
		return utils.ParseError(ctx, station.NewUpdateStationInfoDefault(0), err)
	}

	return station.NewUpdateStationInfoOK()
}

// Delete implementation
func (s Station) DeleteStation(params station.DeleteStationParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_DELETE_STATION, principal.Roles) {
		return station.NewDeleteStationDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	if err := s.dm.DeleteStation(ctx, mcom.DeleteStationRequest{
		StationID: params.ID,
	}); err != nil {
		return utils.ParseError(ctx, station.NewDeleteStationDefault(0), err)
	}

	return station.NewDeleteStationOK()
}

// GetStationStateList implementation.
func (s Station) GetStationStateList(params station.GetStationStateListParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_LIST_STATION_STATE, principal.Roles) {
		return station.NewGetStationStateListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := s.dm.ListStationState(ctx)
	if err != nil {
		return utils.ParseError(ctx, station.NewGetStationStateListDefault(0), err)
	}
	data := make([]*station.GetStationStateListOKBodyDataItems0, len(list))
	for i, stationState := range list {
		data[i] = &station.GetStationStateListOKBodyDataItems0{
			ID:   models.StationState(stationState.Value),
			Name: stationState.Name,
		}
	}
	return station.NewGetStationStateListOK().WithPayload(&station.GetStationStateListOKBody{Data: data})
}

// ListStations implements.
func (s Station) ListStations(params station.ListStationsParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_LIST_STATIONS, principal.Roles) {
		return station.NewListStationsDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	var (
		request = mcom.ListStationIDsRequest{}
	)
	if params.DepartmentOID != nil {
		request = mcom.ListStationIDsRequest{
			DepartmentID: *params.DepartmentOID,
		}
	}

	list, err := s.dm.ListStationIDs(ctx, request)

	if err != nil {
		return utils.ParseError(ctx, station.NewListStationsDefault(0), err)
	}

	return station.NewListStationsOK().WithPayload(&station.ListStationsOKBody{
		Data: parseStationList(list),
	})
}

// ListStationSites implements.
func (s Station) ListStationSites(params station.ListStationSitesParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_LIST_STATION_SITES, principal.Roles) {
		return station.NewListStationSitesDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := s.dm.GetStation(ctx, mcom.GetStationRequest{
		ID: params.StationID,
	})
	if err != nil {
		return utils.ParseError(ctx, station.NewListStationSitesDefault(0), err)
	}

	return station.NewListStationSitesOK().WithPayload(&station.ListStationSitesOKBody{
		Data: parseStationSites(list),
	})
}

// StationForceSignIn implements.
func (s Station) StationForceSignIn(params station.StationForceSignInParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_STATION_FORCE_SIGN_IN, principal.Roles) {
		return station.NewStationForceSignInDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	err := s.dm.SignInStation(ctx, mcom.SignInStationRequest{
		Station: params.StationID,
		Site: mcomModels.SiteID{
			Name:  params.Body.SiteName,
			Index: 0,
		},
		Group:    int32(params.Body.Group),
		WorkDate: time.Time(params.Body.WorkDate),
	}, mcom.ForceSignIn(), mcom.CreateSiteIfNotExists())
	if err != nil {
		return utils.ParseError(ctx, station.NewStationForceSignInDefault(0), err)
	}
	return station.NewStationForceSignInOK()
}

// StationSignOut implements.
func (s Station) StationSignOut(params station.StationSignOutParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_STATION_SIGN_OUT, principal.Roles) {
		return station.NewStationSignOutDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	sites := make([]mcomModels.UniqueSite, len(params.Body.StationSites))
	for i, stationSite := range params.Body.StationSites {
		sites[i] = mcomModels.UniqueSite{
			Station: stationSite.StationID,
			SiteID: mcomModels.SiteID{
				Name:  stationSite.SiteName,
				Index: 0,
			},
		}
	}
	err := s.dm.SignOutStations(ctx, mcom.SignOutStationsRequest{
		Sites: sites,
	})
	if err != nil {
		return utils.ParseError(ctx, station.NewStationSignOutDefault(0), err)
	}

	return station.NewStationSignOutOK()
}

func parseStationList(dataIn mcom.ListStationIDsReply) []*station.ListStationsOKBodyDataItems0 {
	dataOut := make([]*station.ListStationsOKBodyDataItems0, len(dataIn.Stations))
	for i, data := range dataIn.Stations {
		dataOut[i] = &station.ListStationsOKBodyDataItems0{
			StationID: data,
		}
	}
	return dataOut
}

func parseStationSites(dataIn mcom.GetStationReply) []*station.ListStationSitesOKBodyDataItems0 {
	dataOut := make([]*station.ListStationSitesOKBodyDataItems0, len(dataIn.Sites))
	for i, data := range dataIn.Sites {
		dataOut[i] = &station.ListStationSitesOKBodyDataItems0{
			Site: &models.SiteInfo{
				StationID: data.Information.Station,
				SiteName:  data.Information.SiteID.Name,
				SiteIndex: int64(data.Information.SiteID.Index),
			},
			Type:    data.Information.Type.String(),
			SubType: int64(data.Information.SubType),
		}
	}
	return dataOut
}

func parseOrderRequest(dataIn []*station.ListStationInfoParamsBodyOrderRequestItems0, defaultOrderFunc func() []mcom.Order) []mcom.Order {
	length := len(dataIn)
	if length == 0 {
		return defaultOrderFunc()
	}
	dataOut := make([]mcom.Order, length)
	for i, d := range dataIn {
		dataOut[i] = mcom.Order{
			Name:       d.OrderName,
			Descending: d.Descending,
		}
	}
	return dataOut
}

func getStationMaintenanceListInfoByTypeDefaultOrderFunc() []mcom.Order {
	return []mcom.Order{{
		Name:       "id",
		Descending: false,
	}}
}

// parseBoundResources return nil if resources are zero length.
func parseBoundResources(subType mcomSites.SubType, resources ...mcomModels.BoundResource) []*models.BoundResource {
	if len(resources) == 0 {
		return nil
	}
	res := make([]*models.BoundResource, len(resources))
	switch subType {
	case mcomSites.SubType_OPERATOR:
		for i, resource := range resources {
			if resource.Operator != nil {
				res[i] = &models.BoundResource{
					OperatorSite: &models.BoundResourceOperatorSite{
						EmployeeID: resource.Operator.EmployeeID,
						Group:      int64(resource.Operator.Group),
						WorkDate:   strfmt.DateTime(resource.Operator.WorkDate),
					},
				}
			}
		}
	case mcomSites.SubType_MATERIAL:
		for i, resource := range resources {
			if resource.Material != nil {
				var quantity string
				if resource.Material.Quantity != nil {
					quantity = resource.Material.Quantity.String()
				}

				res[i] = &models.BoundResource{
					MaterialSite: &models.BoundResourceMaterialSite{
						ID:         resource.Material.Material.ID,
						Grade:      resource.Material.Material.Grade,
						Quantity:   quantity,
						ResourceID: resource.Material.ResourceID,
					},
				}
			}
		}
	case mcomSites.SubType_TOOL:
		for i, resource := range resources {
			if resource.Tool != nil {
				res[i] = &models.BoundResource{
					ToolSite: &models.BoundResourceToolSite{
						InstalledTime: strfmt.DateTime(resource.Tool.InstalledTime),
						ResourceID:    resource.Tool.ResourceID,
					},
				}
			}
		}
	}
	return res
}
