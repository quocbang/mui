
import { addAccount, defaultAddAccount, defaultUpdateAccount, deleteAccount, getAccountList, getAccountUnauthorizedList, getRoleList, updateAccount } from '@/api/account'
import i18n from '@/lang'
import { getUserDepartmentsInfo, validateRequire } from '@/utils'
import { Form, MessageBox } from 'element-ui'
import { cloneDeep } from 'lodash'
import { Component, Vue } from 'vue-property-decorator'
@Component({
  name: 'accountManagement'
})

export default class extends Vue {
  // Top condition query value
  private departmentOIDValue = ''
  private departmentInfoList: any[] = []
  private accountTable: any[] = []
  private tableKey = 0
  private roleList: any[] = []
  private tempAddData = defaultAddAccount
  private dialogVisible = false
  private dialogStatus = ''
  private userInfoList: any[] = []
  private isPlanAlive = false
  private listLoading = false
  private rules = {
    employeeID: [{ validator: validateRequire }],
    roles: [{ validator: validateRequire }]
  }

  private async getRoleList() {
    let auth: any[] = []
    const { data } = await getRoleList()
    auth = data
    auth.forEach((element, index) => {
      if (element.name === 'ADMINISTRATOR' || element.name === 'LEADER') {
        auth[index].disabled = true
      }
    })
    this.roleList = auth
  }

  private async onGetUserAuthInfo() {
    const { data } = await getAccountUnauthorizedList(this.departmentOIDValue)
    this.userInfoList = data
    // only one auth select
    if (this.userInfoList.length === 1) {
      this.tempAddData.employeeID = this.userInfoList[0].employeeID
    }
  }

  created() {
    this.departmentInfoList = getUserDepartmentsInfo()
    // only one auth select
    if (this.departmentInfoList.length === 1) {
      this.departmentOIDValue = this.departmentInfoList[0].label.OID
      this.getAccountTable()
    }
    this.getRoleList()
  }

  private async getAccountTable() {
    let tempAccountTable: any[] = []
    try {
      this.listLoading = true
      const { data } = await getAccountList(this.departmentOIDValue)
      tempAccountTable = data
      tempAccountTable.forEach((el: any, index: number) => {
        const tempRolesName: any[] = []
        this.$set(tempAccountTable[index], 'rolesName', [])
        this.roleList.forEach((elRoles: any) => {
          if (el.roles.indexOf(elRoles.ID) !== -1) {
            tempRolesName.push(elRoles.name)
          }
        })
        tempAccountTable[index].rolesName = tempRolesName
      })
      this.accountTable = tempAccountTable
      setTimeout(() => {
        this.listLoading = false
        this.isPlanAlive = true
      }, 0.5 * 1000)
    } catch {
      this.listLoading = false
    }
  }

  private async addAccountData() {
    this.onGetUserAuthInfo()
    this.dialogVisible = true
    this.dialogStatus = 'create'
    this.tempAddData = cloneDeep(defaultAddAccount)
    this.tempAddData.departmentOID = this.departmentOIDValue
    this.$nextTick(() => {
      (this.$refs.accountForm as Form).clearValidate()
    })
  }

  private addAccountDataToDB() {
    (this.$refs.accountForm as Form).validate(async valid => {
      if (valid) {
        const data = await addAccount(this.tempAddData)
        if ((data).status === 200) {
          this.dialogVisible = false
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: (this.$t('share.addSuccessfully')).toString(),
            type: 'success',
            duration: 2000
          })
          this.getAccountTable()
        }
      } else {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: (this.$t('message.notify003')).toString(),
          type: 'warning',
          duration: 2000
        })
      }
    })
  }

  private async updateAccountData(rowData: any) {
    this.dialogVisible = true
    this.dialogStatus = 'update'
    this.tempAddData = cloneDeep(defaultAddAccount)
    this.tempAddData.departmentOID = this.departmentOIDValue
    this.tempAddData.employeeID = rowData.employeeID
    this.tempAddData.roles = rowData.roles
    this.$nextTick(() => {
      (this.$refs.accountForm as Form).clearValidate()
    })
  }

  private async updateAccountDataToDB() {
    (this.$refs.accountForm as Form).validate(async valid => {
      if (valid) {
        const tempData = defaultUpdateAccount
        tempData.resetPassword = false
        tempData.roles = this.tempAddData.roles
        const data = await updateAccount(this.tempAddData.employeeID, tempData)
        if (data.status === 200) {
          this.dialogVisible = false
          this.$notify({
            title: (this.$t('share.success')).toString(),
            message: (this.$t('share.updateSuccessfully')).toString(),
            type: 'success',
            duration: 2000
          })
          this.getAccountTable()
        }
      } else {
        this.$notify({
          title: (this.$t('share.errorMessage')).toString(),
          message: (this.$t('message.notify003')).toString(),
          type: 'warning',
          duration: 2000
        })
      }
    })
  }

  private async deleteAccountData(rowData: any) {
    MessageBox.confirm(
      i18n.t('message.notifyDelete').toString(),
      i18n.t('share.prompt').toString(),
      {
        confirmButtonText: i18n.t('share.confirm').toString(),
        cancelButtonText: i18n.t('share.cancel').toString(),
        type: 'warning'
      }
    ).then(async () => {
      const data = await deleteAccount(rowData.employeeID)
      if (data.status === 200) {
        this.$notify({
          title: (this.$t('share.success')).toString(),
          message: (this.$t('share.deleteSuccessfully')).toString(),
          type: 'success',
          duration: 2000
        })
      }
      this.getAccountTable()
    })
  }

  private async resetPassword(rowData: any) {
    MessageBox.confirm(
      i18n.t('message.notify301').toString(),
      i18n.t('share.prompt').toString(),
      {
        confirmButtonText: i18n.t('share.confirm').toString(),
        cancelButtonText: i18n.t('share.cancel').toString(),
        type: 'warning'
      }
    ).then(async () => {
      const tempData = defaultUpdateAccount
      tempData.resetPassword = true
      tempData.roles = rowData.roles
      const data = await updateAccount(rowData.employeeID, tempData)
      if (data.status === 200) {
        this.$notify({
          title: (this.$t('share.success')).toString(),
          message: (this.$t('share.updateSuccessfully')).toString(),
          type: 'success',
          duration: 2000
        })
      }
    })
  }
}
