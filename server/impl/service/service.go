package service

import (
	"github.com/go-openapi/runtime/middleware"

	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
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

// Service Definition.
type Service struct {
	accountAuthorization AccountAuthorization
	legacy               Legacy
	product              Product
	plan                 Plan
	workOrder            WorkOrder
	station              Station
	recipe               Recipe
	resource             Resource
	warehouse            Warehouse
	site                 Site
	carrier              Carrier
	produce              Produce
	ui                   UI
	unspecified          Unspecified
	// add more service
}

//NewService return new RESTful API services.
func NewService(
	accountAuthorization AccountAuthorization,
	legacy Legacy,
	product Product,
	plan Plan,
	workOrder WorkOrder,
	station Station,
	recipe Recipe,
	resource Resource,
	warehouse Warehouse,
	site Site,
	carrier Carrier,
	produce Produce,
	ui UI,
	unspecified Unspecified,

) *Service {
	return &Service{
		accountAuthorization: accountAuthorization,
		legacy:               legacy,
		product:              product,
		plan:                 plan,
		workOrder:            workOrder,
		station:              station,
		recipe:               recipe,
		resource:             resource,
		warehouse:            warehouse,
		site:                 site,
		carrier:              carrier,
		produce:              produce,
		ui:                   ui,
		unspecified:          unspecified,
	}
}

// AccountAuthorization return account authorization services.
func (s *Service) AccountAuthorization() AccountAuthorization {
	return s.accountAuthorization
}

// Legacy return pda platform services.
func (s *Service) Legacy() Legacy {
	return s.legacy
}

// Product return product services.
func (s *Service) Product() Product {
	return s.product
}

// Plan return plan services.
func (s *Service) Plan() Plan {
	return s.plan
}

// WorkOrder return workOrder services.
func (s *Service) WorkOrder() WorkOrder {
	return s.workOrder
}

// Station return station services.
func (s *Service) Station() Station {
	return s.station
}

// Recipe return recipe services.
func (s *Service) Recipe() Recipe {
	return s.recipe
}

// Resource return resource services.
func (s *Service) Resource() Resource {
	return s.resource
}

// Warehouse return warehouse services.
func (s *Service) Warehouse() Warehouse {
	return s.warehouse
}

// Site return site services.
func (s *Service) Site() Site {
	return s.site
}

// Carrier return carrier services.
func (s *Service) Carrier() Carrier {
	return s.carrier
}

// Produce return produce services.
func (s *Service) Produce() Produce {
	return s.produce
}

// UI return ui services.
func (s *Service) UI() UI {
	return s.ui
}

// Unspecified return unspecified services.
func (s *Service) Unspecified() Unspecified {
	return s.unspecified
}

// AccountAuthorization service available function methods.
type AccountAuthorization interface {
	Auth(token string) (*models.Principal, error)
	Login(params account.LoginParams) middleware.Responder
	Logout(params account.LogoutParams) middleware.Responder
	ChangePassword(params account.ChangePasswordParams, principal *models.Principal) middleware.Responder
	GetRoleList(params account.GetRoleListParams, principal *models.Principal) middleware.Responder
	ListAuthorizedAccount(params account.ListAuthorizedAccountParams, principal *models.Principal) middleware.Responder
	ListUnauthorizedAccount(params account.ListUnauthorizedAccountParams, principal *models.Principal) middleware.Responder
	CreateAccountAuthorization(params account.CreateAccountAuthorizationParams, principal *models.Principal) middleware.Responder
	UpdateAccountAuthorization(params account.UpdateAccountAuthorizationParams, principal *models.Principal) middleware.Responder
	DeleteAccount(params account.DeleteAccountParams, principal *models.Principal) middleware.Responder
}

// Legacy service available function methods.
type Legacy interface {
	GetBarcodeInfo(params legacy.GetBarcodeInfoParams, principal *models.Principal) middleware.Responder
	UpdateBarcode(params legacy.UpdateBarcodeParams, principal *models.Principal) middleware.Responder
	GetUpdateBarcodeStatusList(params legacy.GetUpdateBarcodeStatusListParams, principal *models.Principal) middleware.Responder
	GetExtendDays(params legacy.GetExtendDaysParams, principal *models.Principal) middleware.Responder
	GetControlAreaList(params legacy.GetControlAreaListParams, principal *models.Principal) middleware.Responder
	GetHoldReasonList(params legacy.GetHoldReasonListParams, principal *models.Principal) middleware.Responder
}

// Product service available function methods.
type Product interface {
	GetProductTypeByDepartmentList(params product.GetProductTypeByDepartmentListParams, principal *models.Principal) middleware.Responder
	GetProductTypeList(params product.GetProductTypeListParams, principal *models.Principal) middleware.Responder
	GetProductList(params product.GetProductListParams, principal *models.Principal) middleware.Responder
	GetMaterialResourceInfoByType(params product.GetMaterialResourceInfoByTypeParams, principal *models.Principal) middleware.Responder
}

// Plan service available function methods.
type Plan interface {
	GetPlanList(params plan.GetPlanListParams, principal *models.Principal) middleware.Responder
	AddPlan(params plan.AddPlanParams, principal *models.Principal) middleware.Responder
}

// WorkOrder service available function methods.
type WorkOrder interface {
	CreateStationScheduling(params work_order.CreateStationSchedulingParams, principal *models.Principal) middleware.Responder
	UpdateStationScheduling(params work_order.UpdateStationSchedulingParams, principal *models.Principal) middleware.Responder
	GetStationScheduling(params work_order.GetStationSchedulingParams, principal *models.Principal) middleware.Responder
	ListWorkOrders(params work_order.ListWorkOrdersParams, principal *models.Principal) middleware.Responder
	ChangeWorkOrderStatus(params work_order.ChangeWorkOrderStatusParams, principal *models.Principal) middleware.Responder
	GetWorkOrderInformation(params work_order.GetWorkOrderInformationParams, principal *models.Principal) middleware.Responder
	UpdateWorkOrder(params work_order.UpdateWorkOrderParams, principal *models.Principal) middleware.Responder
	CreateWorkOrdersFromFile(params work_order.CreateWorkOrdersFromFileParams, principal *models.Principal) middleware.Responder
	ListWorkOrdersRate(params work_order.ListWorkOrdersRateParams, rincipal *models.Principal) middleware.Responder
}

// Station service available function methods.
type Station interface {
	GetStationList(params station.GetStationListParams, principal *models.Principal) middleware.Responder
	ListStationInfo(params station.ListStationInfoParams, principal *models.Principal) middleware.Responder
	CreateStation(params station.CreateStationParams, principal *models.Principal) middleware.Responder
	UpdateStationInfo(params station.UpdateStationInfoParams, principal *models.Principal) middleware.Responder
	DeleteStation(params station.DeleteStationParams, principal *models.Principal) middleware.Responder
	GetStationStateList(params station.GetStationStateListParams, principal *models.Principal) middleware.Responder
	ListStations(params station.ListStationsParams, principal *models.Principal) middleware.Responder
	ListStationSites(params station.ListStationSitesParams, principal *models.Principal) middleware.Responder
	StationForceSignIn(params station.StationForceSignInParams, principal *models.Principal) middleware.Responder
	StationSignOut(params station.StationSignOutParams, principal *models.Principal) middleware.Responder
}

// Recipe service available function methods.
type Recipe interface {
	GetRecipeList(params recipe.GetRecipeListParams, principal *models.Principal) middleware.Responder
	GetRecipeIDs(params recipe.GetRecipeIDsParams, principal *models.Principal) middleware.Responder
	GetRecipeProcessList(params recipe.GetRecipeProcessListParams, principal *models.Principal) middleware.Responder
}

// Resource service available function methods.
type Resource interface {
	AddMaterial(params resource.AddMaterialParams, principal *models.Principal) middleware.Responder
	SplitMaterial(params resource.SplitMaterialParams, principal *models.Principal) middleware.Responder
	GetMaterialResourceInfo(params resource.GetMaterialResourceInfoParams, principal *models.Principal) middleware.Responder
	ListMaterialStatus(params resource.ListMaterialStatusParams, principal *models.Principal) middleware.Responder
	DownloadMaterialResource(params resource.DownloadMaterialResourceParams, principal *models.Principal) middleware.Responder
	GetToolID(params resource.GetToolIDParams, principal *models.Principal) middleware.Responder
	PrintMaterialResource(params resource.PrintMaterialResourceParams, principal *models.Principal) middleware.Responder
	DownloadPreMaterialResource(params resource.DownloadPreMaterialResourceParams, principal *models.Principal) middleware.Responder
}

// Warehouse service available function methods.
type Warehouse interface {
	GetWarehouseInfo(params warehouse.GetWarehouseInfoParams, principal *models.Principal) middleware.Responder
	WarehouseTransaction(params warehouse.WarehouseTransactionParams, principal *models.Principal) middleware.Responder
}

// Site service available function methods.
type Site interface {
	GetSiteMaterialList(params site.GetSiteMaterialListParams, principal *models.Principal) middleware.Responder
	AutoBindResource(params site.AutoBindSiteResourcesParams, principal *models.Principal) middleware.Responder
	ListType(params site.GetSiteTypeListParams, principal *models.Principal) middleware.Responder
	ListSubType(params site.GetSiteSubTypeListParams, principal *models.Principal) middleware.Responder
	GetSiteInformation(params site.GetSiteInformationParams, principal *models.Principal) middleware.Responder
	GetStationOperator(params site.GetStationOperatorParams, principal *models.Principal) middleware.Responder
}

// Carrier service available function methods.
type Carrier interface {
	GetCarrierList(params carrier.GetCarrierListParams, principal *models.Principal) middleware.Responder
	CreateCarrier(params carrier.CreateCarrierParams, principal *models.Principal) middleware.Responder
	UpdateCarrier(params carrier.UpdateCarrierParams, principal *models.Principal) middleware.Responder
	DeleteCarrier(params carrier.DeleteCarrierParams, principal *models.Principal) middleware.Responder
	DownloadCode39(params carrier.DownloadCode39Params) middleware.Responder
	DownloadQRCode(params carrier.DownloadQRCodeParams) middleware.Responder
}

// Produce service available function methods.
type Produce interface {
	FeedCollect(params produce.FeedCollectParams, principal *models.Principal) middleware.Responder
	MesFeed(params produce.MesFeedParams, principal *models.Principal) middleware.Responder
	MesCollect(params produce.MesCollectParams, principal *models.Principal) middleware.Responder
}

// UI service available function methods.
type UI interface {
	GetProductGroupList(params ui.GetProductGroupListParams, principal *models.Principal) middleware.Responder
	SetStationConfig(params ui.SetStationConfigParams, principal *models.Principal) middleware.Responder
	GetStationConfig(params ui.GetStationConfigParams, principal *models.Principal) middleware.Responder
}

// Unspecified service available function methods.
type Unspecified interface {
	ListDepartmentIDs(params unspecified.ListDepartmentIDsParams, principal *models.Principal) middleware.Responder
}
