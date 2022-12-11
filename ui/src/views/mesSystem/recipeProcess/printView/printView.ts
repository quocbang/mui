import { Component, Vue } from 'vue-property-decorator'
@Component({
  name: 'RecipeProcessPrintView'
})

export default class extends Vue {
  private departmentOIDValue = ''
  private productTypeValue = ''
  private productIDValue = ''
  private recipeValue = ''
  private rowValue = {}
  private stationValue = ''
  private materialDataInfo: any[] = []
  private stationControlStepDataInfo: any[] = []
  private stationControlSliceTable: any[] = []
  private targetData: any

  created() {
    this.targetData = JSON.parse(sessionStorage.getItem('targetData') || '{}')
    this.targetData.row[0].stations = this.targetData.row[0].stations.filter((item: { ID: any }) => item.ID === this.targetData.station)
    this.sliceTable(this.targetData.stationControlStepDataInfo.columns, 5)
  }

  private sliceTable(data: any, num: number) {
    this.stationControlSliceTable = []
    for (let i = 0; i < data.length;) {
      this.stationControlSliceTable.push(data.slice(i, i += num))
    }
  }

  private printContent() {
    window.print()
  }
}
