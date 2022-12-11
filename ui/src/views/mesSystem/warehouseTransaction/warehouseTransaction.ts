import { defaultMaterialCreateInfo, getResourceLocation, updateResourceWarehouse } from '@/api/warehouse'
import { clearObjValue, validateRequire } from '@/utils'
import { Form, Input } from 'element-ui'
import _, { cloneDeep } from 'lodash'
import { Component, Vue } from 'vue-property-decorator'
@Component({
  name: 'warehouseTransaction'
})

export default class extends Vue {
  private tempWarehouseTransactionInfo = defaultMaterialCreateInfo
  private originalData = { warehouseID: '', location: '' }
  private warehouseTransactionRules = {
    ID: [{ validator: validateRequire }],
    newWarehouseID: [{ validator: validateRequire }],
    newLocation: [{ validator: validateRequire }]
  }

  private dataLoading = false

  mounted() {
    (this.$refs.ID as Input).focus()
  }

  created() {
    this.tempWarehouseTransactionInfo = cloneDeep(defaultMaterialCreateInfo)
  }

  private async getBarcodeInfo() {
    if (this.tempWarehouseTransactionInfo.ID !== '') {
      try {
        this.dataLoading = true
        const { data } = await getResourceLocation(this.tempWarehouseTransactionInfo.ID)
        this.originalData.warehouseID = data.warehouseID
        this.originalData.location = data.location
        this.tempWarehouseTransactionInfo.newWarehouseID = data.warehouseID
        this.tempWarehouseTransactionInfo.newLocation = data.location
        this.dataLoading = false
      } catch (e) {
        this.dataLoading = false
        console.log(e)
      }
    }
  }

  private checkNewWarehouseID() {
    if (this.originalData.warehouseID !== this.tempWarehouseTransactionInfo.newWarehouseID) {
      this.tempWarehouseTransactionInfo.newLocation = ''
    }
  }

  private updateWarehouseLocationToDB() {
    (this.$refs.warehouseTransactionDataForm as Form).validate(async valid => {
      if (valid) {
        if (!(this.originalData.warehouseID === this.tempWarehouseTransactionInfo.newWarehouseID &&
          this.originalData.location === this.tempWarehouseTransactionInfo.newLocation)) {
          const updateDate = _.omit(this.tempWarehouseTransactionInfo, ['ID'])
          const data = updateResourceWarehouse(this.tempWarehouseTransactionInfo.ID, updateDate)
          if ((await data).status === 200) {
            this.$notify({
              title: (this.$t('share.success')).toString(),
              message: this.$t('share.updateSuccessfully').toString(),
              type: 'success',
              duration: 2000
            })
            this.$nextTick(() => {
              (this.$refs.ID as Input).focus();
              (this.$refs.warehouseTransactionDataForm as Form).clearValidate()
            })
            clearObjValue(this.tempWarehouseTransactionInfo)
          }
        } else {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notify011').toString(),
            type: 'warning',
            duration: 2000
          })
        }
      } else {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: (this.$t('message.notify003')).toString(),
          type: 'warning',
          duration: 2000
        })
      }
    })
  }
}
