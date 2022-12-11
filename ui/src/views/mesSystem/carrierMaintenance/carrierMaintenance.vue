<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.carrierMaintenance') }}
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
              id="departmentID"
              v-model="departmentOIDValue"
              filterable
              :placeholder="$t('system.departmentID')"
              @change="getCarrierInfo()"
            >
              <el-option
                v-for="item in departmentInfoList"
                :key="item.label.OID"
                :label="item.label.ID"
                :value="item.label.OID"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
    </el-main>
    <!-- Query Info -->
    <div
      class="app-container"
    >
      <div
        v-if="isPlanAlive"
        class="filter-container"
      >
        <el-button
          id="createCarrier"
          class="filter-item"
          type="primary"
          icon="el-icon-circle-plus-outline"
          size="medium"
          @click="createCarrier"
        >
          {{ $t('share.create') }}
        </el-button>
        <el-button
          id="printBarcode"
          class="filter-item"
          type="primary"
          icon="el-icon-circle-plus-outline"
          size="medium"
          @click="printBarcode"
        >
          {{ $t('share.previewPrint') }}
        </el-button>
      </div>
      <el-table
        :key="tableKey"
        v-loading="listLoading"
        :data="carrierList"
        border
        fit
        highlight-current-row
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <!-- // workOrder Table -->
        <el-table-column
          type="selection"
          width="55"
        />
        <el-table-column
          :label="$t('system.carrierID')"
          align="center"
          prop="ID"
        >
          <template slot-scope="{row}">
            <span>{{ row.ID }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('system.allowedMaterial')"
          align="center"
        >
          <template slot-scope="{row}">
            <span>{{ row.allowedMaterial }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('system.updateAt')"
          align="center"
        >
          <template slot-scope="{row}">
            <span
              class="link-type"
            >{{ row.updateAt }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('system.updateBy')"
          align="center"
        >
          <template slot-scope="{row}">
            <span>{{ row.updateBy }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('share.action')"
          align="center"
          class-name="fixed-width"
        >
          <template slot-scope="{row,$index}">
            <el-button
              id="updateCarrierRow"
              type="primary"
              size="small"
              icon="el-icon-edit"
              circle
              @click="updateCarrierRow(row,$index)"
            />

            <el-button
              id="deleteCarrierRow"
              size="mini"
              type="danger"
              icon="el-icon-delete-solid"
              circle
              @click="deleteCarrierRow(row)"
            />
          </template>
        </el-table-column>
      </el-table>
      <pagination
        v-show="total>0"
        :total="total"
        :start.sync="paginationLimit"
        :page.sync="listQuery.page"
        :limit.sync="listQuery.limit"
        @pagination="getCarrierInfo"
      />
    </div>
    <!-- create -->
    <el-dialog
      :title="$t('share.create')"
      :visible.sync="dialogCreateVisible"
    >
      <el-form
        ref="dataCreateForm"
        :rules="dataCreateFormRules"
        :model="tempAddData"
        status-icon
        label-position="left"
        label-width="200px"
        style="margin-left:50px;"
        autocomplete
      >
        <el-form-item
          :label="$t('system.idPrefix')"
          prop="idPrefix"
        >
          <el-input
            id="idPrefix"
            v-model="tempAddData.idPrefix"
            type="text"
            minlength="2"
            maxlength="2"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('system.addCarrierQuantity')"
          prop="quantity"
        >
          <el-input-number
            id="quantity"
            v-model="tempAddData.quantity"
            :min="1"
            :max="9999"
          />
        </el-form-item>
        <el-form-item
          :label="$t('system.allowedMaterial')"
          prop="allowedMaterial"
        >
          <el-input
            id="allowedMaterial"
            v-model="tempAddData.allowedMaterial"
            maxlength="20"
            show-word-limit
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
          @click="createCarrierInfoToDB()"
        >
          {{ $t('share.confirm') }}
        </el-button>
      </div>
    </el-dialog>
    <!-- update -->
    <el-dialog
      :title="$t('share.update')"
      :visible.sync="dialogUpdateVisible"
    >
      <el-form
        ref="dataUpdateForm"
        :model="tempUpdateData"
        status-icon
        label-position="left"
        label-width="200px"
        style="margin-left:50px;"
        autocomplete
      >
        <el-form-item
          :label="$t('system.carrierID')"
        >
          <span>{{ tempUpdateData.ID }}</span>
        </el-form-item>
        <el-form-item
          :label="$t('system.allowedMaterial')"
          prop="allowedMaterial"
        >
          <el-input
            id="allowedMaterial"
            v-model="tempUpdateData.allowedMaterial"
            maxlength="20"
            show-word-limit
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
          @click="UpdateCarrierInfoToDB()"
        >
          {{ $t('share.confirm') }}
        </el-button>
      </div>
    </el-dialog>

    <el-footer />
  </div>
</template>

<script src="./carrierMaintenance.ts" lang="ts"></script>
