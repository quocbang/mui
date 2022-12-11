<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.materialQuery') }}
    </el-header>
    <el-tabs
      v-model="activeName"
      type="border-card"
      @tab-click="handleClick"
    >
      <el-tab-pane
        id="productTypeQuery"
        :key="'productTypeQuery'"
        name="productTypeQuery"
        :label="$t('system.productTypeQuery')"
      >
        <div
          v-if="activeName==='productTypeQuery'"
          style="text-align: center"
        >
          <el-form
            :inline="true"
            class="demo-form-inline"
          >
            <el-form-item
              id="productType"
              :label="$t('system.productType')"
            >
              <el-select
                id="productType"
                v-model="productTypeValue"
                filterable
                :placeholder="$t('system.productType')"
                @change="getGetProductIDList()"
              >
                <el-option
                  v-for="item in productTypeInfoList"
                  :key="item.type"
                  :label="item.type"
                  :value="item.type"
                />
              </el-select>
            </el-form-item>
            <el-form-item
              id="productID"
              :label="$t('system.productID')"
            >
              <el-select
                id="productID"
                v-model="productIDValue"
                filterable
                :placeholder="$t('system.productID')"
              >
                <el-option
                  v-for="item in productIDInfoList"
                  :key="item"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </el-form-item>
            <el-form-item
              id="status"
              :label="$t('share.status')"
            >
              <el-select
                id="status"
                v-model="materialStatusValue"
                filterable
                :placeholder="$t('share.status')"
              >
                <el-option
                  v-for="item in materialStatusList"
                  :key="item.ID"
                  :label="$t('material.status.'+ item.name)"
                  :value="item.ID"
                />
              </el-select>
            </el-form-item>
            <el-form-item
              id="date"
              :label="$t('share.date')"
            >
              <el-date-picker
                v-model="dateValue"
                type="date"
                :placeholder="$t('share.date')"
              />
            </el-form-item>
            <el-form-item>
              <el-button
                id="query"
                class="filter-item"
                type="primary"
                size="medium"
                @click="activeName==='productTypeQuery'?getTypeMaterialInfo():getMaterialLabelCardInfo()"
              >
                {{ $t('share.query') }}
              </el-button>
            </el-form-item>
            <el-form-item>
              <el-button
                id="clearFilter"
                class="filter-item"
                type="success"
                size="medium"
                @click="clearFilter()"
              >
                {{ $t('share.clearFilter') }}
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>
      <el-tab-pane
        id="barcodeQuery"
        :key="'barcodeQuery'"
        name="barcodeQuery"
        :label="$t('system.barcodeQuery')"
      >
        <div v-if="activeName==='barcodeQuery'">
          <el-form
            style="text-align: center"
            :inline="true"
            class="demo-form-inline"
          >
            <el-form-item
              id="materialLabelCard"
              :label="$t('system.materialLabelCard')"
              prop="materialLabelCard"
            >
              <el-input
                ref="materialLabelCard"
                v-model="materialLabelCardValue"
                @input="getMaterialLabelCardInfo()"
              />
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>
    </el-tabs>
    <div
      class="app-container"
    >
      <div
        v-if="isPlanAlive"
        class="filter-container"
      >
        <el-table
          :key="tableKey"
          v-loading="listLoading"
          :data="materialQueryList"
          border
          fit
          highlight-current-row
          style="width: 100%"
        >
          <el-table-column
            :label="$t('system.materialID')"
            align="center"
            prop="materialID"
          >
            <template slot-scope="{row}">
              <span>{{ row.ID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.materialLabelCard')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.resourceID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.quantity')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.quantity }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.minDosage')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.minimumDosage }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('share.status')"
            align="center"
          >
            <template slot-scope="{row}">
              <span><el-tag>{{ $t('material.status.'+ row.statusName) }}</el-tag></span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.reasonForInspection')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>
                <el-tag
                  v-for="item in row.inspections"
                  :key="item.ID"
                  :label="item.remark"
                >
                  {{ item.remark }}
                </el-tag></span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.expirationTime')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.expiredDate | formatDate }}</span>
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
                @click="openDetailInfo(row)"
              />
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('share.action')"
            align="center"
            class-name="fixed-width"
          >
            <template slot-scope="{row}">
              <el-button
                id="showPDF"
                type="danger"
                size="small"
                icon="el-icon-printer"
                circle
                @click="showPDFView(row)"
              />
              <el-button
                id="batchLabelingCard"
                type="primary"
                size="small"
                icon="el-icon-document-copy"
                circle
                @click="batchLabelingCard(row)"
              />
            </template>
          </el-table-column>
        </el-table>
        <pagination
          v-if="activeName==='productTypeQuery'"
          v-show="total>0"
          :total="total"
          :start.sync="paginationLimit"
          :page.sync="listQuery.page"
          :limit.sync="listQuery.limit"
          @pagination="activeName==='productTypeQuery'?getTypeMaterialInfo():getMaterialLabelCardInfo()"
        />
      </div>
      <!-- batchMaterialLabelCard -->
      <el-dialog
        :title="$t('system.batchMaterialLabelCard')"
        :visible.sync="dialogBatchMaterialLabelCardVisible"
      >
        <el-form
          ref="BatchMaterialLabelCardForm"
          :rules="batchMaterialLabelCardFormRules"
          :model="tempBatchMaterialData"
          status-icon
          label-position="left"
          label-width="200px"
          style="margin-left:50px;"
          autocomplete
        >
          <el-form-item
            :label="$t('system.currentQuantity')"
            prop="quantity"
          >
            <span>{{ tempBatchMaterialData.quantity }}</span>
          </el-form-item>

          <el-form-item
            :label="$t('system.materialCardBarcode')"
            prop="resourceID"
          >
            <span>{{ tempBatchMaterialData.resourceID }}</span>
          </el-form-item>
          <el-form-item
            :label="$t('system.batchQuantity')"
            prop="splitQuantity"
          >
            <el-input-number
              v-model="tempBatchMaterialData.splitQuantity"
              :precision="6"
              :min="0"
              :max="tempBatchMaterialData.quantity-0.000001"
            />
          </el-form-item>
          <el-form-item
            :label="$t('system.checkID')"
            prop="inspections"
          >
            <el-checkbox-group
              v-model="tempBatchMaterialData.inspections"
              :max="checkList.length-1"
            >
              <el-checkbox-button
                v-for="item in checkList"
                :key="item.ID"
                :label="item.ID"
              />
            </el-checkbox-group>
          </el-form-item>
          <el-form-item
            :label="
              $t('share.remark')"
            prop="remark"
          >
            <el-input
              v-model="tempBatchMaterialData.remark"
              type="textarea"
              :rows="2"
            />
          </el-form-item>
        </el-form>
        <div
          slot="footer"
          class="dialog-footer"
        >
          <el-button
            id="confirm"
            type="primary"
            @click="getMaterialSplit()"
          >
            {{ $t('share.confirm') }}
          </el-button>
        </div>
      </el-dialog>
      <el-dialog
        :title="$t('share.details')"
        :visible.sync="dialogDetailFormVisible"
      >
        <el-form
          ref="dataForm"
          :model="materialQueryRow"
          status-icon
          label-position="left"
          label-width="200px"
          style="margin-left:50px;"
          autocomplete
        >
          <el-divider />
          <el-form-item :label="$t('system.location')">
            <span>{{ materialQueryRow.warehouse.location }}</span>
          </el-form-item>

          <el-form-item :label="$t('system.carrierID')">
            <span>{{ materialQueryRow.carrierID }}</span>
          </el-form-item>

          <el-form-item :label="$t('system.unit')">
            <span>{{ materialQueryRow.unit }}</span>
          </el-form-item>

          <el-form-item :label="$t('system.createTime')">
            <span>{{ materialQueryRow.createdAt | formatDate }}</span>
          </el-form-item>

          <el-form-item :label="$t('system.changeTime')">
            <span>{{ materialQueryRow.updatedAt | formatDate }}</span>
          </el-form-item>

          <el-form-item :label="$t('system.changeStaff')">
            <span>{{ materialQueryRow.updatedBy }}</span>
          </el-form-item>

          <el-form-item :label="$t('share.remark')">
            <span>{{ materialQueryRow.remark }}</span>
          </el-form-item>
          <el-divider />
        </el-form>
      </el-dialog>
    </div>
  </div>
</template>
<script src="./materialQuery.ts" lang="ts"></script>
