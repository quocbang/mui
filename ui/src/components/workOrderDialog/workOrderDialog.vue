<template>
  <el-dialog
    :visible.sync="dialog"
    :title="$t('share.create')"
    width="70%"
  >
    <el-form
      ref="workOrderDataForm"
      v-loading="workOrderLoading"
      :rules="workOrderRules"
      :model="tempWorkOrderData"
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
        <el-input
          v-model="tempWorkOrderData.productID"
          disabled
        />
      </el-form-item>
      <el-divider />
      <el-form-item
        :label="$t('system.stationID')"
        prop="station"
      >
        <el-select
          v-model="tempWorkOrderData.station"
          filterable
          :disabled="tempWorkOrderData.preWorkOrder ? '' : false"
          class="filter-item"
          @change="getRecipesInfo(tempWorkOrderData.station)"
        >
          <el-option
            v-for="item in stationInfoFormatList"
            :key="item.ID"
            :label="item.ID"
            :value="item.ID"
          />
        </el-select>
      </el-form-item>
      <el-form-item
        :label="$t('recipe.ID')"
        prop="recipe.ID"
      >
        <el-select
          v-model="tempWorkOrderData.recipe.ID"
          filterable
          :disabled="tempWorkOrderData.preWorkOrder ? '' : false"
          class="filter-item"
          @change="getProcessInfoName(tempWorkOrderData.recipe.ID)"
        >
          <el-option
            v-for="(item, index) in recipesList"
            :key="index"
            :label="item"
            :value="item"
          />
        </el-select>
      </el-form-item>
      <el-form-item
        :label="$t('system.processName')"
        prop="recipe.processName"
      >
        <el-select
          v-model="tempWorkOrderData.recipe.processName"
          filterable
          :disabled="tempWorkOrderData.recipe.preWorkOrder ? '' : false"
          class="filter-item"
          @change="getProcessInfoType(tempWorkOrderData.recipe.processName)"
        >
          <el-option
            v-for="(item, index) in processInfoNameList"
            :key="index"
            :label="item"
            :value="item"
          />
        </el-select>
      </el-form-item>
      <el-form-item
        :label="$t('system.processType')"
        prop="recipe.processType"
      >
        <el-select
          v-model="tempWorkOrderData.recipe.processType"
          :disabled="tempWorkOrderData.preWorkOrder ? '' : false"
          class="filter-item"
          @change="getBatchSizeAndBomInfo(tempWorkOrderData.recipe.ID,tempWorkOrderData.recipe.processName ,tempWorkOrderData.recipe.processType)"
        >
          <el-option
            v-for="item in processInfoTypeList"
            :key="item.processInfo.Type"
            :label="item.processInfo.Type"
            :value="item.processInfo.Type"
          />
        </el-select>
      </el-form-item>
      <template v-if="tempWorkOrderData.station!=='' && tempWorkOrderData.recipe.ID!=='' && tempWorkOrderData.recipe.processType!=''">
        <el-form-item
          :label="$t('share.calculation')"
          prop="batchCalculation"
        >
          <el-radio
            v-model="tempWorkOrderData.batchCalculation"
            :disabled="tempWorkOrderData.preWorkOrder ? '' : false"
            :label="false"
          >
            {{ $t('system.batch') }}
          </el-radio>
          <el-radio
            v-model="tempWorkOrderData.batchCalculation"
            :disabled="tempWorkOrderData.preWorkOrder ? '' : false"
            :label="true"
          >
            {{ $t('system.estimateOutput') }}
          </el-radio>
        </el-form-item>
        <el-form-item
          :label="$t('system.batchSize')"
          prop="preBatchSize"
        >
          <el-input
            v-model="tempWorkOrderData.preBatchSize"
            :disabled="disabled"
            label-width="50px"
            @input="calculateQuantity(tempWorkOrderData.batchCount)"
          />
        </el-form-item>
        <template v-if="tempWorkOrderData.batchCalculation===true">
          <el-form-item
            :label="$t('system.batch')"
            prop="batchesCount"
          >
            <el-input-number
              v-model="tempWorkOrderData.batchCount"
              :min="1"
              :step="1"
              step-strictly
              disabled
              label-width="50px"
            />
          </el-form-item>
          <el-form-item
            :label="$t('system.estimateOutput')"
            prop="quantity"
          >
            <el-input-number
              v-model="tempWorkOrderData.quantity"
              :disabled="tempWorkOrderData.preWorkOrder ? '' : false"
              :step="0.1"
              :min="1"
              @input="calculateBatch(tempWorkOrderData.quantity)"
            />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item
            :label="$t('system.batch')"
            prop="batchesCount"
          >
            <el-input-number
              v-model="tempWorkOrderData.batchCount"
              :disabled="tempWorkOrderData.preWorkOrder ? '' : false"
              :min="1"
              :step="1"
              step-strictly
              @input="calculateQuantity(tempWorkOrderData.batchCount)"
            />
          </el-form-item>
          <el-form-item
            :label="$t('system.estimateOutput')"
            prop="quantity"
          >
            <el-input-number
              v-model="tempWorkOrderData.quantity"
              disabled
              :min="1"
              label-width="50px"
              :step="0.1"
            />
          </el-form-item>
        </template>
      </template>

      <el-form-item
        :label="$t('system.estimateDate')"
      >
        <el-date-picker
          v-model="tempWorkOrderData.planDate"
          :disabled="tempWorkOrderData.preWorkOrder ? '' : false"
          type="date"
        />
      </el-form-item>
      <el-form-item
        v-if="targetAction == 'create'"
        :label="$t('system.automaticConstruction')"
      >
        <el-checkbox
          v-model="tempWorkOrderData.preWorkOrder"
          @change="checkValueExists()"
        />
      </el-form-item>

      <el-divider />
      <template v-if="tempWorkOrderData.preWorkOrder == true">
        <el-table
          ref="multipleTable"
          :data="tempBomInfo"
          border
          fit
          highlight-current-row
          style="width: 100%"
          @selection-change="handleSelectionChange"
        >
          <el-table-column
            type="selection"
            width="55"
          />
          <el-table-column
            :label="$t('system.productID')"
            align="center"
            prop="wordId"
          >
            <template slot-scope="{row}">
              <span>{{ row.productID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.stationID')"
            align="center"
          >
            <template slot-scope="{row,$index}">
              <el-select
                v-model="row.station"
                filterable
                @change="getBomRecipesInfo(row,$index)"
              >
                <el-option
                  v-for="item in row.stationInfo"
                  :key="item.ID"
                  :label="item.ID"
                  :value="item.ID"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('recipe.ID')"
            align="center"
          >
            <template slot-scope="{row,$index}">
              <el-select
                ref="bomRecipe"
                v-model="row.recipeID"
                filterable
                @change="getBomProcessInfoName(row,$index)"
              >
                <el-option
                  v-for="(item,index) in BomRecipesList"
                  :key="index"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.processName')"
            align="center"
          >
            <template slot-scope="{row, $index}">
              <el-select
                v-model="row.processName"
                filterable
                @change="getBomProcessInfoType(row.processName,$index)"
              >
                <el-option
                  v-for="(item,index) in BomProcessInfoNameList"
                  :key="index"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.processType')"
            align="center"
          >
            <template slot-scope="{row, $index}">
              <el-select
                v-model="row.processType"
                @change="getBomQuantityInfo(row,$index)"
              >
                <el-option
                  v-for="item in BomProcessInfoTypeList"
                  :key="item.processInfo.Type"
                  :label="item.processInfo.Type"
                  :value="item.processInfo.Type"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.estimateDate')"
          >
            <template slot-scope="{row}">
              <el-date-picker
                v-model="row.planDate"
                type="date"
              />
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.estimateOutput')"
          >
            <template slot-scope="{row}">
              <span>{{ row.totalQuantity }}</span>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </el-form>
    <div
      slot="footer"
      class="dialog-footer"
    >
      <el-button @click="dialog = false">
        {{ $t('share.cancel') }}
      </el-button>
      <el-button
        type="primary"
        @click="targetAction==='create'? createWorkOrderToDB():updateWorkOrderToDB()"
      >
        {{ $t('share.confirm') }}
      </el-button>
    </div>
  </el-dialog>
</template>
<script src="./workOrderDialog.ts" lang="ts"></script>
