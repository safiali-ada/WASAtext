<template>
  <div class="main-container">
    <div class="sidebar">
      <div class="sidebar-header">
        <h2>Conversations</h2>
        <div>
          <button class="btn-icon" @click="showNewChat = true" title="New Chat">+</button>
          <button class="btn-icon" @click="showNewGroup = true" title="New Group">👥</button>
          <button class="btn-icon" @click="goToProfile" title="Profile">⚙️</button>
        </div>
      </div>

      <div class="search-container">
        <input v-model="searchQuery" placeholder="Search users..." @input="searchUsers" />
        <div v-if="searchResults.length > 0" class="search-results">
          <div 
            v-for="user in searchResults" 
            :key="user.id" 
            class="search-result-item"
            @click="startConversation(user)"
          >
            {{ user.username }}
          </div>
        </div>
      </div>

      <div class="conversation-list">
        <div 
          v-for="conv in conversations" 
          :key="conv.id" 
          class="conversation-item"
          :class="{ active: selectedConversation === conv.id }"
          @click="openConversation(conv)"
        >
          <div class="conversation-avatar">{{ getInitial(conv.name) }}</div>
          <div class="conversation-info">
            <div class="conversation-name">{{ conv.name }}</div>
            <div class="conversation-preview" v-if="conv.latestMessage">
              {{ conv.latestMessage.content }}
            </div>
          </div>
          <div class="conversation-time" v-if="conv.latestMessage">
            {{ formatTime(conv.latestMessage.timestamp) }}
          </div>
        </div>

        <div v-if="conversations.length === 0" class="empty-state">
          <p>No conversations yet</p>
          <p style="font-size: 12px;">Search for users to start chatting</p>
        </div>
      </div>
    </div>

    <div class="content">
      <div v-if="!selectedConversation" class="empty-state">
        <p>Select a conversation to start messaging</p>
      </div>
      <router-view v-else />
    </div>

    <!-- New Chat Modal -->
    <div v-if="showNewChat" class="modal-overlay" @click.self="showNewChat = false">
      <div class="modal">
        <h3>New Conversation</h3>
        <input v-model="newChatUsername" placeholder="Search username..." @input="searchForNewChat" />
        <div v-if="newChatResults.length > 0" class="search-results">
          <div 
            v-for="user in newChatResults" 
            :key="user.id" 
            class="search-result-item"
            @click="startConversationFromModal(user)"
          >
            {{ user.username }}
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showNewChat = false">Cancel</button>
        </div>
      </div>
    </div>

    <!-- New Group Modal -->
    <div v-if="showNewGroup" class="modal-overlay" @click.self="showNewGroup = false">
      <div class="modal">
        <h3>Create Group</h3>
        <input v-model="newGroupName" placeholder="Group name" />
        <div class="modal-actions">
          <button class="btn-secondary" @click="showNewGroup = false">Cancel</button>
          <button class="btn-primary" @click="createGroup" :disabled="!newGroupName">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js'

export default {
  name: 'ConversationsView',
  data() {
    return {
      conversations: [],
      selectedConversation: null,
      searchQuery: '',
      searchResults: [],
      showNewChat: false,
      showNewGroup: false,
      newChatUsername: '',
      newChatResults: [],
      newGroupName: ''
    }
  },
  mounted() {
    this.loadConversations()
  },
  methods: {
    async loadConversations() {
      try {
        const userId = localStorage.getItem('wasatext_user_id')
        const response = await axios.get(`/users/${userId}/conversations`)
        this.conversations = response.data || []
      } catch (err) {
        console.error('Error loading conversations:', err)
      }
    },
    async searchUsers() {
      if (this.searchQuery.length < 1) {
        this.searchResults = []
        return
      }
      try {
        const response = await axios.get(`/users?q=${encodeURIComponent(this.searchQuery)}`)
        const userId = localStorage.getItem('wasatext_user_id')
        this.searchResults = (response.data || []).filter(u => u.id !== userId)
      } catch (err) {
        console.error('Error searching users:', err)
      }
    },
    async searchForNewChat() {
      if (this.newChatUsername.length < 1) {
        this.newChatResults = []
        return
      }
      try {
        const response = await axios.get(`/users?q=${encodeURIComponent(this.newChatUsername)}`)
        const userId = localStorage.getItem('wasatext_user_id')
        this.newChatResults = (response.data || []).filter(u => u.id !== userId)
      } catch (err) {
        console.error('Error searching users:', err)
      }
    },
    async startConversation(user) {
      try {
        const response = await axios.post('/conversations', { userId: user.id })
        this.searchQuery = ''
        this.searchResults = []
        await this.loadConversations()
        this.openConversation(response.data)
      } catch (err) {
        console.error('Error starting conversation:', err)
      }
    },
    async startConversationFromModal(user) {
      await this.startConversation(user)
      this.showNewChat = false
      this.newChatUsername = ''
      this.newChatResults = []
    },
    async createGroup() {
      try {
        const response = await axios.post('/groups', { name: this.newGroupName })
        this.showNewGroup = false
        this.newGroupName = ''
        await this.loadConversations()
        this.$router.push(`/conversations/${response.data.id}`)
      } catch (err) {
        console.error('Error creating group:', err)
      }
    },
    openConversation(conv) {
      this.selectedConversation = conv.id
      this.$router.push(`/conversations/${conv.id}`)
    },
    goToProfile() {
      this.$router.push('/profile')
    },
    getInitial(name) {
      return name ? name.charAt(0).toUpperCase() : '?'
    },
    formatTime(timestamp) {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      const now = new Date()
      const diff = now - date
      
      if (diff < 86400000) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      }
      return date.toLocaleDateString()
    }
  }
}
</script>
