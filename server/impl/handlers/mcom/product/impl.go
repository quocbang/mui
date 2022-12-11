package product

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/product"
)

// Product definitions.
type Product struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// NewProduct returns Product service.
func NewProduct(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.Product {
	return Product{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// GetProductTypeByDepartmentList implementation.
func (p Product) GetProductTypeByDepartmentList(params product.GetProductTypeByDepartmentListParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_PRODUCT_TYPE_LIST, principal.Roles) {
		return product.NewGetProductTypeByDepartmentListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	productTypes, err := p.dm.ListProductTypes(ctx, mcom.ListProductTypesRequest{
		DepartmentID: params.DepartmentOID,
	})
	if err != nil {
		return utils.ParseError(ctx, product.NewGetProductTypeByDepartmentListDefault(0), err)
	}
	data := make(models.Products, len(productTypes))
	for i, productType := range productTypes {
		data[i] = &models.ProductsItems0{
			Type: productType,
		}
	}
	return product.NewGetProductTypeByDepartmentListOK().WithPayload(&product.GetProductTypeByDepartmentListOKBody{Data: data})
}

// GetProductTypeList implementation.
func (p Product) GetProductTypeList(params product.GetProductTypeListParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_PRODUCT_TYPE_LIST, principal.Roles) {
		return product.NewGetProductTypeListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	var request mcom.ListProductTypesRequest
	if params.DepartmentOid != nil {
		request.DepartmentID = *params.DepartmentOid
	}
	productTypes, err := p.dm.ListProductTypes(ctx, request)
	if err != nil {
		return utils.ParseError(ctx, product.NewGetProductTypeListDefault(0), err)
	}
	data := make(models.Products, len(productTypes))
	for i, pt := range productTypes {
		data[i] = &models.ProductsItems0{
			Type: pt,
		}
	}
	return product.NewGetProductTypeListOK().WithPayload(&product.GetProductTypeListOKBody{Data: data})
}

// GetProductList implementation.
func (p Product) GetProductList(params product.GetProductListParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_PRODUCT_LIST, principal.Roles) {
		return product.NewGetProductListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	req := mcom.ListProductIDsRequest{
		Type: params.ProductType,
	}
	if params.IsLastProcess != nil {
		req.IsLastProcess = *params.IsLastProcess
	}
	productIDs, err := p.dm.ListProductIDs(ctx, req)
	if err != nil {
		return utils.ParseError(ctx, product.NewGetProductListDefault(0), err)
	}
	return product.NewGetProductListOK().WithPayload(&product.GetProductListOKBody{Data: productIDs})
}

// GetMaterialResourceInfoByType implementation.
func (p Product) GetMaterialResourceInfoByType(params product.GetMaterialResourceInfoByTypeParams, principal *models.Principal) middleware.Responder {
	if !p.hasPermission(kenda.FunctionOperationID_GET_MATERIAL_RESOURCE_INFO_BY_TYPE, principal.Roles) {
		return product.NewGetMaterialResourceInfoByTypeDefault(http.StatusForbidden)
	}

	productID, status, date := "", int64(0), types.TimeNano(0)
	if params.ProductID != nil {
		productID = *params.ProductID
	}
	if params.Status != nil {
		status = *params.Status
	}
	if params.StartDate != nil {
		date = types.TimeNano(time.Time(*params.StartDate).UnixNano())
	}
	pageRequest := mcom.PaginationRequest{}
	if params.Page != nil && params.Limit != nil {
		pageRequest = mcom.PaginationRequest{
			PageCount:      uint(*params.Page),
			ObjectsPerPage: uint(*params.Limit),
		}
	}

	orderRequest := parseOrderRequest(params.Body.OrderRequest, getMaterialResourceInfoByTypeDefaultOrderFunc)

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)
	reply, err := p.dm.ListMaterialResources(ctx, mcom.ListMaterialResourcesRequest{
		ProductID:   productID,
		ProductType: params.ProductType,
		Status:      resources.MaterialStatus(status),
		CreatedAt:   date,
	}.WithPagination(pageRequest).
		WithOrder(orderRequest...))

	if err != nil {
		return utils.ParseError(ctx, product.NewGetMaterialResourceInfoByTypeDefault(0), err)
	}

	return product.NewGetMaterialResourceInfoByTypeOK().WithPayload(&product.GetMaterialResourceInfoByTypeOKBody{
		Data: &product.GetMaterialResourceInfoByTypeOKBodyData{
			Items: parseResourceMaterials(reply.Resources),
			Total: reply.AmountOfData,
		},
	})
}

func getMaterialResourceInfoByTypeDefaultOrderFunc() []mcom.Order {
	return []mcom.Order{{
		Name:       "created_at",
		Descending: false,
	}}
}

func parseOrderRequest(dataIn []*product.GetMaterialResourceInfoByTypeParamsBodyOrderRequestItems0, defaultOrderFunc func() []mcom.Order) []mcom.Order {
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

func parseResourceMaterials(replies []mcom.MaterialReply) models.ResourceMaterials {
	resources := make(models.ResourceMaterials, len(replies))
	for i, resourceReply := range replies {
		warehouseID, warehouseLocation := resourceReply.Warehouse.ID, resourceReply.Warehouse.Location
		resources[i] = &models.ResourceMaterial{
			ID:            resourceReply.Material.ID,
			CarrierID:     resourceReply.Material.CarrierID,
			CreatedAt:     strfmt.DateTime(resourceReply.Material.CreatedAt.Time()),
			CreatedBy:     resourceReply.Material.CreatedBy,
			ExpiredDate:   strfmt.DateTime(resourceReply.Material.ExpiryTime),
			Grade:         models.Grade(resourceReply.Material.Grade),
			Inspections:   inspectionStructType(resourceReply.Material.Inspections),
			MinimumDosage: resourceReply.Material.MinDosage.String(),
			Remark:        resourceReply.Material.Remark,
			ProductType:   resourceReply.Material.Type,
			Quantity:      resourceReply.Material.Quantity.String(),
			Unit:          resourceReply.Material.Unit,
			ResourceID:    resourceReply.Material.ResourceID,
			Status:        models.MaterialStatus(resourceReply.Material.Status),
			UpdatedAt:     strfmt.DateTime(resourceReply.Material.UpdatedAt.Time()),
			UpdatedBy:     resourceReply.Material.UpdatedBy,
			Warehouse: &models.Warehouse{
				ID:       &warehouseID,
				Location: &warehouseLocation,
			},
		}
	}
	return resources
}

func inspectionStructType(dataIn mcomModels.Inspections) []*models.ResourceMaterialInspectionsItems0 {

	dataOut := make([]*models.ResourceMaterialInspectionsItems0, len(dataIn))
	for i, d := range dataIn {
		dataOut[i] = &models.ResourceMaterialInspectionsItems0{
			ID:     int64(d.ID),
			Remark: d.Remark,
		}
	}
	return dataOut
}

type substituteBuilder struct {
	context context.Context
	dm      mcom.DataManager

	products []mcomModels.ProductID
}

func newSubstituteBuilder(ctx context.Context, dm mcom.DataManager, products []mcomModels.ProductID) substituteBuilder {
	return substituteBuilder{
		context:  ctx,
		dm:       dm,
		products: products,
	}
}

func (sb substituteBuilder) Build() (map[mcomModels.ProductID][]string, error) {
	if len(sb.products) == 0 {
		return make(map[mcomModels.ProductID][]string), nil
	}

	substitutionsList, err := sb.dm.ListMultipleSubstitutions(sb.context, mcom.ListMultipleSubstitutionsRequest{ProductIDs: sb.products})
	if err != nil {
		return nil, err
	}

	list := make(map[mcomModels.ProductID][]string, len(sb.products))
	for _, productID := range sb.products {
		productSubstitutions := make([]string, len(substitutionsList.Reply[productID].Substitutions))
		for i, substitution := range substitutionsList.Reply[productID].Substitutions {
			productSubstitutions[i] = substitution.ID + substitution.Grade
		}
		list[productID] = productSubstitutions
	}

	return list, nil
}
