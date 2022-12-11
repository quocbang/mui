/* eslint-disable no-async-promise-executor */
import { Component, Vue } from 'vue-property-decorator'
import Sortable from 'sortablejs'
import { defaultStationScheduleListData, getDepartmentStationList } from '@/api/station'
import { getStationScheduleList, updateWorkOrderSequence } from '@/api/workOrder'
import { preMaterialResourceBarcode } from '@/api/resource'
import { BatchSize, GetDate } from '@/utils'
import moment from 'moment'
import { getAllDepartment } from '@/api/unspecified'
import Decimal from 'decimal.js'
import WorkOrderDialog from '@/components/workOrderDialog/workOrderDialog.vue'

const statusTable = [
  { key: '0', status: 'status.pending' },
  { key: '1', status: 'status.active' },
  { key: '2', status: 'status.closing' },
  { key: '3', status: 'status.closed' },
  { key: '4', status: 'status.skipped' }
]
@Component({
  name: 'stationSchedule',
  components: {
    WorkOrderDialog
  }
})

export default class extends Vue {
  // Query Data
  private departmentOIDValue = ''
  private departmentInfoList: any[] = []
  private dateValue = moment(GetDate(0)).format('YYYY-MM-DD')
  private stationValue = ''
  private stationInfoList: any[] = []
  private isPlanAlive = false
  private sortable: Sortable | null = null
  private multipleSelection: any[] = []
  private scheduleList: any[] = []
  private oldList: any[] = []
  private newList: any[] = []
  private BatchSize = BatchSize

  // dialog Visible
  private dialogDetailFormVisible = false
  private tempStationScheduleListData = defaultStationScheduleListData
  private listLoading = false

  created() {
    this.getAllDepartment()
  }

  private async onDepartmentGetStationList(DepartmentOID: string) {
    try {
      const { data } = await getDepartmentStationList(DepartmentOID)
      this.stationInfoList = data
      this.stationValue = ''
      // only one auth select
      if (this.stationInfoList.length === 1) {
        this.stationValue = this.stationInfoList[0].ID
        this.onGetStationScheduleList(this.stationValue, this.dateValue)
      }
    } catch (e) {
      console.log(e)
    }
  }

  private async getAllDepartment() {
    const { data } = await getAllDepartment()
    this.departmentInfoList = data
    // only one auth select
    if (this.departmentInfoList.length === 1) {
      this.departmentOIDValue = this.departmentInfoList[0].departmentID
      this.onDepartmentGetStationList(this.departmentOIDValue)
    }
  }

  private async onGetStationScheduleList(stationValue: string, date: string) {
    try {
      this.scheduleList = []
      this.listLoading = true
      const { data } = await getStationScheduleList(stationValue, moment(date).format('YYYY-MM-DD'))
      this.scheduleList = data
      this.scheduleList.forEach((el, index) => {
        if (parseInt(el.batchSize) > Object.keys(this.BatchSize).length) {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: (this.$t('message.notifyWorkOrder100')).toString(),
            type: 'warning',
            duration: 2000
          })
        }
        if (el.batchSize === this.BatchSize.PerBatchQuantities) {
          el.batchesQuantity.forEach((item: string) => {
            this.$set(this.scheduleList[index], 'planQuantity', this.scheduleList[index].planQuantity === undefined ? new Decimal(item) : this.scheduleList[index].planQuantity.add(new Decimal(item)))
          })
        }
        statusTable.forEach((element) => {
          if (element.key.indexOf(el.status.toString()) !== -1) {
            this.$set(this.scheduleList[index], 'statusName', element.status)
          }
        })
      })
      this.scheduleList = this.scheduleList.filter(item => item.statusName !== 'closed' && item.statusName !== 'skipped')
      this.listLoading = false
      this.isPlanAlive = true
    } catch {
      this.listLoading = false
    }
    this.oldList = this.scheduleList.map((v) => {
      return {
        ID: v.ID,
        forceToAbort: false,
        sequence: v.sequence
      }
    })

    this.newList = this.oldList.slice()
    this.$nextTick(() => {
      this.setSort()
    })
  }

  private openDetailInfo(row: any) {
    this.tempStationScheduleListData = row
    this.dialogDetailFormVisible = true
  }

  private setSort() {
    const el = (this.$refs.draggableTable as Vue).$el.querySelectorAll('.el-table__body-wrapper > table > tbody')[0] as HTMLElement
    this.sortable = Sortable.create(el, {
      ghostClass: 'sortable-ghost', // Class name for the drop placeholder
      onEnd: evt => {
        if (typeof (evt.oldIndex) !== 'undefined' && typeof (evt.newIndex) !== 'undefined') {
          const targetRow = this.scheduleList.splice(evt.oldIndex, 1)[0]
          this.scheduleList.splice(evt.oldIndex, 0, targetRow)
          // for show the changes, you can delete in you code
          const tempIndex = this.newList.splice(evt.oldIndex, 1)[0]
          this.newList.splice(evt.newIndex, 0, tempIndex)
        }
      }
    })
  }

  private async updateWorkOrder(rowData: any) {
    const refWorkOrder: any = this.$refs.refWorkOrder
    refWorkOrder.updateWorkOrder(rowData)
  }

  private handleSelectionChange(val: any) {
    this.multipleSelection = val
  }

  private async stopWorkOrder() {
    if (this.multipleSelection.length !== 0) {
      const stopWorkOrderList = this.multipleSelection.map((v) => {
        return {
          ID: v.ID,
          forceToAbort: true,
          sequence: v.sequence
        }
      })
      try {
        const data = await updateWorkOrderSequence(stopWorkOrderList)
        if (data.status === 200) {
          await this.onGetStationScheduleList(this.stationValue, this.dateValue)
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: (this.$t('share.updateSuccessfully')).toString(),
            type: 'success',
            duration: 2000
          })
        }
      } catch (e: any) {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: (this.$t('share.fail')).toString() + ': ' + e.errorMessage,
          type: 'warning',
          duration: 2000
        })
      }
    }
  }

  private async confirmOrder() {
    const resultList = this.newList.map((v, i) => {
      return {
        ID: v.ID,
        forceToAbort: false,
        sequence: i + 1
      }
    })
    try {
      const data = await updateWorkOrderSequence(resultList)
      if (data.status === 200) {
        await this.onGetStationScheduleList(this.stationValue, this.dateValue)
        this.$notify({
          title: (this.$t('share.success')).toString(),
          message: (this.$t('share.updateSuccessfully')).toString(),
          type: 'success',
          duration: 2000
        })
      }
    } catch (e: any) {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: (this.$t('share.fail')).toString() + ': ' + e.errorMessage,
        type: 'warning',
        duration: 2000
      })
    }
  }

  private async printBarcode(resourceID: string) {
    try {
      const printParams = {
        fieldName: {
          Station: this.$t('labelFields.materialResource.stationID'),
          NextStation: this.$t('labelFields.materialResource.nextStationID'),
          ProductID: this.$t('labelFields.materialResource.productID'),
          ProductionDate: this.$t('labelFields.materialResource.productionDate'),
          ExpiryDate: this.$t('labelFields.materialResource.expiryDate'),
          Quantity: this.$t('labelFields.materialResource.quantity'),
          ResourceID: this.$t('labelFields.materialResource.resourceID')
        }
      }
      const data = await preMaterialResourceBarcode(resourceID, printParams)
      const file = new Blob([data.data], { type: 'application/pdf' })
      const fileURL = URL.createObjectURL(file)
      window.open(fileURL)
    } catch (e) {
      console.log(e)
    }
  }

  private showPDFView(rowData: any) {
    this.printBarcode(rowData.ID)
  }
}
