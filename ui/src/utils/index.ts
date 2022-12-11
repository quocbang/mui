import i18n from '@/lang'
import { UserModule } from '@/store/modules/user'
// Parse the time to string
export const parseTime = (
  time?: object | string | number | null,
  cFormat?: string
): string | null => {
  if (time === undefined || !time) {
    return null
  }
  const format = cFormat || '{y}-{m}-{d} {h}:{i}:{s}'
  let date: Date
  if (typeof time === 'object') {
    date = time as Date
  } else {
    if (typeof time === 'string') {
      if (/^[0-9]+$/.test(time)) {
        // support "1548221490638"
        time = parseInt(time)
      } else {
        // support safari
        // https://stackoverflow.com/questions/4310953/invalid-date-in-safari
        time = time.replace(new RegExp(/-/gm), '/')
      }
    }
    if (typeof time === 'number' && time.toString().length === 10) {
      time = time * 1000
    }
    date = new Date(time)
  }
  const formatObj: { [key: string]: number } = {
    y: date.getFullYear(),
    m: date.getMonth() + 1,
    d: date.getDate(),
    h: date.getHours(),
    i: date.getMinutes(),
    s: date.getSeconds(),
    a: date.getDay()
  }
  const timeStr = format.replace(/{([ymdhisa])+}/g, (result, key) => {
    const value = formatObj[key]
    // Note: getDay() returns 0 on Sunday
    if (key === 'a') {
      return ['日', '一', '二', '三', '四', '五', '六'][value]
    }
    return value.toString().padStart(2, '0')
  })
  return timeStr
}

// Format and filter json data using filterKeys array
export const formatJson = (filterKeys: any, jsonData: any) =>
  jsonData.map((data: any) => filterKeys.map((key: string) => {
    if (key === 'timestamp') {
      return parseTime(data[key])
    } else {
      return data[key]
    }
  }))

// Check if an element has a class
export const hasClass = (ele: HTMLElement, className: string) => {
  return !!ele.className.match(new RegExp('(\\s|^)' + className + '(\\s|$)'))
}

// Add class to element
export const addClass = (ele: HTMLElement, className: string) => {
  if (!hasClass(ele, className)) ele.className += ' ' + className
}

// Remove class from element
export const removeClass = (ele: HTMLElement, className: string) => {
  if (hasClass(ele, className)) {
    const reg = new RegExp('(\\s|^)' + className + '(\\s|$)')
    ele.className = ele.className.replace(reg, ' ')
  }
}

// Toggle class for the selected element
export const toggleClass = (ele: HTMLElement, className: string) => {
  if (!ele || !className) {
    return
  }
  let classString = ele.className
  const nameIndex = classString.indexOf(className)
  if (nameIndex === -1) {
    classString += '' + className
  } else {
    classString =
      classString.substr(0, nameIndex) +
      classString.substr(nameIndex + className.length)
  }
  ele.className = classString
}

export const GetDate = (addDayCount: number) => {
  const dd = new Date()
  dd.setDate(dd.getDate() + addDayCount)
  const y = dd.getFullYear()
  const m = ('0' + (dd.getMonth() + 1)).substr(-2)
  const d = ('0' + (dd.getDate())).substr(-2)
  return y + '-' + m + '-' + d
}

export const validateRequire = (rule: any, value: string, callback: Function) => {
  const required = (i18n.t('share.required')).toString()
  if (value !== undefined) {
    if (value === '' || value === null || value.length === 0) {
      callback(new Error(required))
    } else {
      callback()
    }
  }
}

export const validateNumber = (rule: any, value: string, callback: Function) => {
  const re = /[0-9]/
  const required = (i18n.t('share.requiredNumber')).toString()
  if (!re.test(value)) {
    callback(new Error(required))
  } else {
    callback()
  }
}

export const clearObjValue = (obj: any) => {
  Object.keys(obj).forEach(key => {
    if (typeof obj[key] === 'object') {
      clearObjValue(obj[key])
    } else if (typeof obj[key] === 'number') {
      obj[key] = 0
    } else {
      obj[key] = ''
    }
  })
}

export const getUserDepartmentsInfo = () => {
  const authorizedDepartmentsString = JSON.parse(UserModule.authorizedDepartments.toString())
  const departmentList = authorizedDepartmentsString.map((el: any, index: any) => {
    return {
      value: index,
      label: el
    }
  }
  )
  return departmentList
}

export const paginationData = (originData: any, listQuery: any) => {
  const rangeData = originData.slice((listQuery.page - 1) * listQuery.pageSize, listQuery.page * listQuery.pageSize)
  return rangeData
}

export const validatePasswordNumberRange = (rule: any, value: string, callback: Function) => {
  const msg = (i18n.t('message.notify113', { min: '6', max: '10' })).toString()
  const len = value.length
  if (len < 6 || len > 10) {
    callback(new Error(msg))
  } else {
    callback()
  }
}

export const validateCarrierPrefix = (rule: any, value: string, callback: Function) => {
  const re = /[A-Z]/
  const msgNumber = (i18n.t('message.notifyCarrier001')).toString()
  const msgCapital = (i18n.t('message.notifyCarrier002')).toString()
  if (value.length !== 2) {
    callback(new Error(msgNumber))
  } else if (!re.test(value)) {
    callback(new Error(msgCapital))
  } else {
    callback()
  }
}

export const validateNumberAndUppercase = (rule: any, value: string, callback: Function) => {
  const re = /^[A-Z0-9]*$/
  const required = (i18n.t('message.notifyValidate100')).toString()
  if (!re.test(value)) {
    callback(new Error(required))
  } else {
    callback()
  }
}

export enum BatchSize {
  PerBatchQuantities = 0,
  FixedQuantity = 1,
  PlanQuantity = 2
}
