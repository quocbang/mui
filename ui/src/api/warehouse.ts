import request from '@/utils/request'
import { IWarehouseTransactionData } from './warehouseTypes'

export const defaultMaterialCreateInfo: IWarehouseTransactionData = {
  ID: '',
  newWarehouseID: '',
  newLocation: ''
}

export const getResourceLocation = (ID: string) =>
  request({
    url: `/warehouse/resource/${ID}`,
    method: 'get'
  })

export const updateResourceWarehouse = (ID: string, data: any) =>
  request({
    url: `/warehouse/resource/${ID}`,
    method: 'put',
    data
  })
