import request from '@/utils/request'
import { IPlanListData, IPlanValue } from './planTypes'

export const defaultPlanListData: IPlanListData = {
  productID: '0',
  dayQuantity: '0',
  weekQuantity: '0',
  stockQuantity: '0',
  conversionDays: '0',
  scheduledQuantity: '0',
  children: [
    {
      productID: '0',
      dayQuantity: '0',
      weekQuantity: '0',
      stockQuantity: '0',
      conversionDays: '0',
      scheduledQuantity: '0'
    },
    {
      productID: '1',
      dayQuantity: '0',
      weekQuantity: '0',
      stockQuantity: '0',
      conversionDays: '0',
      scheduledQuantity: '0'
    }
  ]

}

export const defaultPlanValue: IPlanValue = {
  departmentOID: '',
  date: '',
  productType: '',
  productID: '',
  dayQuantity: ''
}

export const addPlanData = (data: any) =>
  request({
    url: '/plan',
    method: 'post',
    data
  })

export const getPlanList = (departmentOID: string, productType: string, date: string) =>
  request({
    url: `/plans/department-oid/${departmentOID}/product-type/${productType}/date/${date}`,
    method: 'get'
  })
