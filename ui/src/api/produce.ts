import request from '@/utils/request'

export const feed = (stationID: string, data: any) =>
  request({
    url: `/mes/feed/station/${stationID}`,
    method: 'post',
    data
  })

export const collect = (stationID: string, data: any) =>
  request({
    url: `/mes/collect/station/${stationID}`,
    method: 'post',
    data
  })
