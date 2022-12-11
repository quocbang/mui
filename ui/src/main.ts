import Vue, { DirectiveOptions } from 'vue'
import 'normalize.css'
import ElementUI from 'element-ui'
import SvgIcon from 'vue-svgicon'
import '@/styles/element-variables.scss'
import '@/styles/index.scss'
import App from '@/App.vue'
import store from '@/store'
import router from '@/router'
import i18n from '@/lang'
import '@/icons/components'
import '@/permission'
import * as filters from '@/filters'
import * as directives from '@/directives'
import moment from 'moment'

Vue.use(ElementUI, {
  i18n: (key: string, value: string) => i18n.t(key, value)
})
Vue.use(SvgIcon, {
  tagName: 'svg-icon',
  defaultWidth: '1em',
  defaultHeight: '1em'
})
// Register global directives
Object.keys(directives).forEach(key => {
  Vue.directive(key, (directives as { [key: string]: DirectiveOptions })[key])
})

// Register global filter functions
Object.keys(filters).forEach(key => {
  Vue.filter(key, (filters as { [key: string]: Function })[key])
  Vue.filter('formatDate', (value: any) => {
    if (value) {
      return moment(String(value)).format('yyyy-MM-DD HH:mm:ss')
    }
  })
})
Vue.config.productionTip = false

new Vue({
  router,
  store,
  i18n,
  render: (h) => h(App)
}).$mount('#app')
