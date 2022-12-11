<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.warehouseTransaction') }}
    </el-header>
    <el-form
      ref="warehouseTransactionDataForm"
      v-loading="dataLoading"
      :rules="warehouseTransactionRules"
      :model="tempWarehouseTransactionInfo"
      label-position="left"
      label-width="5rem"
      autocomplete
    >
      <el-form-item
        :label="$t('system.materialResourceID')"
        prop="ID"
      >
        <el-input
          ref="ID"
          v-model="tempWarehouseTransactionInfo.ID"
          @input="getBarcodeInfo()"
        />
      </el-form-item>
      <el-form-item
        :label="$t('system.newWarehouse')"
        prop="newWarehouseID"
      >
        <el-input
          v-model="tempWarehouseTransactionInfo.newWarehouseID"
          type="text"
          maxlength="1"
          show-word-limit
          @input="checkNewWarehouseID()"
        />
      </el-form-item>
      <el-form-item
        :label="$t('system.newLocation')"
        prop="newLocation"
      >
        <el-input
          v-model="tempWarehouseTransactionInfo.newLocation"
          type="text"
          maxlength="2"
          show-word-limit
        />
      </el-form-item>
      <el-button
        type="primary"
        @click="updateWarehouseLocationToDB()"
      >
        {{ $t('share.confirm') }}
      </el-button>
    </el-form>
  </div>
</template>
<script src="./warehouseTransaction.ts" lang="ts"></script>
<style src="./warehouseTransaction.css"></style>
