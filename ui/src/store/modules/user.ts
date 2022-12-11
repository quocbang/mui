import { VuexModule, Module, Action, Mutation, getModule } from 'vuex-module-decorators'
import { login, logout } from '@/api/account'
import { getToken, setToken, removeToken, getRoles, setUserInfo, removeUserInfo, getAuthorizedDepartments, getLoginType, getGroup, getWorkDate, getStation, setStation, getSelectStationsInfo, setSelectStationsInfo, removeSelectStationsInfo, getFeedAndCollectMode, setFeedAndCollectMode, removeFeedAndCollectMode } from '@/utils/cookies'
import store from '@/store'
import { PDAModule } from '@/store/modules/pda'
import { stationSignOut } from '@/api/station'

export interface IUserState {
  token: string
  tokenExpiry: string
  name: string
  authorizedDepartments: string // array to jsonData
  avatar: string
  introduction: string
  roles: number[]
  loginType: string
  groups: string
  workDate: string
  selectStationsInfo: string[]
  feedAndCollectMode: string

}

@Module({ dynamic: true, store, name: 'user' })
class User extends VuexModule implements IUserState {
  public token = getToken() || ''
  public tokenExpiry = ''
  public name = ''
  public authorizedDepartments = getAuthorizedDepartments() || ''
  public avatar = ''
  public introduction = ''
  public roles = getRoles() || []
  public loginType = getLoginType() || ''
  public groups = getGroup() || ''
  public workDate = getWorkDate() || ''
  public station = getStation() || ''
  public selectStationsInfo = getSelectStationsInfo() || []
  public feedAndCollectMode = getFeedAndCollectMode() || ''

  @Mutation
  private SET_TOKEN(token: string) {
    this.token = token
  }

  @Mutation
  private SET_TOKEN_EXPIRY(tokenExpiry: string) {
    this.tokenExpiry = tokenExpiry
  }

  @Mutation
  private SET_NAME(name: string) {
    this.name = name
  }

  @Mutation
  private SET_AUTHORIZED_DEPARTMENTS(authorizedDepartments: string) {
    this.authorizedDepartments = authorizedDepartments
  }

  @Mutation
  private SET_AVATAR(avatar: string) {
    this.avatar = avatar
  }

  @Mutation
  private SET_ROLES(roles: number[]) {
    this.roles = roles
  }

  @Mutation
  private SET_LOGIN_TYPE(type: string) {
    this.loginType = type
  }

  @Mutation
  private SET_GROUP(groups: string) {
    this.groups = groups
  }

  @Mutation
  private SET_LOGIN_WORKDATE(workDate: string) {
    this.workDate = workDate
  }

  @Mutation
  private SET_STATION(station: string) {
    this.station = station
  }

  @Mutation
  private SET_FEED_AND_COLLECT_MODE(feedAndCollectMode: string) {
    this.feedAndCollectMode = feedAndCollectMode
  }

  @Mutation
  private SET_SELECT_STATIONS_INFO(selectStationsInfo: string[]) {
    this.selectStationsInfo = selectStationsInfo
  }

  @Action({ rawError: true })
  public async Login(userInfo: { ID: string, password: string, loginType: number, group: number, workDate: string }) {
    const { ID, password, loginType, group, workDate } = userInfo
    const LoginData = {
      ID: ID,
      password: password,
      loginType: loginType
    }
    if (loginType === 2) {
      LoginData.loginType = 0
    }
    const { data } = await login(LoginData)
    const allAuthorizedDepartments = JSON.stringify(data.authorizedDepartments)
    data.roles = data.roles || []
    setToken(data.token)
    setUserInfo(data.roles, allAuthorizedDepartments, loginType.toString(), group.toString(), workDate)
    this.SET_TOKEN(data.token)
    this.SET_TOKEN_EXPIRY(data.tokenExpiry)
    this.SET_AUTHORIZED_DEPARTMENTS(allAuthorizedDepartments)
    this.SET_ROLES(data.roles)
    this.SET_NAME(ID)
    this.SET_AVATAR('https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif')
    this.SET_LOGIN_TYPE(loginType.toString())
    this.SET_GROUP(group.toString())
    this.SET_LOGIN_WORKDATE(workDate)
  }

  @Action({ rawError: true })
  public async Station(station: string) {
    setStation(station)
    this.SET_STATION(station)
  }

  @Action({ rawError: true })
  public async FeedAndCollectMode(feedAndCollectMode: string) {
    setFeedAndCollectMode(feedAndCollectMode)
    this.SET_FEED_AND_COLLECT_MODE(feedAndCollectMode)
  }

  @Action({ rawError: true })
  public async SelectStation(selectStations: string[]) {
    setSelectStationsInfo(selectStations)
    this.SET_SELECT_STATIONS_INFO(selectStations)
  }

  @Action
  public Reset() {
    removeToken()
    removeUserInfo()
    this.SET_TOKEN('')
    this.SET_ROLES([])
    this.SET_AUTHORIZED_DEPARTMENTS('')
    this.SET_LOGIN_TYPE('')
    this.SET_GROUP('')
    this.SET_LOGIN_WORKDATE('')
    this.SET_FEED_AND_COLLECT_MODE('')
  }

  @Action
  public async LogOut() {
    if (this.token === '') {
      throw Error('LogOut: token is undefined!')
    }
    if (this.selectStationsInfo.length !== 0) {
      const operatorStation = { stationSites: this.selectStationsInfo }
      await stationSignOut(operatorStation)
    }
    removeSelectStationsInfo()
    await logout()
    removeToken()
    removeUserInfo()
    this.SET_TOKEN('')
    this.SET_ROLES([])
    this.SET_AUTHORIZED_DEPARTMENTS('')
    this.SET_LOGIN_TYPE('')
    this.SET_GROUP('')
    this.SET_LOGIN_WORKDATE('')
    removeFeedAndCollectMode()
    PDAModule.StationSiteInfoReset()
    PDAModule.WorkOrderInfoReset()
  }
}

export const UserModule = getModule(User)
