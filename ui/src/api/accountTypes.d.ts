export interface IUserRoles {
    name: string
    roles: number[]
}

export interface IAddAccount {
    departmentOID: string
    employeeID: string
    roles: number[]
}

export interface IUpdateAccount {
    roles: number[]
    resetPassword: boolean
}
