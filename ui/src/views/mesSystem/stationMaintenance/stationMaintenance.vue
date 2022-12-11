<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.stationMaintenance') }}
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
              @change="queryStationData()"
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
          v-if="isAlive"
          class="filter-container"
        >
          <el-button
            class="filter-item"
            type="primary"
            icon="el-icon-circle-plus-outline"
            size="medium"
            @click="createStation"
          >
            {{ $t('share.create') }}
          </el-button>
        </div>
        <el-table
          :key="tableKey"
          :data="stationDataList"
          border
          highlight-current-row
          style="width: 100%"
        >
          <el-table-column
            :label="$t('system.stationID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.ID }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.machineCode')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.code }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.stationDescription')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.description }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.lastUpdated')"
            align="center"
          >
            <template
              slot-scope="{row}"
            >
              <span>{{ row.updateAt }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.personnel')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.updateBy }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.stationStatus')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ $t('station.status.' + row.stateName) }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('system.site')"
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
            <template slot-scope="{row,$index}">
              <el-button
                type="primary"
                size="small"
                icon="el-icon-edit"
                circle
                @click="updateStation(row,$index)"
              />
              <el-button
                size="mini"
                type="danger"
                icon="el-icon-delete-solid"
                circle
                @click="deleteStation(row)"
              />
            </template>
          </el-table-column>
        </el-table>
        <pagination
          v-show="total>0"
          :total="total"
          :page.sync="listQuery.page"
          :limit.sync="listQuery.pageSize"
          @pagination="queryStationData"
        />
      </div>
      <!-- create station -->
      <el-dialog
        :title="$t('share.create')"
        :visible.sync="dialogCreateVisible"
      >
        <el-form
          ref="stationDataForm"
          :rules="stationRules"
          :model="tempCreateWorkOrder"
          status-icon
          label-position="left"
          label-width="200px"
          style="margin-left:50px;"
          autocomplete
        >
          <el-form-item
            :label="$t('system.stationID')"
            prop="ID"
          >
            <el-input
              v-model="tempCreateWorkOrder.ID"
              maxlength="32"
              show-word-limit
            />
          </el-form-item>
          <el-form-item
            :label="$t('system.machineCode')"
            prop="code"
          >
            <el-input
              v-model="tempCreateWorkOrder.code"
            />
          </el-form-item>
          <el-form-item
            :label="$t('system.stationDescription')"
          >
            <el-input
              v-model="tempCreateWorkOrder.description"
              type="textarea"
              :rows="2"
            />
          </el-form-item>
          <el-form-item
            :label="$t('system.siteData')"
            align="center"
          />
          <el-divider />
          <div class="filter-container">
            <el-button
              class="filter-item"
              type="primary"
              icon="el-icon-circle-plus-outline"
              size="medium"
              @click="addSitesInfo('add')"
            >
              {{ $t('share.create') }}
            </el-button>
          </div>
          <el-table
            :key="tableKey"
            :data="tempCreateWorkOrder.sites"
            border
            fit
            highlight-current-row
            style="width: 100%"
          >
            <el-table-column
              :label="$t('system.name')"
              align="center"
              prop="name"
            >
              <template slot-scope="{row,$index}">
                <div
                  style="margin-left: -200px;width: 25em;margin-top: 1rem;"
                >
                  <el-form-item
                    :prop="'sites.'+ $index +'.name'"
                    :rules="siteNameRules.name"
                  >
                    <el-input
                      v-model="row.name"
                      size="medium"
                      maxlength="16"
                      show-word-limit
                    />
                  </el-form-item>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('share.ID')"
              align="center"
              prop="index"
            >
              <template slot-scope="{row}">
                <el-input-number
                  v-if="row.subType!==1"
                  id="subType"
                  v-model="row.index"
                  :min="0"
                  :max="32768"
                  size="mini"
                  controls-position="right"
                />
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('share.class')"
              align="center"
              prop="type"
            >
              <template slot-scope="{row}">
                <el-select
                  v-model="row.type"
                  :disabled="(row.type==Type.SLOT && row.subType==SubType.OPERATOR)||(row.type==Type.SLOT && row.subType==SubType.TOOL)"
                  filterable
                >
                  <el-option
                    v-for="item in typeList"
                    :key="item.ID"
                    :label="item.name"
                    :value="item.ID"
                  />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('share.subType')"
              align="center"
              prop="type"
            >
              <template slot-scope="{row,$index}">
                <el-select
                  v-model="row.subType"
                  filterable
                  @change="subTypeRules('create',$index)"
                >
                  <el-option
                    v-for="item in subTypeList"
                    :key="item.ID"
                    :label="$t('site.subType.' + item.name)"
                    :value="item.ID"
                  />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('share.action')"
              align="center"
              class-name="fixed-width"
            >
              <template slot-scope="{row,$index}">
                <el-button
                  size="mini"
                  type="danger"
                  icon="el-icon-delete-solid"
                  circle
                  @click="deleteSitesInfo(row,$index,'add')"
                />
              </template>
            </el-table-column>
          </el-table>
        </el-form>
        <div
          slot="footer"
          class="dialog-footer"
        >
          <el-button
            type="primary"
            @click="createStationInfoToDB()"
          >
            {{ $t('share.confirm') }}
          </el-button>
        </div>
      </el-dialog>
      <!-- update station -->
      <el-dialog
        :title="$t('share.update')"
        :visible.sync="dialogUpdateVisible"
      >
        <el-form
          ref="stationDataForm"
          :rules="stationRules"
          :model="tempUpdateStationData"
          status-icon
          label-position="left"
          label-width="200px"
          style="margin-left:50px;"
          autocomplete
        >
          <el-form-item
            :label="$t('system.stationID')"
            prop="ID"
          >
            <span>{{ tempUpdateStationData.ID }}</span>
          </el-form-item>
          <el-form-item
            :label="$t('system.machineCode')"
            prop="code"
          >
            <span>{{ tempUpdateStationData.code }}</span>
          </el-form-item>
          <el-form-item
            :label="$t('system.stationDescription')"
          >
            <el-input
              v-model="tempUpdateStationData.description"
              type="textarea"
              :rows="2"
            />
          </el-form-item>
          <el-form-item
            :label="$t('share.status')"
          >
            <el-select
              v-model="tempUpdateStationData.state"
              filterable
            >
              <el-option
                v-for="item in stationStatusList"
                :key="item.ID"
                :label="$t('station.status.' + item.name)"
                :value="item.ID"
              />
            </el-select>
          </el-form-item>
          <el-form-item
            :label="$t('system.siteData')"
          />
          <el-divider />
          <div class="filter-container">
            <el-button
              class="filter-item"
              type="primary"
              icon="el-icon-circle-plus-outline"
              size="medium"
              @click="addSitesInfo('update')"
            >
              {{ $t('share.create') }}
            </el-button>
          </div>
          <el-table
            :key="tableKey"
            :data="tempUpdateStationData.sites"
            border
            fit
            highlight-current-row
            style="width: 100%"
          >
            <el-table-column
              :label="$t('system.name')"
              align="center"
              prop="name"
            >
              <template slot-scope="{row,$index}">
                <div v-if="row.exist === undefined">
                  <div
                    style="margin-left: -200px;width: 25em;margin-top: 1rem;"
                  >
                    <el-form-item
                      :prop="'sites.'+ $index +'.name'"
                      :rules="siteNameRules.name"
                    >
                      <el-input
                        v-model="row.name"
                        size="medium"
                        maxlength="16"
                        show-word-limit
                      />
                    </el-form-item>
                  </div>
                </div>
                <div v-else>
                  <span>{{ row.name }}</span>
                </div>
              </template>
            </el-table-column>

            <el-table-column
              :label="$t('share.ID')"
              align="center"
            >
              <template slot-scope="{row}">
                <div v-if="row.exist === undefined">
                  <el-input-number
                    v-if="row.subType!==1"
                    id="subType"
                    v-model="row.index"
                    :min="0"
                    :max="32768"
                    size="mini"
                    controls-position="right"
                  />
                </div>
                <div v-else>
                  <span>{{ row.index }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('share.class')"
              align="center"
            >
              <template slot-scope="{row}">
                <div v-if="row.exist === undefined">
                  <el-select
                    v-model="row.type"
                    :disabled="row.type==2 && row.subType==1"
                    filterable
                  >
                    <el-option
                      v-for="item in typeList"
                      :key="item.ID"
                      :label="item.name"
                      :value="item.ID"
                    />
                  </el-select>
                </div>
                <div v-else>
                  <span>{{ row.typeName }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('share.subType')"
              align="center"
            >
              <template slot-scope="{row,$index}">
                <div v-if="row.exist === undefined">
                  <el-select
                    v-model="row.subType"
                    filterable
                    @change="subTypeRules('update',$index)"
                  >
                    <el-option
                      v-for="item in subTypeList"
                      :key="item.ID"
                      :label="$t('site.subType.' + item.name)"
                      :value="item.ID"
                    />
                  </el-select>
                </div>
                <div v-else>
                  <span>{{ $t('site.subType.' + row.subTypeName) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('share.action')"
              align="center"
              class-name="fixed-width"
            >
              <template slot-scope="{row,$index}">
                <el-button
                  size="mini"
                  type="danger"
                  icon="el-icon-delete-solid"
                  circle
                  @click="deleteSitesInfo(row,$index,'update')"
                />
              </template>
            </el-table-column>
          </el-table>
        </el-form>
        <div
          slot="footer"
          class="dialog-footer"
        >
          <el-button
            type="primary"
            @click="updateStationInfoToDB()"
          >
            {{ $t('share.confirm') }}
          </el-button>
        </div>
      </el-dialog>
      <!-- sites datial info -->
      <el-dialog
        :title="$t('share.details')"
        :visible.sync="dialogDetailVisible"
      >
        <el-table
          :key="tableKey"
          :data="tempDetailData"
          border
          highlight-current-row
          style="width: 100%"
        >
          <el-table-column
            :label="$t('system.name')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('share.ID')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.index }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('share.class')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ row.typeName }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('share.subType')"
            align="center"
          >
            <template slot-scope="{row}">
              <span>{{ $t('site.subType.' + row.subTypeName) }}</span>
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
                @click="openSiteDetailInfo(row)"
              />
            </template>
          </el-table-column>
        </el-table>
      </el-dialog>
      <!-- sites datial information info -->
      <el-dialog
        :title="$t('share.details')"
        :visible.sync="dialogSiteDetailVisible"
        append-to-body
      >
        <el-table
          :key="tableKey"
          :data="tempSiteDetailData"
          border
          highlight-current-row
          style="width: 100%"
        >
          <template v-if="targetSubType === 'OPERATOR'">
            <el-table-column
              :label="$t('system.employeeID')"
              align="center"
            >
              <template slot-scope="{row}">
                <span>{{ row.employeeID }}</span>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.group')"
              align="center"
            >
              <template slot-scope="{row}">
                <span>{{ row.group }}</span>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.workDate')"
              align="center"
            >
              <template slot-scope="{row}">
                <span>{{ row.workDate }}</span>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.expiryTime')"
              align="center"
            >
              <template slot-scope="{row}">
                <span>{{ row.expiryTime }}</span>
              </template>
            </el-table-column>
          </template>
          <template v-else-if="targetSubType === 'MATERIAL'">
            <el-table-column
              :label="$t('system.materialResourceID')"
              align="center"
            >
              <template slot-scope="{row}">
                <span>{{ row.resourceID }}</span>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.materialName')"
              align="center"
            >
              <template slot-scope="{row}">
                <span>{{ row.ID }}</span>
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
          </template>
          <template v-else-if="targetSubType === 'TOOL'">
            <el-table-column
              :label="$t('system.toolBarcode')"
              align="center"
            >
              <template slot-scope="{row}">
                <span>{{ row.resourceID }}</span>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('system.installedTime')"
              align="center"
            >
              <template slot-scope="{row}">
                <span>{{ row.installedTime }}</span>
              </template>
            </el-table-column>
          </template>
        </el-table>
      </el-dialog>
      <!-- sites(COLQUEUE) datial information info -->
      <el-dialog
        :title="$t('share.details')"
        :visible.sync="dialogSiteColqueueDetailVisible"
        append-to-body
      >
        <el-table
          :key="tableKey"
          :data="tempSiteDetailData"
          border
          highlight-current-row
          style="width: 100%"
        >
          <el-table-column type="expand">
            <template slot-scope="{row}">
              <el-table
                :key="tableKey"
                :data="row"
                border
                highlight-current-row
                style="width: 100%"
              >
                <template v-if="targetSubType === 'OPERATOR'">
                  <el-table-column
                    :label="$t('system.employeeID')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.employeeID }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column
                    :label="$t('system.group')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.group }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column
                    :label="$t('system.workDate')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.workDate }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column
                    :label="$t('system.expiryTime')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.expiryTime }}</span>
                    </template>
                  </el-table-column>
                </template>
                <template v-else-if="targetSubType === 'MATERIAL'">
                  <el-table-column
                    :label="$t('system.materialResourceID')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.resourceID }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column
                    :label="$t('system.materialName')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.ID }}</span>
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
                </template>
                <template v-else-if="targetSubType === 'TOOL'">
                  <el-table-column
                    :label="$t('system.toolBarcode')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.resourceID }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column
                    :label="$t('system.installedTime')"
                    align="center"
                  >
                    <template slot-scope="{row}">
                      <span>{{ row.installedTime }}</span>
                    </template>
                  </el-table-column>
                </template>
              </el-table>
            </template>
          </el-table-column>
          <el-table-column
            :label="$t('share.ID')"
            align="center"
          >
            <template slot-scope="{$index}">
              <span>{{ $index + 1 }}</span>
            </template>
          </el-table-column>
        </el-table>
      </el-dialog>
    </el-main>
    <el-footer />
  </div>
</template>
<script src="./stationMaintenance.ts" lang="ts"></script>
<style src="./stationMaintenance.css"></style>
