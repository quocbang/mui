<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.selectWorkOrder') }}
    </el-header>
    <el-form
      :model="tempSelectWorkOrderInfo"
      label-position="left"
      label-width="5rem"
      :rules="rules"
      autocomplete
    >
      <el-form-item
        :label="$t('system.stationID')"
      >
        <el-select
          ref="stationID"
          v-model="tempSelectWorkOrderInfo.stationID"
          filterable
          :placeholder="$t('system.stationID')"
          @change="getWorkOrderListAndStationConfig()"
        >
          <el-option
            v-for="item in productIDList"
            :key="item.stationID"
            :label="item.stationID"
            :value="item.stationID"
          />
        </el-select>
      </el-form-item>
    </el-form>
    <div
      v-if="workOrderInfoList.length==0 &&action==true"
      style="font-size: small;
    text-align: center;
    margin-top: 2rem"
    >
      {{ $t('errorCodes.errorCode_404') }}
    </div>
    <div v-else>
      <el-card
        v-for="item in workOrderInfoList"
        :key="item.workOrderID"
        class="box-card"
      >
        <div
          slot="header"
          style="padding: 10px 20px"
          class="clearfix"
        >
          <el-row
            type="flex"
            justify="space-between"
          >
            <el-col
              :span="20"
              style="margin-top: -12px;"
            >
              <span class="card-title">{{ item.productID }}</span>
            </el-col>
            <el-col :span="4">
              <el-button
                type="primary"
                size="mini"
                style="float: right; width: 30px"
                icon="el-icon-check"
                circle
                @click="operatorSignInCheck(item.workOrderID)"
              />
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
              {{ $t('share.date') }}
            </el-col>
            <el-col :span="14">
              {{ item.date }}
            </el-col>
          </el-row>
          <el-row>
            <el-col
              :span="10"
              class="label-title"
            >
              {{ $t('workOrder.ID') }}
            </el-col>
            <el-col :span="14">
              {{ item.workOrderID }}
            </el-col>
          </el-row>
          <el-row>
            <el-col
              :span="10"
              class="label-title"
            >
              {{ $t('system.estimateOutput') }}
            </el-col>
            <el-col :span="14">
              {{ item.planQuantity }}
            </el-col>
          </el-row>
        </div>
      </el-card>
    </div>
    <el-dialog
      :title="$t('share.prompt')"
      :visible.sync="dialogSelectModelVisible"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
      :fullscreen="true"
      center
    >
      <span>{{ $t('message.notifySelectMode100') }}</span>
      <span
        slot="footer"
        class="dialog-footer"
      >
        <el-button
          type="success"
          @click="selectMode('feed')"
        >{{ $t('system.feed') }}</el-button>
        <el-button
          type="primary"
          @click="selectMode('receipt')"
        >{{ $t('system.receipt') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script src="./selectWorkOrder.ts" lang="ts"></script>
<style src="./selectWorkOrder.scss" lang="scss"></style>
