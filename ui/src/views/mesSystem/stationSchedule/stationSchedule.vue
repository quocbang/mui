<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.stationSchedule') }}
    </el-header>
    <el-main>
      <el-card
        class="box-card"
        style="text-align: center;"
      >
        <el-form
          :inline="true"
          class="demo-form-inline"
        >
          <el-form-item :label="$t('system.departmentID')">
            <el-select
              v-model="departmentOIDValue"
              filterable
              :placeholder="$t('system.departmentID')"
              @change="onDepartmentGetStationList(departmentOIDValue)"
            >
              <el-option
                v-for="item in departmentInfoList"
                :key="item.departmentID"
                :label="item.departmentID"
                :value="item.departmentID"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('system.productionStation')">
            <el-select
              v-model="stationValue"
              filterable
              :placeholder="$t('system.productionStation')"
              @change="onGetStationScheduleList(stationValue, dateValue)"
            >
              <el-option
                v-for="item in stationInfoList"
                :key="item.ID"
                :label="item.ID"
                :value="item.ID"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('share.date')">
            <el-date-picker
              v-model="dateValue"
              type="date"
              :placeholder="$t('share.date')"
              @change="onGetStationScheduleList(stationValue, dateValue)"
            />
          </el-form-item>
        </el-form>
      </el-card>
      <div
        class="app-container"
      >
        <el-table
          ref="draggableTable"
          v-loading="listLoading"
          :data="scheduleList"
          border
          fit
          highlight-current-row
          style=":min-width: 100%"
          @selection-change="handleSelectionChange"
        >
          <el-table-column
            type="selection"
            width="55"
          />
          <el-table-column
            :label="$t('recipe.ID')"
            :min-width="100"
          >
            <template slot-scope="{row}">
              <span>{{ row.recipe.ID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.productID')"
            :min-width="100"
          >
            <template slot-scope="{row}">
              <span>{{ row.productID }}</span>
            </template>
          </el-table-column>

          <el-table-column
            :label="$t('system.batch')"
            :min-width="100"
          >
            <template slot-scope="{row}">
              <div
                v-if="row.batchSize===BatchSize.PerBatchQuantities"
              >
                <span>{{ row.batchesQuantity.length }}</span>
              </div>
              <div v-else-if="row.batchSize==BatchSize.FixedQuantity || row.batchSize==BatchSize.PlanQuantity">
                <span>{{ row.batchCount }}</span>
              </div>
              <div v-else>
                <span>{{ $t("message.notifyWorkOrder100") }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column
            align="center"
            :label="$t('share.status')"
            :min-width="100"
          >
            <template slot-scope="{row}">
              <span>
                <el-tag disable-transitions>{{ $t("workOrder." + row.statusName) }}</el-tag></span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('share.action')"
            align="left"
            class-name="fixed-width"
          >
            <template slot-scope="{row}">
              <el-button
                size="small"
                icon="el-icon-search"
                circle
                @click="openDetailInfo(row)"
              />
              <el-button
                v-if="row.status==0"
                size="small"
                type="primary"
                icon="el-icon-edit"
                circle
                @click="updateWorkOrder(row)"
              />
              <el-button
                v-if="row.status==0"
                type="danger"
                size="small"
                icon="el-icon-printer"
                circle
                @click="showPDFView(row)"
              />
            </template>
          </el-table-column>
        </el-table>
        <div
          v-if="isPlanAlive"
          style="margin-top: 20px"
        >
          <el-button @click="stopWorkOrder()">
            {{ $t('system.stopWorkOrder') }}
          </el-button>
          <el-button @click="confirmOrder()">
            {{ $t('system.confirmOrder') }}
          </el-button>
        </div>
        <!-- details dialog -->
        <el-dialog
          :title="$t('share.details')"
          :visible.sync="dialogDetailFormVisible"
        >
          <el-form
            ref="dataForm"
            :model="tempStationScheduleListData"
            status-icon
            label-position="left"
            label-width="200px"
            style="margin-left:50px;"
            autocomplete
          >
            <el-divider />
            <el-form-item :label="$t('workOrder.ID')">
              <span>{{ tempStationScheduleListData.ID }}</span>
            </el-form-item>

            <el-form-item :label="$t('system.estimateOutput')">
              <span>{{ tempStationScheduleListData.planQuantity }}</span>
            </el-form-item>

            <el-form-item :label="$t('system.estimateDate')">
              <span>{{ tempStationScheduleListData.planDate }}</span>
            </el-form-item>

            <el-form-item :label="$t('system.dispatchStaff')">
              <span>{{ tempStationScheduleListData.updateBy }}</span>
            </el-form-item>

            <el-form-item :label="$t('system.lastUpdated')">
              <span>{{ tempStationScheduleListData.updateAt }}</span>
            </el-form-item>
            <el-divider />
          </el-form>
        </el-dialog>
        <!-- update workOrder dialog -->
        <WorkOrderDialog
          ref="refWorkOrder"
          :department-o-i-d-value="departmentOIDValue"
          @onGetStationScheduleList="onGetStationScheduleList(stationValue, dateValue)"
        />
      </div>
    </el-main>
  </div>
</template>

<script src="./stationSchedule.ts" lang="ts"></script>
<style src="./stationSchedule.css"></style>
