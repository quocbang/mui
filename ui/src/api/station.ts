import request from '@/utils/request'
import { IAddBomWorkOrderData, IAddStationMaintenanceData, IAddWorkOrderData, IStationScheduleListData, IUpdateStationMaintenanceData } from './stationTypes'

export const defaultStationScheduleListData: IStationScheduleListData = {
  ID: '',
  departmentOID: '',
  productID: '',
  recipeID: '',
  station: '',
  batchesQuantity: [],
  planDate: '',
  sequence: 0,
  status: 0,
  statusName: '',
  updateBy: '',
  updateAt: ''
}

export const defaultAddWorkOrderData: IAddWorkOrderData = {
  departmentOID: '',
  recipe: {
    processOID: '',
    processName: '',
    processType: '',
    ID: ''
  },
  station: '',
  batchesQuantity: [],
  planDate: '',
  productID: '',
  batchCalculation: false,
  batchSize: '',
  preBatchSize: '',
  batchCount: 1,
  quantity: 0,
  preWorkOrder: false
}

export const defaultAddBomWorkOrderData: IAddBomWorkOrderData = {
  productID: '',
  departmentOID: '',
  processOID: '',
  recipeID: '',
  station: '',
  batchesQuantity: [],
  planDate: ''
}

export const defaultAddStationMaintenanceData: IAddStationMaintenanceData = {
  ID: '',
  departmentOID: '',
  code: '',
  description: '',
  sites: [
    {
      actionMode: 1,
      name: '',
      index: 0,
      type: 1,
      subType: 1
    }
  ]
}

export const defaultUpdateStationMaintenanceData: IUpdateStationMaintenanceData = {
  code: '',
  description: '',
  state: 0,
  sites: [
    {
      actionMode: 0,
      name: '',
      index: 0,
      type: 1,
      subType: 1
    }
  ]
}

export const getStationMaintenance = (departmentOID: string, params: any) =>
  request({
    url: `/station/maintenance/department-oid/${departmentOID}`,
    method: 'get',
    params
  })

export const addStationMaintenance = (data: any) =>
  request({
    url: '/station/maintenance',
    method: 'post',
    data
  })

export const updateStationMaintenance = (ID: string, data: any) =>
  request({
    url: `/station/maintenance/${ID}`,
    method: 'patch',
    data
  })

export const deleteStationMaintenance = (ID: string) =>
  request({
    url: `/station/maintenance/${ID}`,
    method: 'delete'
  })

export const getStationState = () =>
  request({
    url: '/station/state',
    method: 'get'
  })

export const getDepartmentStationList = (departmentOID: string) =>
  request({
    url: `/station-list/department-oid/${departmentOID}`,
    method: 'get'
  })

export const getStationList = (params: any) =>
  request({
    url: '/production-flow/station',
    method: 'get',
    params
  })

export const getSiteInfoFromStation = (stationID: string) =>
  request({
    url: `/production-flow/site/station/${stationID}`,
    method: 'get'
  })

export const stationSignIn = (stationID: string, data: any) =>
  request({
    url: `/station/${stationID}/sign-in`,
    method: 'post',
    data
  })

export const stationSignOut = (data: any) =>
  request({
    url: '/stations/sign-out',
    method: 'post',
    data
  })
