import { Component, Vue } from 'vue-property-decorator'
import moment from 'moment'
import { getAllDepartment } from '@/api/unspecified'
import { getProductionRate } from '@/api/workOrder'
import XLSX from 'xlsx'
import Pagination from '@/components/Pagination/index.vue'

@Component({
  name: 'ProductionRate',
  components: {
    Pagination
  }
})

export default class extends Vue {
  // Query Data
  private departmentInfoList: any[] = []
  private dateValue= ['', '']
  private departmentID: any = []
  private data: any[] = []
  private tableKey = 0
  private productionRateQuery: any = {}
  private total = 0
  private listQuery = {
    page: 1,
    limit: 10
  }

  created() {
    this.getAllDepartment()
  }

  private async getAllDepartment() {
    const { data } = await getAllDepartment()
    this.departmentInfoList = data
  }

  private async getProductionRate() {
    if (this.departmentID === '' || this.dateValue[0] === '' || this.dateValue[1] === '') {
      this.$notify({
        title: (this.$t('share.errorMessage')).toString(),
        message: (this.$t('message.notify003')).toString(),
        type: 'warning',
        duration: 2000
      })
      return
    }
    const query = {
      workStartDate: moment(this.dateValue[0]).format('YYYY-MM-DD'),
      workEndDate: moment(this.dateValue[1]).format('YYYY-MM-DD')
    }
    const { data } = await getProductionRate(this.departmentID, query)
    if (data.length === 0) {
      this.$notify({
        title: (this.$t('share.status')).toString(),
        message: (this.$t('errorCodes.errorCode_404')).toString(),
        type: 'warning',
        duration: 2000
      })
      return
    }
    const fileName = this.departmentID + '-' + moment(this.dateValue[0]).format('YYYYMMDD') + '-' + moment(this.dateValue[1]).format('YYYYMMDD') + '.xlsx'
    const wsName = 'Sheet1'
    const wb = XLSX.utils.book_new()
    const ws = XLSX.utils.json_to_sheet(data.items)
    ws['!autofilter'] = { ref: 'A1:L1' }
    XLSX.utils.sheet_add_aoa(ws, [[this.$t('system.departmentID'), this.$t('workOrder.ID'), this.$t('system.productID'), this.$t('system.stationID'), this.$t('system.quantity'), this.$t('system.currentQuantity'), this.$t('system.ratio'), this.$t('system.productionTime'), this.$t('system.productionEndTime'), this.$t('system.updateBy'), this.$t('system.createdBy'), this.$t('recipe.ID')]], { origin: 'A1' })
    ws['!cols'] = [{ wch: 16 }, { wch: 29 }, { wch: 18 }, { wch: 21 }, { wch: 10 }, { wch: 10 }, { wch: 13 }, { wch: 15 }, { wch: 15 }, { wch: 13 }, { wch: 13 }, { wch: 22 }]
    XLSX.utils.book_append_sheet(wb, ws, wsName)
    XLSX.writeFile(wb, fileName)
  }

  private async onListWorkOrderRate() {
    try {
      this.productionRateQuery = {}
      this.productionRateQuery.page = this.listQuery.page
      this.productionRateQuery.limit = this.listQuery.limit
      this.productionRateQuery.workStartDate = moment(this.dateValue[0]).format('YYYY-MM-DD')
      this.productionRateQuery.workEndDate = moment(this.dateValue[1]).format('YYYY-MM-DD')
      const { data } = await getProductionRate(this.departmentID, this.productionRateQuery)
      this.data = data.items
      this.total = data.total
    } catch (e) {
      console.log(e)
    }
  }
}
