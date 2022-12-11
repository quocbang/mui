import request from '@/utils/request'
import { IMaterialSplit, IMaterialCreateData, IMaterialQueryInfo, IResourceInfo, IResourceOtherInfo, ISiteBindResource } from './resourceTypes'

export const defaultMaterialCreateInfo: IMaterialCreateData = {
  warehouse: {
    ID: '',
    location: ''
  },
  resource: {
    productType: '',
    productID: '',
    grade: '',
    quantity: '',
    unit: '',
    lotNumber: '',
    productionTime: '',
    expiryTime: ''
  }
}

export const defaultIResourceOtherInfo: IResourceOtherInfo = {
  stationID: '',
  barcodeForBarrelSlot: '',
  siteName: '',
  siteIndex: 0,
  productType: '',
  productID: '',
  materialBarcode: ''
}
export const defaultIResourceInfo: IResourceInfo = {
  ID: '',
  grade: '',
  quantity: '',
  status: '',
  expiredDate: '',
  productType: ''
}

export const defaultISiteBindResource: ISiteBindResource = {
  station: '',
  bindType: 0,
  siteName: '',
  siteIndex: 0,
  resources: [
    {
      resourceID: '',
      productType: '',
      quantity: ''
    }
  ],
  queueOption: {
    head: false,
    tail: false,
    index: 0
  }
}

export const defaultMaterialSplit: IMaterialSplit = {
  productType: '',
  resourceID: '',
  quantity: 0,
  splitQuantity: 0,
  remark: '',
  inspections: []
}

export const defaultMaterialQueryInfo: IMaterialQueryInfo = {
  productType: '',
  ID: '',
  grade: '',
  quantity: '',
  unit: '',
  status: 0,
  expiredDate: '',
  resourceID: '',
  carrierID: '',
  warehouse: {
    ID: '',
    location: ''
  },
  minimumDosage: '',
  inspections: {
    ID: 0,
    remark: ''
  },
  remark: '',
  createdBy: '',
  createdAt: '',
  updatedBy: '',
  updatedAt: ''
}

export const addMaterial = (data: any) =>
  request({
    url: '/resource/material/stock',
    method: 'post',
    data
  })

export const getResourceInfo = (ID: string) =>
  request({
    url: `/resource/material/info/resource-id/${ID}`,
    method: 'get'
  })

export const getMaterialStatus = () =>
  request({
    url: '/resource/material/status',
    method: 'get'
  })

export const materialSplit = (data: any) =>
  request({
    url: '/resource/material/split',
    method: 'post',
    data
  })

export const materialBarcode = (data: any) =>
  request({
    url: '/print/resource/material',
    method: 'post',
    headers: {
      Accept: 'application/pdf'
    },
    responseType: 'arraybuffer',
    data: data
  })

export const getResourceToolInfo = (toolResourceID: any) =>
  request({
    url: `/production-flow/tool-resource/${toolResourceID}`,
    method: 'get'
  })

export const printMaterialResource = (data: any) =>
  request({
    url: '/production-flow/print/material-resource',
    method: 'post',
    data
  })

export const preMaterialResourceBarcode = (workOrderID: string, data: any) =>
  request({
    url: `/print/work-orders/${workOrderID}/pre-material-resource`,
    method: 'post',
    headers: {
      Accept: 'application/pdf'
    },
    responseType: 'arraybuffer',
    data
  })
