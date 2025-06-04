<template>
  <table class="demo-typo-size" style="width: 95%;" align="center">
    <tbody>
      <tr>
        <td style="font-size: 24px;height: 40px;"><img src="../assets/images/logo.png" width="20" height="20"> 悠乐快传</td>
      </tr>
      <tr>
        <td><el-collapse v-model="activeName" accordion>
            <el-collapse-item v-for="(item, index, key) in hostList.host" :key="key" :name="key"
              style="font-size: 5rem;">
              <template #title>
                <el-icon style="color: #303133; padding-right: 5px">
                  <Platform />
                </el-icon>
                <div style="height: 40px;">{{ item.nick }}</div>
              </template>
              <table style="width: 95%;" align="center">
                <tbody>
                  <tr>
                    <td>{{ index }}</td>
                  </tr>
                  <tr>
                    <td><el-upload class="upload-demo" :http-request="uploadFile" :multiple="true"
                        :on-progress="handleProgress" :on-success="handleSuccess" :on-error="handleError">
                        <el-button type="primary">发送文件</el-button>
                      </el-upload></td>
                  </tr>
                  <tr v-for="(share, i, k) in item.share" :key="k">
                    <td width="65%" style="height: 3rem;">{{ share.title }}</td>
                    <td width="25%">{{ bytesToSize(share.size) }}</td>
                    <td width="10%"><el-button type="primary" size="small"
                        @click="downLoad(index, share.shareId)">下载</el-button></td>
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
const pageVisable = ref(false)

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

const progress = ref(0);

// 自定义上传逻辑
const uploadFile = async ({ action, file, onProgress, onSuccess, onError }) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('lastModified', file.lastModified);
  try {
    const response = await axios.post(action + '/upload', formData, {
      onUploadProgress: (event) => {
        const percentCompleted = Math.round((event.loaded * 100) / event.total);
        progress.value = percentCompleted; // 更新进度条
        onProgress({ percent: percentCompleted }); // 调用 Element Plus 的进度回调
      },
    });
    progress.value = 100; // 确保进度条满格
    onSuccess(response.data); // 调用 Element Plus 的成功回调
  } catch (error) {
    progress.value = 0; // 重置进度条
    onError(error); // 调用 Element Plus 的错误回调
  }
};
// 处理上传成功
const handleSuccess = () => {
  progress.value = 0; // 重置进度条
};

// 处理上传失败
const handleError = (error) => {
  progress.value = 0; // 重置进度条
  if (error.response.status == "401") {
    ElMessageBox({
      type: 'error',
      title: '二维码已过期',
      message: '二维码已过期，请重新扫码。',
    });
    return;
  } else {
    ElMessageBox({
      type: 'error',
      title: '网络错误',
      message: '网络错误。',
    });
    return;
  }
};

//const server = "http://192.168.203.99:35286"
const server = "http://" + window.location.hostname + ":35286"
const activeName = ref(0)
const hostList = ref([])

function downLoad(ip, shareId) {
  const url = "http://" + ip + ":35286" + "/share/requestFile";//"http://service-images.bj.bcebos.com/1.xls";
  const sendData = {
    shareId: shareId
  };

  axios({
    method: 'post',
    data: sendData,
    url,
    responseType: 'blob',
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
  axios.post(server + '/host')
    .then(response => {
      if (response.data.code == 0) {
        hostList.value = response.data.data
      } else {
        ElMessageBox({
          type: 'error',
          title: '获取共享主机错误',
          message: response.data.msg,
        })
      }
    })
    .catch(error => {
      if (error.response.status == "401") {
        ElMessageBox({
          type: 'error',
          title: '二维码已过期',
          message: '二维码已过期，请重新扫码。',
        });
        return;
      } else {
        ElMessageBox({
          type: 'error',
          title: '网络错误',
          message: '网络错误。',
        });
        return;
      }
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
</style>
