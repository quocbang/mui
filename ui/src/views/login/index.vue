<template>
  <div class="login-container">
    <el-form
      ref="loginForm"
      :model="loginForm"
      class="login-form"
      autocomplete="on"
      label-position="left"
    >
      <div class="title-container">
        <h3 class="title">
          {{ $t('login.title') }}
        </h3>
        <lang-select class="set-language" />
      </div>
      <el-radio-group
        v-model="loginForm.loginType"
        style="width:100%; margin-bottom:30px;"
        size="medium"
        @change="focusCursor"
      >
        <el-radio-button
          :label="0"
        >
          MES
        </el-radio-button>
        <el-radio-button
          :label="1"
        >
          Windows
        </el-radio-button>
        <el-radio-button
          :label="2"
        >
          PDA
        </el-radio-button>
      </el-radio-group>

      <div
        v-if="loginForm.loginType!==2"
      >
        <el-form-item prop="ID">
          <span class="svg-container">
            <svg-icon name="user" />
          </span>
          <el-input
            ref="ID"
            v-model="loginForm.ID"
            name="ID"
            type="text"
            autocomplete="on"
          />
        </el-form-item>
        <el-form-item prop="password">
          <span class="svg-container">
            <svg-icon name="password" />
          </span>
          <el-input
            :key="passwordType"
            ref="password"
            v-model="loginForm.password"
            :type="passwordType"
            name="password"
            autocomplete="on"
            @keyup.enter.native="handleLogin"
          />
          <span
            class="show-pwd"
            @click="showPwd"
          >
            <svg-icon :name="passwordType === 'password' ? 'eye-off' : 'eye-on'" />
          </span>
        </el-form-item>
      </div>
      <div v-else>
        <el-form-item prop="ID">
          <span class="svg-container">
            <svg-icon name="user" />
          </span>
          <el-input
            ref="ID"
            v-model="loginForm.ID"
            name="ID"
            type="text"
            autocomplete="on"
          />
        </el-form-item>
        <el-form-item prop="password">
          <span class="svg-container">
            <svg-icon name="password" />
          </span>
          <el-input
            :key="passwordType"
            ref="password"
            v-model="loginForm.password"
            :type="passwordType"
            name="password"
            autocomplete="on"
            @keyup.enter.native="handleLogin"
          />
          <span
            class="show-pwd"
            @click="showPwd"
          >
            <svg-icon :name="passwordType === 'password' ? 'eye-off' : 'eye-on'" />
          </span>
        </el-form-item>
        <el-radio-group
          v-model="loginForm.group"
          prop="group"
          style="width:100%; margin-bottom:30px;"
          size="medium"
        >
          <el-radio-button :label="1">
            早
          </el-radio-button>
          <el-radio-button :label="2">
            中
          </el-radio-button>
          <el-radio-button :label="3">
            晚
          </el-radio-button>
        </el-radio-group>
        <el-form-item prop="workDate">
          <span class="svg-container">
            <svg-icon name="table" /></span>
          <el-date-picker
            v-model="loginForm.workDate"
            :placeholder="$t('system.workDate')"
            type="date"
            value-format="yyyy-MM-dd"
            prefix-icon="false"
            clear-icon="true"
          />
        </el-form-item>
      </div>
      <el-button
        :loading="loading"
        type="primary"
        style="width:100%; margin-bottom:30px;"
        @click.native.prevent="handleLogin"
      >
        {{ $t('login.logIn') }}
      </el-button>
    </el-form>
  </div>
</template>

<script src="./index.ts" lang="ts"></script>
<style src="./index.scss" lang="scss"></style>
