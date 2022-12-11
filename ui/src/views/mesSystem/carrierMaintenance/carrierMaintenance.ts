import { addCarrier, defaultICreateCarrierDataInfo, defaultIUpdateCarrierDataInfo, deleteCarrier, getCarrierList, updateCarrier, getBarcodeCode39 } from '@/api/carrier'
import { UserModule } from '@/store/modules/user'
import { validateCarrierPrefix, validateRequire } from '@/utils'
import { Form, MessageBox } from 'element-ui'
import _, { cloneDeep } from 'lodash'
import moment from 'moment'
import { Component, Vue } from 'vue-property-decorator'
import Pagination from '@/components/Pagination/index.vue'
import i18n from '@/lang'

@Component({
  name: 'carrierMaintenance',
  components: {
    Pagination
  }
})

export default class extends Vue {
  // Top condition query value
  private departmentOIDValue = ''
  private departmentInfoList: any[] = []
  private carrierList: any[] = []

  private dialogCreateVisible = false
  private dialogUpdateVisible = false

  private tempAddData = defaultICreateCarrierDataInfo
  private tempUpdateData = defaultIUpdateCarrierDataInfo

  private tableKey = 0
  private total = 0
  private paginationLimit = 10
  private listQuery = {
    page: 1,
    limit: this.paginationLimit
  }

  private listLoading = false
  private isPlanAlive = false
  private multipleSelection: any[] = []
  private carrierQuery: any = {}
  created() {
    this.getUserInfo()
    // only one auth select
    if (this.departmentInfoList.length === 1) {
      this.departmentOIDValue = this.departmentInfoList[0].label.OID
      this.getCarrierInfo()
    }
  }

  // Form Rules
  private dataCreateFormRules = {
    idPrefix: [{ validator: validateRequire }, { validator: validateCarrierPrefix }],
    quantity: [{ validator: validateRequire }]
  }

  private getUserInfo() {
    const authorizedDepartmentsString = JSON.parse(UserModule.authorizedDepartments.toString())
    this.departmentInfoList = authorizedDepartmentsString.map((el: any, index: any) => {
      return {
        value: index,
        label: el
      }
    }
    )
  }

  private createCarrier() {
    if (this.departmentOIDValue !== '') {
      this.dialogCreateVisible = true
      this.tempAddData = cloneDeep(defaultICreateCarrierDataInfo)
      this.$nextTick(() => {
        (this.$refs.dataCreateForm as Form).clearValidate()
      })
    } else {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notify014').toString(),
        type: 'warning',
        duration: 2000
      })
    }
  }

  private async printBarcode() {
    const allSelectID: string[] = []
    if (this.multipleSelection.length !== 0) {
      this.multipleSelection.forEach(element => {
        allSelectID.push(element.ID)
      })

      const data = await getBarcodeCode39(allSelectID)
      // Create a Blob from the PDF Stream
      const file = new Blob([data.data], { type: 'application/pdf' })
      // Build a URL from the file
      const fileURL = URL.createObjectURL(file)
      // Open the URL on new Window
      window.open(fileURL)
    } else {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notifyItemIsNull').toString(),
        type: 'warning',
        duration: 2000
      })
    }
  }

  private async getCarrierInfo() {
    try {
      this.carrierQuery = {}
      this.carrierQuery.page = this.listQuery.page
      this.carrierQuery.limit = this.listQuery.limit
      this.listLoading = true
      const { data } = await getCarrierList(this.departmentOIDValue, this.carrierQuery)
      this.carrierList = data.items
      this.total = data.total
      this.carrierList.forEach((element: any, elementIndex: number) => {
        this.carrierList[elementIndex].updateAt = moment(this.carrierList[elementIndex].updateAt).format('yyyy-MM-DD HH:mm:ss')
      })
      this.listLoading = false
      this.isPlanAlive = true
    } catch (e: any) {
      this.listLoading = false
      this.carrierList = []
    }
  }

  private updateCarrierRow(rowData: any) {
    this.dialogUpdateVisible = true
    this.tempUpdateData = cloneDeep(rowData)
  }

  private async deleteCarrierRow(rowData: any) {
    MessageBox.confirm(
      i18n.t('message.notifyDelete').toString(),
      i18n.t('share.prompt').toString(),
      {
        confirmButtonText: i18n.t('share.confirm').toString(),
        cancelButtonText: i18n.t('share.cancel').toString(),
        type: 'warning'
      }
    ).then(async () => {
      const data = deleteCarrier(rowData.ID)
      if ((await data).status === 200) {
        this.$notify({
          title: (this.$t('share.success')).toString(),
          message: this.$t('share.deleteSuccessfully').toString(),
          type: 'success',
          duration: 2000
        })
        this.getCarrierInfo()
      }
    })
  }

  private createCarrierInfoToDB() {
    (this.$refs.dataCreateForm as Form).validate(async (valid: any) => {
      if (valid) {
        this.tempAddData.departmentOID = this.departmentOIDValue
        const data = addCarrier(this.tempAddData)
        if ((await data).status === 200) {
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: this.$t('share.addSuccessfully').toString(),
            type: 'success',
            duration: 2000
          })
          this.dialogCreateVisible = false
          this.getCarrierInfo()
        }
      }
    })
  }

  private UpdateCarrierInfoToDB() {
    (this.$refs.dataUpdateForm as Form).validate(async (valid: any) => {
      if (valid) {
        const updateDate = _.omit(this.tempUpdateData, ['ID', 'updateAt', 'updateBy'])
        const data = updateCarrier(this.tempUpdateData.ID, updateDate)
        if ((await data).status === 200) {
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: this.$t('share.updateSuccessfully').toString(),
            type: 'success',
            duration: 2000
          })
          this.dialogUpdateVisible = false
          this.getCarrierInfo()
        }
      }
    })
  }

  private handleSelectionChange(val: any) {
    this.multipleSelection = val
  }
}
