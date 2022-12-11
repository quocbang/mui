<template>
  <div class="dashboard-container">
    <div
      v-if="Type!=2"
    >
      <el-button
        type="success"
        @click="$router.push({path: '/materialChanges'})"
      >
        {{ $t('router.materialChanges') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/materialCreate'})"
      >
        {{ $t('router.materialCreate') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/warehouseTransaction'})"
      >
        {{ $t('router.warehouseTransaction') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/productionPlan'})"
      >
        {{ $t('router.productionPlan') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/stationSchedule'})"
      >
        {{ $t('router.stationSchedule') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/recipeProcess'})"
      >
        {{ $t('router.recipeProcess') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/materialMount'})"
      >
        {{ $t('router.materialMount') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/carrierMaintenance'})"
      >
        {{ $t('router.carrierMaintenance') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/accountManagement'})"
      >
        {{ $t('router.accountManagement') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/stationMaintenance'})"
      >
        {{ $t('router.stationMaintenance') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/materialQuery'})"
      >
        {{ $t('router.materialQuery') }}
      </el-button>
      <el-button
        type="success"
        @click="$router.push({path: '/productionRate'})"
      >
        {{ $t('router.productionRate') }}
      </el-button>
    </div>
    <div v-else>
      <!-- PDA -->
      <el-row>
        <el-col
          :span="24"
          align="middle"
        >
          <el-button
            type="danger"
            @click="$router.push({path: '/selectWorkOrder'})"
          >
            {{ $t('router.switchStation') }}
          </el-button>
        </el-col>
        <el-col
          :span="24"
          align="middle"
        >
          <el-button
            type="danger"
            @click="$router.push({path: '/mount'})"
          >
            {{ $t('system.mount') }}
          </el-button>
        </el-col>
        <el-col
          :span="24"
          align="middle"
        >
          <el-button
            type="danger"
            disabled
          >
            {{ $t('router.scanQuery') }}
          </el-button>
        </el-col>
        <el-col
          :span="24"
          align="middle"
        >
          <el-button
            type="danger"
            disabled
          >
            {{ $t('router.menuStatus') }}
          </el-button>
        </el-col>
        <el-col
          :span="24"
          align="middle"
        />
      </el-row>
    </div>
  </div>
</template>
<script src="./index.ts" lang="ts"></script>
<style src="./index.scss" lang="scss"></style>
