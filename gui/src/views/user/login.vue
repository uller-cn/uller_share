<template>
<div class="login-container">
    <div class="login-form">
      <div class="container">
          <div class="left"><img src="../../assets/images/logo.png" width="40" height="40" />
            <span style="font-size: 24px;font-weight: bold;padding-left: 10px;color: black;">悠乐快传</span>
          </div>
        </div>
        <div class="form-group">
          <h2 class="login-title">请扫描下面二维码，在公众号对话中输入“验证码”，来获取登录验证码。</h2>
          <label for="username"><img src="../../assets/images/qrcode.jpg"></label>
        </div>
        <div class="form-group">
          <label>验证码</label>
          <el-popover
            placement="top"
            title="请输入验证码"
            :visible="popoverVisible"
            :width="200"
          >
          <template #reference>
            <el-input type="password" v-model="password" placeholder="请输入验证码"  @keyup.enter="login()" />
          </template>
        </el-popover>
        </div>
        <div>
          <el-button type="primary" @click="login()">
          登录
        </el-button>
        </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue"
import { CodeLogin,GetLocalIp } from "../../../wailsjs/go/main/App";
import { ElMessageBox } from 'element-plus';
import router from '../../router'

const password = ref("");
const popoverVisible = ref(false);
const loginBtnVisible = ref(true);

function login() {
  loginBtnVisible.value = false;
  if(password.value == ""){
    popoverVisible.value = true;
  }else{
    Promise.all([CodeLogin(password.value), GetLocalIp()])
    .then((res) => {
      const token = res[0];
      const localIp = res[1];
      sessionStorage.setItem('token', res);
      router.push({ path: '/',query: { ip: localIp, isSelf: 1 } });
    })
    .catch((err) => {
      loginBtnVisible.value = true;
      ElMessageBox({
        type: 'error',
        title: '登录错误',
        message: err,
      });
    });
  }
  loginBtnVisible.value = true;
}
</script>

<style type="text/css">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f4f4f4;
}

.login-form {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  width: 400px;
}

.login-title {
  text-align: center;
  margin-bottom: 20px;
  color: #333;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  color: #666;
}

.container {
  justify-content: center; 
  margin: 0px;
  height: 60px;
  line-height: 60px;
  display: flex;
  width: 100%;
  align-items: center;
  background-color: #FFFFFF;
}
</style>
