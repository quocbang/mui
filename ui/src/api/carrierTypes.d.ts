export interface ICreateCarrierData {
  departmentOID: string
  idPrefix: string
  quantity: number
  allowedMaterial: string
}

export interface IUpdateCarrierData {
  ID: string
  allowedMaterial: string
}
