import { changePassword } from '@/api/account'
import { clearObjValue, validatePasswordNumberRange, validateRequire } from '@/utils'
import { Form, Input } from 'element-ui'
import { Component, Vue } from 'vue-property-decorator'
@Component({
  name: 'editPassword'
})

export default class extends Vue {
  private editPasswordForm = {
    oldPassword: '',
    newPassword: '',
    checkNewPassword: ''
  }

  private passwordType = 'password'

  // Form Rules
  private rule = {
    oldPassword: [{ validator: validateRequire }, { validator: validatePasswordNumberRange }],
    newPassword: [{ validator: validateRequire }, { validator: validatePasswordNumberRange }],
    checkNewPassword: [{ validator: validateRequire }, { validator: validatePasswordNumberRange }]
  }

  private onClickConfirm() {
    (this.$refs.dataForm as Form).validate(async (valid: any) => {
      if (valid) {
        if (this.editPasswordForm.newPassword !== this.editPasswordForm.checkNewPassword) {
          this.$notify({
            title: (this.$t('share.errorMessage')).toString(),
            message: (this.$t('message.notify112')).toString(),
            type: 'warning',
            duration: 2000
          })
        } else {
          const newDate = {
            currentPassword: this.editPasswordForm.oldPassword,
            newPassword: this.editPasswordForm.newPassword
          }
          const data = await changePassword(newDate)
          if (data.status === 200) {
            clearObjValue(this.editPasswordForm)
            this.$nextTick(() => {
              (this.$refs.dataForm as Form).clearValidate()
            })
            this.$notify({
              title: (this.$t('share.success')).toString(),
              message: this.$t('share.updateSuccessfully').toString(),
              type: 'success',
              duration: 2000
            })
          }
        }
      }
    })
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
}
