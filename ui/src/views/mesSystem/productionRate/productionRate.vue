<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.productionRate') }}
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
              v-model="departmentID"
              filterable
              :placeholder="$t('system.departmentID')"
            >
              <el-option
                v-for="item in departmentInfoList"
                :key="item.departmentID"
                :label="item.departmentID"
                :value="item.departmentID"
              />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-date-picker
              v-model="dateValue"
              type="daterange"
              unlink-panels
              style="width: 600px"
              :range-separator="$t('share.to')"
              :start-placeholder="$t('share.startDate')"
              :end-placeholder="$t('share.endDate')"
              @change="onListWorkOrderRate()"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              id="excel"
              class="filter-item"
              type="primary"
              size="medium"
              @click="getProductionRate()"
            >
              {{ $t('share.excel') }}
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
      <div
        class="app-container"
      >
        <el-table
          :key="tableKey"
          :data="data"
          border
          highlight-current-row
          style="width: 100%"
        >
          <el-table-column
            :label="$t('system.departmentID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.departmentID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('workOrder.ID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.workOrderID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.productID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.productID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.stationID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.station }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.quantity')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.planQuantity }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.currentQuantity')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.collectedQuantity }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.ratio')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.ratio }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.productionTime')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.productionTime }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.productionEndTime')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.productionEndTime }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.updateBy')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.updateBy }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.createdBy')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.createdBy }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('recipe.ID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.recipeID }}</span>
            </template>
          </el-table-column>
        </el-table>
        <pagination
          v-show="total>0"
          :total="total"
          :page.sync="listQuery.page"
          :limit.sync="listQuery.limit"
          @pagination="onListWorkOrderRate"
        />
      </div>
    </el-main>
  </div>
</template>

<script src="./productionRate.ts" lang="ts"></script>
<style src="./productionRate.css"></style>
