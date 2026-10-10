import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'
import 'highlight.js/styles/github.css'
import 'md-editor-v3/lib/style.css'
import '@/assets/style.css'
import App from './App.vue'
import router from './router'
import { installAnalytics } from '@/utils/analytics'

const app = createApp(App)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
// 先注册埋点的路由监听，再启动路由，避免遗漏第一次打开页面的导航。
installAnalytics(router)
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.mount('#app')
