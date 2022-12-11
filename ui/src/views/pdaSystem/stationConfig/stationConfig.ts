/* eslint-disable no-case-declarations */

import { queryProductTypeList } from '@/api/product'
import { getStationConfig, setStationConfig } from '@/api/ui'
import { getStationList } from '@/api/station'
import { validateRequire } from '@/utils'
import { Form } from 'element-ui'
import { Component, Vue } from 'vue-property-decorator'

@Component({
  name: 'setConfig'
})

export default class extends Vue {
  private stationIDList: any[] = []
  private productTypeList: any[] = []
  private collectQuantityFlag = [
    {
      name: 'userDefined',
      type: 0
    }, {
      name: 'PLC',
      type: 1
    }
  ]

  private feedQuantityFlag = [
    {
      name: 'sourceRecipe',
      type: 0
    },
    {
      name: 'userDefined',
      type: 1
    }
  ]

  private stationID = ''
  private tempStationConfigInfo = {
    stationConfig: {
      separateMode: false,
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

  private targetConfig: any = {}

  private rules = {
    'stationConfig.feed.materialResource': [{ validator: validateRequire }],
    'stationConfig.collect.quantity.type': [{ validator: validateRequire }],
    'stationConfig.collect.resource': [{ validator: validateRequire }],
    'stationConfig.collect.carrierResource': [{ validator: validateRequire }],
    'stationConfig.feed.standardQuantity': [{ validator: validateRequire }]
  }

  created() {
    this.getStationList()
    this.getProductTypeList()
  }

  private async getStationConfig() {
    const { data } = await getStationConfig(this.stationID)
    this.tempStationConfigInfo = data
  }

  private async getStationList() {
    const { data } = await getStationList('')
    this.stationIDList = data
  }

  private async getProductTypeList() {
    this.$nextTick(() => {
      (this.$refs.stationConfigInfoForm as Form).clearValidate()
    })
    const { data } = await queryProductTypeList()
    this.productTypeList = data
  }

  private async confirmConfig(mode: string) {
    if (this.stationID === '') {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notify015').toString(),
        type: 'warning',
        duration: 2000
      })
    } else {
      (this.$refs.stationConfigInfoForm as Form).validate(async valid => {
        if (valid) {
          const userDefined = 0
          if (parseInt(this.tempStationConfigInfo.stationConfig.collect.quantity.value) > 0 || this.tempStationConfigInfo.stationConfig.collect.quantity.type !== userDefined) {
            this.tempStationConfigInfo.stationConfig.collect.quantity.value = this.tempStationConfigInfo.stationConfig.collect.quantity.value.toString()
            switch (mode) {
              case 'togetherMode':
                this.setConfig()
                break
              case 'separateMode':
                const operatorFeed = this.tempStationConfigInfo.stationConfig.feed.operatorSites[0]
                const operatorCollect = this.tempStationConfigInfo.stationConfig.collect.operatorSites[0]
                if (operatorFeed.siteName !== '' && operatorCollect.siteName !== '') {
                  operatorFeed.stationID = this.stationID
                  operatorFeed.siteIndex = 0
                  operatorCollect.stationID = this.stationID
                  operatorCollect.siteIndex = 0
                  this.setConfig()
                } else {
                  this.$notify({
                    title: (this.$t('share.errorMessage')).toString(),
                    message: this.$t('message.notifyPDA023').toString(),
                    type: 'warning',
                    duration: 2000
                  })
                }
                break
              default:
                this.$notify({
                  title: (this.$t('share.errorMessage')).toString(),
                  message: this.$t('message.notify008').toString(),
                  type: 'warning',
                  duration: 2000
                })
            }
          } else {
            this.$notify({
              title: (this.$t('share.errorMessage')).toString(),
              message: this.$t('message.notifyPDA010').toString(),
              type: 'warning',
              duration: 2000
            })
          }
        }
      })
    }
  }

  private async setConfig() {
    const data = await setStationConfig(this.stationID, this.tempStationConfigInfo)
    if (data.status === 200) {
      this.$notify({
        title: (this.$t('share.success')).toString(),
        message: (this.$t('share.success')).toString(),
        type: 'success',
        duration: 2000
      })
    }
  }
}
