import { Component, Vue } from 'vue-property-decorator'
import { Form, Input } from 'element-ui'
import { defaultIResourceInfo, defaultIResourceOtherInfo, defaultISiteBindResource, getResourceInfo } from '@/api/resource'
import { getSiteBindResource, updateBindResource } from '@/api/site'
import { clearObjValue, validateRequire } from '@/utils'
import _, { cloneDeep } from 'lodash'
import { IBindResourceInfo } from '@/api/resourceTypes'

@Component({
  name: 'materialMount'
})

export default class extends Vue {
  private tempMaterialMountData = defaultIResourceOtherInfo
  private resourceData = defaultIResourceInfo
  private MaterialBarcodeInfoData: any = []
  private bindResourceData: IBindResourceInfo[] = []
  private siteBindResourceData = defaultISiteBindResource
  private GetAllInfoStatus = 'open'
  private listLoading = false
  private resourceLoading = false
  private originQuantity = 0
  private quantityDisable = true
  private ResourceInfoType = {
    MATERIAL: 1,
    STATION: 2,
    CARRIER: 3,
    SITE: 4
  }

  private activeName = 'simpleVersion'
  private productTypeInfoList: any[] = []
  private materialRules = {
    stationID: [{ validator: validateRequire }],
    barcodeForBarrelSlot: [{ validator: validateRequire }]
  }

  mounted() {
    (this.$refs.stationID as Input).focus()
  }

  created() {
    this.tempMaterialMountData = cloneDeep(defaultIResourceOtherInfo)
  }

  private async getBarcodeForBarrelSlotInfo() {
    this.bindResourceData = []
    const result = this.tempMaterialMountData.barcodeForBarrelSlot.split('/')
    this.tempMaterialMountData.siteName = result[0]
    this.tempMaterialMountData.siteIndex = parseInt(result[1])
    try {
      const { data } = await getSiteBindResource(this.tempMaterialMountData.stationID, this.tempMaterialMountData.siteName, this.tempMaterialMountData.siteIndex)
      this.bindResourceData = data
      this.GetAllInfoStatus = 'close'
    } catch (e) {
      this.GetAllInfoStatus = 'open'
      console.log(e)
    }
  }

  private async getMaterialBarcodeInfo() {
    clearObjValue(this.resourceData)
    this.productTypeInfoList = []
    this.tempMaterialMountData.productType = ''
    this.quantityDisable = true
    if (this.tempMaterialMountData.materialBarcode !== '') {
      try {
        this.resourceLoading = true
        const { data } = await getResourceInfo(this.tempMaterialMountData.materialBarcode)
        this.MaterialBarcodeInfoData = data
        this.productTypeInfoList = data.map((item: any) => { return item.productType })
        if (this.productTypeInfoList.length === 1) {
          this.tempMaterialMountData.productType = this.productTypeInfoList[0]
          this.getMaterialInfo()
        }
      } catch (e) {
        this.resourceLoading = false
        this.GetAllInfoStatus = 'open'
        console.log(e)
      }
    }
  }

  private getMaterialInfo() {
    this.resourceData = this.MaterialBarcodeInfoData.filter((item: { productType: string }) => item.productType === this.tempMaterialMountData.productType)[0]
    this.originQuantity = parseFloat(this.resourceData.quantity)
    this.quantityDisable = false
    this.GetAllInfoStatus = 'close'
  }

  private reset() {
    clearObjValue(this.tempMaterialMountData)
    clearObjValue(this.resourceData)
    this.productTypeInfoList = []
    this.tempMaterialMountData.productType = ''
    this.bindResourceData = []
    this.quantityDisable = true
    this.$nextTick(() => {
      (this.$refs.stationID as Input).focus();
      (this.$refs.materialMountDataForm as Form).clearValidate()
    })
  }

  private handleClick(tab: any) {
    this.activeName = tab.name
    this.tempMaterialMountData.productType = ''
    clearObjValue(this.tempMaterialMountData)
    clearObjValue(this.resourceData)
    this.bindResourceData = []
    this.$nextTick(() => {
      (this.$refs.stationID as Input).focus()
    })
  }

  private resourceAdd() {
    (this.$refs.materialMountDataForm as Form).validate(async valid => {
      if (valid) {
        const currentTime = new Date()
        const tempQuantity = this.resourceData.quantity ? this.resourceData.quantity : 0
        const resultResource = _.omit(this.siteBindResourceData, ['queueOption'])
        resultResource.station = this.tempMaterialMountData.stationID
        resultResource.bindType = 1002
        resultResource.siteName = this.tempMaterialMountData.siteName
        resultResource.siteIndex = this.tempMaterialMountData.siteIndex
        if (this.tempMaterialMountData.productType === '') {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notifyRequired100').toString(),
            type: 'warning',
            duration: 2000
          })
        } else if (this.tempMaterialMountData.materialBarcode === '') {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notifyRequired200').toString(),
            type: 'warning',
            duration: 2000
          })
        } else if (tempQuantity === 0) {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notify004').toString(),
            type: 'warning',
            duration: 2000
          })
        } else if (Date.parse(this.resourceData.expiredDate).valueOf() < Date.parse(currentTime.toString()).valueOf()) {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notifyExpired001').toString(),
            type: 'warning',
            duration: 2000
          })
        } else if (this.tempMaterialMountData.barcodeForBarrelSlot !== '' && this.tempMaterialMountData.materialBarcode !== '') {
          resultResource.resources[0].resourceID = this.tempMaterialMountData.materialBarcode
          resultResource.resources[0].productType = this.tempMaterialMountData.productType
          resultResource.resources[0].quantity = this.resourceData.quantity.toString()
          const data = updateBindResource(resultResource)
          if ((await data).status === 200) {
            this.$notify({
              title: (this.$t('share.success')).toString(),
              message: this.$t('share.updateSuccessfully').toString(),
              type: 'success',
              duration: 2000
            })
            clearObjValue(this.resourceData)
            this.tempMaterialMountData.materialBarcode = ''
            this.productTypeInfoList = []
            this.tempMaterialMountData.productType = ''
            this.quantityDisable = true;
            (this.$refs.materialBarcode as Input).focus()
            this.$nextTick(async () => {
              (this.$refs.materialMountDataForm as Form).clearValidate()
              const { data } = await getSiteBindResource(this.tempMaterialMountData.stationID, this.tempMaterialMountData.siteName, this.tempMaterialMountData.siteIndex)
              this.bindResourceData = data
            })
          }
        } else {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('share.fail').toString(),
            type: 'warning',
            duration: 2000
          })
        }
      }
    })
  }

  private resourceCleanDeviation() {
    (this.$refs.materialMountDataForm as Form).validate(async valid => {
      if (valid) {
        const resultResource = _.omit(this.siteBindResourceData, ['actionMode'], ['queueOption'], ['resources'])
        resultResource.station = this.tempMaterialMountData.stationID
        resultResource.bindType = 1011
        resultResource.siteName = this.tempMaterialMountData.siteName
        resultResource.siteIndex = this.tempMaterialMountData.siteIndex
        if (this.tempMaterialMountData.barcodeForBarrelSlot !== '') {
          const data = updateBindResource(resultResource)
          if ((await data).status === 200) {
            this.$notify({
              title: (this.$t('share.success')).toString(),
              message: this.$t('share.updateSuccessfully').toString(),
              type: 'success',
              duration: 2000
            })
            clearObjValue(this.resourceData)
            this.tempMaterialMountData.materialBarcode = ''
            this.tempMaterialMountData.productType = ''
            this.quantityDisable = true;
            (this.$refs.materialBarcode as Input).focus()
            this.$nextTick(async () => {
              (this.$refs.materialMountDataForm as Form).clearValidate()
              const { data } = await getSiteBindResource(this.tempMaterialMountData.stationID, this.tempMaterialMountData.siteName, this.tempMaterialMountData.siteIndex)
              this.bindResourceData = data
            })
          }
        }
      }
    })
  }
}
