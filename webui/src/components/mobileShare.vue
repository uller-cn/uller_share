<!-- eslint-disable no-empty -->
<template>
  <!-- <div v-if="browserTip" class="browserTip">
    <span style=" padding-left: 20px;">由于{{ browserName }}禁止下载功能，如果您要下载，请点击右上角[…]选择在浏览器打开。</span><img
      src="../assets/images/wechatGoSafari.png" style="width: 80px;height: 80px;float: right;padding-right: 20px;"
      alt="">
  </div> -->
  <table style="width: 100%;background: rgba(0, 0, 0, 0.5);border-radius: 5px;" align="center" v-if="browserTip">
    <tbody>
      <tr>
        <td align="left"><span style="padding-left: 10px;">{{ browserName }}禁止下载功能，如果您要下载，请点击右上角[…]选择在浏览器打开。</span></td>
        <td align="right"><img src="../assets/images/wechatGoSafari.png" style="width: 80px;height: 80px;float: right;"
            alt=""></td>
      </tr>
    </tbody>
  </table>
  <table style="width: 100%;" align="center">
    <tbody>
      <tr>
        <td style="font-size: 24px;height: 40px;"><img src="../assets/images/logo.png" width="20" height="20"> 悠乐快传</td>
      </tr>
      <tr>
        <td><el-collapse v-model="activeName" accordion>
            <el-collapse-item v-for="(item, index, key) in hostList.host" :key="key" :name="key">
              <template #title>
                <div style="display: flex;height: 40px;
  align-items: center;"><el-icon style="color: #303133; padding-right: 5px;padding-left: 5px;" size="20">
                    <Platform />
                  </el-icon>
                  <div>{{ item.nick }}</div>
                </div>
              </template>
              <table style="width: 95%;" align="center">
                <tbody>
                  <tr>
                    <td>{{ index }}</td>
                  </tr>
                  <tr>
                    <td><el-upload v-model:file-list="item.fileList" class="upload-demo" :action="item.action"
                        :headers="uploadHeaders" :data="uploadData" multiple :before-upload="beforeUpload"
                        :on-success="uploadSuccess" :on-error="uploadErr">
                        <el-button type="primary">发送文件</el-button>
                      </el-upload></td>
                  </tr>
                  <tr v-for="(share, i, k) in item.share" :key="k">
                    <td style="width: 50%;word-break: break-all;">{{ share.title }}</td>
                    <td style="width: 40%;word-break: break-all;">{{ bytesToSize(share.size) }}</td>
                    <td style="width: 10%;word-break: break-all;"><el-button type="primary" size="small"
                        @click="downLoad(index, share)">下载</el-button></td>
                  </tr>
                </tbody>
              </table>
            </el-collapse-item>
          </el-collapse></td>
      </tr>
    </tbody>
  </table>
</template>

<script lang="js" setup>
import { ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import axios from 'axios'
import { Platform } from "@element-plus/icons-vue"

const param = window.location.href.split('?')[1]
const sign = ref('')
const time = ref('')
const pageVisable = ref(false);
// const fileList = ref([]);
const uploadHeaders = ref({});
const uploadData = ref({});
var ua = navigator.userAgent.toLowerCase();
const browserTip = ref(false);
const browserName = ref("微信");
if (/MicroMessenger/i.test(ua)) {
  browserName.value = "微信"
  browserTip.value = true;
}
if (ua.includes('dingtalk')) {
  browserName.value = "钉钉"
  browserTip.value = true;
}
//browserTip.value = true;

if (param !== undefined) {
  param.split('&').forEach(param => {
    const [key, value] = param.split('=');
    if (decodeURIComponent(key) == "sign") {
      sign.value = decodeURIComponent(value);
    }
    if (decodeURIComponent(key) == "time") {
      time.value = decodeURIComponent(value);
    }
  });
  if (sign.value == "" || time.value == "") {
    ElMessageBox({
      type: 'error',
      title: '错误',
      message: '请使用pc端二维码扫描进入。',
    });
  } else {
    pageVisable.value = true;
    axios.defaults.headers.common['sign'] = sign.value;
    axios.defaults.headers.common['uller-client-time'] = time.value;
    axios.defaults.headers.common['uller-client'] = 'h5';
  }
} else {
  ElMessageBox({
    type: 'error',
    title: '错误',
    message: '请使用pc端二维码扫描进入。',
  });
}

//const server = "http://192.168.203.99:35286"
const server = "http://" + window.location.hostname + ":35286"
const activeName = ref(0)
const hostList = ref([])

const beforeUpload = (file) => {
  uploadData.value = { 'lastModified': file.lastModified };
  uploadHeaders.value = { 'sign': sign.value, 'uller-client-time': time.value, 'uller-client': 'h5' };
};

const uploadSuccess = (response) => {
  if (response.code != '0') {
    ElMessageBox({
      type: 'error',
      title: '发送文件错误',
      message: response.msg,
    });
  }
};

const uploadErr = () => {
  ElMessageBox({
    type: 'error',
    title: '发送文件错误',
    message: "二维码已过期，请重新扫码",
  });
};

function downLoad(ip, share) {
  const url = "http://" + ip + ":35286" + "/share/requestFile";//"http://service-images.bj.bcebos.com/1.xls";
  const sendData = {
    shareId: share.shareId
  };

  axios({
    method: 'post',
    data: sendData,
    url,
    responseType: 'blob',
    timeout: 30 * 1000
  })
    .then((response) => {
      var contentDisposition = response.headers['content-disposition'];
      var fileName = "default-file-name";
      if (contentDisposition != undefined && contentDisposition != "") {
        var contentDispositionArr = contentDisposition.split('filename=');
        if (contentDispositionArr.length >= 2) {
          fileName = contentDispositionArr[1];
        }
      }

      const blob = new Blob([response.data], { type: 'application/octet-stream' });
      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.setAttribute('download', fileName);
      link.style.display = 'none';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);
    })
    .catch((error) => {
      if (error.response.status == "404") {
        ElMessageBox({
          type: 'error',
          title: '下载错误',
          message: '共享文件已删除，无法下载。',
        });
        return;
      } else if (error.response.status == "410") {
        ElMessageBox({
          type: 'error',
          title: '下载错误',
          message: '共享文件已过期，无法下载。',
        });
        return;
      } else if (error.response.status == "400") {
        ElMessageBox({
          type: 'error',
          title: '下载错误',
          message: '参数错误。',
        });
        return;
      } else if (error.response.status == "401") {
        ElMessageBox({
          type: 'error',
          title: '二维码已过期',
          message: '二维码已过期，请重新扫码。',
        });
        return;
      }
    });
}

function bytesToSize(bytes) {
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  if (bytes === 0) return '0 B';
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return (bytes / Math.pow(1024, i)).toFixed(2) + ' ' + sizes[i]; // toFixed(2)表示保留两位小数
}

if (pageVisable.value == true) {
  axios({
    url: server + '/host',
    method: 'post',
    timeout: 30 * 1000
  })
    .then(response => {
      if (response.data.code == 0) {
        hostList.value = response.data.data;
        for (const index in hostList.value.host) {
          hostList.value.host[index].action = `http://${index}:35286/upload`;
          hostList.value.host[index].fileList = [];
          //console.log(hostList.value.host[index].action);
        }
      } else {
        ElMessageBox({
          type: 'error',
          title: '获取共享主机错误',
          message: response.data.msg,
        })
      }
    })
    .catch(() => {
      ElMessageBox({
        type: 'error',
        title: '二维码已过期',
        message: '二维码已过期，请重新扫码。',
      });
      return;
    });
}
</script>

<style lang="css">
html {
  font-size: 16px;
}

body {
  font-size: 1rem;
}

td {
  height: 1.5rem;
}

.el-collapse-item__header {
  font-size: 1rem !important;
}

.el-collapse-item__wrap {
  font-size: 0.9rem !important;
}

.browserTip {
  /* position: fixed; */
  display: flex;
  top: 0;
  left: 0;
  width: 100%;
  height: 90px;
  background: rgba(0, 0, 0, 0.8);
  color: #fff;
  align-items: center;
  justify-content: center;
  z-index: 999;
}
</style>
