<template>
  <div class="app-container">
    <el-header class="header-style">
      {{ $t('router.editPassword') }}
    </el-header>
    <el-main>
      <el-form
        ref="dataForm"
        :rules="rule"
        :model="editPasswordForm"
        label-position="left"
        label-width="6rem"
      >
        <el-form-item
          :label="$t('login.oldPassword')"
          prop="oldPassword"
        >
          <el-input
            :key="passwordType"
            v-model="editPasswordForm.oldPassword"
            :type="passwordType"
          />
          <span
            class="show-pwd"
            @click="showPwd"
          >
            <svg-icon :name="passwordType === 'password' ? 'eye-off' : 'eye-on'" />
          </span>
        </el-form-item>
        <el-form-item
          prop="newPassword"
          :label="$t('login.newPassword')"
        >
          <el-input
            v-model="editPasswordForm.newPassword"
            :type="passwordType"
          />
        </el-form-item>
        <el-form-item
          prop="checkNewPassword"
          :label="$t('login.checkNewPassword')"
        >
          <el-input
            v-model="editPasswordForm.checkNewPassword"
            :type="passwordType"
          />
        </el-form-item>
      </el-form>
      <el-button
        type="primary"
        @click="onClickConfirm"
      >
        {{ $t('share.confirm') }}
      </el-button>
    </el-main>
  </div>
</template>

<script src="./editPassword.ts" lang="ts"></script>
<style src="./editPassword.css"></style>
