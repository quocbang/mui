package mcom

import (
	"fmt"
	"time"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mui/server/configs"
	accountImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	carrierImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/carrier"
	legacyImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/legacy"
	planImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/plan"
	produceImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/produce"
	productImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/product"
	recipeImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/recipe"
	resourceImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/resource"
	siteImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/site"
	stationImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/station"
	uiImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/ui"
	unspecifiedImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/unspecified"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
	warehouseImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/warehouse"
	workOrderImpl "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/workorder"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils/role"

	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/account"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/carrier"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/legacy"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/plan"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/produce"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/product"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/recipe"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/resource"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/site"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/station"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/ui"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/unspecified"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/warehouse"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/work_order"
)

type ServiceConfig struct {
	TokenLifeTime         time.Duration
	Printers              map[string]string
	FontPath              string
	StationFunctionConfig map[string]configs.FunctionAPIPath
	MesPath               string
}

// RegisterServices register rest api service.
func RegisterServices(dm mcom.DataManager, config ServiceConfig) (*service.Service, error) {
	if config.FontPath == "" {
		return nil, fmt.Errorf("missing font path")
	}

	workOrderService := workOrderImpl.NewWorkOrder(dm, role.HasPermission, workOrderImpl.Config{
		StationFunctionConfig: config.StationFunctionConfig,
	})

	resourceService := resourceImpl.NewResource(dm, role.HasPermission, resourceImpl.Config{
		Printers: config.Printers,
		FontPath: config.FontPath,
	})

	produceService := produceImpl.NewProduce(dm, role.HasPermission, produceImpl.Config{
		Printers: config.Printers,
		FontPath: config.FontPath,
		MesPath:  config.MesPath,
	})

	siteService := siteImpl.NewSite(dm, role.HasPermission, siteImpl.Config{
		StationFunctionConfig: config.StationFunctionConfig,
	})

	return service.NewService(
		accountImpl.NewAuthorization(dm, role.HasPermission, config.TokenLifeTime),
		legacyImpl.NewLegacy(dm, role.HasPermission),
		productImpl.NewProduct(dm, role.HasPermission),
		planImpl.NewPlan(dm, role.HasPermission),
		workOrderService,
		stationImpl.NewStation(dm, role.HasPermission),
		recipeImpl.NewRecipe(dm, role.HasPermission),
		resourceService,
		warehouseImpl.NewWarehouse(dm, role.HasPermission),
		siteService,
		carrierImpl.NewCarrier(dm, role.HasPermission),
		produceService,
		uiImpl.NewUI(dm, role.HasPermission),
		unspecifiedImpl.NewUnspecified(dm, role.HasPermission),
	), nil
}

// RegisterHandlers register real handlers
func RegisterHandlers(dm mcom.DataManager, api *operations.MuiAPI, config ServiceConfig) error {
	s, err := RegisterServices(dm, config)
	if err != nil {
		return err
	}

	api.APIKeyAuth = s.AccountAuthorization().Auth // API key auth

	// account handlers.
	api.AccountLoginHandler = account.LoginHandlerFunc(s.AccountAuthorization().Login)
	api.AccountLogoutHandler = account.LogoutHandlerFunc(s.AccountAuthorization().Logout)
	api.AccountChangePasswordHandler = account.ChangePasswordHandlerFunc(s.AccountAuthorization().ChangePassword)
	api.AccountGetRoleListHandler = account.GetRoleListHandlerFunc(s.AccountAuthorization().GetRoleList)
	api.AccountListAuthorizedAccountHandler = account.ListAuthorizedAccountHandlerFunc(s.AccountAuthorization().ListAuthorizedAccount)
	api.AccountListUnauthorizedAccountHandler = account.ListUnauthorizedAccountHandlerFunc(s.AccountAuthorization().ListUnauthorizedAccount)
	api.AccountCreateAccountAuthorizationHandler = account.CreateAccountAuthorizationHandlerFunc(s.AccountAuthorization().CreateAccountAuthorization)
	api.AccountUpdateAccountAuthorizationHandler = account.UpdateAccountAuthorizationHandlerFunc(s.AccountAuthorization().UpdateAccountAuthorization)
	api.AccountDeleteAccountHandler = account.DeleteAccountHandlerFunc(s.AccountAuthorization().DeleteAccount)

	// operations handler.
	api.CheckServerStatusHandler = operations.CheckServerStatusHandlerFunc(utils.GetServerStatus)

	// legacy handlers.
	api.LegacyGetBarcodeInfoHandler = legacy.GetBarcodeInfoHandlerFunc(s.Legacy().GetBarcodeInfo)
	api.LegacyUpdateBarcodeHandler = legacy.UpdateBarcodeHandlerFunc(s.Legacy().UpdateBarcode)
	api.LegacyGetUpdateBarcodeStatusListHandler = legacy.GetUpdateBarcodeStatusListHandlerFunc(s.Legacy().GetUpdateBarcodeStatusList)
	api.LegacyGetExtendDaysHandler = legacy.GetExtendDaysHandlerFunc(s.Legacy().GetExtendDays)
	api.LegacyGetControlAreaListHandler = legacy.GetControlAreaListHandlerFunc(s.Legacy().GetControlAreaList)
	api.LegacyGetHoldReasonListHandler = legacy.GetHoldReasonListHandlerFunc(s.Legacy().GetHoldReasonList)

	// product handlers.
	api.ProductGetProductTypeByDepartmentListHandler = product.GetProductTypeByDepartmentListHandlerFunc(s.Product().GetProductTypeByDepartmentList)
	api.ProductGetProductTypeListHandler = product.GetProductTypeListHandlerFunc(s.Product().GetProductTypeList)
	api.ProductGetProductListHandler = product.GetProductListHandlerFunc(s.Product().GetProductList)
	api.ProductGetMaterialResourceInfoByTypeHandler = product.GetMaterialResourceInfoByTypeHandlerFunc(s.Product().GetMaterialResourceInfoByType)

	// plan handlers.
	api.PlanGetPlanListHandler = plan.GetPlanListHandlerFunc(s.Plan().GetPlanList)
	api.PlanAddPlanHandler = plan.AddPlanHandlerFunc(s.Plan().AddPlan)

	// work Order handlers.
	api.WorkOrderGetStationSchedulingHandler = work_order.GetStationSchedulingHandlerFunc(s.WorkOrder().GetStationScheduling)
	api.WorkOrderListWorkOrdersHandler = work_order.ListWorkOrdersHandlerFunc(s.WorkOrder().ListWorkOrders)
	api.WorkOrderChangeWorkOrderStatusHandler = work_order.ChangeWorkOrderStatusHandlerFunc(s.WorkOrder().ChangeWorkOrderStatus)
	api.WorkOrderGetWorkOrderInformationHandler = work_order.GetWorkOrderInformationHandlerFunc(s.WorkOrder().GetWorkOrderInformation)
	api.WorkOrderCreateStationSchedulingHandler = work_order.CreateStationSchedulingHandlerFunc(s.WorkOrder().CreateStationScheduling)
	api.WorkOrderUpdateStationSchedulingHandler = work_order.UpdateStationSchedulingHandlerFunc(s.WorkOrder().UpdateStationScheduling)
	api.WorkOrderUpdateWorkOrderHandler = work_order.UpdateWorkOrderHandlerFunc(s.WorkOrder().UpdateWorkOrder)
	api.WorkOrderCreateWorkOrdersFromFileHandler = work_order.CreateWorkOrdersFromFileHandlerFunc(s.WorkOrder().CreateWorkOrdersFromFile)
	api.WorkOrderListWorkOrdersRateHandler = work_order.ListWorkOrdersRateHandlerFunc(s.WorkOrder().ListWorkOrdersRate)

	// station handlers.
	api.StationGetStationListHandler = station.GetStationListHandlerFunc(s.Station().GetStationList)
	api.StationGetStationStateListHandler = station.GetStationStateListHandlerFunc(s.Station().GetStationStateList)
	api.StationListStationInfoHandler = station.ListStationInfoHandlerFunc(s.Station().ListStationInfo)
	api.StationCreateStationHandler = station.CreateStationHandlerFunc(s.Station().CreateStation)
	api.StationUpdateStationInfoHandler = station.UpdateStationInfoHandlerFunc(s.Station().UpdateStationInfo)
	api.StationDeleteStationHandler = station.DeleteStationHandlerFunc(s.Station().DeleteStation)
	api.StationListStationSitesHandler = station.ListStationSitesHandlerFunc(s.Station().ListStationSites)
	api.StationListStationsHandler = station.ListStationsHandlerFunc(s.Station().ListStations)
	api.StationStationForceSignInHandler = station.StationForceSignInHandlerFunc(s.Station().StationForceSignIn)
	api.StationStationSignOutHandler = station.StationSignOutHandlerFunc(s.Station().StationSignOut)

	// recipe handlers.
	api.RecipeGetRecipeListHandler = recipe.GetRecipeListHandlerFunc(s.Recipe().GetRecipeList)
	api.RecipeGetRecipeIDsHandler = recipe.GetRecipeIDsHandlerFunc(s.Recipe().GetRecipeIDs)
	api.RecipeGetRecipeProcessListHandler = recipe.GetRecipeProcessListHandlerFunc(s.Recipe().GetRecipeProcessList)

	// resource handlers.
	api.ResourceAddMaterialHandler = resource.AddMaterialHandlerFunc(s.Resource().AddMaterial)
	api.ResourceSplitMaterialHandler = resource.SplitMaterialHandlerFunc(s.Resource().SplitMaterial)
	api.ResourceGetMaterialResourceInfoHandler = resource.GetMaterialResourceInfoHandlerFunc(s.Resource().GetMaterialResourceInfo)
	api.ResourceListMaterialStatusHandler = resource.ListMaterialStatusHandlerFunc(s.Resource().ListMaterialStatus)
	api.ResourceDownloadMaterialResourceHandler = resource.DownloadMaterialResourceHandlerFunc(s.Resource().DownloadMaterialResource)
	api.ResourceGetToolIDHandler = resource.GetToolIDHandlerFunc(s.Resource().GetToolID)
	api.ResourcePrintMaterialResourceHandler = resource.PrintMaterialResourceHandlerFunc(s.Resource().PrintMaterialResource)
	api.ResourceDownloadPreMaterialResourceHandler = resource.DownloadPreMaterialResourceHandlerFunc(s.Resource().DownloadPreMaterialResource)

	// warehouse handlers.
	api.WarehouseGetWarehouseInfoHandler = warehouse.GetWarehouseInfoHandlerFunc(s.Warehouse().GetWarehouseInfo)
	api.WarehouseWarehouseTransactionHandler = warehouse.WarehouseTransactionHandlerFunc(s.Warehouse().WarehouseTransaction)

	// site handlers.
	api.SiteGetSiteMaterialListHandler = site.GetSiteMaterialListHandlerFunc(s.Site().GetSiteMaterialList)
	api.SiteAutoBindSiteResourcesHandler = site.AutoBindSiteResourcesHandlerFunc(s.Site().AutoBindResource)
	api.SiteGetSiteSubTypeListHandler = site.GetSiteSubTypeListHandlerFunc(s.Site().ListSubType)
	api.SiteGetSiteTypeListHandler = site.GetSiteTypeListHandlerFunc(s.Site().ListType)
	api.SiteGetSiteInformationHandler = site.GetSiteInformationHandlerFunc(s.Site().GetSiteInformation)
	api.SiteGetStationOperatorHandler = site.GetStationOperatorHandlerFunc(s.Site().GetStationOperator)

	// carrier handlers.
	api.CarrierGetCarrierListHandler = carrier.GetCarrierListHandlerFunc(s.Carrier().GetCarrierList)
	api.CarrierCreateCarrierHandler = carrier.CreateCarrierHandlerFunc(s.Carrier().CreateCarrier)
	api.CarrierUpdateCarrierHandler = carrier.UpdateCarrierHandlerFunc(s.Carrier().UpdateCarrier)
	api.CarrierDeleteCarrierHandler = carrier.DeleteCarrierHandlerFunc(s.Carrier().DeleteCarrier)
	api.CarrierDownloadCode39Handler = carrier.DownloadCode39HandlerFunc(s.Carrier().DownloadCode39)
	api.CarrierDownloadQRCodeHandler = carrier.DownloadQRCodeHandlerFunc(s.Carrier().DownloadQRCode)

	// produce handlers.
	api.ProduceFeedCollectHandler = produce.FeedCollectHandlerFunc(s.Produce().FeedCollect)
	api.ProduceMesFeedHandler = produce.MesFeedHandlerFunc(s.Produce().MesFeed)
	api.ProduceMesCollectHandler = produce.MesCollectHandlerFunc(s.Produce().MesCollect)

	// ui handlers.
	api.UISetStationConfigHandler = ui.SetStationConfigHandlerFunc(s.UI().SetStationConfig)
	api.UIGetStationConfigHandler = ui.GetStationConfigHandlerFunc(s.UI().GetStationConfig)
	api.UIGetProductGroupListHandler = ui.GetProductGroupListHandlerFunc(s.UI().GetProductGroupList)

	// unspecified handlers
	api.UnspecifiedListDepartmentIDsHandler = unspecified.ListDepartmentIDsHandlerFunc(s.Unspecified().ListDepartmentIDs)

	// operations handler.
	api.CheckServerStatusHandler = operations.CheckServerStatusHandlerFunc(utils.GetServerStatus)

	return nil
}
