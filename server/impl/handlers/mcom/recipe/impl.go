package recipe

import (
	"context"
	"net/http"
	"sort"
	"sync"

	"github.com/go-openapi/runtime/middleware"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
	"gitlab.kenda.com.tw/kenda/mcom"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/service"
	"gitlab.kenda.com.tw/kenda/mui/server/impl/utils"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/recipe"
)

// Recipe definitions.
type Recipe struct {
	dm mcom.DataManager

	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool
}

// divided by get recipes by requiredRecipeID(only single reply) or by productID (multiple replies)
type recipesGetter func(ctx context.Context) ([]mcom.GetRecipeReply, error)

type recipeReference struct {
	id          string
	processName string
	processType string
}

// NewRecipe returns Recipe service.
func NewRecipe(
	dm mcom.DataManager,
	hasPermission func(id kenda.FunctionOperationID, roles []models.Role) bool) service.Recipe {
	return Recipe{
		dm:            dm,
		hasPermission: hasPermission,
	}
}

// filterList to store recipe information without duplication
type filterList map[recipeReference][]*mcom.RecipeProcessConfig

type childrenRecipe struct {
	Parent   recipeReference
	RecipeID string
}

type childrenProduct struct {
	Parent    recipeReference
	ProductID string
}

type recipesGetterComposer struct {
	dm mcom.DataManager

	list     filterList
	recipes  []childrenRecipe
	products []childrenProduct
}

func newRecipesGetterComposer(dm mcom.DataManager) *recipesGetterComposer {
	return &recipesGetterComposer{
		dm:       dm,
		list:     make(filterList),
		recipes:  []childrenRecipe{},
		products: []childrenProduct{},
	}
}

// GetRecipeList implementation.
func (r Recipe) GetRecipeList(params recipe.GetRecipeListParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_GET_RECIPE_LIST, principal.Roles) {
		return recipe.NewGetRecipeListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	parentRecipes, err := r.dm.ListRecipesByProduct(ctx, mcom.ListRecipesByProductRequest{
		ProductID: params.ProductID,
	}.WithOrder(
		mcom.Order{
			Name:       "stage",
			Descending: true,
		}, mcom.Order{
			Name:       "major",
			Descending: true,
		}, mcom.Order{
			Name:       "minor",
			Descending: true,
		}))
	if err != nil {
		return utils.ParseError(ctx, recipe.NewGetRecipeListDefault(0), err)
	}

	parentRecipes = filterProcessesOutputProduct(params.ProductID, parentRecipes.Recipes)

	childrenRecipes, err := listChildrenRecipes(ctx, parseChildrenRecipesGetters(r.dm, parentRecipes.Recipes))
	if err != nil {
		return utils.ParseError(ctx, recipe.NewGetRecipeListDefault(0), err)
	}

	return recipe.NewGetRecipeListOK().WithPayload(&recipe.GetRecipeListOKBody{Data: composeProcesses(parentRecipes.Recipes, childrenRecipes)})
}

// GetRecipeIDs implementation.
func (r Recipe) GetRecipeIDs(params recipe.GetRecipeIDsParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_GET_RECIPE_IDS, principal.Roles) {
		return recipe.NewGetRecipeIDsDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	recipeList, err := r.dm.ListRecipesByProduct(ctx, mcom.ListRecipesByProductRequest{
		ProductID: params.ProductID,
	}.WithOrder(
		mcom.Order{
			Name:       "stage",
			Descending: true,
		}, mcom.Order{
			Name:       "major",
			Descending: true,
		}, mcom.Order{
			Name:       "minor",
			Descending: true,
		}))
	if err != nil {
		return utils.ParseError(ctx, recipe.NewGetRecipeIDsDefault(0), err)
	}

	data := make([]string, len(recipeList.Recipes))
	for i, list := range recipeList.Recipes {
		data[i] = list.ID
	}
	return recipe.NewGetRecipeIDsOK().WithPayload(&recipe.GetRecipeIDsOKBody{Data: data})
}

// GetRecipeProcessList implementation.
func (r Recipe) GetRecipeProcessList(params recipe.GetRecipeProcessListParams, principal *models.Principal) middleware.Responder {
	if !r.hasPermission(kenda.FunctionOperationID_GET_RECIPE_PROCESS_LIST, principal.Roles) {
		return recipe.NewGetRecipeProcessListDefault(http.StatusForbidden)
	}

	ctx := commonsCtx.WithUserID(params.HTTPRequest.Context(), principal.ID)

	recipeData, err := r.dm.GetRecipe(ctx, mcom.GetRecipeRequest{
		ID: params.RecipeID,
	})
	if err != nil {
		return utils.ParseError(ctx, recipe.NewGetRecipeProcessListDefault(0), err)
	}

	materialSubstitutions, err := newSubstituteBuilder(ctx, r.dm,
		getProcessesMaterials(getAllProcesses(recipeData.Processes)...)).Build()
	if err != nil {
		return utils.ParseError(ctx, recipe.NewGetRecipeProcessListDefault(0), err)
	}

	data := make([]*recipe.GetRecipeProcessListOKBodyDataItems0, len(recipeData.Processes))
	for i, process := range recipeData.Processes {
		optionalFlows := make([]*models.OptionalFlowsProcess, len(process.OptionalFlows))
		for j, optional := range process.OptionalFlows {
			optionalFlows[j] = &models.OptionalFlowsProcess{
				MaxRepetitions: int64(optional.MaxRepetitions),
				Name:           optional.Name,
				Processes:      parseRecipeProcess(materialSubstitutions, optional.Processes...),
			}
		}

		data[i] = &recipe.GetRecipeProcessListOKBodyDataItems0{
			OptionalFlows: optionalFlows,
			RequiredFlows: parseRecipeProcess(materialSubstitutions, process.Info)[0],
		}
	}

	return recipe.NewGetRecipeProcessListOK().WithPayload(&recipe.GetRecipeProcessListOKBody{Data: data})
}

// filterProcessesOutputProduct filters each recipe's processes' output product must be equal to demand productID.
// TODO: will be deprecated in the future after adjust corresponding backend methods.
func filterProcessesOutputProduct(productID string, recipes []mcom.GetRecipeReply) mcom.ListRecipesByProductReply {
	filtered := make([]mcom.GetRecipeReply, 0)
	for _, recipe := range recipes {
		filteredProcesses := make([]*mcom.ProcessEntity, len(recipe.Processes))
		for j, process := range recipe.Processes {
			filteredProcess := &mcom.ProcessEntity{
				OptionalFlows: make([]*mcom.RecipeOptionalFlowEntity, 0),
			}
			for _, optional := range process.OptionalFlows {
				filteredOptionalProcesses := make([]mcom.ProcessDefinition, 0)
				for _, optionalProcess := range optional.Processes {
					if optionalProcess.Output.ID != productID {
						continue
					}
					filteredOptionalProcesses = append(filteredOptionalProcesses, optionalProcess)
				}
				filteredProcess.OptionalFlows = append(filteredProcess.OptionalFlows, &mcom.RecipeOptionalFlowEntity{
					Name:           optional.Name,
					MaxRepetitions: optional.MaxRepetitions,
					Processes:      filteredOptionalProcesses,
				})
			}
			if process.Info.Output.ID == productID {
				filteredProcess.Info = process.Info
			}
			filteredProcesses[j] = filteredProcess
		}
		filtered = append(filtered, mcom.GetRecipeReply{
			ID: recipe.ID,
			Product: mcom.Product{
				ID:   recipe.Product.ID,
				Type: recipe.Product.Type,
			},
			Version: mcom.RecipeVersion{
				Major: recipe.Version.Major,
				Minor: recipe.Version.Minor,
				Stage: recipe.Version.Stage,
			},
			Processes: filteredProcesses,
		})
	}

	return mcom.ListRecipesByProductReply{
		Recipes: filtered,
	}
}

// parseChildrenRecipesGetters parse parent recipe list into mapping list
func parseChildrenRecipesGetters(
	dm mcom.DataManager,
	parentRecipes []mcom.GetRecipeReply) map[recipeReference][]recipesGetter {
	composer := newRecipesGetterComposer(dm)
	for _, r := range parentRecipes {
		for _, process := range r.Processes {
			key := recipeReference{
				id:          r.ID,
				processName: process.Info.Name,
				processType: process.Info.Type,
			}
			composer.AddIfNotExist(key, process.Info.Configs)

			for _, flow := range process.OptionalFlows {
				for _, flowProcess := range flow.Processes {
					key := recipeReference{
						id:          r.ID,
						processName: flowProcess.Name,
						processType: flowProcess.Type,
					}
					composer.AddIfNotExist(key, flowProcess.Configs)
				}
			}
		}
	}

	return composer.Compose()
}

// listChildrenRecipes return children product recipe list
func listChildrenRecipes(ctx context.Context, childrenRecipeGetters map[recipeReference][]recipesGetter) (map[recipeReference][]mcom.GetRecipeReply, error) {
	result := make(map[recipeReference][]mcom.GetRecipeReply, len(childrenRecipeGetters))
	var mu sync.Mutex

	eg, ctx := errgroup.WithContext(ctx)
	for parent, childrenRecipeGetter := range childrenRecipeGetters {
		for _, get := range childrenRecipeGetter {
			parent, get := parent, get
			eg.Go(func() error {
				childrenRecipe, err := get(ctx)
				if err != nil {
					return err
				}
				mu.Lock()
				result[parent] = append(result[parent], childrenRecipe...)
				mu.Unlock()
				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func composeProcesses(parentRecipes []mcom.GetRecipeReply, children map[recipeReference][]mcom.GetRecipeReply) []*models.RecipeProcesses {
	result := make([]*models.RecipeProcesses, len(parentRecipes))
	for i, recipe := range parentRecipes {
		processes := make([]*models.ProcessInfo, 0)
		for _, p := range recipe.Processes {
			if queryRecipe, ok := children[recipeReference{
				id:          recipe.ID,
				processName: p.Info.Name,
				processType: p.Info.Type,
			}]; ok || noChildren(children) {
				processes = append(processes, &models.ProcessInfo{
					OID:      p.Info.OID,
					Name:     p.Info.Name,
					Type:     p.Info.Type,
					Stations: parseStationInfo(p.Info.Configs, queryRecipe),
				})
			}

			for _, optionalFlow := range p.OptionalFlows {
				for _, process := range optionalFlow.Processes {
					if queryRecipe, ok := children[recipeReference{
						id:          recipe.ID,
						processName: process.Name,
						processType: process.Type,
					}]; ok || noChildren(children) {
						processes = append(processes, &models.ProcessInfo{
							OID:      process.OID,
							Name:     process.Name,
							Type:     process.Type,
							Stations: parseStationInfo(process.Configs, queryRecipe),
						})
					}
				}
			}
		}
		result[i] = &models.RecipeProcesses{
			ID:        recipe.ID,
			Processes: processes,
		}
	}

	return result
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

func getProcessesMaterials(processes ...mcom.ProcessDefinition) []mcomModels.ProductID {
	materials := make(map[mcomModels.ProductID]struct{})
	products := []mcomModels.ProductID{}
	for _, process := range processes {
		for _, config := range process.Configs {
			for _, step := range config.Steps {
				for _, material := range step.Materials {
					product := mcomModels.ProductID{
						ID:    material.Name,
						Grade: material.Grade,
					}
					if _, ok := materials[product]; ok {
						continue
					}
					materials[product] = struct{}{}
					products = append(products, product)
				}
			}
		}
	}
	return products
}

// getAllProcesses return recipe processes, include optional flow processes.
func getAllProcesses(processes []*mcom.ProcessEntity) (results []mcom.ProcessDefinition) {
	for _, process := range processes {
		for _, optional := range process.OptionalFlows {
			results = append(results, optional.Processes...)
		}
		results = append(results, process.Info)
	}
	return
}

func parseRecipeProcess(
	materialSubstitutions map[mcomModels.ProductID][]string,
	processes ...mcom.ProcessDefinition) []*models.AllProcessProductInfo {
	results := make([]*models.AllProcessProductInfo, len(processes))

	for k, process := range processes {
		stationsBom := make([]models.StationBOM, 0)
		stationControlList := parseControlTable(process.Configs)
		for _, config := range process.Configs {
			bom := make([]*models.BomData, 0)
			for _, step := range config.Steps {
				for _, material := range step.Materials {
					var high, mid, low string
					if v := material.Value.High; v != nil {
						high = v.String()
					}
					if v := material.Value.Mid; v != nil {
						mid = v.String()
					}
					if v := material.Value.Low; v != nil {
						low = v.String()
					}
					bom = append(bom, &models.BomData{
						ProductID:        material.Name,
						Grade:            material.Grade,
						MaxValue:         high,
						StandardValue:    mid,
						MinValue:         low,
						RequiredRecipeID: material.RequiredRecipeID,
						Substitutes: materialSubstitutions[mcomModels.ProductID{
							ID:    material.Name,
							Grade: material.Grade,
						}],
					})
				}
			}
			for _, station := range config.Stations {
				stationsBom = append(stationsBom, models.StationBOM{
					ID:      station,
					BomList: bom,
					Control: stationControlList[station],
				})
			}
		}
		results[k] = &models.AllProcessProductInfo{
			Name: process.Name,
			Product: models.AllProcessProductInfoProduct{
				ID:   process.Output.ID,
				Type: process.Output.Type,
			},
			Stations: stationsBom,
			Type:     process.Type,
		}
	}
	return results
}

// addIfNotExist will return false if ref is existed; otherwise return true and insert.
func (c *recipesGetterComposer) AddIfNotExist(ref recipeReference, configs []*mcom.RecipeProcessConfig) bool {
	return c.list.addIfNotExist(ref, configs)
}

// addIfNotExist will return false if ref is existed; otherwise return true and insert.
func (f filterList) addIfNotExist(ref recipeReference, configs []*mcom.RecipeProcessConfig) bool {
	if _, ok := f[ref]; ok {
		return false
	}
	f[ref] = configs
	return true
}

// Compose stores each parent-children recipe data.
func (c *recipesGetterComposer) Compose() map[recipeReference][]recipesGetter {
	for key, value := range c.list {
		c.parseRecipeConfigs(key, value)
	}

	result := make(map[recipeReference][]recipesGetter, len(c.recipes)+len(c.products))

	for _, recipe := range c.recipes {
		result[recipe.Parent] = append(result[recipe.Parent], c.newRecipesGetter(recipe.RecipeID))
	}
	for _, p := range c.products {
		result[p.Parent] = append(result[p.Parent], c.newProductRecipesGetter(p.ProductID))
	}

	return result
}

// parseRecipeConfigs parse recipe configs.
// Each step materials divided by two method to find their recipe list;
// 1. By Recipe ID
// 2. By Product ID
func (c *recipesGetterComposer) parseRecipeConfigs(parent recipeReference, configs []*mcom.RecipeProcessConfig) {
	for _, cfg := range configs {
		for _, step := range cfg.Steps {
			for _, material := range step.Materials {
				if material.RequiredRecipeID != "" {
					c.addRecipe(parent, material.RequiredRecipeID)
				} else {
					c.addProduct(parent, material.Name)
				}
			}
		}
	}
}

// list recipes by recipe ID
func (c *recipesGetterComposer) newRecipesGetter(id string) recipesGetter {
	return func(ctx context.Context) ([]mcom.GetRecipeReply, error) {
		r, err := c.dm.GetRecipe(ctx, mcom.GetRecipeRequest{
			ID: id,
		})
		if err != nil {
			return nil, err
		}
		return []mcom.GetRecipeReply{r}, nil
	}
}

// list recipes by product ID
func (c *recipesGetterComposer) newProductRecipesGetter(productID string) recipesGetter {
	return func(ctx context.Context) ([]mcom.GetRecipeReply, error) {
		dataOut, err := c.dm.ListRecipesByProduct(ctx, mcom.ListRecipesByProductRequest{
			ProductID: productID,
		}.WithOrder(
			mcom.Order{
				Name:       "stage",
				Descending: true,
			}, mcom.Order{
				Name:       "major",
				Descending: true,
			}, mcom.Order{
				Name:       "minor",
				Descending: true,
			}))
		if err != nil {
			return nil, err
		}
		return dataOut.Recipes, nil
	}
}

func (c *recipesGetterComposer) addRecipe(parent recipeReference, id string) {
	c.recipes = append(c.recipes, childrenRecipe{
		Parent:   parent,
		RecipeID: id,
	})
}

func (c *recipesGetterComposer) addProduct(parent recipeReference, productID string) {
	c.products = append(c.products, childrenProduct{
		Parent:    parent,
		ProductID: productID,
	})
}

func noChildren(children map[recipeReference][]mcom.GetRecipeReply) bool {
	return len(children) == 0
}

func parseStationInfo(configs []*mcom.RecipeProcessConfig, recipes []mcom.GetRecipeReply) []*models.StationInfo {
	stations := make([]*models.StationInfo, 0)
	for _, cfg := range configs {
		var batchSize string
		if v := cfg.BatchSize; v != nil {
			batchSize = v.String()
		}

		for _, station := range cfg.Stations {
			stations = append(stations, &models.StationInfo{
				ID:        station,
				BatchSize: batchSize,
				BomList:   parseBomList(recipes, parseBom(cfg.Steps)),
			})
		}
	}
	return stations
}

func parseBomList(recipes []mcom.GetRecipeReply, bom map[string]decimal.Decimal) []*models.StationInfoBomListItems0 {
	bomList := make([]*models.StationInfoBomListItems0, 0)
	for productID, quantity := range bom {
		bomRecipes := make([]*models.StationInfoBomListItems0RecipesItems0, 0)
		list := make(filterList)
		for _, recipe := range recipes {
			for _, process := range recipe.Processes {
				if isEqualProduct(productID, process.Info.Output.ID) {
					ref := recipeReference{
						id:          recipe.ID,
						processName: process.Info.Name,
						processType: process.Info.Type,
					}

					if !list.addIfNotExist(ref, process.Info.Configs) {
						continue
					}

					bomProcesses := make([]*models.StationInfoBomListItems0RecipesItems0ProcessesItems0, 0)
					for _, config := range process.Info.Configs {
						var batchSize string
						if v := config.BatchSize; v != nil {
							batchSize = v.String()
						}

						bomStations := make([]*models.StationInfoBomListItems0RecipesItems0ProcessesItems0StationsItems0, 0)
						for _, childrenStation := range config.Stations {
							bomStations = append(bomStations, &models.StationInfoBomListItems0RecipesItems0ProcessesItems0StationsItems0{
								ID:        childrenStation,
								BatchSize: batchSize,
							})
						}

						bomProcesses = append(bomProcesses, &models.StationInfoBomListItems0RecipesItems0ProcessesItems0{
							OID:      process.Info.OID,
							Name:     process.Info.Name,
							Type:     process.Info.Type,
							Stations: bomStations,
						})
					}
					bomRecipes = append(bomRecipes, &models.StationInfoBomListItems0RecipesItems0{
						ID:        recipe.ID,
						Processes: bomProcesses,
					})
				}
			}
		}
		// filter the product which doesn't have any recipe data
		if len(bomRecipes) > 0 {
			bomList = append(bomList, &models.StationInfoBomListItems0{
				ProductID: productID,
				Quantity:  quantity.String(),
				Recipes:   bomRecipes,
			})
		}
	}

	return bomList
}

// parseBom return each material's quantity list
func parseBom(steps []*mcom.RecipeProcessStep) map[string]decimal.Decimal {
	bom := make(map[string]decimal.Decimal)
	for _, step := range steps {
		for _, mat := range step.Materials {
			if qty := mat.Value.Mid; qty != nil {
				total, ok := bom[mat.Name]
				if !ok {
					total = decimal.Zero
				}
				bom[mat.Name] = total.Add(*qty)
			}
		}
	}
	return bom
}

func isEqualProduct(source, target string) bool {
	return source == target
}

func parseControlTable(configs []*mcom.RecipeProcessConfig) map[string]models.StationBOMControl {
	stationControlList := make(map[string]models.StationBOMControl)

	for _, config := range configs {
		columnBuilder := parseColumnBuilder(config.Steps)
		stationControl := models.StationBOMControl{
			Step: models.Table{
				Columns: parseColumns(*columnBuilder),
				Rows:    make([][]*models.TableRowsItems0, len(config.Steps)),
			},
		}

		for i, step := range config.Steps {
			columnsRowData := columnBuilder.build()
			for _, control := range step.Controls {
				_, value, _ := getControlValue(control.Param)
				columnsRowData.set(control.Name, value)
			}

			stationControl.Step.Rows[i] = columnsRowData.done()
		}

		stationControl.Common = make([]*models.StationBOMControlCommonItems0, len(config.CommonControls))
		for i, control := range config.CommonControls {
			max, _, min := getControlValue(control.Param)
			stationControl.Common[i] = &models.StationBOMControlCommonItems0{
				MaxValue: max,
				MinValue: min,
				RowName:  control.Name,
				Unit:     control.Param.Unit,
			}
		}

		for _, station := range config.Stations {
			stationControlList[station] = stationControl
		}
	}
	return stationControlList
}

func (c *columnsRowData) set(key, value string) {
	i, ok := c.columnIndex[key]
	if !ok {
		zap.L().Warn("unexpected not found column",
			zap.String("column_name", key),
			zap.Any("column_map", c.columnIndex))
	}

	c.values[i] = value
}

func (c *columnsRowData) done() []*models.TableRowsItems0 {
	results := make([]*models.TableRowsItems0, len(c.values))
	for i, value := range c.values {
		results[i] = &models.TableRowsItems0{
			Name:  c.indexColumn[i],
			Value: value,
		}
	}

	return results
}

type columnsRowData struct {
	indexColumn map[int]string
	columnIndex map[string]int
	values      []string
}

func (c *columnBuilder) build() *columnsRowData {
	columnIndex := make(map[string]int, len(c.columns))
	for i, columnName := range c.columns {
		columnIndex[columnName] = i
	}

	indexColumn := make(map[int]string, len(columnIndex))
	for name, index := range columnIndex {
		indexColumn[index] = name
	}

	return &columnsRowData{
		indexColumn: indexColumn,
		columnIndex: columnIndex,
		values:      make([]string, len(columnIndex)),
	}
}

type columnBuilder struct {
	columns     []string
	columnsUnit map[string]string
}

func parseColumnBuilder(steps []*mcom.RecipeProcessStep) *columnBuilder {
	controlsUnit := make(map[string]string) // record total control's name in all step. key: control name, value: control unit
	for _, step := range steps {
		for _, control := range step.Controls {
			controlsUnit[control.Name] = control.Param.Unit
		}
	}

	return newColumnBuilder(controlsUnit)
}

func parseColumns(c columnBuilder) []*models.TableColumnsItems0 {
	results := make([]*models.TableColumnsItems0, len(c.columns))
	for i, column := range c.columns {
		results[i] = &models.TableColumnsItems0{
			Name: column,
			Unit: c.columnsUnit[column],
		}
	}

	return results
}

func getControlValue(param *mcom.RecipePropertyParameter) (max string, value string, min string) {
	if v := param.High; v != nil {
		max = v.String()
	}
	if v := param.Low; v != nil {
		min = v.String()
	}
	if v := param.Mid; v != nil {
		value = v.String()
	}
	return
}

// newColumnBuilder creates new columnBuilder
func newColumnBuilder(columnsUnit map[string]string) *columnBuilder {
	columns := sortColumnName(columnsUnit)

	return &columnBuilder{
		columns:     columns,
		columnsUnit: columnsUnit,
	}
}

// sort control name for consistency data
func sortColumnName(columnsUnit map[string]string) []string {
	columns := make([]string, len(columnsUnit))
	count := 0
	for i := range columnsUnit {
		columns[count] = i
		count++
	}
	sort.Strings(columns)

	return columns
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
