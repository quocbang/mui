<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.productionPlan') }}
    </el-header>
    <el-main>
      <!-- Top condition query -->
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
              @change="onDepartmentList(departmentOIDValue)"
            >
              <el-option
                v-for="item in departmentInfoList"
                :key="item.label.OID"
                :label="item.label.ID"
                :value="item.label.OID"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('system.productType')">
            <el-select
              v-model="productTypeValue"
              filterable
              :placeholder="$t('system.productType')"
              @change="onGetPlanList()"
            >
              <el-option
                v-for="item in productTypeInfoList"
                :key="item.type"
                :label="item.type"
                :value="item.type"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('share.date')">
            <el-date-picker
              v-model="dateValue"
              type="date"
              :placeholder="$t('share.date')"
              @change="onGetPlanList()"
            />
          </el-form-item>
        </el-form>
      </el-card>
      <!-- Query plan Info -->
      <div
        class="app-container"
      >
        <!-- // add Plan Button -->
        <div
          v-if="isPlanAlive"
          class="filter-container"
        >
          <el-button
            class="filter-item"
            type="primary"
            icon="el-icon-circle-plus-outline"
            size="medium"
            @click="createPlan"
          >
            {{ $t('share.create') }}
          </el-button>
          <el-button
            class="filter-item"
            type="warning"
            icon="el-icon-upload2"
            size="medium"
            @click="handleUpload"
          >
            <input
              ref="excel-upload-input"
              class="excel-upload-input"
              type="file"
              accept=".xlsx, .xls"
              @change="handleClick"
            >
            {{ $t('share.importExcel') }}
          </el-button>
        </div>
        <!-- // Plan Table -->
        <el-table
          :key="tableKey"
          v-loading="listLoading"
          :data="planList"
          border
          fit
          highlight-current-row
          style="width: 100%"
          row-key="productID"
          :default-sort="{prop: 'productID',order: 'ascending'}"
          :tree-props="{children: 'children', hasChildren: 'hasChildren'}"
        >
          <!-- // workOrder Table -->
          <el-table-column
            :label="$t('system.productID')"
            prop="productID"
            sortable
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.productID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.demandDay')"
            prop="dayQuantity"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.dayQuantity }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.demandWeek')"
            prop="weekQuantity"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.weekQuantity }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.inventory')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.stockQuantity }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.conversionDays')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.conversionDays }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.workAssigned')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.scheduledQuantity }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="$t('system.generateWorkOrder')">
            <template slot-scope="{row}">
              <el-button
                v-if="row.scheduledQuantity!=='' && row.scheduledQuantity!=='-'"
                size="mini"
                type="success"
                icon="el-icon-s-promotion"
                circle
                @click="createWorkOrder(row.productID)"
              />
            </template>
          </el-table-column>
        </el-table>
        <!-- add Plan dialog -->
        <el-dialog
          :visible.sync="dialogCreatePlanFormVisible"
          :title="$t('share.create')"
        >
          <el-form
            ref="planDataForm"
            v-loading="planLoading"
            :rules="planRules"
            :model="tempPlanData"
            status-icon
            label-position="left"
            label-width="200px"
            style="margin-left:50px;"
            autocomplete
          >
            <el-form-item
              :label="$t('system.productID')"
              prop="productID"
            >
              <el-select
                v-model="tempPlanData.productID"
                filterable
                class="filter-item"
              >
                <el-option
                  v-for="(item,index) in productIDList"
                  :key="index"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </el-form-item>
            <el-form-item
              :label="$t('system.demandDay')"
              prop="dayQuantity"
            >
              <el-input
                v-model="tempPlanData.dayQuantity"
                label-width="50px"
              />
            </el-form-item>
          </el-form>
          <div
            slot="footer"
            class="dialog-footer"
          >
            <el-button @click="cancelWorkOrder()">
              {{ $t('share.cancel') }}
            </el-button>
            <el-button
              type="primary"
              @click="createPlanToDB()"
            >
              {{ $t('share.confirm') }}
            </el-button>
          </div>
        </el-dialog>
        <!-- add WorkOrder dialog -->
        <WorkOrderDialog
          ref="refWorkOrder"
          :department-o-i-d-value="departmentOIDValue"
          @onGetPlanList="onGetPlanList"
        />
      </div>
    </el-main>
    <el-footer />
  </div>
</template>

<script src="./productionPlan.ts" lang="ts"></script>
<style src="./productionPlan.css"></style>
