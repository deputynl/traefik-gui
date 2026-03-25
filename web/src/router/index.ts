import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/views/Dashboard.vue'
import StaticConfig from '@/views/StaticConfig.vue'
import DynamicConfig from '@/views/DynamicConfig.vue'
import DynamicEditor from '@/views/DynamicEditor.vue'
import Certificates from '@/views/Certificates.vue'
import DockerLabels from '@/views/DockerLabels.vue'
import AuditLog from '@/views/AuditLog.vue'
import Activity from '@/views/Activity.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/',              component: Dashboard,    name: 'dashboard' },
    { path: '/static',        component: StaticConfig,  name: 'static' },
    { path: '/dynamic',       component: DynamicConfig, name: 'dynamic' },
    { path: '/dynamic/:file', component: DynamicEditor, name: 'dynamic-editor' },
    { path: '/certificates',  component: Certificates,  name: 'certificates' },
    { path: '/docker',        component: DockerLabels,  name: 'docker' },
    { path: '/audit',         component: AuditLog,      name: 'audit' },
    { path: '/activity',      component: Activity,      name: 'activity' },
  ],
})

export default router
