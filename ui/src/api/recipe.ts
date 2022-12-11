import request from '@/utils/request'

export const getProductRecipesList = (productID: string) =>
  request({
    url: `/product/active-recipes/product-id/${encodeURIComponent(productID)}`,
    method: 'get'
  })

export const getProductRecipesProcessList = (recipeID: string) =>
  request({
    url: `/product/recipe-process/recipe-id/${encodeURIComponent(recipeID)}`,
    method: 'get'
  })

export const getProductRecipeIDList = (productID: string) =>
  request({
    url: `/product/recipe-id/product-id/${encodeURIComponent(productID)}`,
    method: 'get'
  })
