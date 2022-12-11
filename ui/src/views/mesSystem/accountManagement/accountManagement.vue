<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.accountManagement') }}
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
              v-model="departmentOIDValue"
              filterable
              :placeholder="$t('system.departmentID')"
              @change="getAccountTable()"
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
      <div
        class="app-container"
      >
        <div
          v-if="isPlanAlive"
          class="filter-container"
        >
          <el-button
            class="filter-item"
            type="primary"
            icon="el-icon-circle-plus-outline"
            size="medium"
            @click="addAccountData()"
          >
            {{ $t('share.create') }}
          </el-button>
        </div>
        <el-table
          :key="tableKey"
          v-loading="listLoading"
          :data="accountTable"
          border
          fit
          highlight-current-row
          style="width: 100%"
        >
          <el-table-column
            :label="$t('system.employeeID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.employeeID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.role')"
            align="center"
          >
            <template slot-scope="{row}">
              <span
                v-for="(item) in row.rolesName"
                :key="item"
              >
                <el-tag>{{ $t("roles." + item) }}</el-tag>
              </span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.resetPassword')"
            align="center"
          >
            <template slot-scope="{row}">
              <el-button
                size="mini"
                type="success"
                icon="el-icon-refresh-left"
                circle
                @click="resetPassword(row)"
              />
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.action')"
            align="center"
          >
            <template slot-scope="{row}">
              <el-button
                size="mini"
                type="primary"
                icon="el-icon-edit"
                circle
                @click="updateAccountData(row)"
              />
              <el-button
                size="mini"
                type="danger"
                icon="el-icon-delete-solid"
                circle
                @click="deleteAccountData(row)"
              />
            </template>
          </el-table-column>
        </el-table>
        <el-dialog
          :title="dialogStatus==='create'?$t('share.create'):$t('share.update')"
          :visible.sync="dialogVisible"
        >
          <el-form
            ref="accountForm"
            :rules="rules"
            :model="tempAddData"
            status-icon
            label-position="left"
            label-width="200px"
            style="margin-left:50px;"
            autocomplete
          >
            <el-form-item
              :label="$t('system.employeeID')"
              prop="employeeID"
            >
              <el-select
                v-model="tempAddData.employeeID"
                filterable
                :placeholder="$t('system.employeeID')"
                :disabled="dialogStatus==='update'"
              >
                <el-option
                  v-for="item in userInfoList"
                  :key="item.employeeID"
                  :label="item.employeeID"
                  :value="item.employeeID"
                />
              </el-select>
            </el-form-item>
            <el-form-item
              :label="$t('system.role')"
              prop="roles"
            >
              <el-select
                v-model="tempAddData.roles"
                filterable
                multiple
                :placeholder="$t('system.role')"
              >
                <el-option
                  v-for="item in roleList"
                  :key="item.ID"
                  :label="$t('roles.' + item.name)"
                  :value="item.ID"
                  :disabled="item.disabled"
                />
              </el-select>
            </el-form-item>
            <span style="color:red">{{ $t('message.notify300') }}</span>
          </el-form>
          <div
            slot="footer"
            class="dialog-footer"
          >
            <el-button
              type="primary"
              @click="dialogStatus==='create'?addAccountDataToDB():updateAccountDataToDB()"
            >
              {{ $t('share.confirm') }}
            </el-button>
          </div>
        </el-dialog>
      </div>
    </el-main>

    <el-footer />
  </div>
</template>

<script src="./accountManagement.ts" lang="ts"></script>
<style src="./accountManagement.css"></style>
