import request from '@/utils/request'

export const getServerStatus = () =>
  request({
    url: '/server/status',
    method: 'get'
  })
