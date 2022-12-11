<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.stationConfig') }}
    </el-header>
    <el-form
      ref="stationConfigInfoForm"
      :model="tempStationConfigInfo"
      label-position="left"
      label-width="5rem"
      :rules="rules"
      autocomplete
    >
      <el-form-item
        :label="$t('system.stationID')"
        prop="stationID"
      >
        <el-select
          v-model="stationID"
          filterable
          :placeholder="$t('system.stationID')"
          @change="getStationConfig()"
        >
          <el-option
            v-for="item in stationIDList"
            :key="item.stationID"
            :label="item.stationID"
            :value="item.stationID"
          />
        </el-select>
      </el-form-item>
      <el-form-item
        :label="$t('system.executionMode')"
        prop="stationConfig.separateMode"
      >
        <el-radio-group
          v-model="tempStationConfigInfo.stationConfig.separateMode"
          size="medium"
        >
          <el-radio-button :label="false">
            {{ $t('system.togetherMode') }}
          </el-radio-button>
          <el-radio-button :label="true">
            {{ $t('system.separateMode') }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>
      <div v-if="tempStationConfigInfo.stationConfig.separateMode==false">
        <el-divider />
        <el-form-item
          :label="$t('system.productType')"
          prop="stationConfig.feed.productType"
        >
          <el-select
            v-model="tempStationConfigInfo.stationConfig.feed.productType"
            filterable
            multiple
            :placeholder="$t('system.productType')"
          >
            <el-option
              v-for="item in productTypeList"
              :key="item.type"
              :label="item.type"
              :value="item.type"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          :label="$t('system.materialResourceID')"
          prop="stationConfig.feed.materialResource"
        >
          <el-radio-group
            v-model="tempStationConfigInfo.stationConfig.feed.materialResource"
            size="medium"
          >
            <el-radio-button :label="true">
              {{ $t('share.show') }}
            </el-radio-button>
            <el-radio-button
              :label="false"
            >
              {{ $t('share.hide') }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          :label="$t('system.receiptQuantity')"
          prop="stationConfig.collect.quantity.type"
        >
          <el-select
            v-model="tempStationConfigInfo.stationConfig.collect.quantity.type"
            filterable
            :placeholder="$t('system.receiptQuantity')"
          >
            <el-option
              v-for="item in collectQuantityFlag"
              :key="item.type"
              :label="$t('share.' + item.name)"
              :value="item.type"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          v-if="tempStationConfigInfo.stationConfig.collect.quantity.type===0"
        >
          <el-input-number
            v-model="tempStationConfigInfo.stationConfig.collect.quantity.value"
            :min="0"
          />
        </el-form-item>
        <el-form-item
          :label="$t('system.receivingBarcode')"
          prop="stationConfig.collect.resource"
        >
          <el-radio-group
            v-model="tempStationConfigInfo.stationConfig.collect.resource"
            size="medium"
          >
            <el-radio-button :label="true">
              {{ $t('share.show') }}
            </el-radio-button>
            <el-radio-button :label="false">
              {{ $t('share.hide') }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          :label="$t('system.carrierBarcode')"
          prop="stationConfig.collect.carrierResource"
        >
          <el-radio-group
            v-model="tempStationConfigInfo.stationConfig.collect.carrierResource"
            size="medium"
          >
            <el-radio-button :label="true">
              {{ $t('share.show') }}
            </el-radio-button>
            <el-radio-button :label="false">
              {{ $t('share.hide') }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-button
          type="success"
          @click="confirmConfig('togetherMode')"
        >
          {{ $t('share.confirm') }}
        </el-button>
      </div>
      <div v-else>
        <el-divider content-position="center">
          {{ $t('system.feed') }}
        </el-divider>
        <el-form-item
          :label="$t('system.productType')"
          prop="stationConfig.feed.productType"
        >
          <el-select
            v-model="tempStationConfigInfo.stationConfig.feed.productType"
            filterable
            multiple
            :placeholder="$t('system.productType')"
          >
            <el-option
              v-for="item in productTypeList"
              :key="item.type"
              :label="item.type"
              :value="item.type"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          :label="$t('system.feedQuantity')"
          prop="stationConfig.feed.standardQuantity"
        >
          <el-select
            v-model="tempStationConfigInfo.stationConfig.feed.standardQuantity"
            filterable
            :placeholder="$t('system.feedQuantity')"
          >
            <el-option
              v-for="item in feedQuantityFlag"
              :key="item.type"
              :label="$t('share.' + item.name)"
              :value="item.type"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          :label="$t('system.operatorSiteName')"
        >
          <el-input
            id="grade"
            v-model="tempStationConfigInfo.stationConfig.feed.operatorSites[0].siteName"
          />
        </el-form-item>
        <el-divider content-position="center">
          {{ $t('system.receipt') }}
        </el-divider>
        <el-form-item
          :label="$t('system.receiptQuantity')"
          prop="stationConfig.collect.quantity.type"
        >
          <el-select
            v-model="tempStationConfigInfo.stationConfig.collect.quantity.type"
            filterable
            :placeholder="$t('system.receiptQuantity')"
          >
            <el-option
              v-for="item in collectQuantityFlag"
              :key="item.type"
              :label="$t('share.' + item.name)"
              :value="item.type"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          v-if="tempStationConfigInfo.stationConfig.collect.quantity.type===0"
        >
          <el-input-number
            v-model="tempStationConfigInfo.stationConfig.collect.quantity.value"
            :min="0"
          />
        </el-form-item>
        <el-form-item
          :label="$t('system.receivingBarcode')"
          prop="stationConfig.collect.resource"
        >
          <el-radio-group
            v-model="tempStationConfigInfo.stationConfig.collect.resource"
            size="medium"
          >
            <el-radio-button :label="true">
              {{ $t('share.show') }}
            </el-radio-button>
            <el-radio-button :label="false">
              {{ $t('share.hide') }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          :label="$t('system.carrierBarcode')"
          prop="stationConfig.collect.carrierResource"
        >
          <el-radio-group
            v-model="tempStationConfigInfo.stationConfig.collect.carrierResource"
            size="medium"
          >
            <el-radio-button :label="true">
              {{ $t('share.show') }}
            </el-radio-button>
            <el-radio-button :label="false">
              {{ $t('share.hide') }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          :label="$t('system.operatorSiteName')"
        >
          <el-input
            id="grade"
            v-model="tempStationConfigInfo.stationConfig.collect.operatorSites[0].siteName"
          />
        </el-form-item>
        <el-button
          type="success"
          @click="confirmConfig('separateMode')"
        >
          {{ $t('share.confirm') }}
        </el-button>
      </div>
    </el-form>
    <div style="margin-bottom: 5rem;" />
  </div>
</template>
<script src="./stationConfig.ts" lang="ts"></script>
<style src="./stationConfig.scss" lang="scss"></style>
