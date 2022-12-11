<template>
  <div
    class="app-container"
  >
    <el-header class="header-style">
      {{ $t('router.recipeProcess') }}
    </el-header>
    <el-main>
      <!-- query data -->
      <el-form
        :model="targetData"
        status-icon
        label-position="left"
        label-width="200px"
        style="margin-left:50px;"
        autocomplete
      >
        <el-divider />
        <el-form-item :label="$t('system.departmentID')">
          <span>{{ targetData.departmentOID }}</span>
        </el-form-item>

        <el-form-item :label="$t('system.productType')">
          <span>{{ targetData.productType }}</span>
        </el-form-item>

        <el-form-item :label="$t('system.lastProductID')">
          <span>{{ targetData.productID }}</span>
        </el-form-item>

        <el-form-item :label="$t('recipe.ID')">
          <span>{{ targetData.recipe }}</span>
        </el-form-item>
        <el-divider />
      </el-form>
      <div
        class="app-container"
      >
        <!-- main data -->
        <el-table
          :data="targetData.row"
          border
          fit
          highlight-current-row
          style="width: 100%"
        >
          <el-table-column
            :label="$t('system.processName')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.processType')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.type }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.stationID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                v-for="(item) in row.stations"
                :key="item.ID"
              >
                <el-tag>{{ item.ID }}</el-tag>
              </span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.productID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                class="link-type"
              >{{ row.product.id }}</span>
            </template>
          </el-table-column>
        </el-table>

        <!-- materialList data -->
        <el-divider content-position="left">
          {{ $t('system.materialList') }}
        </el-divider>
        <el-table
          :data="targetData.materialDataInfo"
          border
          fit
          highlight-current-row
        >
          <el-table-column
            :label="$t('system.productID')"
            prop="name"
          >
            <template slot-scope="{row}">
              <span>{{ row.productID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.materialGrade')"
            prop="name"
          >
            <template slot-scope="{row}">
              <span>{{ row.grade }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.maxValue')"
            prop="name"
          >
            <template slot-scope="{row}">
              <span>{{ row.maxValue }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.standardValue')"
            prop="name"
          >
            <template slot-scope="{row}">
              <span>{{ row.standardValue }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.minValue')"
            prop="name"
          >
            <template slot-scope="{row}">
              <span>{{ row.minValue }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.requiredRecipeID')"
            prop="name"
          >
            <template slot-scope="{row}">
              <span>{{ row.requiredRecipeID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.substituteMaterial')"
            prop="name"
          >
            <template slot-scope="{row}">
              <span>{{ row.substitutesList }}</span>
            </template>
          </el-table-column>
        </el-table>
        <!-- stationControl data -->
        <el-divider content-position="left">
          {{ $t('system.stationControl') }}
        </el-divider>
        <div
          v-for="(item) in stationControlSliceTable"
          :key="item.name"
        >
          <div v-if="targetData.station!==''">
            <el-table
              :data="targetData.stationControlStepDataInfo.newRows"
              highlight-current-row
              style="width: 100%"
            >
              <el-table-column
                type="index"
                width="60"
                :label="$t('share.ID')"
              />
              <el-table-column
                align="center"
                :label="$t('system.operatingStep')"
              >
                <el-table-column
                  v-for="columns in item"
                  id="operatingStep"
                  :key="columns.name"
                  :prop="columns.name"
                  :label="columns.unit =='' ? columns.name : columns.name + '('+ columns.unit +')'"
                />
              </el-table-column>
            </el-table>
          </div>
        </div>
        <el-divider />
        <el-table
          :data="targetData.stationControlCommonDataInfo"
          highlight-current-row
          style="width: 100%"
        >
          <el-table-column
            :label="$t('system.cuttingConditions')"
            align="center"
          >
            <el-table-column
              :label="$t('system.item')"
            >
              <template
                slot-scope="{row}"
              >
                <div v-if="row.unit==''">
                  <span>{{ row.rowName }}</span>
                </div>
                <div v-else>
                  <span>{{ row.rowName + '('+ row.unit +')' }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.max')"
            >
              <template slot-scope="{row}">
                <span>{{ row.maxValue }}</span>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.min')"
            >
              <template slot-scope="{row}">
                <span>{{ row.minValue }}</span>
              </template>
            </el-table-column>
          </el-table-column>
        </el-table>
      </div>
    </el-main>
    <el-button
      id="printPageButton"
      type="primary"
      @click="printContent()"
    >
      {{ $t('share.confirm') }}
    </el-button>
  </div>
</template>

<script src="./printView.ts" lang="ts"></script>
<style src="./printView.css" media="print"></style>
