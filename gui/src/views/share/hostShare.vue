<template>
  <el-table :reserve-selection="true" ref="multipleTableRef" row-key="shareId" :data="tableData" style="width: 100%">
    <el-table-column label="文件名" prop="share.title" width="auto" min-width="50%" />
    <el-table-column label="类型" width="auto" min-width="15%">
      <template #default="scope">
        <el-tag type="primary">{{ scope.row.share.ext }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="有效期" width="auto" min-width="15%">
      <template #default="scope">
        <el-button :type="scope.row.expireTimeBtnColor">
          {{ scope.row.expireTimeBtnTxt }}
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
          </div>
        </el-tooltip>
      </template>
    </el-table-column>
  </el-table>
  <el-pagination v-model:current-page="pageNumber" v-model:page-size="pageSize" :page-sizes="[15, 50, 100]"
    layout="total,sizes, prev, pager, next, jumper" :total=total @size-change="pagerSizeChange"
    @current-change="pagerCurrentChange" style="height: 50px;line-height: 50px;background-color: #ffffff;padding-left: 10px;" >
  </el-pagination>
</template>

<script lang="js" setup>
import { ref,onMounted,onUnmounted,reactive } from "vue"
import { GetHostShareList,DownLoad,DownLoadTaskStop,OpenFile,OpenDir} from "../../../wailsjs/go/main/App";
import { bytesToSize,shareDataExtend } from "../../common/common.js";
import { Loading } from "@element-plus/icons-vue";
import { ElMessageBox } from 'element-plus';
import { useRoute } from "vue-router";
import emitter from '../../event/eventBus';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'

const route = useRoute();
const searchTitle = ref("");
const selectExt = reactive([]);
const tableData = ref([]);
const pageSize = ref(15);
const pageNumber = ref(0);
const total = ref(0)

const searchClickEvent = (data) => {
  //console.log("searchClickEvent",data.title, data.ext)
  searchTitle.value = data.title;
  selectExt.value = data.ext;
  getHostShareList();
};
onMounted(() => {
  emitter.on('searchClick', searchClickEvent);

  EventsOn('DownLoadTaskEvent', (task) => {
    //console.log("DownLoadTaskEvent",task);
    for (var i = 0; i < tableData.value.length; i++) {
      if (tableData.value[i].share.shareId == task.downLoadHistory.share.shareId) {
        tableData.value[i] = shareDataExtend(task.downLoadHistory, 0);
      }
    }
  });
});
onUnmounted(() => {
  emitter.off('searchClick', searchClickEvent);

  EventsOff('DownLoadTaskEvent', () => {

  });
});
searchTitle.value = route.query.title
selectExt.value = route.query.ext

function pagerSizeChange(){
  getHostShareList();
}

function pagerCurrentChange(){
  getHostShareList();
}

const getHostShareList = () => {
  GetHostShareList(searchTitle.value, selectExt.value, pageSize.value, pageNumber.value)
    .then((res) => {
      if (res.total === undefined) {
        tableData.value = [];
        return;
      }
      total.value = res.total;
      tableData.value = [];
      var tmpDownLoadHistory
      for (var i = 0; i < res.downLoadHistory.length; i++) {
        tmpDownLoadHistory = shareDataExtend(res.downLoadHistory[i], 0);
        tableData.value.push(tmpDownLoadHistory);
      }
    })
    .catch((err) => {
      ElMessageBox({
        type: 'error',
        title: 'GetShareList错误',
        message: err,
      })
    });
}

getHostShareList();

const downLoad = (row, event) => {
  if (event == "start" || event == "continue") {
    DownLoad(row, row.share.ip, event)
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

const openFileBtnClick = (row) => {
  //console.log(row)
  OpenFile(row.historyId,row.row.localPath)
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
  //console.log(row)
  OpenDir(row.historyId,row.localPath)
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
</script>