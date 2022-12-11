<template>
  <div class="main-container">
    <navbar />
    <app-main />
    <footer class="fixed-floor grid-content">
      <el-row>
        <el-col :span="6">
          <el-button
            type="primary"
            plain
            class="floor-button"
            @click="$router.push({path: '/selectWorkOrder'})"
          >
            {{ $t('router.switchStation') }}
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button
            type="primary"
            plain
            class="floor-button"
            @click="$router.push({path: '/mount'})"
          >
            {{ $t('system.mount') }}
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button
            type="info"
            plain
            disabled
            class="floor-button"
          >
            {{ $t('router.scanQuery') }}
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button
            type="info"
            plain
            disabled
            class="floor-button"
          >
            {{ $t('router.menuStatus') }}
          </el-button>
        </el-col>
      </el-row>
    </footer>
  </div>
</template>

<script lang="ts">
import { Component } from 'vue-property-decorator'
import { mixins } from 'vue-class-component'
import { AppMain, Navbar, Sidebar } from './components'
import ResizeMixin from './mixin/resize'
import 'default-passive-events'

@Component({
  name: 'pdaLayout',
  components: {
    AppMain,
    Navbar,
    Sidebar
  }
})
export default class extends mixins(ResizeMixin) {

}
</script>
<style lang="scss" scoped>

.fixed-floor{
  width:100%;
  position:fixed;
  bottom:0;
  z-index:1;
}
.floor-button{
width: -webkit-fill-available;
}
</style>
