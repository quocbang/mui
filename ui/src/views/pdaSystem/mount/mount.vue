<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('system.mount') }}
    </el-header>
    <el-form
      ref="mountInfoForm"
      :model="tempMountInfo"
      label-position="left"
      label-width="5rem"
      :rules="rules"
      autocomplete
    >
      <el-form-item
        :label="$t('system.siteLocation')"
        prop="siteLocation"
      >
        <el-input
          ref="siteLocation"
          v-model="tempMountInfo.siteLocation"
          onfocus="this.select()"
          @input="getSiteInfo()"
        />
      </el-form-item>
      <el-form-item
        :label="$t('system.materialResourceID')"
        prop="resourceID"
      >
        <el-input
          v-model="tempMountInfo.resourceID"
          onfocus="this.select()"
          @input="getResourceInfo()"
        />
      </el-form-item>
    </el-form>
    <div style="text-align: center;">
      <el-button
        class="filter-item"
        type="primary"
        @click="bind()"
      >
        {{ $t('system.mount') }}
      </el-button>
      <el-button
        class="filter-item"
        type="primary"
        @click="$router.push({path: '/jobView'})"
      >
        {{ $t('system.gotoHomework') }}
      </el-button>
      <el-button
        class="filter-item"
        type="danger"
        @click="clear()"
      >
        {{ $t('system.uninstall') }}
      </el-button>
    </div>
    <el-divider />
    <el-card
      v-for="item in siteBindInfo"
      :key="item.resourceID"
      class="box-card"
    >
      <div
        slot="header"
        class="clearfix"
      >
        <el-row
          type="flex"
          justify="space-between"
        >
          <el-col :span="24">
            <span class="card-title">{{ item.resourceID }}</span>
          </el-col>
        </el-row>
      </div>
      <div
        class="text item"
        style="line-height: 20px;"
      >
        <el-row>
          <el-col
            :span="10"
            class="label-title"
          >
            {{ $t('system.productID') }}
          </el-col>
          <el-col :span="14">
            {{ item.productID }}
          </el-col>
        </el-row>
        <el-row>
          <el-col
            :span="10"
            class="label-title"
          >
            {{ $t('system.quantity') }}
          </el-col>
          <el-col :span="14">
            {{ item.quantity }}
          </el-col>
        </el-row>
      </div>
    </el-card>
  </div>
</template>
<script src="./mount.ts" lang="ts"></script>
<style src="./mount.scss" lang="scss"></style>
