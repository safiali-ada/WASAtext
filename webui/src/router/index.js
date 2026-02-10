import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ConversationsView from '../views/ConversationsView.vue'
import ChatView from '../views/ChatView.vue'
import ProfileView from '../views/ProfileView.vue'

const routes = [
    {
        path: '/',
        name: 'login',
        component: LoginView
    },
    {
        path: '/conversations',
        name: 'conversations',
        component: ConversationsView,
        meta: { requiresAuth: true }
    },
    {
        path: '/conversations/:id',
        name: 'chat',
        component: ChatView,
        meta: { requiresAuth: true }
    },
    {
        path: '/profile',
        name: 'profile',
        component: ProfileView,
        meta: { requiresAuth: true }
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// Navigation guard for auth
router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('wasatext_token')

    if (to.meta.requiresAuth && !token) {
        next('/')
    } else if (to.name === 'login' && token) {
        next('/conversations')
    } else {
        next()
    }
})

export default router
