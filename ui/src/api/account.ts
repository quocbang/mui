import request from '@/utils/request'
import { IAddAccount, IUpdateAccount, IUserRoles } from './accountTypes'

export const defaultUserRoles: IUserRoles = {
  name: '',
  roles: []
}

export const defaultAddAccount: IAddAccount = {
  departmentOID: '',
  employeeID: '',
  roles: []
}

export const defaultUpdateAccount: IUpdateAccount = {
  roles: [],
  resetPassword: false
}

export const getRoleList = () =>
  request({
    url: '/account/role-list',
    method: 'get'
  })

export const getAccountUnauthorizedList = (departmentOID: string) =>
  request({
    url: `/account/unauthorized/department-oid/${departmentOID}`,
    method: 'get'
  })

export const getAccountList = (departmentOID: string) =>
  request({
    url: `/account/authorized/department-oid/${departmentOID}`,
    method: 'get'
  })

export const addAccount = (data: any) =>
  request({
    url: '/account/authorization',
    method: 'post',
    data
  })

export const updateAccount = (employeeID: string, data: any) =>
  request({
    url: `/account/authorization/${employeeID}`,
    method: 'put',
    data
  })

export const deleteAccount = (employeeID: string) =>
  request({
    url: `/account/authorization/${employeeID}`,
    method: 'delete'
  })

export const login = (data: any) =>
  request({
    url: '/user/login',
    method: 'post',
    data
  })

export const logout = () =>
  request({
    url: '/user/logout',
    method: 'post'
  })

export const changePassword = (data: any) =>
  request({
    url: '/user/change-password',
    method: 'put',
    data
  })
