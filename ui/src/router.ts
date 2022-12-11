import Vue from 'vue'
import Router from 'vue-router'
import Layout from '@/layout/index.vue'
import PdaLayout from '@/layout/pdaLayout.vue'
Vue.use(Router)

/*
  redirect:                      if set to 'noredirect', no redirect action will be trigger when clicking the breadcrumb
  meta: {
    title: 'title'               the name showed in subMenu and breadcrumb (recommend set)
    icon: 'svg-name'             the icon showed in the sidebar
    breadcrumb: false            if false, the item will be hidden in breadcrumb (default is true)
    hidden: true                 if true, this route will not show in the sidebar (default is false)
  }
*/

export default new Router(
  {
    // mode: 'history',  // Enable this if you need.
    scrollBehavior: (to, from, savedPosition) => {
      if (savedPosition) {
        return savedPosition
      } else {
        return { x: 0, y: 0 }
      }
    },
    base: process.env.BASE_URL,
    routes: [
      {
        path: '/login',
        component: () => import('@/views/login/index.vue'),
        meta: {
          hidden: true,
          title: 'Login'
        }
      },
      {
        path: '/404',
        component: () => import('@/views/404.vue'),
        meta: { hidden: true }
      },
      {
        path: '/RecipeProcessPrintView',
        component: () => import('@/views/mesSystem/recipeProcess/printView/printView.vue'),
        meta: {
          hidden: true,
          title: 'RecipeProcessPrintView'
        }
      },
      {
        path: '/',
        component: Layout,
        redirect: '/dashboard',
        children: [
          {
            path: 'dashboard',
            component: () => import('@/views/dashboard/index.vue'),
            meta: {
              title: 'dashboard',
              icon: 'dashboard'
            }
          }
        ]
      },
      {
        path: '/editPassword',
        component: Layout,
        redirect: 'noRedirect',
        meta: {
          hidden: true
        },
        children: [
          {
            path: '',
            component: () => import('@/views/other/editPassword/editPassword.vue'),
            meta: {
              title: 'editPassword',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/materialChanges',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/materialChanges/materialChanges.vue'),
            meta: {
              title: 'materialChanges',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/materialCreate',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/materialCreate/materialCreate.vue'),
            meta: {
              title: 'materialCreate',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/warehouseTransaction',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/warehouseTransaction/warehouseTransaction.vue'),
            meta: {
              title: 'warehouseTransaction',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/productionPlan',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/productionPlan/productionPlan.vue'),
            meta: {
              title: 'productionPlan',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/stationSchedule',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/stationSchedule/stationSchedule.vue'),
            meta: {
              title: 'stationSchedule',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/recipeProcess',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/recipeProcess/recipeProcess.vue'),
            meta: {
              title: 'recipeProcess',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/materialMount',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/materialMount/materialMount.vue'),
            meta: {
              title: 'materialMount',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/carrierMaintenance',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/carrierMaintenance/carrierMaintenance.vue'),
            meta: {
              title: 'carrierMaintenance',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/accountManagement',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/accountManagement/accountManagement.vue'),
            meta: {
              title: 'accountManagement',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/stationMaintenance',
        component: Layout,
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/stationMaintenance/stationMaintenance.vue'),
            meta: {
              title: 'stationMaintenance',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/materialQuery',
        component: Layout,
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/materialQuery/materialQuery.vue'),
            meta: {
              title: 'materialQuery',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/selectWorkOrder',
        component: PdaLayout,
        meta: {
          hidden: true
        },
        children: [
          {
            path: '',
            component: () => import('@/views/pdaSystem/selectWorkOrder/selectWorkOrder.vue'),
            meta: {
              title: 'selectWorkOrder',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/mount',
        component: PdaLayout,
        meta: {
          hidden: true
        },
        children: [
          {
            path: '',
            component: () => import('@/views/pdaSystem/mount/mount.vue'),
            meta: {
              title: 'mount',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/jobView',
        component: PdaLayout,
        meta: {
          hidden: true
        },
        children: [
          {
            path: '',
            component: () => import('@/views/pdaSystem/jobView/jobView.vue'),
            meta: {
              title: 'jobView',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/stationConfig',
        component: PdaLayout,
        meta: {
          hidden: true
        },
        children: [
          {
            path: '',
            component: () => import('@/views/pdaSystem/stationConfig/stationConfig.vue'),
            meta: {
              title: 'stationConfig',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '/productionRate',
        component: Layout,
        redirect: 'noRedirect',
        children: [
          {
            path: '',
            component: () => import('@/views/mesSystem/productionRate/productionRate.vue'),
            meta: {
              title: 'productionRate',
              icon: 'table'
            }
          }
        ]
      },
      {
        path: '*',
        redirect: '/404',
        meta: { hidden: true }
      }
    ]
  })
