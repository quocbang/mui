import { Component, Vue } from 'vue-property-decorator'
import { getProductList, getProductTypeList } from '@/api/product'
import { addMaterial, defaultMaterialCreateInfo } from '@/api/resource'
import { Form } from 'element-ui'
import { clearObjValue, getUserDepartmentsInfo, validateRequire } from '@/utils'
import jstz from 'jstz'
import moment from 'moment-timezone'
import { cloneDeep } from 'lodash'

@Component({
  name: 'materialCreate'
})

export default class extends Vue {
  private departmentOIDValue = ''
  private departmentInfoList: any[] = []
  private productTypeInfoList: any[] = []
  private productInfoList: any[] = []

  private tempMaterialCreateInfo = defaultMaterialCreateInfo
  created() {
    this.tempMaterialCreateInfo = cloneDeep(defaultMaterialCreateInfo)
    // only one auth select
    this.departmentInfoList = getUserDepartmentsInfo()
    if (this.departmentInfoList.length === 1) {
      this.departmentOIDValue = this.departmentInfoList[0].label.OID
      this.onGetProductTypeList(this.departmentOIDValue)
    }
  }

  private materialCreateRules = {
    'resource.productType': [{ validator: validateRequire }],
    'resource.productID': [{ validator: validateRequire }],
    'resource.quantity': [{ validator: validateRequire }],
    'resource.unit': [{ validator: validateRequire }],
    'resource.lotNumber': [{ validator: validateRequire }],
    'resource.productionTime': [{ validator: validateRequire }],
    'resource.expiryTime': [{ validator: validateRequire }],
    'warehouse.ID': [{ validator: validateRequire }],
    'warehouse.location': [{ validator: validateRequire }]
  }

  private async onGetProductTypeList(DepartmentOID: string) {
    try {
      const { data } = await getProductTypeList(DepartmentOID)
      this.productTypeInfoList = data
      this.tempMaterialCreateInfo.resource.productType = ''
      // only one auth select
      if (this.productTypeInfoList.length === 1) {
        this.tempMaterialCreateInfo.resource.productType = this.productTypeInfoList[0].type
        this.onGetProductList(this.tempMaterialCreateInfo.resource.productType)
      }
    } catch (e) {
      console.log(e)
    }
  }

  private async onGetProductList(productType: string) {
    try {
      const { data } = await getProductList(productType)
      this.productInfoList = data
      this.tempMaterialCreateInfo.resource.productID = ''
      // only one auth select
      if (this.productInfoList.length === 1) {
        this.tempMaterialCreateInfo.resource.productID = this.productInfoList[0]
      }
    } catch (e) {
      console.log(e)
    }
  }

  private createMaterialToDB() {
    (this.$refs.materialCreateDataForm as Form).validate(async valid => {
      if (valid) {
        if (parseInt(this.tempMaterialCreateInfo.resource.quantity) > 0) {
          this.tempMaterialCreateInfo.resource.quantity = this.tempMaterialCreateInfo.resource.quantity.toString()
          const timeZone = jstz.determine()
          const timezoneName = timeZone.name()
          this.tempMaterialCreateInfo.resource.productionTime = (moment(new Date(this.tempMaterialCreateInfo.resource.productionTime)).tz(timezoneName).format()).toString()
          this.tempMaterialCreateInfo.resource.expiryTime = (moment(new Date(this.tempMaterialCreateInfo.resource.expiryTime)).tz(timezoneName).format()).toString()
          const { data } = await addMaterial(this.tempMaterialCreateInfo)
          clearObjValue(this.tempMaterialCreateInfo)
          this.$nextTick(() => {
            (this.$refs.materialCreateDataForm as Form).clearValidate()
          })
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: data.resourceID + this.$t('share.addSuccessfully').toString(),
            type: 'success',
            duration: 2000
          })
        } else {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: (this.$t('message.notify004')).toString(),
            type: 'warning',
            duration: 2000
          })
        }
      }
    })
  }
}
