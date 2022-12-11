import request from '@/utils/request'

export const getSubTypeList = () =>
  request({
    url: '/site/sub-type-list',
    method: 'get'
  })

export const getTypeList = () =>
  request({
    url: '/site/type-list',
    method: 'get'
  })

export const getSiteBindResource = (station: string, siteName: string, siteIndex: number) =>
  request({
    url: `/site/material/station/${station}/site-name/${siteName}/site-index/${siteIndex}`,
    method: 'get'
  })

export const updateBindResource = (data: any) =>
  request({
    url: '/site/resources/bind/auto',
    method: 'post',
    data
  })

export const getSiteInfo = (data: any) =>
  request({
    url: '/production-flow/site/information',
    method: 'post',
    data
  })

export const getStationOperator = (stationID: string, data: any) =>
  request({
    url: `/station/${stationID}/operator`,
    method: 'post',
    data
  })
