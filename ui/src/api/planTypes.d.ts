export interface IPlanData {
    productID: string
    dayQuantity: string
    weekQuantity: string
    stockQuantity: string
    conversionDays: string
    scheduledQuantity: string
}

export interface IPlanListData {
    productID: string
    dayQuantity: string
    weekQuantity: string
    stockQuantity: string
    conversionDays: string
    scheduledQuantity: string
    children: IPlanData[]
}

export interface IPlanValue {
    departmentOID: string
    date: string
    productType: string
    productID: string
    dayQuantity: string
}
