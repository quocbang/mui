import request from '@/utils/request'
import { ICreateCarrierData, IUpdateCarrierData } from './carrierTypes'

export const defaultICreateCarrierDataInfo: ICreateCarrierData = {
  departmentOID: '',
  idPrefix: '',
  quantity: 1,
  allowedMaterial: ''
}
export const defaultIUpdateCarrierDataInfo: IUpdateCarrierData = {
  ID: '',
  allowedMaterial: ''
}

export const getCarrierList = (departmentOID: string, params: any) =>
  request({
    url: `/carrier/department-oid/${departmentOID}`,
    method: 'get',
    params
  })

export const addCarrier = (data: any) =>
  request({
    url: '/carrier',
    method: 'post',
    data
  })

export const updateCarrier = (ID: string, data: any) =>
  request({
    url: `/carrier/${ID}`,
    method: 'put',
    data
  })

export const deleteCarrier = (ID: string) =>
  request({
    url: `/carrier/${ID}`,
    method: 'delete'
  })

export const getBarcodeCode39 = (data: any) =>
  request({
    url: '/print/barcode/code39',
    method: 'post',
    headers: {
      Accept: 'application/pdf'
    },
    responseType: 'arraybuffer',
    data: data
  })

export const getBarcodeQrcode = (data: any) =>
  request({
    url: '/print/barcode/qrcode',
    method: 'post',
    headers: {
      Accept: 'application/pdf'
    },
    responseType: 'arraybuffer',
    data
  })
