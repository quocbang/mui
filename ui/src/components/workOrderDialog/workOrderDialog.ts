/* eslint-disable no-async-promise-executor */
import { Component, Prop, Vue } from 'vue-property-decorator'
import { BatchSize, GetDate, validateNumber, validateRequire } from '@/utils/index'
import { defaultAddWorkOrderData } from '@/api/station'
import Decimal from 'decimal.js'
import { Form, Table } from 'element-ui'
import moment from 'moment'
import { createWorkOrder, updateWorkOrder } from '@/api/workOrder'
import { cloneDeep } from 'lodash'
import { getProductRecipesList } from '@/api/recipe'

@Component({
  name: 'WorkOrderDialog'
})
export default class extends Vue {
  private stationList: any = []
  private stationInfoFormatList: any[] = []
  private recipesList: any = []
  private tempWorkOrderData = defaultAddWorkOrderData
  private processInfoList: any[] = []
  private processInfoNameList: any[] = []
  private processInfoTypeList: any[] = []
  private disabled = true
  private tempAddWorkInfo: any = []
  private tempBomInfo: any[] = []
  private BomRecipesList: any[] = []
  private BomProcessInfoList: any[] = []
  private BomProcessInfoNameList: any[] = []
  private BomProcessInfoTypeList: any[] = []
  private multipleSelection: any[] = []
  private BatchSize = BatchSize
  private workOrderLoading = true
  private dateValue = moment(GetDate(0)).format('YYYY-MM-DD')
  private dialog = false
  private targetWorkOrder = ''
  private targetAction = ''
  @Prop() private departmentOIDValue!: string

  private workOrderRules = {
    productID: [{ validator: validateRequire }],
    station: [{ validator: validateRequire }],
    'recipe.ID': [{ validator: validateRequire }],
    'recipe.processName': [{ validator: validateRequire }],
    'recipe.processType': [{ validator: validateRequire }],
    preBatchSize: [{ validator: validateRequire }, { validator: validateNumber }],
    quantity: [{ validator: validateRequire }, { validator: validateNumber }],
    batchCalculation: [{ validator: validateRequire }]
  }

  private getRecipesInfo(station: string) {
    this.stationList = this.stationInfoFormatList.filter(item => item.ID === station)[0]
    const allRecipeID = this.stationList.stationFormatInfo.map((el: { recipeID: any }) => el.recipeID)
    this.recipesList = [...new Set(allRecipeID)]
    this.tempWorkOrderData.recipe.ID = ''
    this.tempWorkOrderData.recipe.processOID = ''
    this.tempWorkOrderData.recipe.processName = ''
    this.tempWorkOrderData.recipe.processType = ''
    // only one auth select
    if (this.recipesList.length === 1) {
      this.tempWorkOrderData.recipe.ID = this.recipesList[0]
      this.getProcessInfoName(this.tempWorkOrderData.recipe.ID)
    }
  }

  private getProcessInfoName(rowDataRecipeID: any) {
    this.processInfoList = this.stationList.stationFormatInfo.filter((item: any) => item.recipeID === rowDataRecipeID)
    const allProcessName: any[] = []
    this.processInfoList.forEach(element => {
      allProcessName.push(element.processInfo.Name)
    })
    this.processInfoNameList = [...new Set(allProcessName)]
    this.tempWorkOrderData.recipe.processName = ''
    this.tempWorkOrderData.recipe.processOID = ''
    this.tempWorkOrderData.recipe.processType = ''
    // only one auth select
    if (this.processInfoNameList.length === 1) {
      this.tempWorkOrderData.recipe.processName = this.processInfoNameList[0]
      this.getProcessInfoType(this.tempWorkOrderData.recipe.processName)
    }
  }

  private getProcessInfoType(processName: any) {
    this.processInfoTypeList = this.processInfoList.filter((item: any) => item.processInfo.Name === processName)
    this.tempWorkOrderData.recipe.processOID = ''
    this.tempWorkOrderData.recipe.processType = ''
    // only one auth select
    if (this.processInfoTypeList.length === 1) {
      this.tempWorkOrderData.recipe.processType = this.processInfoTypeList[0].processInfo.Type
      this.getBatchSizeAndBomInfo(this.tempWorkOrderData.recipe.ID, this.tempWorkOrderData.recipe.processName, this.tempWorkOrderData.recipe.processType)
    }
  }

  private getBatchSizeAndBomInfo(recipeID: string, processName: string, processType: string) {
    this.disabled = true
    this.tempAddWorkInfo = this.stationList.stationFormatInfo.filter((item: { recipeID: string, processInfo: { Name: string, Type: string } }) => item.recipeID === recipeID && item.processInfo.Name === processName && item.processInfo.Type === processType)[0]
    this.tempWorkOrderData.recipe.processOID = this.tempAddWorkInfo.processInfo.OID
    this.tempWorkOrderData.preBatchSize = this.tempAddWorkInfo.batchSize
    if (this.tempWorkOrderData.preBatchSize === '' || this.tempWorkOrderData.preBatchSize === undefined) {
      this.disabled = false
    }
    this.calculateQuantity(this.tempWorkOrderData.batchCount)
    this.getBomInfo()
  }

  private getBomInfo() {
    this.tempBomInfo = this.tempAddWorkInfo.bomList.map((el: any) => {
      return {
        productID: el.productID,
        productType: el.productType,
        quantity: el.quantity,
        recipes: el.recipes,
        planDate: this.tempWorkOrderData.planDate,
        stationInfo: this.recipesDataToStationList(el.recipes)
      }
    })
  }

  // recipesData To StationList
  private recipesDataToStationList(recipesData: any) {
    const DataToStationList: any[] = []
    recipesData.forEach((recipesDataElement: any) => {
      const recipeID = recipesDataElement.ID
      recipesDataElement.processes.forEach((processDataElement: any) => {
        const processInfo = { OID: processDataElement.OID, Name: processDataElement.name, Type: processDataElement.type }
        processDataElement.stations.forEach((stationDataElement: any) => {
          let checkTheSame = false
          if (DataToStationList.length !== 0) {
            for (let i = 0; i < DataToStationList.length; i++) {
              if (DataToStationList[i].ID === stationDataElement.ID) {
                checkTheSame = true
                DataToStationList[i].stationFormatInfo.push({ recipeID: recipeID, processInfo: processInfo, batchSize: stationDataElement.batchSize, bomList: stationDataElement.bomList })
              }
            }
          }
          if (checkTheSame === false) {
            DataToStationList.push({ ID: stationDataElement.ID, stationFormatInfo: [{ recipeID: recipeID, processInfo: processInfo, batchSize: stationDataElement.batchSize, bomList: stationDataElement.bomList }] })
          }
        })
      })
    })
    return DataToStationList
  }

  private calculateQuantity(batchesCountValue: number) {
    if (this.tempWorkOrderData.preBatchSize !== '' && !isNaN(Number(this.tempWorkOrderData.preBatchSize))) {
      this.tempWorkOrderData.quantity = new Decimal(this.tempWorkOrderData.preBatchSize).mul(new Decimal(batchesCountValue)).toNumber()
      this.tempWorkOrderData.batchesQuantity = this.batchesQuantityToArray(this.tempWorkOrderData.quantity, this.tempWorkOrderData.preBatchSize)
    } else {
      this.tempWorkOrderData.quantity = 0
      this.tempWorkOrderData.batchesQuantity = []
    }
  }

  private calculateBatch(quantity: number) {
    if (this.tempWorkOrderData.preBatchSize !== '' && !isNaN(Number(this.tempWorkOrderData.preBatchSize))) {
      this.tempWorkOrderData.batchCount = (Math.ceil(quantity / parseFloat(this.tempWorkOrderData.preBatchSize)))
      this.tempWorkOrderData.batchesQuantity = this.batchesQuantityToArray(this.tempWorkOrderData.quantity, this.tempWorkOrderData.preBatchSize)
    } else {
      this.tempWorkOrderData.batchCount = 1
      this.tempWorkOrderData.batchesQuantity = []
    }
  }

  private batchesQuantityToArray(totalQuantity: any, BatchSize: any) {
    const tempArray: any[] = []
    if (totalQuantity !== Infinity && BatchSize !== 0) {
      const num: number = Math.floor((totalQuantity) / parseFloat(BatchSize))
      let rem = totalQuantity
      for (let i = 0; i < num; i++) {
        tempArray.push(BatchSize.toString())
        rem = new Decimal(rem).sub(new Decimal(BatchSize)).toNumber()
      }
      if (rem !== 0) {
        tempArray.push(rem.toString())
      }
    }
    return tempArray
  }

  private checkValueExists() {
    if (this.tempWorkOrderData.station === '' || this.tempWorkOrderData.recipe.ID === '' || this.tempWorkOrderData.recipe.processName === '' || this.tempWorkOrderData.recipe.processType === '' || this.tempWorkOrderData.quantity === null) {
      this.tempWorkOrderData.preWorkOrder = false
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: (this.$t('message.notify003')).toString(),
        type: 'warning',
        duration: 2000
      })
    }
    this.getBomInfo()
  }

  private getBomRecipesInfo(rowData: any, index: number) {
    const tempData = rowData.stationInfo.filter((item: any) => item.ID === rowData.station)
    this.tempBomInfo[index].BomRecipesInfoList = tempData[0].stationFormatInfo

    const allRecipes: any[] = []
    this.tempBomInfo[index].BomRecipesInfoList.forEach((element: { recipeID: any }) => {
      allRecipes.push(element.recipeID)
    })
    this.BomRecipesList = [...new Set(allRecipes)]
    // clear previous value
    this.toggleSelection(index, false)
    delete this.tempBomInfo[index].recipeID
    delete this.tempBomInfo[index].processName
    delete this.tempBomInfo[index].processType
    delete this.tempBomInfo[index].totalQuantity
  }

  private toggleSelection(index: any, select: boolean) {
    const rows = this.tempBomInfo[index]
    if (this.$refs.multipleTable !== undefined) {
      (this.$refs.multipleTable as Table).toggleRowSelection(rows, select)
    }
  }

  private getBomProcessInfoName(rowDataRecipeID: any, index: number) {
    this.BomProcessInfoList = this.tempBomInfo[index].BomRecipesInfoList.filter((item: any) => item.recipeID === rowDataRecipeID.recipeID)
    const allProcessName: any[] = []
    this.BomProcessInfoList.forEach(element => {
      allProcessName.push(element.processInfo.Name)
    })
    this.BomProcessInfoNameList = [...new Set(allProcessName)]
    delete this.tempBomInfo[index].processType
    delete this.tempBomInfo[index].processName
    delete this.tempBomInfo[index].processOID
    delete this.tempBomInfo[index].totalQuantity
    delete this.tempBomInfo[index].batchesQuantity
  }

  private getBomProcessInfoType(processName: any, index: number) {
    this.BomProcessInfoTypeList = this.BomProcessInfoList.filter((item: any) => item.processInfo.Name === processName)
    delete this.tempBomInfo[index].processType
    delete this.tempBomInfo[index].processOID
    delete this.tempBomInfo[index].totalQuantity
    delete this.tempBomInfo[index].batchesQuantity
    this.toggleSelection(index, false)
  }

  private getBomQuantityInfo(rowData: any, index: number) {
    delete this.tempBomInfo[index].totalQuantity
    delete this.tempBomInfo[index].batchesQuantity
    this.tempBomInfo[index].processOID = this.tempBomInfo[index].BomRecipesInfoList[0].processInfo.OID
    const bomBatchSize = this.tempBomInfo[index].BomRecipesInfoList.filter((item: { recipeID: any }) => item.recipeID === rowData.recipeID)[0].batchSize
    if (bomBatchSize === undefined || bomBatchSize === '' || bomBatchSize === '0') {
      this.toggleSelection(index, false)
      this.tempBomInfo[index].processType = ''
      this.tempBomInfo[index].processOID = ''
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: (this.$t('message.notify100')).toString(),
        type: 'warning',
        duration: 2000
      })
    } else {
      this.tempBomInfo[index].bomBatchSize = bomBatchSize
      // parseFloat
      const bomQuantity = rowData.quantity
      const userQuantity = this.tempWorkOrderData.quantity
      this.tempBomInfo[index].totalQuantity = new Decimal(bomQuantity).mul(new Decimal(new Decimal(userQuantity).div(new Decimal(this.tempWorkOrderData.preBatchSize)).toNumber())).toNumber()
      this.tempBomInfo[index].batchesQuantity = this.batchesQuantityToArray(this.tempBomInfo[index].totalQuantity, bomBatchSize)
      this.toggleSelection(index, true)
    }
  }

  // create WorkOrder
  public async createWorkOrder(productID: string) {
    this.targetAction = 'create'
    this.dialog = true
    this.tempWorkOrderData = cloneDeep(defaultAddWorkOrderData)
    this.tempWorkOrderData.productID = productID
    this.tempWorkOrderData.planDate = this.dateValue
    this.$nextTick(() => {
      (this.$refs.workOrderDataForm as Form).clearValidate()
    })
    try {
      const { data } = await getProductRecipesList(productID)
      this.stationInfoFormatList = this.recipesDataToStationList(data)
      this.workOrderLoading = false
    } catch {
      this.workOrderLoading = false
    }
    // only one auth select
    if (this.stationInfoFormatList.length === 1) {
      this.tempWorkOrderData.station = this.stationInfoFormatList[0].ID
      this.getRecipesInfo(this.tempWorkOrderData.station)
    }
  }

  private async updateWorkOrder(rowData: any) {
    this.targetAction = 'update'
    this.dialog = true
    this.tempWorkOrderData = cloneDeep(defaultAddWorkOrderData)
    this.tempWorkOrderData.productID = rowData.productID
    this.tempWorkOrderData.planDate = this.dateValue
    this.$nextTick(() => {
      (this.$refs.workOrderDataForm as Form).clearValidate()
    })
    try {
      const { data } = await getProductRecipesList(rowData.productID)
      this.stationInfoFormatList = this.recipesDataToStationList(data)
    } catch {
      this.workOrderLoading = false
    }
    this.getRecipesInfo(rowData.station)
    this.getProcessInfoName(rowData.recipe.ID)
    this.getProcessInfoType(rowData.recipe.processName)
    this.getBatchSizeAndBomInfo(rowData.recipe.ID, rowData.recipe.processName, rowData.recipe.processType)

    this.rowDataToFormat(rowData)
    this.workOrderLoading = false
    this.$nextTick(() => {
      (this.$refs.workOrderDataForm as Form).clearValidate()
    })
  }

  private rowDataToFormat(rowData: any) {
    this.targetWorkOrder = rowData.ID
    rowData.batchCalculation = !(rowData.batchSize === this.BatchSize.FixedQuantity)
    this.tempWorkOrderData = cloneDeep(rowData)
    this.tempWorkOrderData.preBatchSize = this.tempAddWorkInfo.batchSize
    if (rowData.batchSize === this.BatchSize.PerBatchQuantities) {
      this.tempWorkOrderData.batchCount = this.tempWorkOrderData.batchesQuantity.length
      this.tempWorkOrderData.quantity = new Decimal(this.tempWorkOrderData.preBatchSize).mul(new Decimal(this.tempWorkOrderData.batchCount)).toNumber()
    } else {
      this.tempWorkOrderData.quantity = rowData.planQuantity
    }
  }

  private createWorkOrderToDB() {
    (this.$refs.workOrderDataForm as Form).validate(async valid => {
      // check data is not undefined
      let formatFormatData = false
      let toDbFormatData: any = []
      if (this.multipleSelection.length !== 0) {
        const toDB: any[] = []
        this.multipleSelection.forEach(element => {
          element.batchSize = BatchSize.PlanQuantity
          element.batchCount = element.batchesQuantity.length
          toDB.push(element)
        })
        toDbFormatData = this.returnWorkOrderAPiFormat(toDB, false)
        toDbFormatData.forEach((element: any) => {
          if (formatFormatData === false) {
            if (element.recipe.processName === undefined || element.recipe.processName === '' ||
              element.recipe.processType === undefined || element.recipe.processType === '' ||
              element.recipe.ID === undefined || element.recipe.ID === '' ||
              element.station === undefined || element.station === '' ||
              element.planDate === undefined) {
              formatFormatData = true
              this.$notify({
                title: (this.$t('share.errorMessage')).toString(),
                message: (this.$t('message.notify003')).toString(),
                type: 'warning',
                duration: 2000
              })
            }
          }
        })
      }

      if (formatFormatData === false) {
        let parentWorkOrderID = ''
        if (valid) {
          await new Promise(async (resolve) => {
            this.tempWorkOrderData.batchSize = (this.tempWorkOrderData.batchCalculation === false) ? this.BatchSize.FixedQuantity.toString() : this.BatchSize.PlanQuantity.toString()
            const { data } = await createWorkOrder(this.returnWorkOrderAPiFormat([this.tempWorkOrderData], true))
            parentWorkOrderID = data[0]
            resolve('patent success')
          })
          await new Promise(async (resolve) => {
            if (toDbFormatData.length !== 0) {
              toDbFormatData.forEach((element: any) => {
                element.parentID = parentWorkOrderID
              })
              await createWorkOrder(toDbFormatData)
              resolve('children success')
            } else { resolve('children is null') }
          })
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: this.$t('share.addSuccessfully').toString(),
            type: 'success',
            duration: 2000
          })
          this.dialog = false
          this.$emit('onGetPlanList')
        }
      }
    })
  }

  private updateWorkOrderToDB() {
    (this.$refs.workOrderDataForm as Form).validate(async valid => {
      if (valid) {
        this.tempWorkOrderData.batchSize = (this.tempWorkOrderData.batchCalculation === false) ? this.BatchSize.FixedQuantity.toString() : this.BatchSize.PlanQuantity.toString()
        const data = await updateWorkOrder(this.targetWorkOrder, this.returnWorkOrderAPiFormat([this.tempWorkOrderData], false)[0])
        if ((await data).status === 200) {
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: this.$t('share.updateSuccessfully').toString(),
            type: 'success',
            duration: 2000
          })
        }
        this.dialog = false
        this.$emit('onGetStationScheduleList')
      }
    })
  }

  private handleSelectionChange(val: any) {
    this.multipleSelection = val
  }

  private returnWorkOrderAPiFormat(data: any, parent: boolean) {
    if (parent) {
      return data.map((el: any) => {
        return {
          departmentOID: this.departmentOIDValue,
          recipe: {
            processOID: el.recipe.processOID,
            processName: el.recipe.processName,
            processType: el.recipe.processType,
            ID: el.recipe.ID
          },
          station: el.station,
          batchSize: parseInt(el.batchSize),
          batchCount: el.batchCount,
          planQuantity: el.quantity.toString(),
          planDate: moment(el.planDate).format('YYYY-MM-DD')
        }
      })
    } else {
      return data.map((el: any) => {
        return {
          departmentOID: this.departmentOIDValue,
          recipe: {
            processOID: el.processOID,
            processName: el.processName,
            processType: el.processType,
            ID: el.recipeID
          },
          station: el.station,
          batchSize: parseInt(el.batchSize),
          batchCount: el.batchCount,
          planQuantity: el.quantity.toString(),
          planDate: moment(el.planDate).format('YYYY-MM-DD')
        }
      })
    }
  }
}
