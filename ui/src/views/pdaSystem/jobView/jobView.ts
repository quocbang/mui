import { Component, Vue } from 'vue-property-decorator'
import Pagination from '@/components/Pagination/index.vue'
import { printMaterialResource } from '@/api/resource'
import { PDAModule } from '@/store/modules/pda'
import { collect, feed } from '@/api/produce'
import { changeWorkOrderStatus, getWorkOrderInfo } from '@/api/workOrder'
import { getStationConfig } from '@/api/ui'
import { UserModule } from '@/store/modules/user'
import { Form, Input, MessageBox } from 'element-ui'
import i18n from '@/lang'
import { cloneDeep } from 'lodash'

@Component({
  name: 'jobView',
  components: {
    Pagination
  }
})

export default class extends Vue {
  private workOrderInfo = {
    productID: '',
    planQuantity: '0',
    currentQuantity: 0
  }

  private getAllWorkOrderInfo: any

  private closeWorkOrderRemark = [{
    name: 'normal',
    code: 0
  }, {
    name: 'changes',
    code: 1
  }, {
    name: 'materialUnusual',
    code: 2
  }, {
    name: 'stationFailure',
    code: 3
  }, {
    name: 'other',
    code: 4
  }]

  private closeWorkOrderReason = ''
  private dialogCloseWorkOrder = false
  private stationID = ''
  private stationSiteList: any
  private targetSite: any
  private targetResource: any
  private resourceInfo: any[] = []
  private bindSuccessTag = false
  private fullscreenLoading = false
  private tempBindInfo = {
    station: '',
    bindType: 0,
    siteName: '',
    siteIndex: 0,
    resources: [{
      resourceID: '',
      productType: '',
      quantity: ''
    }],
    resourceType: 0
  }

  private feedAndCollect = {
    stationID: '',
    feed: {
      batch: 0,
      source: [
        {
          siteInfo: {
            stationID: '',
            siteName: '',
            siteIndex: 0
          },
          quantity: 0
        }
      ]
    },
    collect: {
      group: 1,
      workDate: '',
      sequence: 0,
      resourceID: '',
      quantity: 0,
      carrierResource: '',
      print: false
    }
  }

  private tempStationConfigInfo = {
    stationConfig: {
      separateMode: true,
      feed: {
        productType: [
        ],
        materialResource: true,
        standardQuantity: 0,
        operatorSites: [
          {
            stationID: '',
            siteName: '',
            siteIndex: 0
          }
        ]
      },
      collect: {
        resource: true,
        carrierResource: true,
        quantity: {
          type: 0,
          value: ''
        },
        operatorSites: [
          {
            stationID: '',
            siteName: '',
            siteIndex: 0
          }
        ]
      }
    }
  }

  private tempFeed = {
    stationID: '',
    workOrderID: '',
    batch: 0,
    resource: [
      {
        ID: '',
        quantity: ''
      }
    ],
    forceFeed: {
      force: false
    },
    closeBatch: true
  }

  private tempCollectOrFeedAndCollect = {
    workOrderID: '',
    stationID: '',
    sequence: 0,
    resourceID: '',
    quantity: '',
    carrierResource: '',
    print: true,
    feedResourceIDs: [],
    forceCollect: {
      force: false
    }
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

  private standardQuantity = {
    RECIPE: 0,
    USER: 1
  }

  private modelTabsValue = ''

  async created() {
    await this.getStationConfig()
    this.setData()
    this.tabClick()
  }

  mounted() {
    if (UserModule.feedAndCollectMode === undefined) {
      this.modelTabsValue = 'togetherMode'
    } else {
      this.modelTabsValue = UserModule.feedAndCollectMode
    }
  }

  private tabClick() {
    switch (this.modelTabsValue) {
      case 'togetherMode':
        this.$nextTick(() => {
          (this.$refs.focusTogetherMode as Input).focus()
        })
        break
      case 'feed':
        this.$nextTick(() => {
          ((this.$refs.focusFeed_0 as any)[0] as Input).focus()
        })
        break
      case 'receipt':
        this.$nextTick(() => {
          (this.$refs.focusReceipt as Input).focus()
        })

        break
    }
  }

  private async setData() {
    if (typeof PDAModule.workOrderInfo !== 'object') {
      this.getAllWorkOrderInfo = JSON.parse(PDAModule.workOrderInfo.toString())
    } else {
      this.getAllWorkOrderInfo = PDAModule.workOrderInfo
    }
    this.stationID = UserModule.station
    if (this.getAllWorkOrderInfo.workOrderID === undefined) {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyPDA022').toString(),
        type: 'warning',
        duration: 2000
      })
    } else {
      this.setUIInfo()
    }
  }

  private setUIInfo() {
    this.workOrderInfo.planQuantity = this.getAllWorkOrderInfo.planQuantity
    this.workOrderInfo.productID = this.getAllWorkOrderInfo.productID
    this.workOrderInfo.currentQuantity = this.getAllWorkOrderInfo.currentQuantity
  }

  private addSequenceAndBatch() {
    this.getAllWorkOrderInfo.currentBatch = this.getAllWorkOrderInfo.currentBatch + 1
    this.getAllWorkOrderInfo.collectSequence = this.getAllWorkOrderInfo.collectSequence + 1
  }

  private reduceSequenceAndBatch() {
    this.getAllWorkOrderInfo.currentBatch = this.getAllWorkOrderInfo.currentBatch - 1
    this.getAllWorkOrderInfo.collectSequence = this.getAllWorkOrderInfo.collectSequence - 1
  }

  private async getStationConfig() {
    const { data } = await getStationConfig(UserModule.station)
    this.tempStationConfigInfo = data
    this.tempCollectOrFeedAndCollect.quantity = this.tempStationConfigInfo.stationConfig.collect.quantity.value
  }

  private async collectOrFeedAndCollectMode(mode: string) {
    if (this.getAllWorkOrderInfo.workOrderID === undefined) {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyPDA022').toString(),
        type: 'warning',
        duration: 2000
      })
    } else if (parseInt(this.tempCollectOrFeedAndCollect.quantity) <= 0 ||
      this.tempCollectOrFeedAndCollect.quantity === undefined) {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyPDA010').toString(),
        type: 'warning',
        duration: 2000
      })
    } else {
      this.fullscreenLoading = true
      this.tempCollectOrFeedAndCollect.workOrderID = this.getAllWorkOrderInfo.workOrderID
      this.tempCollectOrFeedAndCollect.stationID = this.stationID
      this.tempCollectOrFeedAndCollect.forceCollect.force = false
      if (this.tempCollectOrFeedAndCollect.resourceID === '' && this.tempCollectOrFeedAndCollect.carrierResource === '') { this.tempCollectOrFeedAndCollect.print = true } else { this.tempCollectOrFeedAndCollect.print = false }
      this.addSequenceAndBatch()
      this.tempCollectOrFeedAndCollect.sequence = this.getAllWorkOrderInfo.collectSequence
      const newTempCollectOrFeedAndCollect: any = cloneDeep(this.tempCollectOrFeedAndCollect)
      // because this.newTempCollectOrFeedAndCollect.quantity cannot be converted to string, so new newTempCollectOrFeedAndCollect replace
      newTempCollectOrFeedAndCollect.quantity = newTempCollectOrFeedAndCollect.quantity.toString()
      this.runFeedAndCollect(newTempCollectOrFeedAndCollect, mode)
    }
  }

  private async runFeedAndCollect(info: any, mode: string) {
    const { data: collectMsg } = await collect(this.stationID, info)
    this.fullscreenLoading = false
    if (collectMsg === undefined) {
      this.reduceSequenceAndBatch()
      return
    }
    if (collectMsg.mesResponse.success) {
      await this.$notify({
        title: (this.$t('share.success')).toString(),
        message: (this.$t('system.' + mode)).toString() + (this.$t('share.success')).toString(),
        type: 'success',
        duration: 2000
      })
      this.updateWorkOrderInfo()
      if (collectMsg.print !== undefined) {
        if (!collectMsg.print.success) {
          this.$notify({
            title: (this.$t('errorCodes.errorCode_100500')).toString(),
            message: (this.$t('errorCodes.' + 'errorCode_' + collectMsg.print.error.code)).toString(),
            type: 'warning',
            duration: 2000
          })
        }
      }
    } else {
      if (collectMsg.mesResponse.enableForce) {
        let allError = ''
        collectMsg.mesResponse.error.forEach((errorItem: any) => {
          allError = allError + this.$t('errorCodes.' + 'errorCode_' + errorItem.code) + '</br>'
        })
        await MessageBox.confirm(
          i18n.t('message.notifyJobViewForce').toString() + '</br>' + i18n.t('message.notifyJobViewErrMsg').toString() + '</br>' + allError,
          i18n.t('share.prompt').toString(),
          {
            dangerouslyUseHTMLString: true,
            confirmButtonText: i18n.t('share.confirm').toString(),
            cancelButtonText: i18n.t('share.cancel').toString(),
            type: 'warning'
          }
        ).then(async () => {
          info.forceCollect.force = true
          const { data: forceCollectMsg } = await collect(this.stationID, info)
          if (forceCollectMsg.mesResponse.success) {
            await this.$notify({
              title: (this.$t('share.success')).toString(),
              message: (this.$t('system.' + mode)).toString() + (this.$t('share.success')).toString(),
              type: 'success',
              duration: 2000
            })
            this.updateWorkOrderInfo()
            if (forceCollectMsg.print !== undefined) {
              if (!forceCollectMsg.print.success) {
                this.$notify({
                  title: (this.$t('errorCodes.errorCode_100500')).toString(),
                  message: (this.$t('errorCodes.' + 'errorCode_' + forceCollectMsg.print.error.code)).toString(),
                  type: 'warning',
                  duration: 2000
                })
              }
            }
          } else {
            this.reduceSequenceAndBatch()
            let allError = ''
            forceCollectMsg.mesResponse.error.forEach((errorItem: { code: string }) => {
              allError = allError + this.$t('errorCodes.' + 'errorCode_' + errorItem.code) + '</br>'
            })
            this.$notify({
              dangerouslyUseHTMLString: true,
              title: (this.$t('share.errorMessage')).toString(),
              message: allError,
              type: 'warning',
              duration: 2000
            })
          }
        }).catch(() => {
          this.reduceSequenceAndBatch()
        })
      } else {
        let allError = ''
        collectMsg.mesResponse.error.forEach((errorItem: { code: string }) => {
          allError = allError + this.$t('errorCodes.' + 'errorCode_' + errorItem.code) + '</br>'
        })
        this.$notify({
          dangerouslyUseHTMLString: true,
          title: (this.$t('share.errorMessage')).toString(),
          message: allError,
          type: 'warning',
          duration: 2000
        })
      }
    }
  }

  private async feed() {
    this.fullscreenLoading = true
    let ok = true
    const newTempFeed: any = cloneDeep(this.tempFeed)
    this.tempFeed.resource.forEach((item: any, index: number) => {
      // because this.resource[index].quantity cannot be converted to string, so new newTempFeed replace
      newTempFeed.resource[index].quantity = item.quantity.toString()
      if (item.ID === '' || item.ID === undefined) {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: this.$t('message.notifyRequired200').toString(),
          type: 'warning',
          duration: 5000
        })
        ok = false
      } else if (parseInt(item.quantity) <= 0 &&
        this.tempStationConfigInfo.stationConfig.feed.standardQuantity === this.standardQuantity.USER) {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: this.$t('message.notifyPDA024').toString(),
          type: 'warning',
          duration: 2000
        })
        ok = false
      }
    })
    if (ok) {
      newTempFeed.workOrderID = this.getAllWorkOrderInfo.workOrderID
      newTempFeed.stationID = this.stationID
      newTempFeed.forceFeed.force = false
      this.addSequenceAndBatch()
      newTempFeed.batch = this.getAllWorkOrderInfo.currentBatch
      const { data: feedMsg } = await feed(this.stationID, newTempFeed)
      if (feedMsg === undefined) {
        this.reduceSequenceAndBatch()
        return
      }
      if (feedMsg.success) {
        this.$notify({
          title: (this.$t('share.success')).toString(),
          message: (this.$t('system.feed').toString()) + (this.$t('share.success')).toString(),
          type: 'success',
          duration: 2000
        })
        this.updateWorkOrderInfo()
      } else {
        if (feedMsg.enableForce) {
          let allError = ''
          feedMsg.error.forEach((errorItem: { code: string }) => {
            allError = allError + this.$t('errorCodes.' + 'errorCode_' + errorItem.code) + '</br>'
          })
          MessageBox.confirm(
            i18n.t('message.notifyJobViewForce').toString() + '</br>' + i18n.t('message.notifyJobViewErrMsg').toString() + '</br>' + allError,
            i18n.t('share.prompt').toString(),
            {
              dangerouslyUseHTMLString: true,
              confirmButtonText: i18n.t('share.confirm').toString(),
              cancelButtonText: i18n.t('share.cancel').toString(),
              type: 'warning'
            }
          ).then(async () => {
            newTempFeed.forceFeed.force = true
            newTempFeed.batch = this.getAllWorkOrderInfo.currentBatch
            const { data } = await feed(this.stationID, newTempFeed)
            const forceFeedMsg = data
            if (forceFeedMsg.success) {
              this.$notify({
                title: (this.$t('share.success')).toString(),
                message: (this.$t('system.feed').toString()) + (this.$t('share.success')).toString(),
                type: 'success',
                duration: 2000
              })
              this.updateWorkOrderInfo()
            } else {
              this.reduceSequenceAndBatch()
              let allError = ''
              forceFeedMsg.error.forEach((errorItem: { code: string }) => {
                allError = allError + this.$t('errorCodes.' + 'errorCode_' + errorItem.code) + '</br>'
              })
              this.$notify({
                dangerouslyUseHTMLString: true,
                title: (this.$t('share.errorMessage')).toString(),
                message: allError,
                type: 'warning',
                duration: 2000
              })
            }
          }).catch(() => {
            this.reduceSequenceAndBatch()
          })
        } else {
          let allError = ''
          feedMsg.error.forEach((errorItem: { code: string }) => {
            allError = allError + this.$t('errorCodes.' + 'errorCode_' + errorItem.code) + '</br>'
          })
          this.$notify({
            dangerouslyUseHTMLString: true,
            title: (this.$t('share.errorMessage')).toString(),
            message: allError,
            type: 'warning',
            duration: 2000
          })
        }
      }
    }
    this.fullscreenLoading = false
  }

  public async updateWorkOrderInfo() {
    const { data } = await getWorkOrderInfo(this.getAllWorkOrderInfo.workOrderID)
    this.getAllWorkOrderInfo = data
    await PDAModule.WorkOrderInfo(this.getAllWorkOrderInfo)
    this.setUIInfo()
  }

  private async print() {
    this.fullscreenLoading = true
    const printData = { workOrderID: this.getAllWorkOrderInfo.workOrderID, sequence: this.getAllWorkOrderInfo.collectSequence }
    const data = await printMaterialResource(printData)
    if (data.status === 200) {
      this.$notify({
        title: (this.$t('share.success')).toString(),
        message: (this.$t('share.success')).toString(),
        type: 'success',
        duration: 2000
      })
    }
    this.fullscreenLoading = false
  }

  private closeWorkOrder() {
    this.dialogCloseWorkOrder = true
  }

  private async checkReason() {
    this.dialogCloseWorkOrder = false
    const closeData = {
      type: 1,
      remark: this.closeWorkOrderReason
    }
    const data = await changeWorkOrderStatus(this.getAllWorkOrderInfo.workOrderID, closeData)
    if (data.status === 200) {
      this.$notify({
        title: (this.$t('share.success')).toString(),
        message: (this.$t('share.updateSuccessfully')).toString(),
        type: 'success',
        duration: 2000
      })
      PDAModule.WorkOrderInfoReset()
      this.$router.push({ path: '/selectWorkOrder' })
    }
  }

  private createResource() {
    this.tempFeed.resource.push({
      ID: '',
      quantity: '0'
    })
  }

  private deleteResource(index: number) {
    if (this.tempFeed.resource.length > 1) {
      this.tempFeed.resource.splice(index, 1)
    } else {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyRequired200').toString(),
        type: 'warning',
        duration: 2000
      })
    }
  }
}
