<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.recipeProcess') }}
    </el-header>
    <el-main>
      <el-card
        class="box-card"
        style="text-align: center"
      >
        <el-form
          :inline="true"
          class="demo-form-inline"
        >
          <el-form-item :label="$t('system.departmentID')">
            <el-select
              id="departmentID"
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
              id="productType"
              v-model="productTypeValue"
              filterable
              :placeholder="$t('system.productType')"
              @change="onProductIDList()"
            >
              <el-option
                v-for="item in productTypeInfoList"
                :key="item.type"
                :label="item.type"
                :value="item.type"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('system.lastProductID')">
            <el-select
              id="lastProductID"
              v-model="productIDValue"
              filterable
              :placeholder="$t('system.lastProductID')"
              @change="onRecipeList()"
            >
              <el-option
                v-for="item in productIDInfoList"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('recipe.ID')">
            <el-select
              id="recipeID"
              v-model="recipeValue"
              filterable
              :placeholder="$t('recipe.ID')"
              @change="onRecipeProcessList()"
            >
              <el-option
                v-for="item in recipeInfoList"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
      <div
        class="app-container"
      >
        <el-table
          :key="tableKey"
          v-loading="listLoading"
          :data="recipeProcessTable"
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
              <span>{{ row.requiredFlows.name }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.processType')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.requiredFlows.type }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.stationID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                v-for="(item) in row.requiredFlows.stations"
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
              >{{ row.requiredFlows.product.id }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('share.details')"
            align="center"
          >
            <template slot-scope="{row}">
              <el-button
                size="small"
                icon="el-icon-search"
                circle
                @click="openDetailInfo(row.requiredFlows)"
              />
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.hasOptionFlow')"
            align="center"
          >
            <template
              slot-scope="{row}"
            >
              <el-button
                v-if="row.optionalFlows.length!==0"
                id="hasOptionFlow"
                size="small"
                icon="el-icon-search"
                circle
                @click="openOptionFlowDetailInfo(row)"
              />
            </template>
          </el-table-column>
        </el-table>
        <!-- openDetailInfo -->
        <el-dialog
          :title="$t('share.details')"
          :visible.sync="dialogDetailVisible"
        >
          <el-form
            :inline="true"
            class="demo-form-inline"
          >
            <el-form-item :label="$t('system.stationID')">
              <el-select
                id="stationID"
                v-model="stationValue"
                filterable
                :placeholder="$t('system.stationID')"
                @change="onDetailInfo(stationValue,stationList)"
              >
                <el-option
                  v-for="item in stationList"
                  :key="item.ID"
                  :label="item.ID"
                  :value="item.ID"
                />
              </el-select>
            </el-form-item>
          </el-form>
          <el-divider content-position="left">
            {{ $t('system.materialList') }}
          </el-divider>
          <el-table
            :key="tableKey"
            :data="materialDataInfo"
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
          <el-divider content-position="left">
            {{ $t('system.stationControl') }}
          </el-divider>
          <div v-if="stationValue!==''">
            <el-table
              :data="stationControlStepDataInfo.newRows"
              highlight-current-row
              style="width: 100%"
            >
              <el-table-column
                align="center"
                :label="$t('system.operatingStep')"
              >
                <el-table-column
                  v-for="columns in stationControlStepDataInfo.columns"
                  id="operatingStep"
                  :key="columns.name"
                  :prop="columns.name"
                  :label="columns.unit =='' ? columns.name : columns.name + '('+ columns.unit +')'"
                />
              </el-table-column>
            </el-table>
            <el-table
              :data="stationControlCommonDataInfo"
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
          <el-divider />
          <el-button
            type="primary"
            @click="sendDataInfo"
          >
            {{ $t('share.previewPrint') }}
          </el-button>
        </el-dialog>
        <!-- openOptionFlowInfo -->
        <el-dialog
          :title="$t('system.optionFlow')"
          :visible.sync="dialogOptionFlowDetailVisible"
        >
          <el-table
            :key="tableKey"
            :data="anewProcessDataInfo"
            border
            fit
            highlight-current-row
          >
            <el-table-column type="expand">
              <template slot-scope="{row}">
                <el-table
                  :key="tableKey"
                  :data="row.processes"
                  border
                  fit
                  highlight-current-row
                  style="width: 100%"
                >
                  <el-table-column
                    :label="$t('system.processName')"
                    align="center"
                    prop="processID"
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
                        v-for="(item) in row.stationsList"
                        :key="item"
                      >
                        <el-tag>{{ item }}</el-tag>
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column
                    :label="$t('system.productID')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.product.id }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column
                    :label="$t('share.details')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <el-button
                        id="OptionFlowList"
                        size="small"
                        icon="el-icon-search"
                        circle
                        @click="openDetailInfo(row)"
                      />
                    </template>
                  </el-table-column>
                </el-table>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.name')"
              prop="name"
            >
              <template slot-scope="{row}">
                <span>{{ row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.maxRun')"
              align="center"
            >
              <template slot-scope="{row}">
                <span
                  class="link-type"
                >{{ row.maxRepetitions }}</span>
              </template>
            </el-table-column>
          </el-table>
        </el-dialog>
      </div>
    </el-main>

    <el-footer />
  </div>
</template>

<script src="./recipeProcess.ts" lang="ts"></script>
<style src="./recipeProcess.css"></style>
