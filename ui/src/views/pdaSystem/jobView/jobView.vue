<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.jobView') }}
    </el-header>
    <el-card class="box-card">
      <el-row>
        <el-col
          :span="12"
          style="font-size: 14px;"
        >
          <el-row>
            <el-col
              :span="12"
              class="label-title"
            >
              {{ $t('system.productID') }}
            </el-col>
            <el-col :span="12">
              {{ workOrderInfo.productID }}
            </el-col>
          </el-row>
          <el-row>
            <el-col
              :span="12"
              class="label-title"
            >
              {{ $t('system.estimateOutput') }}
            </el-col>
            <el-col :span="12">
              {{ workOrderInfo.planQuantity }}
            </el-col>
          </el-row>
          <el-row>
            <el-col
              :span="12"
              class="label-title"
            >
              {{ $t('system.currentQuantity') }}
            </el-col>
            <el-col :span="12">
              {{ workOrderInfo.currentQuantity }}
            </el-col>
          </el-row>
        </el-col>
        <el-col
          :span="12"
        >
          <el-button
            type="primary"
            size="mini"
            style="float: right; width: 30px"
            icon="el-icon-close"
            circle
            @click="closeWorkOrder()"
          />
        </el-col>
      </el-row>
    </el-card>
    <el-divider />
    <el-tabs
      v-model="modelTabsValue"
      type="border-card"
      @tab-click="tabClick()"
    >
      <el-tab-pane
        v-if="tempStationConfigInfo.stationConfig.separateMode === false"
        name="togetherMode"
        :label="$t('system.togetherMode')"
      >
        <el-form
          ref="refDataForm"
          :model="tempCollectOrFeedAndCollect"
          label-position="left"
          label-width="5rem"
          autocomplete
        >
          <el-form-item
            v-if="tempStationConfigInfo.stationConfig.feed.materialResource==true"
            :label="$t('system.materialResourceID')"
            prop="feedResourceIDs[0]"
          >
            <el-input
              ref="focusTogetherMode"
              v-model="tempCollectOrFeedAndCollect.feedResourceIDs[0]"
              onfocus="this.select()"
            />
          </el-form-item>

          <el-form-item
            v-if="tempStationConfigInfo.stationConfig.collect.resource==true"
            :label="$t('system.receivingBarcode')"
          >
            <el-input
              v-model="tempCollectOrFeedAndCollect.resourceID"
              onfocus="this.select()"
            />
          </el-form-item>
          <el-form-item
            :label="$t('system.receiptQuantity')"
          >
            <el-input-number
              v-model="tempCollectOrFeedAndCollect.quantity"
              :min="0"
            />
          </el-form-item>
          <el-form-item
            v-if="tempStationConfigInfo.stationConfig.collect.carrierResource==true"

            :label="$t('system.carrierBarcode')"
            prop="carrierResource"
          >
            <el-input
              v-model="tempCollectOrFeedAndCollect.carrierResource"
              onfocus="this.select()"
            />
          </el-form-item>
        </el-form>
        <div style="text-align: center;">
          <el-button
            v-loading.fullscreen.lock="fullscreenLoading"
            class="filter-item"
            type="primary"
            @click="collectOrFeedAndCollectMode('togetherMode')"
          >
            {{ $t('system.togetherMode') }}
          </el-button>
          <el-button
            v-loading.fullscreen.lock="fullscreenLoading"
            class="filter-item"
            type="primary"
            @click="print()"
          >
            {{ $t('share.print') }}
          </el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane
        v-if="tempStationConfigInfo.stationConfig.separateMode === true"
        name="feed"
        :label="$t('system.feed')"
      >
        <el-button
          id="createResource"
          class="filter-item"
          type="primary"
          icon="el-icon-circle-plus-outline"
          size="medium"
          @click="createResource()"
        >
          {{ $t('share.create') }}
        </el-button>
        <el-card class="box-card">
          <div>
            <el-form
              :model="tempFeed"
              label-position="left"
              label-width="5rem"
              autocomplete
            >
              <el-card
                v-for="(item, index) in tempFeed.resource"
                :key="index"

                class="box-card"
              >
                <div
                  slot="header"
                  class="clearfix"
                >
                  <el-button
                    type="danger"
                    size="mini"
                    icon="el-icon-close"
                    @click="deleteResource(index)"
                  />
                </div>

                <el-form-item
                  v-if="tempStationConfigInfo.stationConfig.feed.materialResource==true"
                  :label="$t('system.materialResourceID')"
                >
                  <el-input
                    :ref="'focusFeed_'+index"
                    v-model="item.ID"
                    onfocus="this.select()"
                  />
                </el-form-item>
                <el-form-item
                  v-if="tempStationConfigInfo.stationConfig.feed.standardQuantity==1"
                  :label="$t('system.feedQuantity')"
                  prop="quantity"
                  size="mini"
                >
                  <el-input-number
                    v-model="item.quantity"
                    :min="0"
                  />
                </el-form-item>
              </el-card>
            </el-form>
          </div>
        </el-card>
        <div style="text-align: center;">
          <el-button
            v-loading.fullscreen.lock="fullscreenLoading"
            class="filter-item"
            type="primary"
            @click="feed()"
          >
            {{ $t('system.feed') }}
          </el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane
        v-if="tempStationConfigInfo.stationConfig.separateMode === true"
        name="receipt"
        :label="$t('system.receipt')"
      >
        <el-form
          ref="refDataForm"
          :model="tempCollectOrFeedAndCollect"
          label-position="left"
          label-width="5rem"
          autocomplete
        >
          <el-form-item
            v-if="tempStationConfigInfo.stationConfig.collect.resource==true"
            :label="$t('system.receivingBarcode')"
            prop="resourceID"
          >
            <el-input
              ref="focusReceipt"
              v-model="tempCollectOrFeedAndCollect.resourceID"
              onfocus="this.select()"
            />
          </el-form-item>
          <el-form-item
            :label="$t('system.receiptQuantity')"
            prop="quantity"
          >
            <el-input-number
              v-model="tempCollectOrFeedAndCollect.quantity"
              :min="0"
            />
          </el-form-item>
          <el-form-item
            v-if="tempStationConfigInfo.stationConfig.collect.carrierResource==true"

            :label="$t('system.carrierBarcode')"
            prop="carrierResource"
          >
            <el-input
              v-model="tempCollectOrFeedAndCollect.carrierResource"
              onfocus="this.select()"
            />
          </el-form-item>
        </el-form>
        <div style="text-align: center;">
          <el-button
            v-loading.fullscreen.lock="fullscreenLoading"
            class="filter-item"
            type="primary"
            @click="collectOrFeedAndCollectMode('receipt')"
          >
            {{ $t('system.receipt') }}
          </el-button>
          <el-button
            v-loading.fullscreen.lock="fullscreenLoading"
            class="filter-item"
            type="primary"
            @click="print()"
          >
            {{ $t('share.print') }}
          </el-button>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog
      :visible.sync="dialogCloseWorkOrder"
    >
      <el-select
        v-model="closeWorkOrderReason"
        :placeholder="$t('system.closeWorkOrderReason')"
      >
        <el-option
          v-for="item in closeWorkOrderRemark"
          :key="item.code"
          :label="$t('closeWorkOrderRemark.'+item.name)"
          :value="item.code"
        />
      </el-select>
      <el-divider />
      <el-button
        id="clearFilter"
        class="filter-item"
        type="primary"
        size="medium"
        @click="checkReason()"
      >
        {{ $t('share.confirm') }}
      </el-button>
    </el-dialog>
  </div>
</template>
<script src="./jobView.ts" lang="ts"></script>
<style src="./jobView.scss" lang="scss"></style>
