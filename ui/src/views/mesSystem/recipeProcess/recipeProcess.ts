import { getFinalProductList, getProductTypeList } from '@/api/product'
import { getProductRecipeIDList, getProductRecipesProcessList } from '@/api/recipe'
import { getUserDepartmentsInfo } from '@/utils'
import { Component, Vue } from 'vue-property-decorator'
import Pagination from '@/components/Pagination/index.vue'
@Component({
  name: 'RecipeProcess',
  components: {
    Pagination
  }
})

export default class extends Vue {
  // Top condition query value
  private departmentOIDValue = ''
  private productTypeValue = ''
  private productIDValue = ''
  private recipeValue = ''
  private departmentInfoList: any[] = []
  private productTypeInfoList: any[] = []
  private productIDInfoList: any[] = []
  private recipeInfoList: any[] = []
  private recipeProcessTable: any[] = []
  private anewProcessDataInfo: any[] = []
  private materialDataInfo: any[] = []
  private stationControlStepDataInfo: any[] = []
  private stationControlCommonDataInfo: any[] = []
  private stationList: any[] = []
  private stationValue = ''
  private rowValue = {}
  private dialogStatus = ''
  private dialogDetailVisible = false
  private dialogOptionFlowDetailVisible = false
  private tableKey = 0
  private total = 0
  private listQuery = {
    page: 1,
    pageSize: 10
  }

  private listLoading = false

  private textMap = {
    optionFlow: this.$t('system.optionFlow')
  }

  created() {
    this.departmentInfoList = getUserDepartmentsInfo()
    // only one auth select
    if (this.departmentInfoList.length === 1) {
      this.departmentOIDValue = this.departmentInfoList[0].label.OID
      this.onDepartmentList()
    }
  }

  private async onDepartmentList() {
    try {
      const { data } = await getProductTypeList(this.departmentOIDValue)
      this.productTypeInfoList = data
      this.productTypeValue = ''
      this.productIDInfoList = []
      this.productIDValue = ''
      this.recipeProcessTable = []
      this.recipeValue = ''
      this.recipeInfoList = []
      // only one auth select
      if (this.productTypeInfoList.length === 1) {
        this.productTypeValue = this.productTypeInfoList[0].type
        this.onProductIDList()
      }
    } catch (e) {
      console.log(e)
    }
  }

  private async onProductIDList() {
    const { data } = await getFinalProductList(this.productTypeValue)
    this.productIDInfoList = data
    this.productIDValue = ''
    this.recipeProcessTable = []
    this.recipeValue = ''
    this.recipeInfoList = []
    // only one auth select
    if (this.productIDInfoList.length === 1) {
      this.productIDValue = this.productIDInfoList[0]
      this.onRecipeList()
    }
  }

  private async onRecipeList() {
    const { data } = await getProductRecipeIDList(this.productIDValue)
    this.recipeInfoList = data
    this.recipeProcessTable = []
    this.recipeValue = ''
    // only one auth select
    if (this.recipeInfoList.length === 1) {
      this.recipeValue = this.recipeInfoList[0]
      this.onRecipeProcessList()
    }
  }

  private async onRecipeProcessList() {
    try {
      this.listLoading = true
      const { data } = await getProductRecipesProcessList(this.recipeValue)
      this.recipeProcessTable = data
      this.recipeProcessTable.forEach((el: any, index: number) => {
        this.recipeProcessTable[index].requiredFlows.stationsList = this.recipeProcessTable[index].requiredFlows.stations.join()
      })
      this.listLoading = false
    } catch (e: any) {
      this.listLoading = false
      this.recipeProcessTable = []
    }
  }

  private openOptionFlowDetailInfo(row: any) {
    this.anewProcessDataInfo = row.optionalFlows
    this.anewProcessDataInfo.forEach((el: any, index: number) => {
      el.processes.forEach((elProcesses: any, indexProcesses: number) => {
        const allStations: any[] = []
        elProcesses.stations.forEach((stationsEl: any) => {
          allStations.push(stationsEl.ID)
        })
        this.anewProcessDataInfo[index].processes[indexProcesses].stationsList = allStations
      })
    })
    this.dialogOptionFlowDetailVisible = true
    this.dialogStatus = 'optionFlow'
  }

  private openDetailInfo(row: any) {
    this.rowValue = row
    // MaterialDetailInfo
    this.stationValue = ''
    this.materialDataInfo = []
    // StationControlDetailInfo
    this.stationControlStepDataInfo = []
    this.stationControlCommonDataInfo = []
    // shareInfo
    this.dialogDetailVisible = true
    this.stationList = row.stations

    // only one auth select
    if (this.stationList.length === 1) {
      this.stationValue = this.stationList[0].ID
      this.onDetailInfo(this.stationValue, this.stationList)
    }
  }

  private onDetailInfo(station: string, stationList: any) {
    // Material Data
    this.onMaterialDataInfoInfo(station, stationList)
    // StationControl Data
    this.onStationControlDetailInfo(station, stationList)
  }

  private onMaterialDataInfoInfo(station: string, stationList: any) {
    const data = stationList.filter((item: { ID: string }) => item.ID === station)[0].bomList
    data.forEach((element: any, index: number) => {
      data[index].substitutesList = element.substitutes.join()
    })
    this.materialDataInfo = data
  }

  private onStationControlDetailInfo(station: string, stationList: any) {
    const data = stationList.filter((item: { ID: string }) => item.ID === this.stationValue)[0].control
    this.stationControlStepDataInfo = data.step
    this.stationControlCommonDataInfo = data.common
    const newRow: any[] = []
    const rowColumns: any[] = []
    data.step.columns.forEach((el: { name: any }) => {
      rowColumns.push(el.name)
    })

    data.step.rows.forEach((stepElement: any, stepIndex: number) => {
      const obj: any = {}
      stepElement.forEach((el: any, index: number) => {
        const key = rowColumns[index]
        obj[key] = el.value
      })
      newRow.push(obj)
      if (data.step.rows.length === stepIndex + 1) {
        data.step.newRows = newRow
      }
    })
    this.stationControlStepDataInfo = data.step
  }

  private sendDataInfo() {
    if (this.stationValue !== '') {
      const targetRow = [this.rowValue]
      const targetData = {
        departmentOID: this.departmentOIDValue,
        productType: this.productTypeValue,
        productID: this.productIDValue,
        recipe: this.recipeValue,
        row: targetRow,
        station: this.stationValue,
        materialDataInfo: this.materialDataInfo,
        stationControlStepDataInfo: this.stationControlStepDataInfo,
        stationControlCommonDataInfo: this.stationControlCommonDataInfo
      }
      const routeData = this.$router.resolve({
        path: '/RecipeProcessPrintView'
      })
      sessionStorage.setItem('targetData', JSON.stringify(targetData))
      window.open(routeData.href, '_blank')
    } else {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: (this.$t('message.notify015')).toString(),
        type: 'warning',
        duration: 2000
      })
    }
  }
}
