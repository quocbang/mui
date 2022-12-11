/* eslint-disable no-async-promise-executor */
import { Component, Vue, Watch } from 'vue-property-decorator'
import { Route } from 'vue-router'
import { Dictionary } from 'vue-router/types/router'
import { Form as ElForm, Input } from 'element-ui'
import { UserModule } from '@/store/modules/user'
import LangSelect from '@/components/LangSelect/index.vue'
import moment from 'moment'
import { GetDate } from '@/utils'
@Component({
  name: 'Login',
  components: {
    LangSelect
  }
})

export default class extends Vue {
  private loginForm = {
    ID: '',
    password: '',
    loginType: 0, // # MES=0 #WINDOWS/AD=1 #PDA=2
    group: 1,
    workDate: moment(GetDate(0)).format('YYYY-MM-DD')
  }

  private passwordType = 'password'
  private loading = false
  private redirect?: string
  private otherQuery: Dictionary<string> = {}

  @Watch('$route', { immediate: true })
  private onRouteChange(route: Route) {
    // TODO: remove the "as Dictionary<string>" hack after v4 release for vue-router
    // See https://github.com/vuejs/vue-router/pull/2050 for details
    const query = route.query as Dictionary<string>
    if (query) {
      this.redirect = query.redirect
      this.otherQuery = this.getOtherQuery(query)
    }
  }

  mounted() {
    if (this.loginForm.ID === '') {
      (this.$refs.ID as Input).focus()
    } else if (this.loginForm.password === '') {
      (this.$refs.password as Input).focus()
    }
    const screenWidth = screen.width
    let setWidth = 360
    if (window.config.PDAWidth !== undefined) {
      setWidth = window.config.PDAWidth
    }
    if (screenWidth <= setWidth) {
      this.loginForm.loginType = 2
    }
  }

  private showPwd() {
    if (this.passwordType === 'password') {
      this.passwordType = ''
    } else {
      this.passwordType = 'password'
    }
    this.$nextTick(() => {
      (this.$refs.password as Input).focus()
    })
  }

  private focusCursor() {
    (this.$refs.ID as Input).focus()
  }

  private handleLogin() {
    (this.$refs.loginForm as ElForm).validate(async (valid: boolean) => {
      if (valid) {
        this.loading = true
        try {
          let checkTag = false
          checkTag = this.checkLogInData()
          if (checkTag === true) {
            await UserModule.Login(this.loginForm)
            // eslint-disable-next-line
            this.$router.push({ path: '/' }, () => { })
            setTimeout(() => {
              this.loading = false
            }, 0.5 * 1000)
          }
        } catch (e) {
          console.log(e)
          this.loading = false
        }
      } else {
        return false
      }
    })
    this.loading = false
  }

  private getOtherQuery(query: Dictionary<string>) {
    return Object.keys(query).reduce((acc, cur) => {
      if (cur !== 'redirect') {
        acc[cur] = query[cur]
      }
      return acc
    }, {} as Dictionary<string>)
  }

  private checkLogInData() {
    if (this.loginForm.loginType !== 2) {
      this.loginForm.group = 0
      this.loginForm.workDate = ''
    } else {
      if (this.loginForm.workDate === '') {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: (this.$t('message.notify016')).toString(),
          type: 'warning',
          duration: 2000
        })
        return false
      }
    }
    return true
  }
}
