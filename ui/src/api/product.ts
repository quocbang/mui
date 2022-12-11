import request from '@/utils/request'

export const getProductTypeList = (departmentOID: string) =>
  request({
    url: `/product/active-product-types/department-oid/${departmentOID}`,
    method: 'get'
  })

export const queryProductTypeList = () =>
  request({
    url: '/product/active-product-types',
    method: 'get'
  })

export const getProductList = (productType: string) =>
  request({
    url: `/product/active-products/product-type/${productType}`,
    method: 'get'
  })

export const getFinalProductList = (productType: string) =>
  request({
    url: `/product/active-products/product-type/${productType}` + '?is_last_process=true',
    method: 'get'
  })

export const getMaterialInfo = (productType: string, params: any) =>
  request({
    url: `/resource/material/info/product-type/${productType}`,
    method: 'get',
    params
  })
