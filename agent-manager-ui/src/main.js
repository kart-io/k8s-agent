import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'

// VXETable
import VXETable from 'vxe-table'
import 'vxe-table/lib/style.css'
import VXETablePluginAntd from 'vxe-table-plugin-antd'
import 'vxe-table-plugin-antd/dist/style.css'

// App & Router
import App from './App.vue'
import router from './router'

// Directives
import permissionDirective from './directives/permission'

// Styles
import './assets/styles/main.scss'

// VXETable 使用插件
VXETable.use(VXETablePluginAntd)

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(Antd)
app.use(VXETable)
app.use(permissionDirective)

app.mount('#app')
