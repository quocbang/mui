import { Component, Vue } from 'vue-property-decorator'
import { defaultBarcodeData, defaultBarcodeInfo, getBarcodeControlArea, getBarcodeExpiredDate, getBarcodeInfo, getBarcodeReasonList, getBarcodeUpdateStatusList, updateBarcodeInfo } from '@/api/legacy'
import { Form, Input } from 'element-ui'
import { ICodeStatus } from '@/api/types'

const stageStatus: ICodeStatus[] = [
  { code: '0', description: 'stage.production' },
  { code: '1', description: 'stage.trialRun' },
  { code: '2', description: 'stage.experimental' }
]

@Component({
  name: 'PDA-materialChanges'
})

export default class extends Vue {
  private form = defaultBarcodeData
  private formInfo = defaultBarcodeInfo
  private materialBarcode = ''
  private radioItem = 'changeStatus'
  private changeStatusList: ICodeStatus[] = []
  private reasonList = []
  private controlAreaList: ICodeStatus[] = []
  private allStage = stageStatus
  private GetAllInfoStatus = 'open'
  private getExtendDays = 0
  private dataLoading = false

  created() {
    this.getControlAreaList()
  }

  mounted() {
    (this.$refs.materialBarcode as Input).focus()
  }

  private async getBarcodeInfoList() {
    if (this.materialBarcode !== '' && (this.materialBarcode.length === 10 || this.materialBarcode.length === 16)) {
      try {
        this.dataLoading = true
        const { data } = await getBarcodeInfo(this.materialBarcode)
        this.formInfo = data.material
        this.getStatusList()
        this.getExtendDate()
        this.getReasonList()
        this.GetAllInfoStatus = 'close'
        this.dataLoading = false
      } catch (e) {
        this.GetAllInfoStatus = 'open'
        this.dataLoading = false
        console.log(e)
      }
    }
  }

  private async getStatusList() {
    if (this.materialBarcode !== '') {
      try {
        const { data } = await getBarcodeUpdateStatusList(this.materialBarcode)
        this.changeStatusList = data
        this.form.newStatus = this.changeStatusList[0].code
      } catch (e) {
        this.GetAllInfoStatus = 'open'
        console.log(e)
      }
    }
  }

  private async getExtendDate() {
    if (this.materialBarcode !== '') {
      try {
        const { data } = await getBarcodeExpiredDate(this.materialBarcode)
        this.getExtendDays = data.extendDay
        this.form.extendDays = this.getExtendDays
      } catch (e) {
        this.GetAllInfoStatus = 'open'
        console.log(e)
      }
    }
  }

  private async getReasonList() {
    const { data } = await getBarcodeReasonList()
    this.reasonList = data
  }

  private async getControlAreaList() {
    const { data } = await getBarcodeControlArea()
    this.controlAreaList = data
    this.form.controlArea = this.controlAreaList[0].code
    this.form.productCate = (this.$t('recipe.' + this.allStage[0].description)).toString()
  }

  private async updateBarcodeInfoDate() {
    (this.$refs.dataForm as Form).validate(async (valid) => {
      if (valid) {
        try {
          if (this.radioItem === 'changeStatus') {
            this.form.extendDays = 0
          }
          await updateBarcodeInfo(this.materialBarcode, this.form)
          this.saveClear()
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: this.$t('share.addSuccessfully').toString(),
            type: 'success',
            duration: 2000
          })
        } catch (e) {
          console.log(e)
        }
      }
    })
  }

  private validateRequire = (rule: any, value: string, callback: Function) => {
    const required = (this.$t('share.required')).toString()
    if (value === '' || value === undefined) {
      callback(new Error(required))
    } else {
      callback()
    }
  }

  public rule = {
    productCate: [{ validator: this.validateRequire }],
    newStatus: [{ validator: this.validateRequire }],
    holdReason: [{ validator: this.validateRequire }],
    controlArea: [{ validator: this.validateRequire }]
  }

  private resetBarcodeData() {
    this.form.extendDays = 0
    this.form.holdReason = ''
    this.form.newStatus = ''
    this.form.productCate = (this.$t('recipe.' + this.allStage[0].description)).toString()
    this.formInfo = defaultBarcodeInfo
    this.materialBarcode = ''
    this.radioItem = 'changeStatus'
    this.changeStatusList = []
    this.reasonList = []
    this.allStage = stageStatus
    this.GetAllInfoStatus = 'open'
    this.getExtendDays = 0
    this.$nextTick(() => {
      (this.$refs.materialBarcode as Input).focus();
      (this.$refs.dataForm as Form).clearValidate()
    })
  }

  private saveClear() {
    this.form.extendDays = 0
    this.form.holdReason = ''
    this.form.newStatus = ''

    this.formInfo = defaultBarcodeInfo
    this.formInfo.materialBarcode = ''
    this.formInfo.status = ''
    this.formInfo.barcode = ''
    this.formInfo.productID = ''
    this.formInfo.grade = ''
    this.formInfo.inventory = ''
    this.formInfo.expiredAt = ''
    this.formInfo.createdBy = ''
    this.formInfo.createdAt = 0
    this.formInfo.updateAt = 0
    this.materialBarcode = ''
    this.radioItem = 'changeStatus'
    this.changeStatusList = []
    this.reasonList = []

    this.GetAllInfoStatus = 'open'
    this.getExtendDays = 0
    this.$nextTick(() => {
      (this.$refs.materialBarcode as Input).focus();
      // (this.$refs.materialBarcode as Input).select();
      (this.$refs.dataForm as Form).clearValidate()
    })
  }

  private radioItemChange() {
    this.clearReason()
    if (this.radioItem === 'changeStatus') {
      // this.form.newStatus = this.changeStatusList[0].code
      this.form.extendDays = 0
    } else {
      this.form.extendDays = this.getExtendDays
      if (this.formInfo.status === 'HOLD') {
        this.form.newStatus = 'AVAL'
      } else {
        this.form.newStatus = ''
      }
      this.form.holdReason = ''
    }
  }

  private clearReason() {
    if (this.form.newStatus !== 'HOLD') {
      this.form.holdReason = ''
    }
  }
}
