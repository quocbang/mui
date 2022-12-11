import request from '@/utils/request'

export const setStationConfig = (stationID: string, data: any) =>
  request({
    url: `/production-flow/config/station/${stationID}`,
    method: 'post',
    data
  })

export const getStationConfig = (stationID: string) =>
  request({
    url: `/production-flow/config/station/${stationID}`,
    method: 'get'
  })

export const getProductGroupsList = (departmentOID: string, productType: string) =>
  request({
    url: `/product/groups/department-oid/${departmentOID}/product-type/${productType}`,
    method: 'get'
  })
