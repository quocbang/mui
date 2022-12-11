export interface ISites {
  index: number[]
  name: string
  type: number
}

export interface IBarcode {
  materialBarcode: string
  barcode: string
  productID: string
  grade: string
  inventory: string
  status: string
  expiredAt: string
  createdBy: string
  createdAt: number
  updateAt: number
}
export interface IBarcodeDate {
  newStatus: string
  holdReason: string
  extendDays: number
  productCate: string
  controlArea: string
}

export interface ICodeStatus {
  code: string
  description: string
}
