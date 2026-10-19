import { h } from 'vue'
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
        // The chat is nested so the conversation list stays visible (and keeps refreshing)
        // next to the open conversation.
        path: '/conversations',
        component: ConversationsView,
        meta: { requiresAuth: true },
        children: [
            {
                path: '',
                name: 'conversations',
                component: {
                    name: 'NoChatSelected',
                    render: () => h('div', { class: 'empty-state' }, [h('p', 'Select a conversation to start messaging')])
                }
            },
            {
                path: ':id',
                name: 'chat',
                component: ChatView,
                meta: { requiresAuth: true }
            }
        ]
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
