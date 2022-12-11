<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.materialChanges') }}
    </el-header>
    <el-form
      ref="dataForm"
      v-loading="dataLoading"
      :rules="rule"
      :model="form"
      label-position="left"
      label-width="5rem"
      autocomplete
    >
      <el-form-item
        v-if="materialBarcode ==='' || GetAllInfoStatus === 'open'"
        :label="$t('system.materialResourceID')"
      >
        <el-input
          ref="materialBarcode"
          v-model="materialBarcode"
          @input="getBarcodeInfoList()"
        />
      </el-form-item>
      <el-form-item
        v-else
        :label="$t('system.materialResourceID')"
      >
        <el-input
          v-model="materialBarcode"
          disabled
        />
      </el-form-item>
      <el-form-item :label="$t('system.materialID')">
        <el-input
          v-model="formInfo.productID"
          disabled
        />
      </el-form-item>
      <el-form-item
        :label="$t('system.stage')"
        prop="productCate"
      >
        <el-select
          v-model="form.productCate"
          filterable
        >
          <el-option
            v-for="item in allStage"
            :key="item.code"
            :label="$t('recipe.'+ item.description)"
            :value="$t('recipe.'+ item.description)"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('system.expirationTime')">
        <el-input
          v-model="formInfo.expiredAt"
          disabled
        />
      </el-form-item>
      <el-form-item :label="$t('share.status')">
        <el-input
          v-model="formInfo.status"
          disabled
        />
      </el-form-item>
      <el-form-item>
        <el-radio-group
          v-model="radioItem"
          @change="radioItemChange"
        >
          <el-radio label="changeStatus">
            {{ $t('system.statusChange') }}
          </el-radio>
          <el-radio label="addDate">
            {{ $t('system.increaseDeadline') }}
          </el-radio>
        </el-radio-group>
      </el-form-item>
      <template
        v-if="radioItem==='changeStatus'"
      >
        <el-form-item prop="newStatus">
          <el-select
            v-model="form.newStatus"
            filterable
            @change="radioItemChange"
          >
            <el-option
              v-for="changeStatusItem in changeStatusList"
              :key="changeStatusItem.code"
              :label="changeStatusItem.description"
              :value="changeStatusItem.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          v-if="form.newStatus === 'HOLD'"
          :label="$t('system.reasonForChange')"
          prop="holdReason"
        >
          <el-select
            v-model="form.holdReason"
          >
            <el-option
              v-for="reasonItem in reasonList"
              :key="reasonItem.description"
              :label="reasonItem.description"
              :value="reasonItem.description"
            />
          </el-select>
        </el-form-item>
      </template>
      <template v-else>
        <el-form-item>
          <el-input
            v-model="form.extendDays"
            disabled
          />
        </el-form-item>
      </template>

      <el-form-item
        :label="$t('system.controlArea')"
        prop="controlArea"
      >
        <el-select
          v-model="form.controlArea"
          filterable
        >
          <el-option
            v-for="controlAreaItem in controlAreaList"
            :key="controlAreaItem.code"
            :label="controlAreaItem.description"
            :value="controlAreaItem.code"
          />
        </el-select>
      </el-form-item>
    </el-form>
    <el-button
      v-if="GetAllInfoStatus === 'open' "
      type="primary"
      disabled
      @click="updateBarcodeInfoDate()"
    >
      {{ $t('share.confirm') }}
    </el-button>
    <el-button
      v-else
      type="primary"
      @click="updateBarcodeInfoDate()"
    >
      {{ $t('share.confirm') }}
    </el-button>
    <el-button
      type="primary"
      @click="resetBarcodeData()"
    >
      {{ $t('system.reset') }}
    </el-button>
  </div>
</template>
<script src="./materialChanges.ts" lang="ts"></script>
<style src="./materialChanges.css"></style>
