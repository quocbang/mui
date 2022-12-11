import { getSubTypeList, getTypeList } from '@/api/site'
import { addStationMaintenance, defaultAddStationMaintenanceData, defaultUpdateStationMaintenanceData, deleteStationMaintenance, getStationMaintenance, getStationState, updateStationMaintenance } from '@/api/station'
import i18n from '@/lang'
import { UserModule } from '@/store/modules/user'
import { validateNumberAndUppercase, validateRequire } from '@/utils'
import { SubType, Type } from '@/utils/sites'
import { Form, MessageBox } from 'element-ui'
import _, { cloneDeep } from 'lodash'
import Pagination from '@/components/Pagination/index.vue'
import moment from 'moment'
import { Component, Vue } from 'vue-property-decorator'

@Component({
  name: 'stationMaintenance',
  components: {
    Pagination
  }
})

export default class extends Vue {
  private departmentInfoList: any[] = []
  private departmentOIDValue = ''
  private stationDataList: any[] = []
  private subTypeList: any[] = []
  private typeList: any[] = []
  private stationStatusList: any[] = []
  private isAlive = false
  private tableKey = 0
  private tempDetailData = {}
  private tempSiteDetailData: any[] = []
  private tempCreateWorkOrder = defaultAddStationMaintenanceData
  private tempUpdateStationData = defaultUpdateStationMaintenanceData
  private targetSubType = ''
  private targetType = ''
  private deleteSite: any[] = []
  // dialog Visible
  private dialogCreateVisible = false
  private dialogUpdateVisible = false
  private dialogDetailVisible = false
  private dialogSiteDetailVisible = false
  private dialogSiteColqueueDetailVisible = false
  private listLoading = false
  private total = 0
  private targetSiteData: any
  private stationMaintenanceQuery: any = {}
  private listQuery = {
    page: 1,
    limit: 10
  }

  private Type = Type
  private SubType = SubType

  // Form Rules
  private stationRules = {
    ID: [{ validator: validateRequire }],
    code: [{ validator: validateRequire }]
  }

  private siteNameRules = {
    name: [{ validator: validateNumberAndUppercase }]
  }

  created() {
    this.getUserInfo()
    this.getTypeInfoList()
    this.getSubTypeInfoList()
    this.getStationStatusList()
    // only one auth select
    if (this.departmentInfoList.length === 1) {
      this.departmentOIDValue = this.departmentInfoList[0].label.OID
      this.queryStationData()
    }
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

  private async getSubTypeInfoList() {
    const { data } = await getSubTypeList()
    this.subTypeList = data
  }

  private async getTypeInfoList() {
    const { data } = await getTypeList()
    this.typeList = data
  }

  private async getStationStatusList() {
    const { data } = await getStationState()
    this.stationStatusList = data
  }

  private async queryStationData() {
    try {
      this.stationMaintenanceQuery = {}
      this.stationMaintenanceQuery.page = this.listQuery.page
      this.stationMaintenanceQuery.limit = this.listQuery.limit
      this.listLoading = true
      let actionModeError = false
      this.isAlive = true
      const { data } = await getStationMaintenance(this.departmentOIDValue, this.stationMaintenanceQuery)
      this.stationDataList = data.items
      this.total = data.total
      this.stationDataList.forEach((element: any, elementIndex: number) => {
        if ((element.sites.filter((item: { actionMode: number }) => item.actionMode !== 0)).length !== 0) {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: this.$t('message.notify008').toString(),
            type: 'warning',
            duration: 2000
          })
          this.stationDataList = []
          actionModeError = true
        }
        if (actionModeError === false) {
          this.stationDataList[elementIndex].updateAt = moment(this.stationDataList[elementIndex].updateAt).format('yyyy-MM-DD HH:mm:ss')
          this.stationDataList[elementIndex].stateName = this.stationStatusList.filter(item => item.ID === element.state)[0].name
          element.sites.forEach((el: any, index: number) => {
            this.stationDataList[elementIndex].sites[index].subTypeName = this.subTypeList.filter((item: { ID: any }) => item.ID === el.subType)[0].name
            this.stationDataList[elementIndex].sites[index].typeName = this.typeList.filter((item: { ID: any }) => item.ID === el.type)[0].name
          })
        }
        this.listLoading = false
      })
    } catch {
      this.listLoading = false
    }
  }

  private createStation() {
    this.dialogCreateVisible = true
    this.tempCreateWorkOrder = cloneDeep(defaultAddStationMaintenanceData)
    this.tempCreateWorkOrder.sites.splice(0, 1) // delete default
    this.$nextTick(() => {
      (this.$refs.stationDataForm as Form).clearValidate()
    })
  }

  private async createStationInfoToDB() {
    // Combination site array
    const arrCombination: any = []
    let empty = false
    this.tempCreateWorkOrder.sites.forEach(element => {
      arrCombination.push(element.name + '_' + element.index)
      if (element.name === '' || element.index === undefined || element.subType === undefined) {
        empty = true
      }
    })
    // Judgment repetition
    const repeat = arrCombination.filter((element: any, index: any, arr: string | any[]) => {
      return arr.indexOf(element) !== index
    })
    if (empty === false) {
      if (repeat.length > 0) {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: this.$t('message.notify400').toString(),
          type: 'warning',
          duration: 2000
        })
      } else {
        (this.$refs.stationDataForm as Form).validate(async valid => {
          if (valid) {
            this.tempCreateWorkOrder.departmentOID = this.departmentOIDValue
            const data = addStationMaintenance(this.tempCreateWorkOrder)
            if ((await data).status === 200) {
              this.$notify({
                title: (this.$t('share.success')).toString(),
                message: this.$t('share.addSuccessfully').toString(),
                type: 'success',
                duration: 2000
              })
              this.dialogCreateVisible = false
              this.queryStationData()
            }
          }
        })
      }
    } else {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notify003').toString(),
        type: 'warning',
        duration: 2000
      })
    }
  }

  private updateStation(rowData: any, index: number) {
    this.dialogUpdateVisible = true
    this.stationDataList[index].sites.forEach((element: any, elementIndex: number) => {
      this.stationDataList[index].sites[elementIndex].exist = true
    })
    this.tempUpdateStationData = cloneDeep(this.stationDataList[index])
    this.targetSiteData = this.stationDataList[index].sites
  }

  private async updateStationInfoToDB() {
    // Combination site array
    const arrCombination: any = []
    let empty = false
    this.tempUpdateStationData.sites.forEach(element => {
      arrCombination.push(element.name + '_' + element.index)
      if (element.name === '' || element.index === undefined || element.subType === undefined) {
        empty = true
      }
    })
    // Judgment repetition
    const repeat = arrCombination.filter((element: any, index: any, arr: string | any[]) => {
      return arr.indexOf(element) !== index
    })
    if (empty === false) {
      if (repeat.length > 0) {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: this.$t('message.notify012').toString(),
          type: 'warning',
          duration: 2000
        })
      } else {
        (this.$refs.stationDataForm as Form).validate(async valid => {
          if (valid) {
            const tempData: any = cloneDeep(this.tempUpdateStationData)
            const updateDataPick = _.pick(tempData, ['code'], ['description'], ['state'], ['sites'])
            updateDataPick.sites.forEach((element: any, index: number) => {
              updateDataPick.sites[index] = _.pick(element, ['actionMode'], ['name'], ['index'], ['type'], ['subType'])
            })

            // remove origin site
            const finalDate: any = cloneDeep(updateDataPick)
            const resultSites = finalDate.sites.filter((item: { actionMode: number }) => item.actionMode !== 0)
            finalDate.sites = resultSites
            const data = updateStationMaintenance(tempData.ID, finalDate)
            if ((await data).status === 200) {
              this.$notify({
                title: (this.$t('share.success')).toString(),
                message: this.$t('share.updateSuccessfully').toString(),
                type: 'success',
                duration: 2000
              })
              this.dialogUpdateVisible = false
              this.queryStationData()
            }
          }
        })
      }
    } else {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: this.$t('message.notify003').toString(),
        type: 'warning',
        duration: 2000
      })
    }
  }

  private async deleteStation(rowData: any) {
    MessageBox.confirm(
      i18n.t('message.notifyDelete').toString(),
      i18n.t('share.prompt').toString(),
      {
        confirmButtonText: i18n.t('share.confirm').toString(),
        cancelButtonText: i18n.t('share.cancel').toString(),
        type: 'warning'
      }
    ).then(async () => {
      const data = deleteStationMaintenance(rowData.ID)
      if ((await data).status === 200) {
        this.$notify({
          title: (this.$t('share.success')).toString(),
          message: this.$t('share.updateSuccessfully').toString(),
          type: 'success',
          duration: 2000
        })
        this.queryStationData()
      }
    }).catch(e => e)
  }

  private addSitesInfo(uiType: string) {
    const obj: any = {}
    obj.actionMode = 1
    obj.name = ''
    obj.index = 0
    obj.type = 1
    if (uiType === 'add') {
      this.tempCreateWorkOrder.sites.push(obj)
    } else {
      this.tempUpdateStationData.sites.push(obj)
    }
  }

  private deleteSitesInfo(rowData: any, index: number, uiType: string) {
    MessageBox.confirm(
      i18n.t('message.notifyDelete').toString(),
      i18n.t('share.prompt').toString(),
      {
        confirmButtonText: i18n.t('share.confirm').toString(),
        cancelButtonText: i18n.t('share.cancel').toString(),
        type: 'warning'
      }
    ).then(async () => {
      if (uiType === 'add') {
        this.tempCreateWorkOrder.sites.splice(index, 1)
      } else {
        this.tempUpdateStationData.sites.splice(index, 1)
        rowData.actionMode = 2
        const newRowData = _.pick(rowData, ['actionMode'], ['name'], ['index'], ['type'], ['subType'])
        const tempData: any = cloneDeep(this.tempUpdateStationData)
        const updateDataPick = _.pick(tempData, ['code'], ['description'], ['state'], ['sites'])
        updateDataPick.sites = [newRowData]
        if (this.targetSiteData.filter((item: { name: any, index: any }) => item.name === newRowData.name && item.index === newRowData.index).length > 0) {
          const data = updateStationMaintenance(tempData.ID, updateDataPick)
          if ((await data).status === 200) {
            this.$notify({
              title: (this.$t('share.success')).toString(),
              message: this.$t('share.updateSuccessfully').toString(),
              type: 'success',
              duration: 2000
            })
          }
          this.queryStationData()
        } else {
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: this.$t('share.updateSuccessfully').toString(),
            type: 'success',
            duration: 2000
          })
        }
      }
    }).catch(e => e)
  }

  private openDetailInfo(rowData: any) {
    this.dialogDetailVisible = true
    this.tempDetailData = rowData.sites
  }

  private openSiteDetailInfo(rowData: any) {
    this.tempSiteDetailData = []
    this.targetSubType = rowData.subTypeName
    this.targetType = rowData.typeName
    if (this.targetType !== 'COLQUEUE') {
      this.dialogSiteDetailVisible = true
    } else {
      this.dialogSiteColqueueDetailVisible = true
    }
    switch (this.targetType) {
      case 'CONTAINER':
        if (rowData.content.container !== undefined) {
          switch (this.targetSubType) {
            case 'OPERATOR':
              this.tempSiteDetailData = rowData.content.container.map((item: { operatorSite: any }) => { return item.operatorSite })
              break
            case 'MATERIAL':
              this.tempSiteDetailData = rowData.content.container.map((item: { materialSite: any }) => { return item.materialSite })
              break
            case 'TOOL':
              this.tempSiteDetailData = rowData.content.container.map((item: { toolSite: any }) => { return item.toolSite })
              break
          }
        }
        break
      case 'COLLECTION':
        if (rowData.content.collection !== undefined) {
          switch (this.targetSubType) {
            case 'OPERATOR':
              this.tempSiteDetailData = rowData.content.collection.map((item: { operatorSite: any }) => { return item.operatorSite })
              break
            case 'MATERIAL':
              this.tempSiteDetailData = rowData.content.collection.map((item: { materialSite: any }) => { return item.materialSite })
              break
            case 'TOOL':
              this.tempSiteDetailData = rowData.content.collection.map((item: { toolSite: any }) => { return item.toolSite })
              break
          }
        }
        break
      case 'QUEUE':
        if (rowData.content.queue !== undefined) {
          switch (this.targetSubType) {
            case 'OPERATOR':
              this.tempSiteDetailData = rowData.content.queue.map((item: { operatorSite: any }) => { return item.operatorSite })
              break
            case 'MATERIAL':
              this.tempSiteDetailData = rowData.content.queue.map((item: { materialSite: any }) => { return item.materialSite })
              break
            case 'TOOL':
              this.tempSiteDetailData = rowData.content.queue.map((item: { toolSite: any }) => { return item.toolSite })
              break
          }
        }
        break
      case 'SLOT':
        if (rowData.content.slot !== undefined) {
          switch (this.targetSubType) {
            case 'OPERATOR':
              this.tempSiteDetailData.push(rowData.content.slot.operatorSite)
              break
            case 'MATERIAL':
              this.tempSiteDetailData.push(rowData.content.slot.materialSite)
              break
            case 'TOOL':
              this.tempSiteDetailData.push(rowData.content.slot.toolSite)
              break
          }
        }
        break
      case 'COLQUEUE':
        if (rowData.content.colqueue !== undefined) {
          switch (this.targetSubType) {
            case 'OPERATOR':
              rowData.content.colqueue.forEach((item: any) => {
                const subTypeItem = item.map((el: any) => {
                  return el.operatorSite
                })
                this.tempSiteDetailData.push(subTypeItem)
              })
              break
            case 'MATERIAL':
              rowData.content.colqueue.forEach((item: any) => {
                const subTypeItem = item.map((el: any) => {
                  return el.materialSite
                })
                this.tempSiteDetailData.push(subTypeItem)
              })
              break
            case 'TOOL':
              rowData.content.colqueue.forEach((item: any) => {
                const subTypeItem = item.map((el: any) => {
                  return el.toolSite
                })
                this.tempSiteDetailData.push(subTypeItem)
              })
              break
          }
        }
        break
    }
  }

  private subTypeRules(name: string, index: number) {
    if (name === 'create') {
      if (this.tempCreateWorkOrder.sites[index].subType === SubType.OPERATOR) {
        this.tempCreateWorkOrder.sites[index].type = Type.SLOT
        this.tempCreateWorkOrder.sites[index].index = 0
      } else if (this.tempCreateWorkOrder.sites[index].subType === SubType.TOOL) {
        this.tempCreateWorkOrder.sites[index].type = Type.SLOT
      }
    } else {
      if (this.tempUpdateStationData.sites[index].subType === SubType.OPERATOR) {
        this.tempUpdateStationData.sites[index].type = Type.SLOT
        this.tempUpdateStationData.sites[index].index = 0
      } else if (this.tempUpdateStationData.sites[index].subType === SubType.TOOL) {
        this.tempUpdateStationData.sites[index].type = Type.SLOT
      }
    }
  }
}
