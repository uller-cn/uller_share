<template>
  <div class="menu-container">
    <div class="menu-item-title" style="padding-top: 10px;">
      <el-icon style="color: #303133; padding-right: 5px" size="20">
        <HomeFilled />
      </el-icon>
      <el-text class="mx-1">我的电脑</el-text>
    </div>
    <div class="menu-item" :class="myShareLinkActive?'menu-active':''">
      <el-icon style="color: #303133; padding-right: 5px">
        <Platform />
      </el-icon>
      <RouterLink :to="{ path: '/share', query: { ip: ip, isSelf: 1 } }" class="a"  @click.native="routerLinkClick(ip)">
        <el-badge :value="shareCount" class="mark" type="primary">
          <el-tooltip popper-class="tooltip" class="box-item" effect="light" placement="right-start" @show="toolTipOver">
            <template #content>
              <div style="width: 300px;font-size: 16px; padding-left: 20px;padding-right: 20px;">
                <div style="height: 40px;font-weight: bold;text-align: left;display: flex;">扫码互传</div>
                <div style="height: 40px;line-height: 40px; gap: 4px;text-align: left;">电脑名称：<span v-if="hostNameTxt"
                    @click="clickHostName()">{{ nick }}</span> <el-icon v-if="hostNameEditIcon"
                    @click="clickHostName()">
                    <Edit />
                  </el-icon>
                  <el-input v-if="hostNameInput" v-model="nick" ref="hostNameInputRef" placeholder="{{ nick }}"
                    style="width: 200px;" @keyup.enter="editHostName()" @blur="hostNameInputBlur" />
                </div>
                <div style="height: 40px;line-height: 40px; text-align: left;">IP:{{ ip }}</div>
                <div style="height: 280px;display: flex;align-items: left;justify-content: center;">
                  <QRCodeVue3 :width="240" :height="240" :value="localQRStr" :dotsOptions="{
                    type: 'extra-rounded',
                    color: '#26249a',
                  }" />
                </div>
              </div>
            </template>
            <el-text class="mx-1">我的共享</el-text>
          </el-tooltip>
        </el-badge>
      </RouterLink>
    </div>
    <div class="menu-item">
      <el-icon style="color: #303133;padding-right: 5px">
        <Download />
      </el-icon>
      <RouterLink :to="{ path: '/history' }" class="a">
        <el-text class="mx-1">历史记录</el-text>
      </RouterLink>
    </div>
    <div class="menu-item-title">
      <el-icon style="color: #303133; padding-right: 5px" size="20">
        <Menu />
      </el-icon>
      <el-text class="mx-1">共享主机</el-text>
      <el-tooltip popper-class="tooltip" effect="light" content="新增网络主机"
        placement="right-start">
        <el-icon size="15" enterable="true" style="margin-left: 10px;" @click="syncHostSharePopoverClick($event)"><Plus style="color: #303133" /></el-icon>
      </el-tooltip>
      <el-popover v-model:visible="syncHostShareVisible" trigger="manual" virtual-triggering :virtual-ref="tableRowClick"
        placement="bottom" width="340">
        <el-input v-model="syncIp" style="width: 180px" placeholder="IPV4地址" @keyup.enter="syncHostShare()" />
        <el-button type="primary" style="margin-right: 10px;margin-left: 10px;" :loading="syncHostShareBtnLoading" @click="syncHostShare()">查找</el-button>
      </el-popover>
      <el-tooltip popper-class="tooltip" effect="light" content="刷新网络主机"
        placement="right-start">
        <el-icon size="15" enterable="true" :class="[syncHostLoad]" style="margin-left: 10px;" :disabled="syncHostDisabled" @click="syncHost"><Refresh style="color: #303133" /></el-icon>
      </el-tooltip>
    </div>
    <div style="text-align: left; margin-left: 15px;margin-top: 10px;margin-bottom: 10px;">
      <el-input v-model="searchValue" style="width: 80%" size="small" placeholder="主机搜索" :prefix-icon="Search"
        @input="debouncedHandleInput" />
    </div>
    <div v-for="[k, value] in shareHost" :key="k" class="menu-item" :class="value.acvtive?'menu-active':''">
      <el-icon style="color: #303133; padding-right: 5px">
        <Platform />
      </el-icon>
      <RouterLink :to="{ path: '/share', query: { ip: k, isSelf: 0 } }" class="a" @click.native="routerLinkClick(k)">
        <el-badge :value="value.shareCount" class="mark" type="primary">
          <el-tooltip popper-class="tooltip" effect="light" placement="right-start">
            <template #content>
              <div style="width: 240px;font-size: 16px; padding-left: 20px;padding-right: 20px;margin-top: 15px;">
                <div style="height: 40px;font-weight: bold;text-align: left;display: flex;">主机信息</div>
                <div style="height: 40px;text-align: left;">电脑名称：{{ value.nick }}</div>
                <div style="height: 40px;text-align: left;">IP:{{ k }}</div>
              </div>
            </template>
          <el-text class="mx-1">{{ value.nick }}</el-text>
        </el-tooltip>
        </el-badge>
      </RouterLink>
    </div>
  </div>
</template>

<script lang="js" setup>
import { ref, onMounted, onUnmounted, computed, nextTick } from "vue"
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { Platform, Menu, Refresh, Search, Download, HomeFilled, Edit,Plus } from "@element-plus/icons-vue"
import { ElNotification } from 'element-plus'
import QRCodeVue3 from "qrcode-vue3";
import { GetHttpSign, EditHostName, GetHostList,SyncHost,GetLocalInfo,SyncHostShare } from "../../wailsjs/go/main/App";
import { ElMessageBox } from 'element-plus';

const searchValue = ref('');
const shareHost = ref(new Map());
const hostNameInput = ref(false);
const hostNameTxt = ref(true)
const hostNameInputRef = ref(null);
const hostNameEditIcon = ref(true);
const syncHostLoad = ref("");
const syncHostDisabled = ref(false);
const httpSigin = ref('');
const httpTime = ref(0);
const shareCount = ref(0);
const ip = ref("");
const nick = ref("");
const syncHostShareVisible = ref(false);
const tableRowClick = ref(null);
const syncIp=ref("");
const syncHostShareBtnLoading=ref(false);
const myShareLinkActive=ref(true);

//const localQRStr = computed(() => "http://" + ip.value + ":35287/#/mobileShare?sign="+httpSigin.value+"&time="+httpTime.value);
const localQRStr = computed(() => {
  //console.log(`http://${ip.value}:35287/#/mobileShare?sign=${httpSigin.value}&time=${httpTime.value}`);
  return `http://${ip.value}:35287/#/mobileShare?sign=${httpSigin.value}&time=${httpTime.value}`;
});

onMounted(() => {
  GetLocalInfo()
  .then((res) => {
    ip.value = res["ip"];
    nick.value = res["nick"];
    shareCount.value = res["shareCount"];
  })
  .catch((err) => {
    ElMessageBox({
      type: 'error',
      title: '获取主机信息错误',
      message: err,
    })
  });
  EventsOn('HostJoinGroup', (data) => {
    //console.log("HostJoinGroup",data);
    if(!shareHost.value.has(data.ip)){
      data.acvtive=false;
      hostOnlineNify('主机上线通知', "主机：" + data.nick + "，ip地址：" + data.ip + "，已上线");
      shareHost.value.set(data.ip, data);
    }
  });
  EventsOn('HostLeaveGroup', (data) => {
    if(shareHost.value.has(data)){
      hostOnlineNify('主机下线通知', "主机：" + shareHost.value.get(data).nick + "，ip地址：" + data + "，已下线");
      shareHost.value.delete(data);
    }
  });
  EventsOn('FindHost', (data) => {
    shareHost.value.set(data.ip, data);
  });
  EventsOn('NewShareCount', (data) => {
    shareHost.value.set(data.ip, data);
  });
  EventsOn('LocalShareCount', (count) => {
    shareCount.value = count;
  });
  EventsOn('DelShare', (delShare) => {
    //console.log("DelShare",shareHost.value.get(delShare.ip),shareHost.value.get(delShare.ip).shareCount);
    shareHost.value.get(delShare.ip).shareCount=shareHost.value.get(delShare.ip).shareCount-1;
    //shareHost.value.set(delShare.ip, shareHost.value.get(delShare.ip).shareCount-1);
  });
  EventsOn('EditHostName', (data) => {
    shareHost.value.get(data.ip).nick=data.nick;
  });
});

onUnmounted(() => {
  EventsOff('HostJoinGroup', () => {

  });
  EventsOff('HostLeaveGroup', () => {

  });
  EventsOff('FindHost', () => {

  });
  EventsOff('NewShareCount', () => {

  });
  EventsOff('LocalShareCount', () => {

  });
  EventsOff('DelShare', () => {

  });
  EventsOff('EditHostName', () => {

  });
});

const routerLinkClick=(ipKey)=>{
  shareHost.value.forEach((value, key) => {
    shareHost.value.get(key).acvtive=false;
  });
  if(ipKey==ip.value){
    myShareLinkActive.value=true;
  }else{
    shareHost.value.get(ipKey).acvtive=true;
    myShareLinkActive.value=false;
  }
}

const hostOnlineNify = (title, msg) => {
  ElNotification({
    title: title,
    message: msg,
    position: 'bottom-left',
  })
}

function syncHostShare() {
  syncHostShareBtnLoading.value=true;
  if(isIPv4(syncIp.value) == false && isIPv6(syncIp.value) == false){
    ElMessageBox({
      type: 'error',
      title: 'ip地址格式错误',
      message: "您输入的ip地址不正确，请重新输入",
    });
    return
  }
  SyncHostShare(syncIp.value)
    .then((res) => {
      syncHostShareBtnLoading.value=false;
    })
    .catch((err) => {
      syncHostShareBtnLoading.value=false;
      ElMessageBox({
        type: 'error',
        title: '主机不存在',
        message: "未找到共享主机："+syncIp.value,
      })
    });
}

const syncHostSharePopoverClick = (event) => {
  tableRowClick.value = event.target;
  syncHostShareVisible.value = true;
};

const isIPv4 = (ipStr)=> {
  const ipv4Regex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
  return ipv4Regex.test(ipStr);
};

const isIPv6 = (ipStr)=> {
  const ipv6Regex = /^(?:(?:[0-9a-fA-F]{1,4}:){7}(?:[0-9a-fA-F]{1,4}|:))|(?:::(?:[0-9a-fA-F]{1,4}:){0,5}(?:[0-9a-fA-F]{1,4}|:))|(?:(?:[0-9a-fA-F]{1,4}:){1,6}:)|(?:(?:[0-9a-fA-F]{1,4}:){1,5}:(?:[0-9a-fA-F]{1,4}:)[0-9a-fA-F]{1,4})|(?:(?:[0-9a-fA-F]{1,4}:){1,4}:(?:[0-9a-fA-F]{1,4}:){1,2}[0-9a-fA-F]{1,4})|(?:(?:[0-9a-fA-F]{1,4}:){1,3}:(?:[0-9a-fA-F]{1,4}:){1,3}[0-9a-fA-F]{1,4})|(?:(?:[0-9a-fA-F]{1,4}:){1,2}:(?:[0-9a-fA-F]{1,4}:){1,4}[0-9a-fA-F]{1,4})|(?::(?:[0-9a-fA-F]{1,4}:){1,5}[0-9a-fA-F]{1,4})|(?:(?:[0-9a-fA-F]{1,4}:){1,7}:)|(?:(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4})|(?:(?:[0-9a-fA-F]{1,4}:){1,5}:(?:[0-9a-fA-F]{1,4}:)[0-9a-fA-F]{1,4})|(?:(?:[0-9a-fA-F]{1,4}:){1,4}:(?:[0-9a-fA-F]{1,4}:){1,2}[0-9a-fA-F]{1,4})|(?:(?:[0-9a-fA-F]{1,4}:){1,3}:(?:[0-9a-fA-F]{1,4}:){1,3}[0-9a-fA-F]{1,4})|(?:(?:[0-9a-fA-F]{1,4}:){1,2}:(?:[0-9a-fA-F]{1,4}:){1,4}[0-9a-fA-F]{1,4})|(?::(?:[0-9a-fA-F]{1,4}:){1,5}[0-9a-fA-F]{1,4})|(?:::(?:[0-9a-fA-F]{1,4}:){0,5}(?:[0-9a-fA-F]{1,4}|:))$/i;
  return ipv6Regex.test(ipStr);
}

const toolTipOver = () => {
  GetHttpSign()
    .then((res) => {
      httpSigin.value = res["sign"];
      httpTime.value = res["uller-client-time"];
      //console.log("toolTipOver",localQRStr.value);
    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: '获取二维码错误',
        message: err,
      })
    });
}

toolTipOver();

function syncHost(){
  syncHostDisabled.value=true;
    syncHostLoad.value='is-loading';
    SyncHost(ip.value)
    .then((res) => {
      syncHostDisabled.value=false;
      syncHostLoad.value='';
    })
    .catch((err) => {
      syncHostDisabled.value=false;
      ElMessageBox({
        type: 'error',
        title: '同步远程主机数据错误',
        message: err,
      })
    });
}

function editHostName() {
  EditHostName(nick.value)
    .then(() => {

    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: '获取二维码错误',
        message: err,
      })
    });
}

function hostNameInputBlur() {
  hostNameEditIcon.value = true;
  hostNameInput.value = false;
  hostNameTxt.value = true;
  EditHostName(nick.value)
    .then(() => {

    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: '获取二维码错误',
        message: err,
      })
    });
}

function clickHostName() {
  if (hostNameInput.value == false) {
    hostNameEditIcon.value = false;
    hostNameTxt.value = false;
    hostNameInput.value = true;
    nextTick(() => {
      hostNameInputRef.value.focus();
    });
  }
}

// 防抖函数
const debounce = (fn, delay = 1000) => {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      fn.apply(this, args);
    }, delay);
  };
};

const debouncedHandleInput = debounce(searchChange, 1000);

function searchChange() {
  GetHostList(searchValue.value)
    .then((data) => {
      if (data == null) {
        shareHost.value.clear();
      } else {
        for (var i = 0; i < data.length; i++) {
          //console.log(data[i].ip);
          shareHost.value.set(data[i].ip, data[i]);
        }
        //shareHost.value.set(data.ip, data);
      }
    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: '获取二维码错误',
        message: err,
      })
    });
}
</script>

<style lang="css">
.tooltip {
  color: black;
}

.menu-item-title {
  border-radius: 5px;
  margin-left: 0.5rem;
  margin-right: 0.5rem;
  padding-left: 5px;
  padding-right: 5px;
  height: 40px;
  text-align: center;
  cursor: pointer;
  display: flex;
  position: relative;
  white-space: nowrap;
  list-style: none;
  align-items: center;
}

.menu-item {
  border-radius: 5px;
  margin-left: 0.5rem;
  margin-right: 0.5rem;
  padding-left: 15px;
  padding-right: 15px;
  height: 40px;
  text-align: center;
  cursor: pointer;
  display: flex;
  position: relative;
  white-space: nowrap;
  list-style: none;
  align-items: center;
  justify-content: start;
}

.menu-item:hover {
  background-color: #ecf5ff;
}

.menu-active {
  background-color: #ecf5ff;
}

.menu-container {
  min-height: calc(100vh - 90px);
  width: 100%;
  background-color: #ffffff;
}

.demo-progress .el-progress--line {
  max-width: 600px;
}
.is-loading{

}
</style>
