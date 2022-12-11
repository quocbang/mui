package site

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"

	"gitlab.kenda.com.tw/kenda/mui/server/configs"
	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	mesageModels "gitlab.kenda.com.tw/kenda/mui/server/models"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/site"
)

// Site definition.
type Site struct {
	dm            mcom.DataManager
	config        Config
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

type Config struct {
	StationFunctionConfig map[string]configs.FunctionAPIPath
}

// NewSite initialize Site.
func NewSite(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool,
	config Config) service.Site {
	return Site{
		dm:            dm,
		hasPermission: hasPermission,
		config:        config,
	}
}

// GetSiteMaterialList implementation.
func (s Site) GetSiteMaterialList(params site.GetSiteMaterialListParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_GET_SITE_MATERIAL_LIST, principal.Roles) {
		return site.NewGetSiteMaterialListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	materials, err := s.dm.ListSiteMaterials(ctx, mcom.ListSiteMaterialsRequest{
		Station: params.Station,
		Site: mcomModels.SiteID{
			Name:  params.SiteName,
			Index: int16(params.SiteIndex),
		},
	})
	if err != nil {
		return utils.ParseError(ctx, site.NewGetSiteMaterialListDefault(0), err)
	}

	queryResourceIDs := make([]string, len(materials))
	bindData := make([]*models.BindMaterialData, len(materials))
	for i, material := range materials {
		resourceID := material.ResourceID
		bindData[i] = &models.BindMaterialData{
			Grade:      models.Grade(material.Grade),
			ProductID:  material.ID,
			Quantity:   "",
			ResourceID: resourceID,
		}
		if material.Quantity != nil {
			bindData[i].Quantity = material.Quantity.String()
		}

		queryResourceIDs[i] = resourceID
	}

	return site.NewGetSiteMaterialListOK().WithPayload(&site.GetSiteMaterialListOKBody{Data: bindData})
}

// AutoBindResource implementation.
func (s Site) AutoBindResource(params site.AutoBindSiteResourcesParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_BIND_RESOURCE, principal.Roles) {
		return site.NewAutoBindSiteResourcesDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	// 檢查機台，工位是否存在
	stationSite, err := s.dm.GetSite(ctx, mcom.GetSiteRequest{
		StationID: *params.Body.Station,
		SiteName:  *params.Body.SiteName,
		SiteIndex: int16(*params.Body.SiteIndex),
	})
	if err != nil {
		return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), err)
	}

	var (
		Material  = int64(0)
		Tool      = int64(1)
		forceBind = false

		recipeConfig = mcom.RecipeProcessConfig{}
	)

	if params.Body.ForceBind != nil {
		forceBind = params.Body.ForceBind.Force
	}

	// if work order not empty & bind type not clear
	if params.Body.WorkOrderID != "" && bindClearCheck(bindtype.BindType(*params.Body.BindType)) {
		// get RecipeID & Process
		getWorkOrder, err := s.dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{
			ID: params.Body.WorkOrderID,
		})
		if err != nil {
			return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), err)
		}

		// get Recipe
		getRecipe, err := s.dm.GetProcessDefinition(ctx, mcom.GetProcessDefinitionRequest{
			RecipeID:    getWorkOrder.RecipeID,
			ProcessName: getWorkOrder.Process.Name,
			ProcessType: getWorkOrder.Process.Type,
		})
		if err != nil {
			return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), err)
		}

		for _, data := range getRecipe.Configs {
			for _, configStation := range data.Stations {
				if configStation == *params.Body.Station {
					recipeConfig = *data
					break
				}
			}
		}
	}

	switch params.Body.ResourceType {
	case Material:
		bindOption := mcomModels.BindOption{}
		if params.Body.QueueOption != nil {
			if params.Body.QueueOption.Head {
				bindOption.Head = true
			} else if params.Body.QueueOption.Tail {
				bindOption.Tail = true
			} else {
				queueIndex := uint16(params.Body.QueueOption.Index)
				bindOption.QueueIndex = &queueIndex
			}
		}
		detail := mcom.MaterialBindRequestDetailV2{
			Type: bindtype.BindType(*params.Body.BindType),
			Site: mcomModels.UniqueSite{
				Station: *params.Body.Station,
				SiteID: mcomModels.SiteID{
					Name:  *params.Body.SiteName,
					Index: int16(*params.Body.SiteIndex),
				},
			},
			Option: bindOption,
		}

		materialList := []*mesageModels.SiteBindingStateResourcesItems0{}

		// if bind is clear, not to find material data
		if bindClearCheck(bindtype.BindType(*params.Body.BindType)) {
			bindMaterials, err := validateAndParseBindMaterials(ctx, s.dm, stationSite.Attributes, params.Body.Resources, forceBind)
			if err != nil {
				return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), err)
			}

			for _, data := range bindMaterials {
				materialList = append(materialList, &mesageModels.SiteBindingStateResourcesItems0{
					ID: data.ResourceID,
				})
			}
			detail.Resources = bindMaterials

			if params.Body.WorkOrderID != "" && !forceBind {

				// if material not in recipe,return error
				for _, bindMaterialResource := range detail.Resources {
					if !materialCheck(bindMaterialResource.Material.ID, recipeConfig.Steps) {
						return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), mcomErrors.Error{
							Code:    mcomErrors.Code_RESOURCE_WORKORDER_RESOURCE_UNEXPECTED,
							Details: "material resource not in recipe",
						})
					}
				}
			}
		} else {
			detail.Resources = []mcom.BindMaterialResource{}
		}

		if err := s.dm.MaterialResourceBindV2(ctx, mcom.MaterialResourceBindRequestV2{
			Details: []mcom.MaterialBindRequestDetailV2{detail},
		}); err != nil {
			return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), err)
		}

		if apiConfig := s.config.StationFunctionConfig[*params.Body.Station]; apiConfig.BindResourceAPIPath != "" {

			var (
				resourceBindRequest mesageModels.NotifyBindResourceRequestBody
			)
			siteMaterial, notOK := materialResourceList(stationSite, recipeConfig.Steps)
			materialList = append(materialList, siteMaterial...)
			// check material clear
			if bindClearCheck(bindtype.BindType(*params.Body.BindType)) {

				resourceBindRequest = mesageModels.NotifyBindResourceRequestBody{
					BindType: int64(*params.Body.BindType),
					Site: &mesageModels.Site{
						Station: *params.Body.Station,
						Name:    *params.Body.SiteName,
						Index:   *params.Body.SiteIndex,
					},
					CurrentState: &mesageModels.SiteBindingState{
						Resources: materialList,
						NotOK:     notOK,
					},
				}
			} else {
				resourceBindRequest = mesageModels.NotifyBindResourceRequestBody{
					BindType: int64(*params.Body.BindType),
					Site: &mesageModels.Site{
						Station: *params.Body.Station,
						Name:    *params.Body.SiteName,
						Index:   *params.Body.SiteIndex,
					},
					CurrentState: &mesageModels.SiteBindingState{
						Resources: []*mesageModels.SiteBindingStateResourcesItems0{},
						NotOK:     true,
					},
				}
			}

			// send bindResource request to MES agent
			if err := handlerUtils.SendMesAgePOSTRequest(resourceBindRequest, apiConfig.BindResourceAPIPath); err != nil {
				commonsCtx.Logger(ctx).Error("failed to send the request to MES Agent", zap.Error(err))
			}
		}

	case Tool:

		toolBindResource := mcom.ToolResource{}
		if bindClearCheck(bindtype.BindType(*params.Body.BindType)) {
			getToolID, err := s.dm.GetToolResource(ctx, mcom.GetToolResourceRequest{
				ResourceID: params.Body.Resources[0].ResourceID,
			})

			if err != nil {
				e, ok := mcomErrors.As(err)
				if !ok {
					return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), err)
				}
				if e.Code == mcomErrors.Code_RESOURCE_NOT_FOUND && forceBind {
					toolBindResource = mcom.ToolResource{
						ResourceID: params.Body.Resources[0].ResourceID,
					}
				} else {
					return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), err)
				}
			} else {
				if params.Body.WorkOrderID != "" && !forceBind {
					if !toolCheck(getToolID.ToolID, recipeConfig.Tools) {
						return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), mcomErrors.Error{
							Code:    mcomErrors.Code_RESOURCE_WORKORDER_RESOURCE_UNEXPECTED,
							Details: "tool resource not in recipe",
						})
					}
				}

				toolBindResource = mcom.ToolResource{
					ResourceID: params.Body.Resources[0].ResourceID,
					ToolID:     getToolID.ToolID,
				}
			}
		}

		detail := mcom.ToolBindRequestDetailV2{
			Type: bindtype.BindType(*params.Body.BindType),
			Site: mcomModels.UniqueSite{
				SiteID: mcomModels.SiteID{
					Name:  *params.Body.SiteName,
					Index: int16(*params.Body.SiteIndex),
				},

				Station: *params.Body.Station,
			},
			Resource: toolBindResource,
		}

		if err := s.dm.ToolResourceBindV2(ctx, mcom.ToolResourceBindRequestV2{
			Details: []mcom.ToolBindRequestDetailV2{detail},
		}); err != nil {
			return utils.ParseError(ctx, site.NewAutoBindSiteResourcesDefault(0), err)
		}
	}

	return site.NewAutoBindSiteResourcesOK()
}

// ListType implementation.
func (s Site) ListType(params site.GetSiteTypeListParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_SITE_TYPE_LIST, principal.Roles) {
		return site.NewGetSiteTypeListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := s.dm.ListSiteType(ctx)
	if err != nil {
		return utils.ParseError(ctx, site.NewGetSiteTypeListDefault(0), err)
	}

	data := make([]*site.GetSiteTypeListOKBodyDataItems0, len(list))
	for i, siteType := range list {
		data[i] = &site.GetSiteTypeListOKBodyDataItems0{
			ID:   models.SiteType(siteType.Value),
			Name: siteType.Name,
		}
	}

	return site.NewGetSiteTypeListOK().WithPayload(&site.GetSiteTypeListOKBody{Data: data})
}

// ListSubType implementation.
func (s Site) ListSubType(params site.GetSiteSubTypeListParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_SITE_SUBTYPE_LIST, principal.Roles) {
		return site.NewGetSiteSubTypeListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := s.dm.ListSiteSubType(ctx)
	if err != nil {
		return utils.ParseError(ctx, site.NewGetSiteSubTypeListDefault(0), err)
	}

	data := make([]*site.GetSiteSubTypeListOKBodyDataItems0, len(list))
	for i, siteSubType := range list {
		data[i] = &site.GetSiteSubTypeListOKBodyDataItems0{
			ID:   models.SiteSubType(siteSubType.Value),
			Name: siteSubType.Name,
		}
	}

	return site.NewGetSiteSubTypeListOK().WithPayload(&site.GetSiteSubTypeListOKBody{Data: data})
}

// GetSiteInformation implements.
func (s Site) GetSiteInformation(params site.GetSiteInformationParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_GET_SITE_INFORMATION, principal.Roles) {
		return site.NewGetSiteInformationDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	list, err := s.dm.GetSite(ctx, mcom.GetSiteRequest{
		StationID: params.Body.Site.StationID,
		SiteName:  params.Body.Site.SiteName,
		SiteIndex: int16(params.Body.Site.SiteIndex),
	})
	if err != nil {
		return utils.ParseError(ctx, site.NewGetSiteInformationDefault(0), err)
	}

	return site.NewGetSiteInformationOK().WithPayload(&site.GetSiteInformationOKBody{
		Data: &site.GetSiteInformationOKBodyData{
			Type:    list.Attributes.Type.String(),
			SubType: int64(list.Attributes.SubType),
		},
	})
}

// GetStationOperator implements.
func (s Site) GetStationOperator(params site.GetStationOperatorParams, principal *models.Principal) middleware.Responder {
	if !s.hasPermission(kenda.FunctionOperationID_GET_STATION_OPERATOR, principal.Roles) {
		return site.NewGetStationOperatorDefault(http.StatusForbidden)
	}

	var operatorID string

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	siteMcom, err := s.dm.GetSite(ctx, mcom.GetSiteRequest{
		StationID: params.Body.Site.StationID,
		SiteName:  params.Body.Site.SiteName,
		SiteIndex: int16(params.Body.Site.SiteIndex),
	})
	if err != nil {
		if e, ok := mcomErrors.As(err); ok {
			// if first time sign in the station, the site is not found,
			// maybe get the error about station site not found,
			// but it means no operator, so response empty operator instead of the error
			if e.Code != mcomErrors.Code_STATION_SITE_NOT_FOUND {
				return utils.ParseError(ctx, site.NewGetStationOperatorDefault(0), err)
			}
		} else {
			return utils.ParseError(ctx, site.NewGetStationOperatorDefault(0), err)
		}
	} else {
		if siteMcom.Attributes.SubType != sites.SubType_OPERATOR || siteMcom.Attributes.Type != sites.Type_SLOT {
			return utils.ParseError(ctx, site.NewGetStationOperatorDefault(0), mcomErrors.Error{
				Code: mcomErrors.Code_STATION_SITE_SUB_TYPE_MISMATCH,
			})
		}

		operatorID = siteMcom.Content.Slot.Operator.Current().EmployeeID
	}

	return site.NewGetStationOperatorOK().WithPayload(&site.GetStationOperatorOKBody{
		Data: &site.GetStationOperatorOKBodyData{
			OperatorID: operatorID,
		},
	})
}

func bindClearCheck(target bindtype.BindType) bool {
	data := target.String()
	return data[len(data)-5:] != "CLEAR"
}

// check bind material resource in recipe
func materialCheck(target string, dataIn []*mcom.RecipeProcessStep) bool {
	for _, stepData := range dataIn {
		for _, materialData := range stepData.Materials {
			if target == materialData.Name {
				return true
			}
		}
	}
	return false
}

// check bind tool resource in recipe
func toolCheck(target string, dataIn []*mcom.RecipeTool) bool {
	for _, data := range dataIn {
		if target == data.ID {
			return true
		}
	}
	return false
}

func validateAndParseBindMaterials(
	ctx context.Context,
	dm mcom.DataManager,
	siteAttributes mcom.SiteAttributes,
	bindResources []*models.BindResource,
	forceBind bool) ([]mcom.BindMaterialResource, error) {
	// return bindMaterials
	listMaterialResourceIdentitiesRequest := make([]mcom.GetMaterialResourceIdentityRequest, len(bindResources))
	for i, requestResource := range bindResources {
		listMaterialResourceIdentitiesRequest[i] = mcom.GetMaterialResourceIdentityRequest{
			ResourceID:  requestResource.ResourceID,
			ProductType: requestResource.ProductType,
		}
	}

	listMaterialResourcesReply, err := dm.ListMaterialResourceIdentities(ctx, mcom.ListMaterialResourceIdentitiesRequest{
		Details: listMaterialResourceIdentitiesRequest,
	})
	if err != nil {
		return nil, err
	}

	bindMaterials := make([]mcom.BindMaterialResource, len(listMaterialResourcesReply.Replies))
	for i, materialResource := range listMaterialResourcesReply.Replies {
		quantity := decimal.Zero
		if bindResources[i].Quantity != "" {
			strQuantity := bindResources[i].Quantity
			quantity, err = decimal.NewFromString(strQuantity)
			if err != nil {
				return nil, mcomErrors.Error{
					Code:    mcomErrors.Code_INVALID_NUMBER,
					Details: "invalid_number=" + bindResources[i].Quantity,
				}
			}
		}

		if materialResource != nil {
			if bindResources[i].Quantity != "" {
				if materialResource.Material.Quantity.LessThan(quantity) ||
					materialResource.Material.Quantity.LessThanOrEqual(decimal.Zero) {
					return nil, mcomErrors.Error{
						Code: mcomErrors.Code_RESOURCE_MATERIAL_SHORTAGE,
						Details: fmt.Sprintf("not enough resource quantity to bind, index=%d, storage=%s, demand=%s",
							i, materialResource.Material.Quantity.String(), bindResources[i].Quantity),
					}
				}
			}

			if materialResource.Material.ExpiryTime.Before(time.Now()) {
				return nil, mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_EXPIRED,
					Details: fmt.Sprintf("resource expired, index=%d", i),
				}
			}

			if materialResource.Material.Status != resources.MaterialStatus_AVAILABLE {
				return nil, mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_UNAVAILABLE,
					Details: "material resource not available",
				}
			}

			// 檢查指定工位是否允許綁定指定產品代號
			if err := siteAttributes.LimitHandler(materialResource.Material.ID); err != nil {
				return nil, err
			}

			bindMaterials[i] = mcom.BindMaterialResource{
				Material: mcomModels.Material{
					ID:    materialResource.Material.ID,
					Grade: materialResource.Material.Grade,
				},
				ResourceID:  materialResource.Material.ResourceID,
				ProductType: materialResource.Material.Type,
				Warehouse: mcom.Warehouse{
					ID:       materialResource.Warehouse.ID,
					Location: materialResource.Warehouse.Location,
				},
				Status:     materialResource.Material.Status,
				ExpiryTime: types.ToTimeNano(materialResource.Material.ExpiryTime),
			}
			if bindResources[i].Quantity != "" {
				bindMaterials[i].Quantity = &quantity
			}

		} else {
			if !forceBind {
				return nil, mcomErrors.Error{
					Code:    mcomErrors.Code_RESOURCE_NOT_FOUND,
					Details: bindResources[i].ResourceID,
				}
			}

			bindMaterials[i] = mcom.BindMaterialResource{
				ResourceID: bindResources[i].ResourceID,
			}
		}
	}
	return bindMaterials, nil
}

func materialResourceList(dataIn mcom.GetSiteReply, recipe []*mcom.RecipeProcessStep) ([]*mesageModels.SiteBindingStateResourcesItems0, bool) {
	dataOut := []*mesageModels.SiteBindingStateResourcesItems0{}
	notOK := false
	if dataIn.Attributes.SubType == sites.SubType_MATERIAL {
		switch dataIn.Attributes.Type {
		case sites.Type_SLOT:
			m := dataIn.Content.Slot
			if m != nil && m.Material != nil {
				dataOut = append(dataOut, &mesageModels.SiteBindingStateResourcesItems0{
					ID: m.Material.ResourceID,
				})
				notOK = handlerUtils.SiteMaterialNotOKCheck(mcomModels.BoundResource(*m), recipe)
			}

		case sites.Type_CONTAINER, sites.Type_COLLECTION:
			var (
				content []mcomModels.BoundResource
			)
			if dataIn.Attributes.Type == sites.Type_CONTAINER {
				content = *dataIn.Content.Container
			} else {
				content = *dataIn.Content.Collection
			}

			for _, m := range content {
				dataOut = append(dataOut, &mesageModels.SiteBindingStateResourcesItems0{
					ID: m.Material.ResourceID,
				})
				if !notOK {
					notOK = handlerUtils.SiteMaterialNotOKCheck(mcomModels.BoundResource(m), recipe)
				}
			}

		case sites.Type_QUEUE:
			content := dataIn.Content.Queue

			for _, m := range *content {
				dataOut = append(dataOut, &mesageModels.SiteBindingStateResourcesItems0{
					ID: m.Material.ResourceID,
				})
				if !notOK {
					notOK = handlerUtils.SiteMaterialNotOKCheck(mcomModels.BoundResource(m), recipe)
				}
			}

		case sites.Type_COLQUEUE:
			content := dataIn.Content.Colqueue

			for _, m := range *content {
				for _, n := range m {
					dataOut = append(dataOut, &mesageModels.SiteBindingStateResourcesItems0{
						ID: n.Material.ResourceID,
					})
					if !notOK {
						notOK = handlerUtils.SiteMaterialNotOKCheck(mcomModels.BoundResource(n), recipe)
					}
				}
			}
		}
	}
	return dataOut, notOK
}
