export interface IResources {
  resourceID: string
  productType: string
  quantity: string
}

export interface IQueueOption {
  head: boolean
  tail: boolean
  index: number
}
export interface IResourceOtherInfo {
  stationID: string
  barcodeForBarrelSlot: string
  siteName: string
  siteIndex: number
  productType: string
  productID: string
  materialBarcode: string
}

export interface IResourceInfo {
  ID: string
  grade: string
  quantity: string
  status: string
  expiredDate: string
  productType: string
}

export interface IBindResourceInfo {
  resourceID: string
  productID: string
  grade: string
  quantity: string
}

export interface ISiteBindResource {
  station: string
  bindType: number
  siteName: string
  siteIndex: number
  resources: IResources[]
  queueOption: IQueueOption
}

export interface IMaterialCreateWarehouse {
  ID: string
  location: string
}
export interface IMaterialCreateResource {
  productType: string
  productID: string
  grade: string
  quantity: string
  unit: string
  lotNumber: string
  productionTime: string
  expiryTime: string
}

export interface IMaterialCreateData {
  warehouse: IMaterialCreateWarehouse
  resource: IMaterialCreateResource
}

export interface IMaterialSplit {
  productType: string
  resourceID: string
  quantity: number
  splitQuantity: number
  remark: string
  inspections: any[]
}

export interface IMaterialQueryInfo {
  productType: string
  ID: string
  grade: string
  quantity: string
  unit: string
  status: number
  expiredDate: string
  resourceID: string
  carrierID: string
  warehouse: {
    ID: string
    location: string
  }
  minimumDosage: string
  inspections: {
    ID: number
    remark: string
  }
  remark: string
  createdBy: string
  createdAt: string
  updatedBy: string
  updatedAt: string
}
