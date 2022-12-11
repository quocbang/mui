import { Component, Vue } from 'vue-property-decorator'
import { validateRequire } from '@/utils'
import { getResourceInfo, getResourceToolInfo } from '@/api/resource'
import { getSiteBindResource, updateBindResource, getSiteInfo } from '@/api/site'
import { UserModule } from '@/store/modules/user'
import { Form, Input } from 'element-ui'
import _, { cloneDeep } from 'lodash'
import { PDAModule } from '@/store/modules/pda'

@Component({
  name: 'mount'
})

export default class extends Vue {
  private stationID = UserModule.station
  private siteBindInfo: any[] = []
  private siteType = ''
  private siteSubType = 0

  private tempMountInfo = {
    resourceID: '',
    siteLocation: ''
  }

  private bindTypeList = [
    {
      type: 'CONTAINER',
      code: 1001
    },
    {
      type: 'SLOT',
      code: 1101
    },
    {
      type: 'COLLECTION',
      code: 1201
    },
    {
      type: 'QUEUE',
      code: 2101
    },
    {
      type: 'COLQUEUE',
      code: 2201
    }
  ]

  private clearTypeList = [
    {
      type: 'CONTAINER',
      code: 1010
    },
    {
      type: 'SLOT',
      code: 1110
    },
    {
      type: 'COLLECTION',
      code: 1210
    },
    {
      type: 'QUEUE',
      code: 2110
    },
    {
      type: 'COLQUEUE',
      code: 2210
    }
  ]

  private tempBindInfo = {
    station: this.stationID,
    bindType: 0,
    siteName: '',
    siteIndex: 0,
    workOrderID: '',
    resources: [{
      resourceID: '',
      productType: ''
    }],
    forceBind: {
      force: false
    },
    resourceType: 0
  }

  private tempInfo: any

  private rules = {
    siteLocation: [{ validator: validateRequire }],
    resourceID: [{ validator: validateRequire }]
  }

  private getAllWorkOrderInfo: any
  private subTypeMaterials = 2
  private subTypeTools = 3

  mounted() {
    (this.$refs.siteLocation as Input).focus()
    if (typeof PDAModule.workOrderInfo !== 'object') {
      this.getAllWorkOrderInfo = JSON.parse(PDAModule.workOrderInfo.toString())
    } else {
      this.getAllWorkOrderInfo = PDAModule.workOrderInfo
    }
  }

  private async getResourceInfo() {
    if (this.tempMountInfo.resourceID !== '') {
      if (this.siteSubType === this.subTypeMaterials) {
        const { data } = await getResourceInfo(this.tempMountInfo.resourceID)
        if (data === undefined) {
          this.tempBindInfo.forceBind.force = true
          this.tempInfo = cloneDeep(this.tempBindInfo)
          this.tempBindInfo.resources[0].resourceID = this.tempMountInfo.resourceID
          this.tempInfo = _.omit(this.tempBindInfo, ['resourceType'])
        }
        if (data.length > 1) {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notifyPDA002').toString(),
            type: 'warning',
            duration: 2000
          })
        } else {
          this.tempBindInfo.resources[0].resourceID = data[0].resourceID
          this.tempBindInfo.resources[0].productType = data[0].productType
          this.tempBindInfo.forceBind.force = true
          this.tempBindInfo.resourceType = 0
          this.tempInfo = cloneDeep(this.tempBindInfo)
        }
      } else if (this.siteSubType === this.subTypeTools) {
        await getResourceToolInfo(this.tempMountInfo.resourceID)
        this.tempBindInfo.resources[0].resourceID = this.tempMountInfo.resourceID
        this.tempBindInfo.resourceType = 1
        this.tempInfo = cloneDeep(this.tempBindInfo)
        this.tempInfo.resources[0] = _.omit(this.tempInfo.resources[0], ['productType'])
      } else {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: this.$t('message.notifyPDA002').toString(),
          type: 'warning',
          duration: 2000
        })
      }
    }
  }

  private async getSiteInfo() {
    if (this.stationID === '') {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyPDA022').toString(),
        type: 'warning',
        duration: 2000
      })
    } else {
      if (this.tempMountInfo.siteLocation !== '') {
        const request = {
          site: {
            stationID: this.stationID,
            siteName: this.tempMountInfo.siteLocation,
            siteIndex: 0
          }
        }
        const { data } = await getSiteInfo(request)
        if (data.subType !== this.subTypeMaterials && data.subType !== this.subTypeTools) {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notifyPDA001').toString(),
            type: 'warning',
            duration: 2000
          })
          this.tempMountInfo.siteLocation = ''
        } else {
          this.siteType = data.type
          this.siteSubType = data.subType
          this.getSiteBindData()
        }
      }
    }
  }

  private async getSiteBindData() {
    const { data } = await getSiteBindResource(this.stationID, this.tempMountInfo.siteLocation, 0)
    this.siteBindInfo = data
  }

  private async bind() {
    (this.$refs.mountInfoForm as Form).validate(async valid => {
      if (valid) {
        this.tempInfo.siteName = this.tempMountInfo.siteLocation
        this.tempInfo.bindType = this.bindTypeList.filter(item => item.type === this.siteType)[0].code
        this.tempInfo.workOrderID = this.getAllWorkOrderInfo.workOrderID
        if (this.siteType === 'QUEUE' || this.siteType === 'COLQUEUE') {
          const queueData: any = cloneDeep(this.tempInfo)
          queueData.queueOption = {
            head: true
          }
          const data = await updateBindResource(queueData)
          if (data.status === 200) {
            this.$notify({
              title: (this.$t('share.success')).toString(),
              message: (this.$t('share.updateSuccessfully')).toString(),
              type: 'success',
              duration: 2000
            })
          }
        } else {
          const data = await updateBindResource(this.tempInfo)
          if (data.status === 200) {
            this.$notify({
              title: (this.$t('share.success')).toString(),
              message: (this.$t('share.updateSuccessfully')).toString(),
              type: 'success',
              duration: 2000
            })
          }
        }
        this.getSiteBindData()
      }
    })
  }

  private async clear() {
    if (this.tempMountInfo.siteLocation !== '') {
      this.tempInfo = cloneDeep(this.tempBindInfo)
      this.tempInfo.siteName = this.tempMountInfo.siteLocation
      this.tempInfo.bindType = this.clearTypeList.filter(item => item.type === this.siteType)[0].code
      this.tempInfo.workOrderID = this.getAllWorkOrderInfo.workOrderID
      if (this.siteSubType === this.subTypeMaterials) {
        this.tempInfo.resourceType = 0
      } else if (this.siteSubType === this.subTypeTools) {
        this.tempInfo.resourceType = 1
      }
      const clearData = _.omit(this.tempInfo, ['resources', 'forceBind'])
      const data = await updateBindResource(clearData)
      if (data.status === 200) {
        this.$notify({
          title: (this.$t('share.success')).toString(),
          message: (this.$t('share.updateSuccessfully')).toString(),
          type: 'success',
          duration: 2000
        })
      }
      this.getSiteBindData()
    } else {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyPDA023').toString(),
        type: 'warning',
        duration: 2000
      })
    }
  }
}
