import { createRouter, createWebHashHistory } from 'vue-router'
import index from './views/index.vue'
import share from './views/share/share.vue'
import history from './views/share/history.vue'
import hostShare from './views/share/hostShare.vue'

const routes = [
    {
        path: '/',
        redirect: (to) => {
            const query = to.query;
            return { path: '/share', query };
        },
        component: index,
        children: [{
            path: '/share',
            name: 'share',
            component: share,
        }, {
            path: '/host/share',
            name: 'hostShare',
            component: hostShare,
        }, {
            path: '/history',
            name: 'history',
            component: history,
        }],
    }
]

const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

export default router