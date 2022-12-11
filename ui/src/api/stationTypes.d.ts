export interface IStationScheduleData {
    departmentOID: string
    productType: string
    productID: string
    recipeID: string
    station: string
    batchesQuantity: string[]
    planDate: string
}

export interface IUpdateStationSchedule {
    ID: string
    forceToAbort: boolean
    sequence: number
}

export interface IStationScheduleListData {
    ID: string
    departmentOID: string
    productID: string
    recipeID: string
    station: string
    batchesQuantity: string[]
    planDate: string
    sequence: number
    status: number
    statusName: string // UI need
    updateBy: string
    updateAt: string
}

export interface IRecipeData {
    processOID: string
    processName: string
    processType: string
    ID: string
}
export interface IAddWorkOrderData {
    productID: string
    departmentOID: string
    recipe: IRecipeData
    station: string
    batchesQuantity: string[]
    planDate: string
    batchCalculation: boolean
    batchSize: string
    preBatchSize: string
    batchCount: number
    quantity: number
    preWorkOrder: boolean
}
export interface IAddBomWorkOrderData {
    productID: string
    departmentOID: string
    processOID: string
    recipeID: string
    station: string
    batchesQuantity: string[]
    planDate: string
}
export interface ISitesInfo {
    actionMode: number
    name: string
    index: number
    type: number
    subType: number
}
export interface IAddStationMaintenanceData {
    ID: string
    departmentOID: string
    code: string
    description: string
    sites: ISitesInfo[]
}

export interface IUpdateStationMaintenanceData {
    code: string
    description: string
    state: number
    sites: ISitesInfo[]
}
