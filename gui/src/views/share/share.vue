<template>
  <div style="
      width: 100%;
      height: 60px;
      line-height: 60px;
      text-align: left;
      background-color: #ffffff;
    ">
    <el-input v-model="searchTitle" placeholder="文件搜索" style="width: 200px; margin-right: 10px; margin-left: 10px"
      @keyup.enter="getShareList()" />
    <el-select v-model="selectExt" clearable placeholder="类型" multiple style="width: 140px; margin-right: 10px">
      <el-option v-for="item in ext" :key="item.value" :label="item.value" :value="item.value">
        <div class="flex" style="text-align: left">
          <el-tag type="primary">{{ item.value }}</el-tag>
        </div>
      </el-option>
      <template #tag>
        <el-tag type="primary" v-for="item in selectExt" :key="item">{{
          item
        }}</el-tag>
      </template>
    </el-select>
    <el-button type="primary" style="margin-right: 10px" @click="getShareList()">搜索</el-button>
  </div>
  <div style="
      width: 100%;
      height: 60px;
      line-height: 60px;
      text-align: left;
      margin-top: 15px;
      margin-bottom: 10px;
      background-color: #ffffff;
      align-items: center;
      display: flex;
    ">
    <el-button type="primary" style="margin-left: 10px" @click="syncHostShare()">
      <el-icon>
        <Refresh />
      </el-icon>
    </el-button>
    <el-button type="primary" @click="addShare" v-if="isSelf == 1">
      分享<el-icon class="el-icon-right shake" placement="right-start">
      </el-icon>
    </el-button>
    <el-popover v-if="isSelf==0" placement="bottom" :width="550" trigger="click" :visible="uploadPopoverVisible">
      <template #reference>
        <el-upload v-if="isSelf==0" :show-file-list="false" :on-change="uploadChange" :on-success="handleSuccess" :auto-upload="false" :disabled="upLoadBtnLoding"
          :multiple="true" :file-list="fileList" v-model:file-list="fileList" :on-error="handleError"
          style="display: flex;flex-direction: column;align-items: center;height: 32px;padding-left: 10px;padding-right: 10px;">
          <el-button type="primary" v-if="isSelf==0" :loading="upLoadBtnLoding">发送文件</el-button>
        </el-upload>
      </template>
      <el-table :data="fileList" max-height="300px">
        <el-table-column width="200" property="name" label="文件名" />
        <el-table-column label="大小" width="100">
          <template #default="scope">
            {{ bytesToSize(scope.row.size) }}
          </template>
        </el-table-column>
        <el-table-column label="上传进度" width="150">
          <template #default="scope">
            <el-progress :text-inside="true" status="success" :percentage="scope.row.percentage" :stroke-width="15"
              style="width: 120px;"></el-progress>
          </template>
        </el-table-column>
        <el-table-column label="删除" width="60">
          <template #default="scope">
            <el-icon @click="uploadDel(scope.row.uid)" style="cursor:pointer;">
              <Delete />
            </el-icon>
          </template>
        </el-table-column>
      </el-table>
      <div style="height: 50px;align-items: center;justify-content: center;display: flex;"><el-button type="primary" :loading="upLoadBtnLoding" @click="uploadFile">发送</el-button></div>
    </el-popover>
    <el-popconfirm title="确认要删除吗?" v-if="batchDelBtn" @confirm="tableDelMultiple()" confirm-button-text="确认"
      cancel-button-text="取消">
      <template #reference>
        <el-button type="primary" v-if="batchDelBtn" :disabled="batchDisabled">
          删除所选<el-icon class="el-icon-right shake" placement="right-start">
          </el-icon>
        </el-button>
      </template>
    </el-popconfirm>
  </div>
  <el-table :reserve-selection="true" ref="multipleTableRef" @selection-change="tableSelectionChange"
    row-key="share.shareId" :data="tableData" style="width: 100%">
    <el-table-column type="selection" width="55" />
    <el-table-column label="文件名" prop="share.title" width="auto" min-width="50%" />
    <el-table-column label="类型" width="auto" min-width="15%">
      <template #default="scope">
        <el-tag type="primary">{{ scope.row.share.ext }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="有效期" width="auto" min-width="15%">
      <template #default="scope">
        <el-button :type="scope.row.expireTimeBtnColor" @click="expireTimePopoverClick(scope.row, $event)">
          {{ scope.row.expireTimeBtnTxt }}
          <el-icon v-if="isSelf == 1" class="el-icon--right"><arrow-down /></el-icon>
        </el-button>
      </template>
    </el-table-column>
    <el-table-column label="大小" width="auto" min-width="15%">
      <template #default="scope">
        {{ bytesToSize(scope.row.share.size) }}
      </template>
    </el-table-column>
    <el-table-column label="操作" width="auto" min-width="20%">
      <template #default="scope">
        <el-tooltip popper-class="tooltip" :disabled="scope.row.downLoadTooltipDisabled" class="box-item" effect="light"
          placement="right-start">
          <template #content>
            <div style="line-height:40px;height: 40px;text-align: left;font-weight: bold;">
              {{ scope.row.downLoadTooltipTitle }}</div>
            <div style="height: 30px;"><el-progress :text-inside="true" :status="scope.row.downLoadTooltipStatus"
                :percentage="scope.row.downLoadTooltipPercentage" class="progress" :stroke-width="20"
                style="width: 120px;"></el-progress></div>
          </template>
          <div style="display: flex;align-items: center;justify-content: flex-start;gap: 8px;">
            <el-button v-if="scope.row.downLoadBtnShow" size="small" :loading="scope.row.downLoadBtnLoad"
              @click="downLoad(scope.row, scope.row.downLoadBtnEvent)">
              <!-- <el-icon class="is-loading" v-if="scope.row.downLoadBtnLoad" style="color: #303133;" size="10">
                <Loading />
              </el-icon> -->
              {{ scope.row.downLoadBtnTxt }}
            </el-button>
            <el-dropdown v-if="scope.row.openFileBtnShow" split-button type="primary" size="small"
              @command="openFolderBtnClick" @click="openFileBtnClick(scope.row)">
              打开文件
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="scope.row">打开文件夹</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-popconfirm title="确认要删除吗?" v-if="isSelf == 1" @confirm="tableDel(scope.$index, scope.row)"
              confirm-button-text="确认" cancel-button-text="取消">
              <template #reference>
                <el-button v-if="isSelf == 1" size="small" type="danger" :icon="Delete">
                </el-button>
              </template>
            </el-popconfirm>
          </div>
        </el-tooltip>
      </template>
    </el-table-column>
  </el-table>
  <el-popover v-model:visible="popoverVisible" :virtual-ref="tableRowClick" virtual-triggering placement="bottom"
    width="280">
    <el-input-number v-model="expireTimeNum" :min="1" :max="100" @change="expireTimeInputChange" size="small"
      style="width: 80px" />
    <el-select v-model="expireTimeUnit" :teleported="false" size="small" style="width: 80px">
      <el-option label="分钟" value="分钟" />
      <el-option label="小时" value="小时" />
      <el-option label="天" value="天" />
      <el-option label="永久" value="永久" />
    </el-select>
    <el-button type="primary" style="margin-right: 10px" @click="editShare()" size="small">修改</el-button>
  </el-popover>
  <el-pagination v-model:current-page="pageNumber" v-model:page-size="pageSize" :page-sizes="[15, 50, 100]"
    layout="total,sizes, prev, pager, next" :total=total @size-change="pagerSizeChange"
    @current-change="pagerCurrentChange"
    style="height: 50px;line-height: 50px;background-color: #ffffff;padding-left: 10px;">
  </el-pagination>
</template>

<script lang="js" setup>
import { ref, watch, onMounted, onUnmounted } from "vue"
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { AddShare, GetShareList, GetShareExt, DelShare, EditShare, DownLoad, OpenDir, OpenFile, DownLoadTaskStop, SyncHostShare, GetHttpSign } from "../../../wailsjs/go/main/App";
import { useRoute } from "vue-router";
import { bytesToSize, shareDataExtend } from "../../common/common.js";
import { Delete, ArrowDown, Refresh, Loading } from "@element-plus/icons-vue";
import { ElMessageBox, ElNotification } from 'element-plus';
import axios from 'axios'

const multipleTableRef = ref(null);
const multipleSelection = ref([]);
const route = useRoute();
const searchTitle = ref("");
const isSelf = ref(0);
const ip = ref("");
const popoverVisible = ref(false);
const tableRowClick = ref(null);
const expireTimeUnit = ref("小时");
const expireTimeNum = ref(1);
const tableData = ref([]);
const ext = ref([]);
const selectExt = ref([]);
const editShareId = ref(0);
const pageSize = ref(15);
const pageNumber = ref(0);
const total = ref(0);
const batchDisabled = ref(true);
const batchDelBtn = ref(false);
const fileList = ref([]);
const upLoadBtnLoding = ref(false)
const uploadPopoverVisible = ref(false);
isSelf.value = route.query.isSelf;
ip.value = route.query.ip;
if (isSelf.value == 1) {
  batchDelBtn.value = true;
} else {
  batchDelBtn.value = false;
}

onMounted(() => {
  EventsOn('NewShare', (downLoadHistory) => {
    var tmpDownLoadHistory = shareDataExtend(downLoadHistory, isSelf.value);
    tableData.value.push(tmpDownLoadHistory);
    hostOnlineNify('新增共享文件', "新增共享文件：" + downLoadHistory.title);
  });
  EventsOn('DelShare', (delShare) => {
    getShareList();
  });
  EventsOn('DownLoadTaskEvent', (task) => {
    for (var i = 0; i < tableData.value.length; i++) {
      if (tableData.value[i].share.shareId == task.downLoadHistory.share.shareId) {
        tableData.value[i] = shareDataExtend(task.downLoadHistory, isSelf);
      }
    }
  });
});

onUnmounted(() => {
  EventsOff('NewShare', () => {

  });
  EventsOff('DelShare', () => {

  });
  EventsOff('DownLoadTaskEvent', () => {

  });
});

const uploadChange = (uploadFile) => {
  //console.log(uploadFile);
  fileList.value.push(uploadFile);
  uploadPopoverVisible.value = true;
}

const uploadDel = (uid) => {
  fileList.value.forEach((item, index) => {
    if (item.uid == uid) {
      fileList.value.splice(index, 1)
    }
    if (fileList.value.length == 0) {
      uploadPopoverVisible.value = false;
    }
  });
}

const uploadFile = async () => {
  upLoadBtnLoding.value = true;
  var finish = 0;
  fileList.value.forEach((item, index) => {
    //console.log(item);
    const formData = new FormData();
    formData.append('file', item.raw);
    formData.append('lastModified', item.raw.lastModified);

    GetHttpSign()
      .then((res) => {
        fileList.value[index].percentage = 0;
        axios.post("http://" + ip.value + ":35286" + '/upload', formData, {
          headers: {
            'sign': res["sign"],
            'uller-client-time': res["uller-client-time"],
          },
          onUploadProgress: (event) => {
            const percentCompleted = Math.round((event.loaded * 100) / event.total);
            fileList.value[index].percentage = percentCompleted;
            if (fileList.value[index].percentage == 100) {
              finish++;
              if (finish >= fileList.value.length) {
                setTimeout(() => {
                  uploadPopoverVisible.value = false;
                  upLoadBtnLoding.value = false;
                  fileList.value = [];
                }, 1000);
              }
            }
          },
        })
          .catch((err) => {
            fileList.value[index].percentage = 0;
            ElMessageBox({
              type: 'error',
              title: '发送文件错误',
              message: err,
            })
          });
      })
      .catch((err) => {
        fileList.value[index].percentage = 0;
        ElMessageBox({
          type: 'error',
          title: '发送文件错误',
          message: err,
        })
      });
  });
};
// 处理上传成功
const handleSuccess = () => {
  
};

// 处理上传失败
const handleError = (err) => {
  ElMessageBox({
    type: 'error',
    title: '发送文件错误',
    message: err,
  })
  uploadProgress.value = 0; // 重置进度条
};

const hostOnlineNify = (title, msg) => {
  ElNotification({
    title: title,
    message: msg,
    position: 'bottom-left',
  })
}

watch(
  () => route.query.ip, (newVal, oldVal) => {
    if (newVal !== oldVal) {
      isSelf.value = route.query.isSelf;
      ip.value = route.query.ip;
      if (isSelf.value == 1) {
        batchDelBtn.value = true;
      } else {
        batchDelBtn.value = false;
      }
      getShareList();
    }
  }
);

GetShareExt(String(ip.value))
  .then((res) => {
    if (res === null) {
      return
    }
    for (var i = 0; i < res.length; i++) {
      ext.value.push({ value: res[i], label: res[i] })
    }
  })
  .catch((err) => {
    ElMessageBox({
      type: 'error',
      title: '获取分享文件类型错误',
      message: err,
    })
  });

const expireTimeInputChange = (value) => {
  expireTimeNum.value = value;
};

const expireTimePopoverClick = (row, event) => {
  if (isSelf.value == 1) {
    expireTimeNum.value = 1;
    expireTimeUnit.value = "小时";
    popoverVisible.value = true;
    tableRowClick.value = event.target;
    editShareId.value = row.share.shareId
  }
};

const openFileBtnClick = (row) => {
  var localPath = "";
  if (isSelf.value == 1) {
    localPath = row.share.localPath;
  } else {
    localPath = row.localPath;
  }
  OpenFile(row.historyId, localPath)
    .then(() => {
    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: err,
        message: err,
      })
      getShareList();
    });
}

const openFolderBtnClick = (row) => {
  var localPath = "";
  if (isSelf.value == 1) {
    localPath = row.share.localPath;
  } else {
    localPath = row.localPath;
  }
  if (localPath != "") {
    OpenDir(row.historyId, localPath)
      .then(() => {
      })
      .catch((err) => {
        ElMessageBox({
          type: 'error',
          title: err,
          message: err,
        })
        getShareList();
      });
  }
}

const downLoad = (row, event) => {
  if (event == "start" || event == "continue") {
    row.downLoadBtnLoad=true;
    DownLoad(row, row.share.ip, event)
      .then((res) => {
        for (var i = 0; i < tableData.value.length; i++) {
          //console.log(tableData.value[i]);
          if (res.share.shareId == tableData.value[i].share.shareId) {
            tableData.value[i].historyId = res.historyId
          }
        }
        row.downLoadBtnLoad=false;
      })
      .catch((err) => {
        row.downLoadBtnLoad=false;
        if (err == "dialog canceled") {
          return
        }
        ElMessageBox({
          type: 'error',
          title: '下载文件错误',
          message: err,
        })
      });
  } else if (event == "stop") {
    DownLoadTaskStop(row.share.shareId)
      .then((res) => {

      })
      .catch((err) => {
        if (err == "dialog canceled") {
          return
        }
        ElMessageBox({
          type: 'error',
          title: '停止下载文件错误',
          message: err,
        })
      });
  }
};

const downLoadBatch = () => {
  var downLoadHistory = []
  for (var i = 0; i < multipleSelection.value.length; i++) {
    downLoadHistory.push(multipleSelection.value[i])
  }
  DownLoadBatch(downLoadHistory, ip.value);
};

function pagerSizeChange() {
  getShareList();
}

function pagerCurrentChange() {
  getShareList();
}


function syncHostShare() {
  if (isSelf.value == 0) {
    //console.log("isSelf.value",isSelf.value)
    SyncHostShare(ip.value)
      .then((res) => {

      })
      .catch((err) => {
        ElMessageBox({
          type: 'error',
          title: '同步远程主机数据错误',
          message: err,
        })
      });
  }
  getShareList();
}

function getShareList() {
  //console.log("getShareList",ip.value);
  GetShareList(String(ip.value), searchTitle.value, selectExt.value, pageSize.value, pageNumber.value)
    .then((res) => {
      if (res.total === undefined) {
        tableData.value = [];
        return;
      }
      total.value = res.total;
      tableData.value = [];
      var tmpDownLoadHistory
      for (var i = 0; i < res.downLoadHistory.length; i++) {
        tmpDownLoadHistory = shareDataExtend(res.downLoadHistory[i], isSelf.value);
        tableData.value.push(tmpDownLoadHistory);
      }
    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: '获取共享文件列表错误',
        message: err,
      })
    });
}

getShareList();

function editShare() {
  if (editShareId.value != 0) {
    EditShare(String(editShareId.value), String(expireTimeUnit.value), String(expireTimeNum.value))
      .then((res) => {
        popoverVisible.value = false;
        getShareList();
      })
      .catch((err) => {
        ElMessageBox({
          type: 'error',
          title: '修改共享文件有效期错误',
          message: err,
        })
      });
  }
}

function addShare() {
  AddShare()
    .then((res) => {
      getShareList();
    })
    .catch((err) => {
      if (err == "dialog canceled") {
        return
      }
      ElMessageBox({
        type: 'error',
        title: '新增共享文件错误',
        message: err,
      })
    });
}

const tableDel = (index, row) => {
  var shareIds = [String(row.share.shareId)]
  DelShare(shareIds)
    .then((res) => {
      getShareList();
    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: '删除共享文件错误',
        message: err,
      })
    });
};

const tableSelectionChange = (val) => {
  if (val.length > 0) {
    batchDisabled.value = false;
  } else {
    batchDisabled.value = true;
  }
  multipleSelection.value = val;
  //console.log("tableSelectionChange", multipleSelection.value)
};

const tableDelMultiple = () => {
  const row = multipleSelection.value
  var shareIds = []
  for (let i = 0; i < row.length; i++) {
    shareIds.push(String(row[i].share.shareId))
  }
  //console.log("tableDelMultiple", shareIds)
  if (shareIds.length > 0) {
    DelShare(shareIds)
      .then((res) => {
        getShareList();
      })
      .catch((err) => {
        ElMessageBox({
          type: 'error',
          title: '批量删除共享文件错误',
          message: err,
        })
      });
  }
}
</script>