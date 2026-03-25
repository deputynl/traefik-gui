import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/views/Dashboard.vue'
import StaticConfig from '@/views/StaticConfig.vue'
import DynamicConfig from '@/views/DynamicConfig.vue'
import DynamicEditor from '@/views/DynamicEditor.vue'
import Certificates from '@/views/Certificates.vue'
import DockerLabels from '@/views/DockerLabels.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/',             component: Dashboard,    name: 'dashboard' },
    { path: '/static',       component: StaticConfig,  name: 'static' },
    { path: '/dynamic',         component: DynamicConfig, name: 'dynamic' },
    { path: '/dynamic/:file',   component: DynamicEditor, name: 'dynamic-editor' },
    { path: '/certificates', component: Certificates,  name: 'certificates' },
    { path: '/docker',       component: DockerLabels,  name: 'docker' },
  ],
})

export default router
