<template>
  <div class="container">
    <div class="left"><img src="../assets/images/logo.png" width="40" height="40" style="padding-left: 10px;" />
      <span style="font-size: 24px;font-weight: bold;padding-left: 10px;color: black;padding-right: 10px;">悠乐快传<span style="font-size: 10px;">v{{ version }}</span></span>
    </div>
    <div class="search">
      <el-input v-model="searchTitle" style="width: 40%" placeholder="全局搜索" :prefix-icon="Search" @keyup.enter="getHostShareList()" />
      <el-select v-model="selectExt" clearable placeholder="类型" multiple style="width: 140px;margin-left: 10px;">
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
      <el-button type="primary" style="margin-left: 10px" @click="getHostShareList()">搜索</el-button>
    </div>
  </div>
</template>

<script lang="js" setup>
import { Search } from '@element-plus/icons-vue'
import { ref,defineProps  } from "vue"
import router from '../router'
import { GetShareExt} from "../../wailsjs/go/main/App";
import emitter from '../event/eventBus';

const props = defineProps({
  version: {
    type: String,
    required: true
  }
});


const searchTitle = ref('');
const ext = ref([]);
const selectExt = ref([]);

function getHostShareList(){
  //console.log("searchClick",selectExt.value)
  emitter.emit('searchClick', {title:searchTitle.value,ext:selectExt.value});
  router.push({ path: '/host/share', query: { title: searchTitle.value,ext:selectExt.value }});
}

GetShareExt()
  .then((res) => {
    if (res === null) {
      return
    }
    //console.log("GetShareExt", res)
    for (var i = 0; i < res.length; i++) {
      ext.value.push({ value: res[i], label: res[i] })
    }
  })
  .catch((err) => {
    ElMessageBox({
      type: 'error',
      title: 'GetShareExt错误',
      message: err,
    })
  });
</script>

<style type="text/css">
.shake:hover {
  animation: shake 800ms ease-in-out;
}

@keyframes shake {

  20%,
  80% {
    transform: translate3d(-1px, 0, 0);
  }

  30%,
  70% {
    transform: translate3d(+2px, 0, 0);
  }

  50% {
    transform: translate3d(-3px, 0, 0);
  }
}

.container {
  margin: 0px;
  height: 80px;
  line-height: 80px;
  display: flex;
  width: 100%;
  align-items: center;
  background-color: #FFFFFF;
}

.left {
  width: 300px;
  height: 80px;
  float: left;
  display: flex;
  justify-content: center;
  align-items: center;
}

.search {
  line-height: 80px;
  width: 100%;
  text-align: left;
}
</style>
