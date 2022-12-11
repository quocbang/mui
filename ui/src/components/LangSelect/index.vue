<template>
  <el-dropdown
    trigger="click"
    class="international"
    @command="handleSetLanguage"
  >
    <div>
      <svg-icon
        name="language"
        class="international-icon"
      />
    </div>
    <el-dropdown-menu slot="dropdown">
      <el-dropdown-item
        :disabled="language==='tw'"
        command="tw"
      >
        {{ $t('language.traditionalChinese') }}
      </el-dropdown-item>
      <el-dropdown-item
        :disabled="language==='cn'"
        command="cn"
      >
        {{ $t('language.simplifiedChinese') }}
      </el-dropdown-item>
      <el-dropdown-item
        :disabled="language==='en'"
        command="en"
      >
        {{ $t('language.english') }}
      </el-dropdown-item>
      <el-dropdown-item
        :disabled="language==='vi'"
        command="vi"
      >
        {{ $t('language.vietnamese') }}
      </el-dropdown-item>
    </el-dropdown-menu>
  </el-dropdown>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { AppModule } from '@/store/modules/app'
import settings from '../../settings'
@Component({
  name: 'Login'
})
export default class extends Vue {
  get language() {
    return AppModule.language
  }

  private handleSetLanguage(lang: string) {
    this.$i18n.locale = lang
    AppModule.SetLanguage(lang)
    document.documentElement.lang = lang
    const title = this.$route.meta.title ? `${this.$t(`route.${this.$route.meta.title}`)} - ${settings.title}` : `${settings.title}`
    document.title = title
    this.$notify({
      title: (this.$t('share.success')).toString(),
      message: this.$t('share.changeLanguageTips').toString(),
      type: 'success',
      duration: 2000
    })
  }
}
</script>
