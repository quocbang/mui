import { defaultMaterialSplit, defaultMaterialQueryInfo, getMaterialStatus, getResourceInfo, materialSplit, materialBarcode } from '@/api/resource'
import { getProductList, queryProductTypeList, getMaterialInfo } from '@/api/product'
import { Component, Vue } from 'vue-property-decorator'
import Pagination from '@/components/Pagination/index.vue'
import moment from 'moment-timezone'
import jstz from 'jstz'
import { Form, Input } from 'element-ui'
import { validateRequire } from '@/utils'
import _, { cloneDeep } from 'lodash'

@Component({
  name: 'materialQuery',
  components: {
    Pagination
  }
})

export default class extends Vue {
  private activeName = 'productTypeQuery'
  private productTypeInfoList: any[] = []
  private productTypeValue = ''
  private productIDInfoList: any[] = []
  private productIDValue = ''
  private materialStatusList: any[] = []
  private materialStatusValue = ''
  private materialQueryList: any[] = []
  private dateValue = ''
  private status = ''
  private listLoading = false
  private materialLabelCardValue = ''
  private isPlanAlive = false
  private tableKey = 0
  private total = 0
  private paginationLimit = 10
  private listQuery = {
    page: 1,
    limit: this.paginationLimit
  }

  private materialQuery: any = {}
  private dialogBatchMaterialLabelCardVisible = false
  private tempBatchMaterialData = defaultMaterialSplit
  private batchMaterialLabelCardFormRules = {
    batchQuantity: [{ validator: validateRequire }],
    inspections: [{ validator: validateRequire }],
    remark: [{ validator: validateRequire }]
  }

  private dialogDetailFormVisible = false
  private materialQueryRow = defaultMaterialQueryInfo
  private checkList: any[] = []
  private materialBarcodeInfo = {
    productType: '',
    resourceID: ''
  }

  created() {
    this.getProductTypeList()
    this.getMaterialStatusList()
  }

  private async getMaterialStatusList() {
    try {
      const { data } = await getMaterialStatus()
      this.materialStatusList = data
      // only one auth select
      if (this.materialStatusList.length === 1) {
        this.materialStatusValue = this.materialStatusList[0].ID
      }
    } catch (e) {
      console.log(e)
    }
  }

  private async getProductTypeList() {
    try {
      const { data } = await queryProductTypeList()
      this.productTypeInfoList = data
      this.productTypeValue = ''
      // only one auth select
      if (this.productTypeInfoList.length === 1) {
        this.productTypeValue = this.productTypeInfoList[0].type
        this.getGetProductIDList()
      }
    } catch (e) {
      console.log(e)
    }
  }

  private async getGetProductIDList() {
    try {
      const { data } = await getProductList(this.productTypeValue)
      this.productIDInfoList = data
      this.productIDValue = ''
      // only one auth select
      if (this.productIDInfoList.length === 1) {
        this.productIDValue = this.productIDInfoList[0]
      }
    } catch (e) {
      console.log(e)
    }
  }

  private handleClick(tab: any) {
    this.activeName = tab.name
    if (this.activeName === 'barcodeQuery') {
      this.$nextTick(() => {
        (this.$refs.materialLabelCard as Input).focus()
      })
    }
    this.listQuery = {
      page: 1,
      limit: this.paginationLimit
    }
    this.total = 0
    this.materialQuery = {}
    this.productTypeValue = ''
    this.productIDValue = ''
    this.materialStatusValue = ''
    this.dateValue = ''
    this.materialLabelCardValue = ''
    this.materialQueryList = []
  }

  private async getTypeMaterialInfo() {
    if (this.productTypeValue !== '') {
      this.listLoading = true
      this.materialQuery = {}
      this.materialQueryList = []
      if (this.productIDValue !== '') {
        this.materialQuery.productID = this.productIDValue
      }
      if (this.materialStatusValue !== '') {
        this.materialQuery.status = this.materialStatusValue
      }
      if (this.dateValue !== '') {
        const timeZone = jstz.determine()
        const timezoneName = timeZone.name()
        this.materialQuery.startDate = (moment(new Date(this.dateValue)).tz(timezoneName).format()).toString()
      }
      this.materialQuery.page = this.listQuery.page
      this.materialQuery.limit = this.listQuery.limit

      try {
        const { data } = await getMaterialInfo(this.productTypeValue, this.materialQuery)
        this.materialQueryList = data.items
        this.total = data.total
        this.isPlanAlive = true
        this.setStatus()
      } catch (e) {
        this.listLoading = false
      }
    } else {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyRequired100').toString(),
        type: 'warning',
        duration: 2000
      })
    }
  }

  private async getMaterialLabelCardInfo() {
    if (this.materialLabelCardValue !== '') {
      this.listLoading = true
      this.materialQueryList = []
      try {
        const { data } = await getResourceInfo(this.materialLabelCardValue)
        this.isPlanAlive = true
        this.total = data.length
        this.materialQueryList = data
        this.setStatus()
      } catch (e) {
        this.listLoading = false
      }
    }
  }

  private setStatus() {
    this.materialQueryList.forEach((el: any, index: number) => {
      this.materialStatusList.forEach((elStatus: any) => {
        if (elStatus.ID === el.status) {
          this.$set(this.materialQueryList[index], 'statusName', elStatus.name)
        }
      })
    })
    this.listLoading = false
  }

  private clearFilter() {
    this.productIDValue = ''
    this.materialStatusValue = ''
    this.dateValue = ''
  }

  private batchLabelingCard(rowData: any) {
    this.tempBatchMaterialData = cloneDeep(defaultMaterialSplit)
    this.checkList = rowData.inspections
    if (this.checkList.length < 2) {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyMaterial002').toString(),
        type: 'warning',
        duration: 2000
      })
    } else {
      this.dialogBatchMaterialLabelCardVisible = true
      this.tempBatchMaterialData.quantity = rowData.quantity
      this.tempBatchMaterialData.resourceID = rowData.resourceID
      this.tempBatchMaterialData.productType = rowData.productType
      this.$nextTick(() => {
        (this.$refs.BatchMaterialLabelCardForm as Form).clearValidate()
      })
    }
  }

  private showPDFView(rowData: any) {
    this.materialBarcodeInfo.productType = rowData.productType
    this.materialBarcodeInfo.resourceID = rowData.resourceID
    this.printBarcode(this.materialBarcodeInfo)
  }

  private getMaterialSplit() {
    (this.$refs.BatchMaterialLabelCardForm as Form).validate(async valid => {
      if (valid) {
        if (this.tempBatchMaterialData.splitQuantity !== 0) {
          const materialSplitData = _.omit(this.tempBatchMaterialData, ['quantity'])
          try {
            const { data } = await materialSplit(materialSplitData)
            this.$notify({
              title: (this.$t('share.success')).toString(),
              message: data.resourceID + this.$t('share.addSuccessfully').toString(),
              type: 'success',
              duration: 2000
            })
            if (this.activeName === 'productTypeQuery') {
              this.getTypeMaterialInfo()
            } else {
              this.getMaterialLabelCardInfo()
            }
            // printBarcode
            this.materialBarcodeInfo.productType = this.tempBatchMaterialData.productType
            this.materialBarcodeInfo.resourceID = data.resourceID
            this.printBarcode(this.materialBarcodeInfo)
          } catch (e) {
            console.log(e)
          }
          this.dialogBatchMaterialLabelCardVisible = false
        } else {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notifyMaterial001').toString(),
            type: 'warning',
            duration: 2000
          })
        }
      }
    })
  }

  private async printBarcode(materialBarcodeInfo: object) {
    try {
      const data = await materialBarcode(materialBarcodeInfo)
      const file = new Blob([data.data], { type: 'application/pdf' })
      const fileURL = URL.createObjectURL(file)
      window.open(fileURL)
    } catch (e) {
      console.log(e)
    }
  }

  private openDetailInfo(rowData: any) {
    this.materialQueryRow = rowData
    this.dialogDetailFormVisible = true
  }
}
