<template>
  <div style="
      width: 100%;
      height: 60px;
      line-height: 60px;
      text-align: left;
      background-color: #ffffff;
    ">
    <el-input v-model="searchTitle" placeholder="文件搜索" style="width: 200px; margin-right: 10px; margin-left: 10px"
      @keyup.enter="getHistoryList()" />
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
    <el-button type="primary" style="margin-right: 10px" @click="getHistoryList()">搜索</el-button>
  </div>
  <div style="
      width: 100%;
      height: 60px;
      line-height: 60px;
      text-align: left;
      margin-top: 15px;
      margin-bottom: 10px;
      background-color: #ffffff;
    ">
    <el-button type="primary" style="margin-left: 10px" @click="getHistoryList()">
      <el-icon>
        <Refresh />
      </el-icon>
    </el-button>
    <el-popconfirm title="确认要删除吗?" @confirm="tableDelMultiple()" confirm-button-text="确认" cancel-button-text="取消">
      <template #reference>
        <el-button type="primary" :disabled="batchDisabled">
          删除所选<el-icon class="el-icon-right shake" placement="right-start">
          </el-icon>
        </el-button>
      </template>
    </el-popconfirm>
  </div>
  <el-tabs style="background-color: #ffffff;" @tab-click="tabsClick" v-model="activeTab">
    <el-tab-pane name="downLoad">
      <template #label>
        <span style="width: 100px;">
          下载历史
        </span>
      </template>
    </el-tab-pane>
    <el-tab-pane name="upload">
      <template #label>
        <span style="width: 100px;">
          其他主机上传
        </span>
      </template>
    </el-tab-pane>
  </el-tabs>
  <el-table :reserve-selection="true" ref="multipleTableRef" @selection-change="tableSelectionChange"
    row-key="share.shareId" :data="tableData" style="width: 100%">
    <el-table-column type="selection" width="55" />
    <el-table-column label="文件名" prop="title" width="auto" min-width="30%" />
    <el-table-column label="来源" prop="ip" width="auto" min-width="20%" />
    <el-table-column label="类型" width="auto" min-width="15%">
      <template #default="scope">
        <el-tag type="primary">{{ scope.row.ext }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="进度" width="auto" min-width="15%">
      <template #default="scope">
        <el-tooltip popper-class="tooltip" :disabled="scope.row.downLoadTooltipDisabled" class="box-item" effect="light"
          placement="right-start" style="">
          <template #content>
            <div style="line-height:40px;height: 40px;text-align: left;font-weight: bold;">
              {{ scope.row.downLoadTooltipTitle }}</div>
            <div style="height: 30px;"><el-progress :text-inside="true" :status="scope.row.downLoadTooltipStatus"
                :percentage="scope.row.downLoadTooltipPercentage" class="progress" :stroke-width="20"
                style="width: 120px;"></el-progress></div>
          </template>
          <div style="height: 30px;width: 120px;"><el-progress :text-inside="true"
              :status="scope.row.downLoadTooltipStatus" :percentage="scope.row.downLoadTooltipPercentage"
              class="progress" :stroke-width="20" style="width: 120px;"></el-progress></div>
        </el-tooltip>
      </template>
    </el-table-column>
    <el-table-column label="大小" width="auto" min-width="15%">
      <template #default="scope">
        {{ bytesToSize(scope.row.size) }}
      </template>
    </el-table-column>
    <el-table-column label="操作" width="auto" min-width="15%">
      <template #default="scope">
        <div style="display: flex;align-items: center;justify-content: flex-start;gap: 8px;">
          <el-button v-if="scope.row.downLoadBtnShow" size="small"
            @click="downLoad(scope.row, scope.row.downLoadBtnEvent)">
            <el-icon class="is-loading" v-if="scope.row.downLoadBtnLoad" style="color: #303133;" size="10">
              <Loading />
            </el-icon>
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
          <el-popconfirm title="确认要删除吗?" @confirm="tableDel(scope.$index, scope.row)" confirm-button-text="确认"
            cancel-button-text="取消">
            <template #reference>
              <el-button size="small" type="danger" :icon="Delete">
              </el-button>
            </template>
          </el-popconfirm>
        </div>
      </template>
    </el-table-column>
  </el-table>
  <el-pagination v-model:current-page="pageNumber" v-model:page-size="pageSize" :page-sizes="[15, 50, 100]"
    layout="total,sizes, prev, pager, next" :total=total @size-change="pagerSizeChange"
    @current-change="pagerCurrentChange"
    style="height: 50px;line-height: 50px;background-color: #ffffff;padding-left: 10px;">
  </el-pagination>
</template>

<script lang="js" setup>
import { ref, onMounted } from "vue"
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { GetHistoryList, GetShareExt, DelHistory, OpenDir, OpenFile, DownLoad, DownLoadTaskStop } from "../../../wailsjs/go/main/App";
import { bytesToSize, shareDataExtend } from "../../common/common.js";
import { Delete, Refresh, Loading } from "@element-plus/icons-vue";
import { ElMessageBox } from 'element-plus';

const multipleTableRef = ref(null);
const multipleSelection = ref([]);
const searchTitle = ref("");
const isSelf = ref(0);
const tableData = ref([]);
const ext = ref([]);
const selectExt = ref([]);
const pageSize = ref(15);
const pageNumber = ref(0);
const total = ref(0);
const batchDisabled = ref(true);
const activeTab = ref("downLoad");

onMounted(() => {
  EventsOn('DownLoadTaskEvent', (task) => {
    for (var i = 0; i < tableData.value.length; i++) {
      if (tableData.value[i].share.shareId == task.downLoadHistory.share.shareId) {
        tableData.value[i] = shareDataExtend(task.downLoadHistory, isSelf);
      }
    }
  });
});

GetShareExt()
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

  const openFileBtnClick = (row) => {
  var localPath = "";
  if (isSelf.value == 1) {
    localPath = row.share.localPath;
  } else {
    localPath = row.localPath;
  }
  //console.log("openFileBtnClick",localPath);
  OpenFile(row.historyId,localPath)
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
  //console.log("openFolderBtnClick",row);
  var localPath = "";
  if (isSelf.value == 1) {
    localPath = row.share.localPath;
  } else {
    localPath = row.localPath;
  }
  //console.log("openFolderBtnClick",localPath)
  if (localPath != "") {
    OpenDir(row.historyId,localPath)
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
    DownLoad(row, row.ip, event)
      .then((res) => {
        for (var i = 0; i < tableData.value.length; i++) {
          //console.log(tableData.value[i]);
          if (res.share.shareId == tableData.value[i].share.shareId) {
            tableData.value[i].historyId = res.historyId
          }
        }
      })
      .catch((err) => {
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

function pagerSizeChange() {
  getHistoryList();
}

function pagerCurrentChange() {
  getHistoryList();
}

function tabsClick() {
  //console.log("activeTab",activeTab.value);
  getHistoryList();
}

function getHistoryList() {
  var sType = 1;
  if (activeTab.value == "upload") {
    sType = 0;
  }
  GetHistoryList(searchTitle.value, selectExt.value, sType, pageSize.value, pageNumber.value)
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
      //console.log("tableData.value", tableData.value)
    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: '获取共享文件列表错误',
        message: err,
      })
    });
}

getHistoryList();

const tableDel = (index, row) => {
  var historyIds = [String(row.historyId)]
  DelHistory(historyIds)
    .then((res) => {
      getHistoryList();
    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: '删除历史记录错误',
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
};

const tableDelMultiple = () => {
  const row = multipleSelection.value
  var historyId = []
  for (let i = 0; i < row.length; i++) {
    historyId.push(String(row[i].historyId))
  }
  //console.log("tableDelMultiple", historyId)
  if (historyId.length > 0) {
    DelHistory(historyId)
      .then((res) => {
        getHistoryList();
      })
      .catch((err) => {
        ElMessageBox({
          type: 'error',
          title: '批量删除历史记录错误',
          message: err,
        })
      });
  }
}
</script>