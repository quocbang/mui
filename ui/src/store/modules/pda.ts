import { VuexModule, Module, Action, Mutation, getModule } from 'vuex-module-decorators'
import store from '@/store'
import { getStationSiteInfo, getWorkOrderInfo, removeStationSiteInfo, removeWorkOrderInfo, setStationSiteInfo, setWorkOrderInfo } from '@/utils/cookies'
export interface IPDAState {
  workOrderInfo: any
  stationSiteInfo: any

}

@Module({ dynamic: true, store, name: 'pad' })
class PDA extends VuexModule implements IPDAState {
  public workOrderInfo = getWorkOrderInfo() || {}
  public stationSiteInfo = getStationSiteInfo() || {}

  @Mutation
  private SET_WORKORDER_INFO(workOrderInfo: any) {
    this.workOrderInfo = workOrderInfo
  }

  @Action({ rawError: true })
  public async WorkOrderInfo(workOrderInfo: any) {
    setWorkOrderInfo(workOrderInfo)
    this.SET_WORKORDER_INFO(workOrderInfo)
  }

  @Action
  public WorkOrderInfoReset() {
    removeWorkOrderInfo()
    this.SET_WORKORDER_INFO([])
  }

  @Mutation
  private SET_STATION_SITE_INFO(stationSiteInfo: any) {
    this.stationSiteInfo = stationSiteInfo
  }

  @Action({ rawError: true })
  public async StationSiteInfo(stationSiteInfo: any) {
    setStationSiteInfo(stationSiteInfo)
    this.SET_STATION_SITE_INFO(stationSiteInfo)
  }

  @Action
  public StationSiteInfoReset() {
    removeStationSiteInfo()
    this.SET_WORKORDER_INFO([])
  }
}

export const PDAModule = getModule(PDA)
