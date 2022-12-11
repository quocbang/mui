<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.materialMount') }}
    </el-header>
    <el-tabs
      v-model="activeName"
      type="border-card"
      @tab-click="handleClick"
    >
      <el-tab-pane
        :key="'simpleVersion'"
        name="simpleVersion"
        :label="$t('system.simpleVersion')"
      >
        <div v-if="activeName==='simpleVersion'">
          <el-form
            ref="materialMountDataForm"
            :rules="materialRules"
            :model="tempMaterialMountData"
            label-position="left"
            label-width="5rem"
            autocomplete
          >
            <el-form-item
              :label="$t('system.stationID')"
              prop="stationID"
            >
              <el-input
                ref="stationID"
                v-model="tempMaterialMountData.stationID"
              />
            </el-form-item>
            <el-form-item
              :label="$t('system.barcodeForBarrelSlot')"
              prop="barcodeForBarrelSlot"
            >
              <el-input
                ref="barcodeForBarrelSlot"
                v-model="tempMaterialMountData.barcodeForBarrelSlot"
                @input="getBarcodeForBarrelSlotInfo()"
              />
            </el-form-item>
            <el-form-item
              :label="$t('system.materialResourceID')"
              prop="materialBarcode"
            >
              <el-input
                ref="materialBarcode"
                v-model="tempMaterialMountData.materialBarcode"
                @input="getMaterialBarcodeInfo()"
              />
            </el-form-item>
            <el-form-item
              :label="$t('system.productType')"
              prop="productType"
            >
              <el-select
                v-model="tempMaterialMountData.productType"
                filterable
                :placeholder="$t('system.productType')"
                @change="getMaterialInfo()"
              >
                <el-option
                  v-for="item in productTypeInfoList"
                  :key="item"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('system.materialID')">
              <el-input
                v-model="resourceData.ID"
                disabled
              />
            </el-form-item>
            <el-form-item
              :label="$t('system.materialQuantity')"
            >
              <el-input-number
                v-model="resourceData.quantity"
                :min="0"
                :max="originQuantity"
                :disabled="quantityDisable"
              />
            </el-form-item>
            <el-form-item :label="$t('system.expirationTime')">
              <el-date-picker
                v-model="resourceData.expiredDate"
                disabled
                type="datetime"
                value-format="yyyy-MM-dd HH:mm:ss"
              />
            </el-form-item>
          </el-form>
          <el-button
            class="filter-item"
            type="primary"
            icon="el-icon-circle-plus-outline"
            @click="resourceAdd()"
          >
            {{ $t('mount.active.add') }}
          </el-button>
          <el-button
            class="filter-item"
            type="primary"
            @click="resourceCleanDeviation()"
          >
            {{ $t('mount.active.cleanDeviation') }}
          </el-button>
          <el-button
            type="primary"
            @click="reset()"
          >
            {{ $t('system.reset') }}
          </el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane
        :key="'fullVersion'"
        name="fullVersion"
        :label="$t('system.fullVersion')"
      >
        <div v-if="activeName==='fullVersion'">
          <el-form
            ref="materialMountDataForm"
            :rules="materialRules"
            :model="tempMaterialMountData"
            label-position="left"
            label-width="5rem"
            autocomplete
          >
            <el-form-item
              :label="$t('system.stationID')"
              prop="stationID"
            >
              <el-input
                ref="stationID"
                v-model="tempMaterialMountData.stationID"
              />
            </el-form-item>
            <el-form-item
              :label="$t('system.barcodeForBarrelSlot')"
              prop="barcodeForBarrelSlot"
            >
              <el-input
                ref="barcodeForBarrelSlot"
                v-model="tempMaterialMountData.barcodeForBarrelSlot"
                @input="getBarcodeForBarrelSlotInfo()"
              />
            </el-form-item>
            <el-form-item :label="$t('system.bucketName')">
              <el-input
                v-model="tempMaterialMountData.siteName"
                disabled
              />
            </el-form-item>
            <el-form-item :label="$t('system.tankNumber')">
              <el-input
                v-model="tempMaterialMountData.siteIndex"
                disabled
              />
            </el-form-item>
            <el-form-item
              :label="$t('system.materialResourceID')"
              prop="materialBarcode"
            >
              <el-input
                ref="materialBarcode"
                v-model="tempMaterialMountData.materialBarcode"
                @input="getMaterialBarcodeInfo()"
              />
            </el-form-item>
            <el-form-item
              :label="$t('system.productType')"
              prop="productType"
            >
              <el-select
                v-model="tempMaterialMountData.productType"
                filterable
                :placeholder="$t('system.productType')"
                @change="getMaterialInfo()"
              >
                <el-option
                  v-for="item in productTypeInfoList"
                  :key="item"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('system.materialID')">
              <el-input
                v-model="resourceData.ID"
                disabled
              />
            </el-form-item>
            <el-form-item :label="$t('system.materialGrade')">
              <el-input
                v-model="resourceData.grade"
                disabled
              />
            </el-form-item>
            <el-form-item
              :label="$t('system.materialQuantity')"
            >
              <el-input-number
                v-model="resourceData.quantity"
                :min="0"
                :max="originQuantity"
                :disabled="quantityDisable"
              />
            </el-form-item>
            <el-form-item :label="$t('system.expirationTime')">
              <el-date-picker
                v-model="resourceData.expiredDate"
                disabled
                type="datetime"
                value-format="yyyy-MM-dd HH:mm:ss"
              />
            </el-form-item>
          </el-form>
          <div class="filter-container">
            <el-button
              class="filter-item"
              type="primary"
              icon="el-icon-circle-plus-outline"
              @click="resourceAdd()"
            >
              {{ $t('mount.active.add') }}
            </el-button>
            <el-button
              type="primary"
              @click="reset()"
            >
              {{ $t('system.reset') }}
            </el-button>
          </div>
          <el-divider />
          <h4> {{ $t('system.bucketMaterialInformation') }}</h4>
          <el-table
            ref="draggableTable"
            v-loading="listLoading"
            :data="bindResourceData"
            border
            fit
            highlight-current-row
            style=":min-width: 100%"
          >
            <el-table-column
              :label="$t('system.materialResourceID')"
              :min-width="100"
            >
              <template slot-scope="{row}">
                <span>{{ row.resourceID }}</span>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.materialID')"
              :min-width="100"
            >
              <template slot-scope="{row}">
                <span>{{ row.productID }}</span>
              </template>
            </el-table-column>

            <el-table-column
              :label="$t('system.materialGrade')"
              :min-width="100"
            >
              <template slot-scope="{row}">
                <span>{{ row.grade }}</span>
              </template>
            </el-table-column>

            <el-table-column
              :label="$t('system.materialQuantity')"
              :min-width="100"
            >
              <template slot-scope="{row}">
                <span>{{ row.quantity }}</span>
              </template>
            </el-table-column>
          </el-table>
          <div class="filter-container">
            <el-button
              class="filter-item"
              type="primary"
              @click="resourceCleanDeviation()"
            >
              {{ $t('mount.active.cleanDeviation') }}
            </el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
<script src="./materialMount.ts" lang="ts"></script>
<style src="./materialMount.css"></style>
