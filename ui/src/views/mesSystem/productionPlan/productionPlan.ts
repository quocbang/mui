/* eslint-disable no-async-promise-executor */
import { Component, Vue } from 'vue-property-decorator'
import { BatchSize, GetDate, validateNumber, validateRequire } from '@/utils/index'
import { Form } from 'element-ui'
import { UserModule } from '@/store/modules/user'
import { addPlanData, defaultPlanValue, getPlanList } from '@/api/plan'
import { getProductList, getProductTypeList } from '@/api/product'
import { getProductGroupsList } from '@/api/ui'
import { cloneDeep } from 'lodash'
import { uploadWorkOrders } from '@/api/workOrder'
import { IPlanData } from '@/api/planTypes'
import moment from 'moment'
import Pagination from '@/components/Pagination/index.vue'
import WorkOrderDialog from '@/components/workOrderDialog/workOrderDialog.vue'

@Component({
  name: 'productionPlan',
  components: {
    Pagination,
    WorkOrderDialog
  }
})

export default class extends Vue {
  // Top condition query value
  private departmentOIDValue = ''
  private productTypeValue = ''
  private dateValue = moment(GetDate(0)).format('YYYY-MM-DD')
  private departmentInfoList: any[] = []
  private productTypeInfoList: any[] = []
  // receive Info List
  private originalPlanData: IPlanData[] = []
  private productIDList: any[] = []
  private groupsInfoList: any[] = []
  private planList: any[] = []
  private multipleSelection: any[] = []
  private isPlanAlive = false
  private tableKey = 0
  private listLoading = false
  private planLoading = true
  private BatchSize = BatchSize
  private uploadWorkOrdersResponses: any
  private listQuery = {
    page: 1,
    pageSize: 10
  }

  // dialog Visible
  private dialogCreatePlanFormVisible = false
  // private dialogWorkOrderVisible = false
  private tempPlanData = defaultPlanValue

  // Form Rules
  private planRules = {
    productID: [{ validator: validateRequire }],
    dayQuantity: [{ validator: validateRequire }, { validator: validateNumber }]
  }

  created() {
    this.getUserInfo()
    // only one auth select
    if (this.departmentInfoList.length === 1) {
      this.departmentOIDValue = this.departmentInfoList[0].label.OID
      this.onDepartmentList(this.departmentOIDValue)
    }
  }

  private createWorkOrder(productID: string) {
    const refWorkOrder: any = this.$refs.refWorkOrder
    refWorkOrder.createWorkOrder(productID)
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

  private async getProductGroupsInfo() {
    if (this.departmentOIDValue === '' || this.productTypeValue === '') {
      return
    }
    try {
      const { data } = await getProductGroupsList(this.departmentOIDValue, this.productTypeValue)
      this.groupsInfoList = data
    } catch (e) {
      console.log(e)
    }
  }

  private async onDepartmentList(DepartmentOID: string) {
    try {
      const { data } = await getProductTypeList(DepartmentOID)
      this.productTypeInfoList = data
      this.productTypeValue = ''
      // only one auth select
      if (this.productTypeInfoList.length === 1) {
        this.productTypeValue = this.productTypeInfoList[0].type
        this.onGetPlanList()
      }
    } catch (e) {
      console.log(e)
    }
  }

  private async onGetPlanList() {
    if (this.departmentOIDValue === '' || this.productTypeValue === '' || this.dateValue === '') {
      return
    }
    try {
      this.getProductGroupsInfo()
      this.listLoading = true
      const { data } = await getPlanList(this.departmentOIDValue, this.productTypeValue, moment(this.dateValue).format('YYYY-MM-DD'))
      this.originalPlanData = data
      // originalPlanData To UI Data
      const resultArray: { productID: any, dayQuantity: string, weekQuantity: string, stockQuantity: string, scheduledQuantity: string, children: any[] }[] = []
      // all parent for groups
      const parentArray = this.groupsInfoList.map((el) => el.parent)
      // all children for groups
      const allChildren = this.groupsInfoList.map((el) => el.children)
      const childrenArray: any[] = []
      allChildren.forEach((children) => {
        children.forEach((item: any) => childrenArray.push(item))
      })
      const allGroupsData = parentArray.concat(childrenArray)
      // add conversionDays Data to originalPlanData
      const tempOriginalPlanData: any[] = this.originalPlanData
      this.originalPlanData.forEach((el, index) => {
        tempOriginalPlanData[index].conversionDays = (parseFloat(el.stockQuantity) / parseFloat(el.dayQuantity)).toString()
      })
      // find originalPlanData parent
      const originalPlanDataParent = tempOriginalPlanData.filter((el) => parentArray.includes(el.productID) || (!allGroupsData.includes(el.productID)))
      // find originalPlanData children
      const OriginalPlanDataChildren = tempOriginalPlanData.filter(
        (el) => !parentArray.includes(el.productID)
      )
      // push parent to resultArray
      originalPlanDataParent.forEach((parent: any) => {
        resultArray.push(parent)
      })
      // find children's parent
      OriginalPlanDataChildren.forEach((children: any) => {
        // who is parent
        const parentObj = this.groupsInfoList.find(el => el.children.includes(children.productID)) == null ? null
          : this.groupsInfoList.find(el => el.children.includes(children.productID))
        // if parentObj is not null, parent push into children
        if (parentObj != null) {
          const findParentByResultArray = resultArray.find(el => el.productID === parentObj.parent)
          // if originalPlanData not find parent ,generate parent from groups
          if (findParentByResultArray == null) {
            const newObj = {
              productID: parentObj.parent,
              dayQuantity: '-',
              weekQuantity: '-',
              stockQuantity: '-',
              conversionDays: '-',
              scheduledQuantity: '-',
              children: [children]
            }
            resultArray.push(newObj)
          } else {
            // if have children,but originalPlanData not have, children is null
            if (findParentByResultArray.children == null) {
              findParentByResultArray.children = []
            }
            findParentByResultArray.children.push(children)
          }
        }
      })
      // result Data
      this.planList = resultArray
      this.listLoading = false
      this.isPlanAlive = true
    } catch (e) {
      this.listLoading = false
      console.log(e)
    }
  }

  // create Plan
  private async createPlan() {
    if (this.productTypeValue === '') {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: (this.$t('message.notify002')).toString(),
        type: 'warning',
        duration: 2000
      })
      return
    }
    this.tempPlanData = cloneDeep(defaultPlanValue)
    this.dialogCreatePlanFormVisible = true
    this.planLoading = true
    try {
      const { data } = await getProductList(this.productTypeValue)
      const planListProductID = this.originalPlanData.map((el) => el.productID)
      const filterProductID = data.filter((el: any) => !planListProductID.includes(el))
      this.productIDList = filterProductID
      this.$nextTick(() => {
        (this.$refs.planDataForm as Form).clearValidate()
      })
      this.planLoading = false
    } catch {
      this.planLoading = false
    }
    // only one auth select
    if (this.productIDList.length === 1) {
      this.tempPlanData.productID = this.productIDList[0]
    }
    this.$nextTick(() => {
      (this.$refs.planDataForm as Form).clearValidate()
    })
  }

  private async createPlanToDB() {
    (this.$refs.planDataForm as Form).validate(async valid => {
      if (valid) {
        this.tempPlanData.departmentOID = this.departmentOIDValue
        this.tempPlanData.productType = this.productTypeValue
        this.tempPlanData.date = moment(this.dateValue).format('YYYY-MM-DD')
        try {
          await addPlanData(this.tempPlanData)
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: this.$t('share.addSuccessfully').toString(),
            type: 'success',
            duration: 2000
          })
        } catch (e: any) {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: (this.$t('share.fail')).toString() + ': ' + e.errorMessage,
            type: 'warning',
            duration: 2000
          })
        }

        await new Promise((resolve) => {
          this.getProductGroupsInfo()
          resolve('getPlan success')
        })
        await new Promise((resolve) => {
          this.onGetPlanList()
          resolve('getPlanList success')
        })
        this.dialogCreatePlanFormVisible = false
      }
    })
  }

  private handleUpload() {
    (this.$refs['excel-upload-input'] as HTMLInputElement).click()
  }

  private handleClick(e: MouseEvent) {
    const files = (e.target as HTMLInputElement).files
    if (files) {
      const rawFile = files[0] // only use files[0]
      const isLt1MB = rawFile.size / 1024 / 1024 < 1
      if (!this.isExcel(rawFile)) {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: (this.$t('message.notifyWorkOrder102')).toString(),
          type: 'warning',
          duration: 2000
        })
        return false
      }
      if (!isLt1MB) {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: (this.$t('message.notifyWorkOrder103')).toString(),
          type: 'warning',
          duration: 2000
        })
        return false
      }
      this.uploadExcel(rawFile)
    }
  }

  private isExcel(file: File) {
    return /\.(xlsx|xls)$/.test(file.name)
  }

  private async uploadExcel(file: File) {
    (this.$refs['excel-upload-input'] as HTMLInputElement).value = ''
    const formData = new FormData()
    formData.append('uploadFile', file)
    const { data } = await uploadWorkOrders(this.departmentOIDValue, formData)
    this.uploadWorkOrdersResponses = data
    if (this.uploadWorkOrdersResponses !== undefined) {
      if (this.uploadWorkOrdersResponses.failData.length === 0) {
        this.$notify({
          title: (this.$t('share.success')).toString(),
          message: this.$t('share.updateSuccessfully').toString(),
          type: 'success',
          duration: 2000
        })
      } else {
        let allFailMsg = ''
        this.uploadWorkOrdersResponses.failData.forEach((item: any) => {
          allFailMsg = allFailMsg + this.$t('share.the') + item.index + this.$t('share.row') + item.columns.toString() + this.$t('share.error') + '</br>'
        })
        this.$alert(allFailMsg, (this.$t('share.errorMessage')).toString(), {
          dangerouslyUseHTMLString: true,
          confirmButtonText: (this.$t('share.confirm')).toString()
        })
      }
    }
  }
}
