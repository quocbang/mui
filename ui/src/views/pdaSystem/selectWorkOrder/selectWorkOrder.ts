/* eslint-disable no-async-promise-executor */
import { Component, Vue } from 'vue-property-decorator'
import { validateRequire } from '@/utils'
import { UserModule } from '@/store/modules/user'
import { changeWorkOrderStatus, getWorkOrderList, getWorkOrderInfo } from '@/api/workOrder'
import { getStationList, stationSignIn } from '@/api/station'
import { getStationOperator } from '@/api/site'
import { PDAModule } from '@/store/modules/pda'
import { MessageBox } from 'element-ui'
import i18n from '@/lang'
import { getStationConfig } from '@/api/ui'

@Component({
  name: 'selectWorkOrder'
})

export default class extends Vue {
  private workOrderInfoList: any[] = []
  private productIDList: any[] = []
  private workOrderInfo: any
  private action = false
  private jobViewRedirect = false
  private tempSelectWorkOrderInfo = {
    stationID: '',
    siteName: '',
    siteIndex: 0
  }

  private selectStationInfo: any[] = []
  private stationConfig: any
  private dialogSelectModelVisible = false

  private rules = {
    stationID: [{ validator: validateRequire }]
  }

  private tempWorkOrderID = ''
  private loadWorkOrderOK = false

  created() {
    this.getStationList()
  }

  mounted() {
    UserModule.FeedAndCollectMode('togetherMode')
    const stationIDEl: any = this.$refs.stationID
    stationIDEl.focus()
  }

  private async getWorkOrderListAndStationConfig() {
    const { data } = await getStationConfig(this.tempSelectWorkOrderInfo.stationID)
    this.stationConfig = data.stationConfig
    if (this.stationConfig.separateMode === true) {
      this.dialogSelectModelVisible = true
    } else {
      this.runCheckAndGetInfo()
    }
  }

  private async runCheckAndGetInfo() {
    await this.workOrderListInfo()
    await this.checkWorkOrderStatus()
    if (this.jobViewRedirect === true && this.workOrderInfoList.length !== 0) {
      await this.operatorSignInCheck(this.tempWorkOrderID)
    }
  }

  private selectMode(mode: string) {
    this.dialogSelectModelVisible = false
    UserModule.FeedAndCollectMode(mode)
    this.runCheckAndGetInfo()
  }

  public async operatorSignInCheck(workOrderID: string) {
    const operatorID = await this.getStationInfo()
    if (operatorID !== UserModule.name && operatorID !== '') {
      MessageBox.confirm(
        i18n.t('message.notifyForceStationSignIn').toString(),
        i18n.t('share.prompt').toString(),
        {
          confirmButtonText: i18n.t('share.confirm').toString(),
          cancelButtonText: i18n.t('share.cancel').toString(),
          type: 'warning'
        }
      ).then(async () => {
        await this.operatorSignIn(workOrderID)
      })
    } else {
      await this.operatorSignIn(workOrderID)
    }
  }

  private async operatorSignIn(workOrderID: any) {
    this.selectStationInfo = UserModule.selectStationsInfo
    await UserModule.Station(this.tempSelectWorkOrderInfo.stationID)
    const operatorInfo = {
      siteName: '',
      group: parseInt(UserModule.groups),
      workDate: UserModule.workDate
    }
    switch (UserModule.feedAndCollectMode) {
      case 'feed':
        operatorInfo.siteName = this.stationConfig.feed.operatorSites[0].siteName
        this.tempSelectWorkOrderInfo.siteName = this.stationConfig.feed.operatorSites[0].siteName
        break
      case 'receipt':
        operatorInfo.siteName = this.stationConfig.collect.operatorSites[0].siteName
        this.tempSelectWorkOrderInfo.siteName = this.stationConfig.collect.operatorSites[0].siteName
        break
    }
    this.selectStationInfo.push(this.tempSelectWorkOrderInfo)
    this.selectStationInfo = [...new Set(this.selectStationInfo.map(item => JSON.stringify(item)))].map(item => JSON.parse(item))
    UserModule.SelectStation(this.selectStationInfo)
    const data = await stationSignIn(this.tempSelectWorkOrderInfo.stationID, operatorInfo)
    if (data.status === 200) {
      await this.loadWorkOrder(workOrderID)
      this.$router.push({ path: '/jobView' })
    }
  }

  public async workOrderListInfo() {
    this.action = true
    const query = {
      workDate: UserModule.workDate
    }
    const { data } = await getWorkOrderList(this.tempSelectWorkOrderInfo.stationID, query)
    this.workOrderInfoList = data
  }

  private async getStationInfo() {
    const stationInfo = {
      site: {
        stationID: this.tempSelectWorkOrderInfo.stationID,
        siteName: '',
        siteIndex: 0
      }
    }
    const { data } = await getStationOperator(this.tempSelectWorkOrderInfo.stationID, stationInfo)
    return data.operatorID
  }

  private async getStationList() {
    const { data } = await getStationList('')
    this.productIDList = data
  }

  private async loadWorkOrder(workOrderID: any) {
    const statusPending = 0
    await new Promise(async (resolve) => {
      const { data } = await getWorkOrderInfo(workOrderID)
      this.workOrderInfo = data
      await PDAModule.WorkOrderInfo(this.workOrderInfo)
      resolve('getWorkOrderInfo success')
    })
    if (this.workOrderInfo.workOrderStatus === statusPending) {
      await new Promise(async (resolve) => {
        const data = await changeWorkOrderStatus(workOrderID, {
          type: 0,
          remark: 0
        })
        if (data.status === 200) {
          this.loadWorkOrderOK = true
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: (this.$t('share.success')).toString(),
            type: 'success',
            duration: 2000
          })
        }
        resolve('changeWorkOrderStatus success')
      })
    } else {
      this.loadWorkOrderOK = true
      this.$notify({
        title: (this.$t('share.success')).toString(),
        message: (this.$t('share.success')).toString(),
        type: 'success',
        duration: 2000
      })
    }
  }

  private async checkWorkOrderStatus() {
    const statusActive = 1
    const statusClosing = 2
    const loadStatusActiveData = this.workOrderInfoList.filter((item: { workOrderStatus: number }) => item.workOrderStatus === statusActive || item.workOrderStatus === statusClosing)
    if (loadStatusActiveData.length === 1) {
      this.tempWorkOrderID = loadStatusActiveData[0].workOrderID
      this.jobViewRedirect = true
    } else if (loadStatusActiveData.length > 1) {
      this.workOrderInfoList = loadStatusActiveData
    }
  }
}
