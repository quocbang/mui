import request from '@/utils/request'

export const getAllDepartment = () =>
  request({
    url: '/departments',
    method: 'get'
  })
