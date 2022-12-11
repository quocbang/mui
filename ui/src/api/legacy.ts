import request from '@/utils/request'
import { IBarcode, IBarcodeDate } from './types'

export const defaultBarcodeInfo: IBarcode = {
  materialBarcode: '',
  barcode: '',
  productID: '',
  grade: '',
  inventory: '',
  status: '',
  expiredAt: '',
  createdBy: '',
  createdAt: 0,
  updateAt: 0
}

export const defaultBarcodeData: IBarcodeDate = {
  newStatus: '',
  holdReason: '',
  extendDays: 0,
  productCate: '',
  controlArea: ''
}

export const getBarcodeInfo = (ID: string) =>
  request({
    url: `/pda/barcode/${ID}`,
    method: 'get'
  })

export const updateBarcodeInfo = (ID: string, data: any) =>
  request({
    url: `/pda/barcode/${ID}`,
    method: 'put',
    data
  })
export const getBarcodeUpdateStatusList = (ID: string) =>
  request({
    url: `/pda/barcode/update-status-list/ID/${ID}`,
    method: 'get'
  })

export const getBarcodeExpiredDate = (ID: string) =>
  request({
    url: `/pda/barcode/extend-expired-date/ID/${ID}`,
    method: 'get'
  })

export const getBarcodeControlArea = () =>
  request({
    url: '/pda/barcode/control-area',
    method: 'get'
  })

export const getBarcodeReasonList = () =>
  request({
    url: '/pda/barcode/reason-list',
    method: 'get'
  })
