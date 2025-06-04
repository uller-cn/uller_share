import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import router from './router'
//import { createPinia } from 'pinia'
import './style.css';
import zhCn from 'element-plus/es/locale/lang/zh-cn';

localStorage.removeItem('token');

const app = createApp(App);
//const piniaStore = createPinia()

app.use(ElementPlus, { locale: zhCn });
app.use(router);
//app.use(piniaStore)
app.mount('#app');
