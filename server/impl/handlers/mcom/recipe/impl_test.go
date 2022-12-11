package recipe

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErrors "gitlab.kenda.com.tw/kenda/mcom/errors"
	mcomModels "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/mock"
	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"

	"gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/account"
	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/restapi/operations/recipe"
)

const (
	userID = "tester"

	testProductTypeA = "TYPE-A"
	testProductTypeB = "TYPE-B"
	testProductTypeC = "TYPE-C"
	testProductTypeD = "TYPE-D"
	testProductTypeE = "TYPE-E"

	testProduct1Parent    = "PRODUCT-A"
	testProduct1Children1 = "PRODUCT-B"
	testProduct1Children2 = "PRODUCT-C"
	testProduct1Children3 = "PRODUCT-D"
	testProduct2Parent    = "PRODUCT-W"
	testProduct2Children1 = "PRODUCT-X"
	testProduct2Children2 = "PRODUCT-Y"
	testProduct2Children3 = "PRODUCT-Z"

	testRecipeA = "Recipe-A"
	testRecipeB = "Recipe-B"

	testProcessAOID  = "PROCESS001OID"
	testProcessA     = "PROCESS-A"
	testProcessAType = "PROCESS"

	testProcessAOptionalProcessOID = "OPTIONAL-PROCESS001OID"
	testProcessAOptional           = "OPTIONAL-PROCESS-A"
	testProcessAOptionalType       = "PROCESS"

	testMaterialAID           = "MATERIAL-A"
	testMaterialAGrade        = "X"
	testMaterialAType         = "NATURAL_RUBBER"
	testMaterialARecipeID     = "CUTA"
	testMaterialAProcessAOID  = "MATAPROCESS001"
	testMaterialAProcessA     = "MATAPROCESS-A"
	testMaterialAProcessAType = "PROCESS"

	testMaterialBID           = "MATERIAL-B"
	testMaterialBGrade        = ""
	testMaterialBType         = "COMPOUND_INGREDIENTS"
	testMaterialBRecipeID     = "CMP001"
	testMaterialBProcessAOID  = "MATBPROCESS001"
	testMaterialBProcessA     = "MATBPROCESS-A"
	testMaterialBProcessAType = "PROCESS"

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
	testStationA = "STATION-A"
	testStationB = "STATION-B"

	testBatchSizeDecimal = types.Decimal.NewFromInt16(10)

	testMaterialAValue    = types.Decimal.NewFromInt16(10)
	testMaterialAMaxValue = types.Decimal.NewFromInt16(15)
	testMaterialAMinValue = types.Decimal.NewFromInt16(5)

	testMaterialBValue    = types.Decimal.NewFromInt16(25)
	testMaterialBMaxValue = types.Decimal.NewFromInt16(50)
	testMaterialBMinValue = types.Decimal.NewFromInt16(0)

	testControlTemperature = types.Decimal.NewFromInt16(180)
	testHighOilTemperature = types.Decimal.NewFromInt16(270)
	testLowOilTemperature  = types.Decimal.NewFromInt16(100)
)

func TestRecipe_GetRecipeList(t *testing.T) {
	assert := assert.New(t)

	var (
		getParentRecipeScripts = []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testProduct1Parent,
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
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: testRecipeA,
								Product: mcom.Product{
									ID:   testProduct1Parent,
									Type: testProductTypeA,
								},
								Version: mcom.RecipeVersion{},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  testProcessAOID,
											Name: testProcessA,
											Type: testProcessAType,
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{testStationA, testStationB},
													BatchSize: testBatchSizeDecimal,
													Unit:      "",
													Steps: []*mcom.RecipeProcessStep{
														{
															Materials: []*mcom.RecipeMaterial{
																{
																	Name:  testMaterialAID,
																	Grade: testMaterialAGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialAValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: "",
																},
																{
																	Name:  testMaterialBID,
																	Grade: testMaterialBGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialBValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: testMaterialBRecipeID,
																},
															},
														},
													},
												},
											},
											Output: mcom.OutputProduct{
												ID:   testProduct1Parent,
												Type: testProductTypeA,
											},
										},
										OptionalFlows: []*mcom.RecipeOptionalFlowEntity{
											{
												Name: "optional1",
												Processes: []mcom.ProcessDefinition{
													{
														OID:  testProcessAOptionalProcessOID,
														Name: testProcessAOptional,
														Type: testProcessAOptionalType,
														Configs: []*mcom.RecipeProcessConfig{
															{
																Stations:  []string{testStationA},
																BatchSize: testBatchSizeDecimal,
																Unit:      "",
																Steps: []*mcom.RecipeProcessStep{
																	{
																		Materials: []*mcom.RecipeMaterial{
																			{
																				Name:  testMaterialAID,
																				Grade: testMaterialAGrade,
																				Value: mcom.RecipeMaterialParameter{
																					High: nil,
																					Mid:  testMaterialAValue,
																					Low:  nil,
																					Unit: "",
																				},
																				Site:             "",
																				RequiredRecipeID: "",
																			},
																		},
																	},
																},
															},
														},
														Output: mcom.OutputProduct{
															ID:   testProduct1Parent,
															Type: testProductTypeA,
														},
													},
												},
												MaxRepetitions: 2,
											},
										},
									},
									{ // repeated process
										Info: mcom.ProcessDefinition{
											OID:  testProcessAOID,
											Name: testProcessA,
											Type: testProcessAType,
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{testStationA, testStationB},
													BatchSize: testBatchSizeDecimal,
													Unit:      "",
													Steps: []*mcom.RecipeProcessStep{
														{
															Materials: []*mcom.RecipeMaterial{
																{
																	Name:  testMaterialAID,
																	Grade: testMaterialAGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialAValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: "",
																},
																{
																	Name:  testMaterialBID,
																	Grade: testMaterialBGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialBValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: testMaterialBRecipeID,
																},
															},
														},
													},
												},
											},
											Output: mcom.OutputProduct{
												ID:   testProduct1Parent,
												Type: testProductTypeA,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
	)

	httpRequestWithHeader := httptest.NewRequest("GET", "/product/active-recipes/product-type/{productType}/product-id/{productID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	{ // list recipe error
		scripts := []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testProduct1Parent,
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
						}),
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_PRODUCT_ID_NOT_FOUND,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)
		params := recipe.GetRecipeListParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeList(params, principal).(*recipe.GetRecipeListDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_PRODUCT_ID_NOT_FOUND),
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // list recipe internal server error
		scripts := []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testProduct1Parent,
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
						}),
				},
				Output: mock.Output{
					Error: errors.New(testInternalServerError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)
		params := recipe.GetRecipeListParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeList(params, principal).(*recipe.GetRecipeListDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // list recipe normal
		monkey.PatchInstanceMethod(reflect.TypeOf(new(recipesGetterComposer)), "Compose", func(composer *recipesGetterComposer) map[recipeReference][]recipesGetter {
			g1 := func(ctx context.Context) ([]mcom.GetRecipeReply, error) {
				return []mcom.GetRecipeReply{
					{
						ID: testMaterialARecipeID,
						Product: mcom.Product{
							ID:   testMaterialAID,
							Type: testMaterialAType,
						},
						Version: mcom.RecipeVersion{},
						Processes: []*mcom.ProcessEntity{
							{
								Info: mcom.ProcessDefinition{
									OID:  testMaterialAProcessAOID,
									Name: testMaterialAProcessA,
									Type: testMaterialAProcessAType,
									Configs: []*mcom.RecipeProcessConfig{
										{
											Stations:  []string{testStationA},
											BatchSize: testBatchSizeDecimal,
											Unit:      "",
											Steps:     []*mcom.RecipeProcessStep{},
										},
									},
									Output: mcom.OutputProduct{
										ID:   testMaterialAID,
										Type: testMaterialAType,
									},
								},
							},
						},
					},
				}, nil
			}
			return map[recipeReference][]recipesGetter{
				{
					id:          testRecipeA,
					processName: testProcessA,
					processType: testProcessAType,
				}: {g1}}
		})
		monkey.Patch(composeProcesses, func(parentRecipes []mcom.GetRecipeReply, children map[recipeReference][]mcom.GetRecipeReply) []*models.RecipeProcesses {
			return nil
		})
		dm, err := mock.New(getParentRecipeScripts)
		assert.NoError(err)
		params := recipe.GetRecipeListParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		_, ok := p.GetRecipeList(params, principal).(*recipe.GetRecipeListOK)
		assert.True(ok)
		assert.NoError(dm.Close())
		monkey.UnpatchAll()
	}
	{ // recipe not found
		monkey.PatchInstanceMethod(reflect.TypeOf(new(recipesGetterComposer)), "Compose", func(composer *recipesGetterComposer) map[recipeReference][]recipesGetter {
			return map[recipeReference][]recipesGetter{
				{
					id:          testRecipeA,
					processName: testProcessA,
					processType: testProcessAType,
				}: {
					func(ctx context.Context) ([]mcom.GetRecipeReply, error) {
						return nil, mcomErrors.Error{
							Code: mcomErrors.Code_RECIPE_NOT_FOUND,
						}
					},
				},
			}
		})
		dm, err := mock.New(getParentRecipeScripts)
		assert.NoError(err)
		params := recipe.GetRecipeListParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeList(params, principal).(*recipe.GetRecipeListDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_RECIPE_NOT_FOUND),
			}), reply)
		}
		assert.NoError(dm.Close())
		monkey.UnpatchAll()
	}
	{ // internal error
		monkey.PatchInstanceMethod(reflect.TypeOf(new(recipesGetterComposer)), "Compose", func(composer *recipesGetterComposer) map[recipeReference][]recipesGetter {
			return map[recipeReference][]recipesGetter{
				{
					id:          testRecipeA,
					processName: testProcessA,
					processType: testProcessAType,
				}: {
					func(ctx context.Context) ([]mcom.GetRecipeReply, error) {
						return nil, errors.New(testInternalServerError)
					},
				},
			}
		})
		dm, err := mock.New(getParentRecipeScripts)
		assert.NoError(err)
		params := recipe.GetRecipeListParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeList(params, principal).(*recipe.GetRecipeListDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), reply)
		}
		assert.NoError(dm.Close())
		monkey.UnpatchAll()
	}
	{ // forbidden access
		dm, err := mock.New([]mock.Script{})
		assert.NoError(err)
		defer dm.Close()
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetRecipeList(recipe.GetRecipeListParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}, principal).(*recipe.GetRecipeListDefault)
		assert.True(ok)
		assert.Equal(recipe.NewGetRecipeListDefault(http.StatusForbidden), rep)
	}
}

func TestRecipe_GetRecipeIDsList(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("GET", "/product/recipe-id/product-type/{productType}/product-id/{productID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	{ // normal case
		scripts := []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testProduct1Parent,
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
						}),
				},
				Output: mock.Output{
					Response: mcom.ListRecipesByProductReply{
						Recipes: []mcom.GetRecipeReply{
							{
								ID: testRecipeA,
								Product: mcom.Product{
									ID:   testProduct1Parent,
									Type: testProductTypeA,
								},
								Version: mcom.RecipeVersion{},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  testProcessAOID,
											Name: testProcessA,
											Type: testProcessAType,
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{testStationA, testStationB},
													BatchSize: testBatchSizeDecimal,
													Unit:      "",
													Steps: []*mcom.RecipeProcessStep{
														{
															Materials: []*mcom.RecipeMaterial{
																{
																	Name:  testMaterialAID,
																	Grade: testMaterialAGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialAValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: "",
																},
																{
																	Name:  testMaterialBID,
																	Grade: testMaterialBGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialBValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: testMaterialBRecipeID,
																},
															},
														},
													},
												},
											},
											Output: mcom.OutputProduct{
												ID:   testProduct1Parent,
												Type: testProductTypeA,
											},
										},
										OptionalFlows: []*mcom.RecipeOptionalFlowEntity{
											{
												Name: "optional1",
												Processes: []mcom.ProcessDefinition{
													{
														OID:  testProcessAOptionalProcessOID,
														Name: testProcessAOptional,
														Type: testProcessAOptionalType,
														Configs: []*mcom.RecipeProcessConfig{
															{
																Stations:  []string{testStationA},
																BatchSize: testBatchSizeDecimal,
																Unit:      "",
																Steps: []*mcom.RecipeProcessStep{
																	{
																		Materials: []*mcom.RecipeMaterial{
																			{
																				Name:  testMaterialAID,
																				Grade: testMaterialAGrade,
																				Value: mcom.RecipeMaterialParameter{
																					High: nil,
																					Mid:  testMaterialAValue,
																					Low:  nil,
																					Unit: "",
																				},
																				Site:             "",
																				RequiredRecipeID: "",
																			},
																		},
																	},
																},
															},
														},
														Output: mcom.OutputProduct{
															ID:   testProduct1Parent,
															Type: testProductTypeA,
														},
													},
												},
												MaxRepetitions: 2,
											},
										},
									},
									{ // repeated process
										Info: mcom.ProcessDefinition{
											OID:  testProcessAOID,
											Name: testProcessA,
											Type: testProcessAType,
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{testStationA, testStationB},
													BatchSize: testBatchSizeDecimal,
													Unit:      "",
													Steps: []*mcom.RecipeProcessStep{
														{
															Materials: []*mcom.RecipeMaterial{
																{
																	Name:  testMaterialAID,
																	Grade: testMaterialAGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialAValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: "",
																},
																{
																	Name:  testMaterialBID,
																	Grade: testMaterialBGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialBValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: testMaterialBRecipeID,
																},
															},
														},
													},
												},
											},
											Output: mcom.OutputProduct{
												ID:   testProduct1Parent,
												Type: testProductTypeA,
											},
										},
									},
								},
							},
							{
								ID: testRecipeB,
								Product: mcom.Product{
									ID:   testProduct2Parent,
									Type: testProductTypeB,
								},
								Version: mcom.RecipeVersion{},
								Processes: []*mcom.ProcessEntity{
									{
										Info: mcom.ProcessDefinition{
											OID:  testProcessAOID,
											Name: testProcessA,
											Type: testProcessAType,
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{testStationA, testStationB},
													BatchSize: testBatchSizeDecimal,
													Unit:      "",
													Steps: []*mcom.RecipeProcessStep{
														{
															Materials: []*mcom.RecipeMaterial{
																{
																	Name:  testMaterialAID,
																	Grade: testMaterialAGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialAValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: "",
																},
																{
																	Name:  testMaterialBID,
																	Grade: testMaterialBGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialBValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: testMaterialBRecipeID,
																},
															},
														},
													},
												},
											},
											Output: mcom.OutputProduct{
												ID:   testProduct1Parent,
												Type: testProductTypeA,
											},
										},
										OptionalFlows: []*mcom.RecipeOptionalFlowEntity{
											{
												Name: "optional1",
												Processes: []mcom.ProcessDefinition{
													{
														OID:  testProcessAOptionalProcessOID,
														Name: testProcessAOptional,
														Type: testProcessAOptionalType,
														Configs: []*mcom.RecipeProcessConfig{
															{
																Stations:  []string{testStationA},
																BatchSize: testBatchSizeDecimal,
																Unit:      "",
																Steps: []*mcom.RecipeProcessStep{
																	{
																		Materials: []*mcom.RecipeMaterial{
																			{
																				Name:  testMaterialAID,
																				Grade: testMaterialAGrade,
																				Value: mcom.RecipeMaterialParameter{
																					High: nil,
																					Mid:  testMaterialAValue,
																					Low:  nil,
																					Unit: "",
																				},
																				Site:             "",
																				RequiredRecipeID: "",
																			},
																		},
																	},
																},
															},
														},
														Output: mcom.OutputProduct{
															ID:   testProduct1Parent,
															Type: testProductTypeA,
														},
													},
												},
												MaxRepetitions: 2,
											},
										},
									},
									{ // repeated process
										Info: mcom.ProcessDefinition{
											OID:  testProcessAOID,
											Name: testProcessA,
											Type: testProcessAType,
											Configs: []*mcom.RecipeProcessConfig{
												{
													Stations:  []string{testStationA, testStationB},
													BatchSize: testBatchSizeDecimal,
													Unit:      "",
													Steps: []*mcom.RecipeProcessStep{
														{
															Materials: []*mcom.RecipeMaterial{
																{
																	Name:  testMaterialAID,
																	Grade: testMaterialAGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialAValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: "",
																},
																{
																	Name:  testMaterialBID,
																	Grade: testMaterialBGrade,
																	Value: mcom.RecipeMaterialParameter{
																		High: nil,
																		Mid:  testMaterialBValue,
																		Low:  nil,
																		Unit: "",
																	},
																	Site:             "",
																	RequiredRecipeID: testMaterialBRecipeID,
																},
															},
														},
													},
												},
											},
											Output: mcom.OutputProduct{
												ID:   testProduct1Parent,
												Type: testProductTypeA,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)
		params := recipe.GetRecipeIDsParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeIDs(params, principal).(*recipe.GetRecipeIDsOK)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeIDsOK().WithPayload(&recipe.GetRecipeIDsOKBody{Data: []string{testRecipeA, testRecipeB}}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // bad request
		scripts := []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: "",
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
						}),
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code:    mcomErrors.Code_PRODUCT_ID_NOT_FOUND,
						Details: "empty product id",
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)
		params := recipe.GetRecipeIDsParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   "",
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeIDs(params, principal).(*recipe.GetRecipeIDsDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeIDsDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_PRODUCT_ID_NOT_FOUND),
				Details: "empty product id",
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // internal error
		scripts := []mock.Script{
			{
				Name: mock.FuncListRecipesByProduct,
				Input: mock.Input{
					Request: mcom.ListRecipesByProductRequest{
						ProductID: testProduct1Parent,
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
						}),
				},
				Output: mock.Output{
					Error: errors.New(testInternalServerError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)
		params := recipe.GetRecipeIDsParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeIDs(params, principal).(*recipe.GetRecipeIDsDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeIDsDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // forbidden access
		dm, err := mock.New([]mock.Script{})
		assert.NoError(err)
		defer dm.Close()
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetRecipeIDs(recipe.GetRecipeIDsParams{
			HTTPRequest: httpRequestWithHeader,
			ProductID:   testProduct1Parent,
		}, principal).(*recipe.GetRecipeIDsDefault)
		assert.True(ok)
		assert.Equal(recipe.NewGetRecipeIDsDefault(http.StatusForbidden), rep)
	}
}

func TestRecipe_GetRecipeProcessList(t *testing.T) {
	assert := assert.New(t)

	httpRequestWithHeader := httptest.NewRequest("GET", "/product/recipe-process/recipe-id/{recipeID}", nil)
	httpRequestWithHeader.Header.Set(account.AuthorizationKey, "token-for-tester")

	materials1 := []*mcom.RecipeMaterial{
		{
			Name:  testMaterialAID,
			Grade: testMaterialAGrade,
			Value: mcom.RecipeMaterialParameter{
				High: testMaterialAMaxValue,
				Mid:  testMaterialAValue,
				Low:  testMaterialAMinValue,
				Unit: "",
			},
			Site:             "",
			RequiredRecipeID: "",
		},
		{
			Name:  testMaterialBID,
			Grade: testMaterialBGrade,
			Value: mcom.RecipeMaterialParameter{
				High: testMaterialBMaxValue,
				Mid:  testMaterialBValue,
				Low:  testMaterialBMinValue,
				Unit: "",
			},
			Site:             "",
			RequiredRecipeID: testMaterialBRecipeID,
		},
	}
	materials1Opts := []*mcom.RecipeMaterial{
		{
			Name:  testMaterialAID,
			Grade: testMaterialAGrade,
			Value: mcom.RecipeMaterialParameter{
				High: testMaterialAMaxValue,
				Mid:  testMaterialAValue,
				Low:  testMaterialAMinValue,
				Unit: "",
			},
			Site:             "",
			RequiredRecipeID: "",
		},
	}
	materials2 := []*mcom.RecipeMaterial{
		{
			Name:  testMaterialAID,
			Grade: testMaterialAGrade,
			Value: mcom.RecipeMaterialParameter{
				High: testMaterialAMaxValue,
				Mid:  testMaterialAValue,
				Low:  testMaterialAMinValue,
				Unit: "",
			},
			Site:             "",
			RequiredRecipeID: "",
		},
		{
			Name:  testMaterialBID,
			Grade: testMaterialBGrade,
			Value: mcom.RecipeMaterialParameter{
				High: testMaterialBMaxValue,
				Mid:  testMaterialBValue,
				Low:  testMaterialBMinValue,
				Unit: "",
			},
			Site:             "",
			RequiredRecipeID: testMaterialBRecipeID,
		},
	}
	{ // normal case
		scripts := []mock.Script{
			{
				Name: mock.FuncGetRecipe,
				Input: mock.Input{
					Request: mcom.GetRecipeRequest{
						ID: testRecipeA,
					},
				},
				Output: mock.Output{
					Response: mcom.GetRecipeReply{
						ID: testRecipeA,
						Product: mcom.Product{
							ID:   testProduct1Parent,
							Type: testProductTypeA,
						},
						Version: mcom.RecipeVersion{},
						Processes: []*mcom.ProcessEntity{
							{
								Info: mcom.ProcessDefinition{
									OID:  testProcessAOID,
									Name: testProcessA,
									Type: testProcessAType,
									Configs: []*mcom.RecipeProcessConfig{
										{
											Stations:  []string{testStationA, testStationB},
											BatchSize: testBatchSizeDecimal,
											Unit:      "",
											Steps: []*mcom.RecipeProcessStep{
												{
													Materials: materials1,
													Controls: []*mcom.RecipeProperty{
														{
															Name: "TEMPERATURE",
															Param: &mcom.RecipePropertyParameter{
																High: nil,
																Mid:  testControlTemperature,
																Low:  nil,
																Unit: "Celcius",
															},
														},
													},
												},
											},
											CommonControls: []*mcom.RecipeProperty{
												{
													Name: "OIL_TEMPERATURE",
													Param: &mcom.RecipePropertyParameter{
														High: testHighOilTemperature,
														Mid:  nil,
														Low:  testLowOilTemperature,
														Unit: "Celcius",
													},
												},
											},
										},
									},
									Output: mcom.OutputProduct{
										ID:   testProduct1Parent,
										Type: testProductTypeA,
									},
								},
								OptionalFlows: []*mcom.RecipeOptionalFlowEntity{
									{
										Name: "optional1",
										Processes: []mcom.ProcessDefinition{
											{
												OID:  testProcessAOptionalProcessOID,
												Name: testProcessAOptional,
												Type: testProcessAOptionalType,
												Configs: []*mcom.RecipeProcessConfig{
													{
														Stations:  []string{testStationA},
														BatchSize: testBatchSizeDecimal,
														Unit:      "",
														Steps: []*mcom.RecipeProcessStep{
															{
																Materials: materials1Opts,
															},
														},
													},
												},
												Output: mcom.OutputProduct{
													ID:   testProduct1Parent,
													Type: testProductTypeA,
												},
											},
										},
										MaxRepetitions: 2,
									},
								},
							},
							{
								Info: mcom.ProcessDefinition{
									OID:  testMaterialBProcessAOID,
									Name: testMaterialBProcessA,
									Type: testMaterialBProcessAType,
									Configs: []*mcom.RecipeProcessConfig{
										{
											Stations:  []string{testStationB},
											BatchSize: testBatchSizeDecimal,
											Unit:      "",
											Steps: []*mcom.RecipeProcessStep{
												{
													Materials: materials2,
													Controls: []*mcom.RecipeProperty{
														{
															Name: "TEMPERATURE",
															Param: &mcom.RecipePropertyParameter{
																High: nil,
																Mid:  testControlTemperature,
																Low:  nil,
																Unit: "Celcius",
															},
														},
													},
												},
											},
										},
									},
									Output: mcom.OutputProduct{
										ID:   testMaterialBID,
										Type: testMaterialBType,
									},
								},
							},
						},
					},
				},
			},
			{
				Name: mock.FuncListMultipleSubstitutions,
				Input: mock.Input{
					Request: mcom.ListMultipleSubstitutionsRequest{
						ProductIDs: []mcomModels.ProductID{
							{
								ID:    testMaterialAID,
								Grade: testMaterialAGrade,
							},
							{
								ID:    testMaterialBID,
								Grade: testMaterialBGrade,
							},
						},
					},
				},
				Output: mock.Output{
					Response: mcom.ListMultipleSubstitutionsReply{
						Reply: map[mcomModels.ProductID]mcom.ListSubstitutionsReply{
							{
								ID:    testMaterialAID,
								Grade: testMaterialAGrade,
							}: {},
							{
								ID:    testMaterialBID,
								Grade: testMaterialBGrade,
							}: {
								Substitutions: []mcomModels.Substitution{
									{
										ID:         testMaterialAID,
										Grade:      testMaterialAGrade,
										Proportion: decimal.Decimal{},
									},
								},
							},
						},
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		params := recipe.GetRecipeProcessListParams{
			HTTPRequest: httpRequestWithHeader,
			RecipeID:    testRecipeA,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeProcessList(params, principal).(*recipe.GetRecipeProcessListOK)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeProcessListOK().WithPayload(&recipe.GetRecipeProcessListOKBody{
				Data: []*recipe.GetRecipeProcessListOKBodyDataItems0{
					{
						OptionalFlows: []*models.OptionalFlowsProcess{
							{
								MaxRepetitions: 2,
								Name:           "optional1",
								Processes: []*models.AllProcessProductInfo{
									{
										Name: testProcessAOptional,
										Product: models.AllProcessProductInfoProduct{
											ID:   testProduct1Parent,
											Type: testProductTypeA,
										},
										Stations: []models.StationBOM{
											{
												ID: testStationA,
												BomList: []*models.BomData{
													{
														Substitutes:      []string{},
														ProductID:        testMaterialAID,
														Grade:            testMaterialAGrade,
														MaxValue:         testMaterialAMaxValue.String(),
														StandardValue:    testMaterialAValue.String(),
														MinValue:         testMaterialAMinValue.String(),
														RequiredRecipeID: "",
													},
												},
												Control: models.StationBOMControl{
													Common: []*models.StationBOMControlCommonItems0{},
													Step: models.Table{
														Columns: []*models.TableColumnsItems0{},
														Rows: [][]*models.TableRowsItems0{
															{},
														},
													},
												},
											},
										},
										Type: testProcessAOptionalType,
									},
								},
							},
						},
						RequiredFlows: &models.AllProcessProductInfo{
							Name: testProcessA,
							Product: models.AllProcessProductInfoProduct{
								ID:   testProduct1Parent,
								Type: testProductTypeA,
							},
							Stations: []models.StationBOM{
								{
									ID: testStationA,
									BomList: []*models.BomData{
										{
											Substitutes:      []string{},
											ProductID:        testMaterialAID,
											Grade:            testMaterialAGrade,
											MaxValue:         testMaterialAMaxValue.String(),
											StandardValue:    testMaterialAValue.String(),
											MinValue:         testMaterialAMinValue.String(),
											RequiredRecipeID: "",
										},
										{
											Substitutes:      []string{testMaterialAID + testMaterialAGrade},
											ProductID:        testMaterialBID,
											Grade:            testMaterialBGrade,
											MaxValue:         testMaterialBMaxValue.String(),
											StandardValue:    testMaterialBValue.String(),
											MinValue:         testMaterialBMinValue.String(),
											RequiredRecipeID: testMaterialBRecipeID,
										},
									},
									Control: models.StationBOMControl{
										Common: []*models.StationBOMControlCommonItems0{
											{
												MaxValue: testHighOilTemperature.String(),
												MinValue: testLowOilTemperature.String(),
												RowName:  "OIL_TEMPERATURE",
												Unit:     "Celcius",
											},
										},
										Step: models.Table{
											Columns: []*models.TableColumnsItems0{
												{
													Name: "TEMPERATURE",
													Unit: "Celcius",
												},
											},
											Rows: [][]*models.TableRowsItems0{
												{
													{
														Name:  "TEMPERATURE",
														Value: "180",
													},
												},
											},
										},
									},
								},
								{
									ID: testStationB,
									BomList: []*models.BomData{
										{
											Substitutes:      []string{},
											ProductID:        testMaterialAID,
											Grade:            testMaterialAGrade,
											MaxValue:         testMaterialAMaxValue.String(),
											StandardValue:    testMaterialAValue.String(),
											MinValue:         testMaterialAMinValue.String(),
											RequiredRecipeID: "",
										},
										{
											Substitutes:      []string{testMaterialAID + testMaterialAGrade},
											ProductID:        testMaterialBID,
											Grade:            testMaterialBGrade,
											MaxValue:         testMaterialBMaxValue.String(),
											StandardValue:    testMaterialBValue.String(),
											MinValue:         testMaterialBMinValue.String(),
											RequiredRecipeID: testMaterialBRecipeID,
										},
									},
									Control: models.StationBOMControl{
										Common: []*models.StationBOMControlCommonItems0{
											{
												MaxValue: testHighOilTemperature.String(),
												MinValue: testLowOilTemperature.String(),
												RowName:  "OIL_TEMPERATURE",
												Unit:     "Celcius",
											},
										},
										Step: models.Table{
											Columns: []*models.TableColumnsItems0{
												{
													Name: "TEMPERATURE",
													Unit: "Celcius",
												},
											},
											Rows: [][]*models.TableRowsItems0{
												{
													{
														Name:  "TEMPERATURE",
														Value: "180",
													},
												},
											},
										},
									},
								},
							},
							Type: testProcessAType,
						},
					},
					{
						OptionalFlows: []*models.OptionalFlowsProcess{},
						RequiredFlows: &models.AllProcessProductInfo{
							Name: testMaterialBProcessA,
							Product: models.AllProcessProductInfoProduct{
								ID:   testMaterialBID,
								Type: testMaterialBType,
							},
							Stations: []models.StationBOM{
								{
									ID: testStationB,
									BomList: []*models.BomData{
										{
											Substitutes:      []string{},
											ProductID:        testMaterialAID,
											Grade:            testMaterialAGrade,
											MaxValue:         testMaterialAMaxValue.String(),
											StandardValue:    testMaterialAValue.String(),
											MinValue:         testMaterialAMinValue.String(),
											RequiredRecipeID: "",
										},
										{
											Substitutes:      []string{testMaterialAID + testMaterialAGrade},
											ProductID:        testMaterialBID,
											Grade:            testMaterialBGrade,
											MaxValue:         testMaterialBMaxValue.String(),
											StandardValue:    testMaterialBValue.String(),
											MinValue:         testMaterialBMinValue.String(),
											RequiredRecipeID: testMaterialBRecipeID,
										},
									},
									Control: models.StationBOMControl{
										Common: []*models.StationBOMControlCommonItems0{},
										Step: models.Table{
											Columns: []*models.TableColumnsItems0{
												{
													Name: "TEMPERATURE",
													Unit: "Celcius",
												},
											},
											Rows: [][]*models.TableRowsItems0{
												{
													{
														Name:  "TEMPERATURE",
														Value: "180",
													},
												},
											},
										},
									},
								},
							},
							Type: testMaterialBProcessAType,
						},
					},
				},
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // insufficient request on list substitutes
		scripts := []mock.Script{
			{
				Name: mock.FuncGetRecipe,
				Input: mock.Input{
					Request: mcom.GetRecipeRequest{
						ID: testRecipeA,
					},
				},
				Output: mock.Output{
					Response: mcom.GetRecipeReply{
						ID: testRecipeA,
						Product: mcom.Product{
							ID:   testProduct1Parent,
							Type: testProductTypeA,
						},
						Version: mcom.RecipeVersion{},
						Processes: []*mcom.ProcessEntity{
							{
								Info: mcom.ProcessDefinition{
									OID:  testProcessAOID,
									Name: testProcessA,
									Type: testProcessAType,
									Configs: []*mcom.RecipeProcessConfig{
										{
											Stations:  []string{testStationA, testStationB},
											BatchSize: testBatchSizeDecimal,
											Unit:      "",
											Steps: []*mcom.RecipeProcessStep{
												{Materials: materials1},
											},
										},
									},
									Output: mcom.OutputProduct{
										ID:   testProduct1Parent,
										Type: testProductTypeA,
									},
								},
								OptionalFlows: []*mcom.RecipeOptionalFlowEntity{
									{
										Name: "optional1",
										Processes: []mcom.ProcessDefinition{
											{
												OID:  testProcessAOptionalProcessOID,
												Name: testProcessAOptional,
												Type: testProcessAOptionalType,
												Configs: []*mcom.RecipeProcessConfig{
													{
														Stations:  []string{testStationA},
														BatchSize: testBatchSizeDecimal,
														Unit:      "",
														Steps: []*mcom.RecipeProcessStep{
															{Materials: materials1Opts},
														},
													},
												},
												Output: mcom.OutputProduct{
													ID:   testProduct1Parent,
													Type: testProductTypeA,
												},
											},
										},
										MaxRepetitions: 2,
									},
								},
							},
							{
								Info: mcom.ProcessDefinition{
									OID:  testMaterialBProcessAOID,
									Name: testMaterialBProcessA,
									Type: testMaterialBProcessAType,
									Configs: []*mcom.RecipeProcessConfig{
										{
											Stations:  []string{testStationB},
											BatchSize: testBatchSizeDecimal,
											Unit:      "",
											Steps: []*mcom.RecipeProcessStep{
												{Materials: materials2},
											},
										},
									},
									Output: mcom.OutputProduct{
										ID:   testMaterialBID,
										Type: testMaterialBType,
									},
								},
							},
						},
					},
				},
			},
			{
				Name: mock.FuncListMultipleSubstitutions,
				Input: mock.Input{
					Request: mcom.ListMultipleSubstitutionsRequest{
						ProductIDs: []mcomModels.ProductID{
							{
								ID:    testMaterialAID,
								Grade: testMaterialAGrade,
							},
							{
								ID:    testMaterialBID,
								Grade: testMaterialBGrade,
							},
						},
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code: mcomErrors.Code_INSUFFICIENT_REQUEST,
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		params := recipe.GetRecipeProcessListParams{
			HTTPRequest: httpRequestWithHeader,
			RecipeID:    testRecipeA,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeProcessList(params, principal).(*recipe.GetRecipeProcessListDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeProcessListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code: int64(mcomErrors.Code_INSUFFICIENT_REQUEST),
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // internal error on list substitutes
		scripts := []mock.Script{
			{
				Name: mock.FuncGetRecipe,
				Input: mock.Input{
					Request: mcom.GetRecipeRequest{
						ID: testRecipeA,
					},
				},
				Output: mock.Output{
					Response: mcom.GetRecipeReply{
						ID: testRecipeA,
						Product: mcom.Product{
							ID:   testProduct1Parent,
							Type: testProductTypeA,
						},
						Version: mcom.RecipeVersion{},
						Processes: []*mcom.ProcessEntity{
							{
								Info: mcom.ProcessDefinition{
									OID:  testProcessAOID,
									Name: testProcessA,
									Type: testProcessAType,
									Configs: []*mcom.RecipeProcessConfig{
										{
											Stations:  []string{testStationA, testStationB},
											BatchSize: testBatchSizeDecimal,
											Unit:      "",
											Steps: []*mcom.RecipeProcessStep{
												{Materials: materials1},
											},
										},
									},
									Output: mcom.OutputProduct{
										ID:   testProduct1Parent,
										Type: testProductTypeA,
									},
								},
								OptionalFlows: []*mcom.RecipeOptionalFlowEntity{
									{
										Name: "optional1",
										Processes: []mcom.ProcessDefinition{
											{
												OID:  testProcessAOptionalProcessOID,
												Name: testProcessAOptional,
												Type: testProcessAOptionalType,
												Configs: []*mcom.RecipeProcessConfig{
													{
														Stations:  []string{testStationA},
														BatchSize: testBatchSizeDecimal,
														Unit:      "",
														Steps: []*mcom.RecipeProcessStep{
															{Materials: materials1Opts},
														},
													},
												},
												Output: mcom.OutputProduct{
													ID:   testProduct1Parent,
													Type: testProductTypeA,
												},
											},
										},
										MaxRepetitions: 2,
									},
								},
							},
							{
								Info: mcom.ProcessDefinition{
									OID:  testMaterialBProcessAOID,
									Name: testMaterialBProcessA,
									Type: testMaterialBProcessAType,
									Configs: []*mcom.RecipeProcessConfig{
										{
											Stations:  []string{testStationB},
											BatchSize: testBatchSizeDecimal,
											Unit:      "",
											Steps: []*mcom.RecipeProcessStep{
												{Materials: materials2},
											},
										},
									},
									Output: mcom.OutputProduct{
										ID:   testMaterialBID,
										Type: testMaterialBType,
									},
								},
							},
						},
					},
				},
			},
			{
				Name: mock.FuncListMultipleSubstitutions,
				Input: mock.Input{
					Request: mcom.ListMultipleSubstitutionsRequest{
						ProductIDs: []mcomModels.ProductID{
							{
								ID:    testMaterialAID,
								Grade: testMaterialAGrade,
							},
							{
								ID:    testMaterialBID,
								Grade: testMaterialBGrade,
							},
						},
					},
				},
				Output: mock.Output{
					Error: errors.New(testInternalServerError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)

		params := recipe.GetRecipeProcessListParams{
			HTTPRequest: httpRequestWithHeader,
			RecipeID:    testRecipeA,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeProcessList(params, principal).(*recipe.GetRecipeProcessListDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeProcessListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // bad request
		scripts := []mock.Script{
			{
				Name: mock.FuncGetRecipe,
				Input: mock.Input{
					Request: mcom.GetRecipeRequest{
						ID: "",
					},
				},
				Output: mock.Output{
					Error: mcomErrors.Error{
						Code:    mcomErrors.Code_RECIPE_NOT_FOUND,
						Details: "empty recipe id",
					},
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)
		params := recipe.GetRecipeProcessListParams{
			HTTPRequest: httpRequestWithHeader,
			RecipeID:    "",
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeProcessList(params, principal).(*recipe.GetRecipeProcessListDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeProcessListDefault(http.StatusBadRequest).WithPayload(&models.Error{
				Code:    int64(mcomErrors.Code_RECIPE_NOT_FOUND),
				Details: "empty recipe id",
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // internal error
		scripts := []mock.Script{
			{
				Name: mock.FuncGetRecipe,
				Input: mock.Input{
					Request: mcom.GetRecipeRequest{
						ID: testRecipeA,
					},
				},
				Output: mock.Output{
					Error: errors.New(testInternalServerError),
				},
			},
		}
		dm, err := mock.New(scripts)
		assert.NoError(err)
		params := recipe.GetRecipeProcessListParams{
			HTTPRequest: httpRequestWithHeader,
			RecipeID:    testRecipeA,
		}
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return true
		})
		reply, ok := p.GetRecipeProcessList(params, principal).(*recipe.GetRecipeProcessListDefault)
		if assert.True(ok) {
			assert.Equal(recipe.NewGetRecipeProcessListDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Details: testInternalServerError,
			}), reply)
		}
		assert.NoError(dm.Close())
	}
	{ // forbidden access
		dm, err := mock.New([]mock.Script{})
		assert.NoError(err)
		defer dm.Close()
		p := NewRecipe(dm, func(id kenda.FunctionOperationID, roles []models.Role) bool {
			return false
		})
		rep, ok := p.GetRecipeProcessList(recipe.GetRecipeProcessListParams{
			HTTPRequest: httpRequestWithHeader,
			RecipeID:    testRecipeA,
		}, principal).(*recipe.GetRecipeProcessListDefault)
		assert.True(ok)
		assert.Equal(recipe.NewGetRecipeProcessListDefault(http.StatusForbidden), rep)
	}
}
