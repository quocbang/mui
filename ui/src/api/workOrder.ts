import request from '@/utils/request'

export const createWorkOrder = (data: any) =>
  request({
    url: '/work-orders',
    method: 'post',
    data
  })

export const updateWorkOrderSequence = (data: any) =>
  request({
    url: '/work-orders',
    method: 'put',
    data
  })

export const updateWorkOrder = (id: string, data: any) =>
  request({
    url: `/work-orders/${id}`,
    method: 'put',
    data
  })

export const getStationScheduleList = (station: string, date: string) =>
  request({
    url: `/schedulings/station/${station}/date/${date}`,
    method: 'get'
  })

export const getWorkOrderList = (stationID: string, params: any) =>
  request({
    url: `/production-flow/work-orders/station/${stationID}`,
    method: 'get',
    params
  })

export const changeWorkOrderStatus = (workOrderID: string, data: any) =>
  request({
    url: `/production-flow/status/work-order/${workOrderID}`,
    method: 'put',
    data
  })

export const getWorkOrderInfo = (workOrderID: string) =>
  request({
    url: `/production-flow/work-order/${workOrderID}/information`,
    method: 'get'
  })

export const getProductionRate = (departmentID: string, params: any) =>
  request({
    url: `/work-orders-rate/department/${departmentID}`,
    method: 'get',
    params
  })
  
export const uploadWorkOrders = (department: string, data: any) =>
  request({
    url: `/work-orders/upload/department/${department}`,
    method: 'post',
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    data: data
  })
