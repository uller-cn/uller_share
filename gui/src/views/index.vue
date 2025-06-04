<template>
    <el-container v-loading="loading" class="common-layout">
      <el-header style="width: 100%; padding: 0px;height: 80px;"><top :version="version"></top></el-header>
      <el-container>
        <el-aside style="width: 240px;background-color: #F8F9FB;padding-top: 7px;"><left v-if="isInit"></left></el-aside>
        <el-main style="background-color: #F8F9FB;padding: 10px;"><container></container></el-main>
      </el-container>
    </el-container>
</template>

<script lang="js" setup>
  import top from './header.vue'
  import left from './left.vue'
  import container from './main.vue'
  import 'element-plus/dist/index.css'
  import { ref} from 'vue'
  import { Init } from '../../wailsjs/go/main/App'
  import { ElMessageBox } from 'element-plus'
  
  const version = ref('')
  // import { useRoute } from "vue-router";
  // const route = useRoute();
  // console.log("route.query.ip",route.query.ip);
  // sessionStorage.removeItem('token');

  const isInit = ref(false)
  const loading = ref(true)
    Init().then((res) => {
      isInit.value = true
      loading.value = false;
      version.value = res["version"]
      if(res["timeErr"] == "1"){
        ElMessageBox({
          type: 'error',
          title: '错误',
          message: "您的电脑时间不准确，使用不准确时间会影响系统使用，请校准。",
        })
      }
    }).catch((err) => {
      loading.value = false;
      ElMessageBox({
        type: 'error',
        title: '错误',
        message: err,
      })
    })
  
</script>
<style type="text/css">
.common-layout{
  width: 100%;
  height: 100%;
  margin: 0px;
  padding: 0px;
}
</style>
