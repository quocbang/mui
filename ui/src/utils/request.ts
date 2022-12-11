import axios from 'axios'
import { MessageBox, Notification } from 'element-ui'
import { UserModule } from '@/store/modules/user'
import i18n from '@/lang'

declare global {
  interface Window {
    config: any
  }
}

const service = axios.create({
  baseURL: window.config.ApiUrl,
  withCredentials: false
})
// Request interceptors
service.interceptors.request.use(
  (config) => {
    config.headers['Content-Type'] = 'application/json;charset=UTF-8'
    if (UserModule.token) {
      config.headers['x-mui-auth-key'] = UserModule.token
    }
    return config
  },
  (error) => {
    console.log(error)
    Promise.reject(error)
  }
)

// Response interceptors
service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (response.status !== 200) {
      return Promise.reject(new Error(res.message || 'Error'))
    } else {
      // because somme data need other information, so return response
      if ((response.status === 200 && response.data.length === 0) || response.data.byteLength !== undefined) {
        return response
      } else {
        return response.data
      }
    }
  },
  (error) => {
    if (error.response.status === 401) {
      MessageBox.confirm(
        i18n.t('message.notify005').toString(),
        i18n.t('login.logOut').toString(),
        {
          confirmButtonText: i18n.t('login.logInAgain').toString(),
          cancelButtonText: i18n.t('share.cancel').toString(),
          type: 'warning'
        }
      ).then(() => {
        UserModule.Reset()
        location.reload() // To prevent bugs from vue-router
      })
    } else {
      if (error.response.status !== undefined) {
        const firstCode = error.response.status.toString().substr(0, 1)
        const code = 'errorCode_' + error.response.data.code
        if (firstCode === '4') {
          if (error.response.status === 403) {
            Notification.warning({ title: i18n.t('share.errorMessage').toString() + ': ' + error.response.status, message: i18n.t('errorCodes.errorCode_403').toString() })
          } else if (error.response.status === 404) {
            Notification.warning({ title: i18n.t('share.errorMessage').toString() + ': ' + error.response.status, message: i18n.t('errorCodes.errorCode_404').toString() })
          } else if (error.response.status === 408) {
            Notification.warning({ title: i18n.t('share.errorMessage').toString() + ': ' + error.response.status, message: i18n.t('errorCodes.errorCode_408').toString() })
          } else {
            if (error.response.data.code !== undefined) {
              Notification.warning({ title: i18n.t('share.errorMessage').toString() + ': ' + error.response.data.code, message: i18n.t('errorCodes.' + code).toString() })
            } else {
              Notification.warning({ title: i18n.t('share.errorMessage').toString() + ': ' + error.response.status, message: i18n.t('errorCodes.400').toString() })
            }
          }
        } else if (firstCode === '5') {
          Notification.warning({ title: i18n.t('share.errorMessage').toString() + ': ' + error.response.data.code, message: i18n.t('message.notify008').toString() })
        } else {
          Notification.warning({ title: i18n.t('share.errorMessage').toString() + ': ' + error.response.data.code, message: i18n.t('errorCodes.' + code).toString() })
        }
        return error.response.status
      }
    }

    return Promise.reject(error)
  }
)

export default service
