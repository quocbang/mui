import { Component, Vue } from 'vue-property-decorator'
import { UserModule } from '@/store/modules/user'

@Component({
  name: 'Dashboard'
})
export default class extends Vue {
  private Type = 0

  private loginType() {
    return UserModule.loginType
  }

  created() {
    this.Type = Number(this.loginType())
  }
}
