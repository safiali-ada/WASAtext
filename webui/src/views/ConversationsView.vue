<template>
  <div class="main-container" :class="{ 'chat-open': !!selectedConversation }">
    <div class="sidebar">
      <div class="sidebar-header">
        <div>
          <h2>Conversations</h2>
          <small class="current-user">Logged in as {{ username }}</small>
        </div>
        <div class="header-buttons">
          <button class="btn-icon" @click="openNewChat" title="New chat">✉️</button>
          <button class="btn-icon" @click="openNewGroup" title="New group">👥</button>
          <button class="btn-icon" @click="goToProfile" title="Profile">⚙️</button>
        </div>
      </div>

      <ErrorMsg :msg="error" @close="error = ''" />

      <div class="search-container">
        <input v-model="searchQuery" placeholder="Search users..." @input="searchUsers" />
        <div v-if="searchResults.length > 0" class="search-results">
          <div
            v-for="user in searchResults"
            :key="user.id"
            class="search-result-item user-row"
            @click="startConversation(user)"
          >
            <UserAvatar :url="user.photoUrl" :name="user.username" :size="28" />
            <span>{{ user.username }}</span>
          </div>
        </div>
        <div v-else-if="searchQuery && searchDone" class="search-empty">No users found</div>
      </div>

      <div class="conversation-list">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          class="conversation-item"
          :class="{ active: selectedConversation === conv.id }"
          @click="openConversation(conv)"
        >
          <UserAvatar class="conversation-avatar" :url="conv.photoUrl" :name="conv.name" :size="45" />
          <div class="conversation-info">
            <div class="conversation-name">
              <span v-if="conv.type === 'group'" title="Group">👥 </span>{{ conv.name }}
            </div>
            <div class="conversation-preview" v-if="conv.latestMessage">
              <span v-if="conv.latestMessage.senderId === userId">You: </span>
              <span v-if="conv.latestMessage.type === 'photo'">📷 </span>
              {{ conv.latestMessage.content || (conv.latestMessage.type === 'photo' ? 'Photo' : '') }}
            </div>
            <div class="conversation-preview" v-else>No messages yet</div>
          </div>
          <div class="conversation-time" v-if="conv.latestMessage">
            {{ formatDateTime(conv.latestMessage.timestamp) }}
          </div>
        </div>

        <div v-if="loaded && conversations.length === 0" class="empty-state">
          <p>No conversations yet</p>
          <p style="font-size: 12px;">Search for users to start chatting</p>
        </div>
      </div>
    </div>

    <div class="content">
      <router-view :key="$route.params.id || 'none'" @changed="loadConversations" />
    </div>

    <!-- New Chat Modal: lists every WASAText user -->
    <div v-if="showNewChat" class="modal-overlay" @click.self="showNewChat = false">
      <div class="modal">
        <h3>New Conversation</h3>
        <input v-model="newChatFilter" placeholder="Filter users..." />
        <div class="search-results modal-list">
          <div
            v-for="user in filteredAllUsers(newChatFilter)"
            :key="user.id"
            class="search-result-item user-row"
            @click="startConversationFromModal(user)"
          >
            <UserAvatar :url="user.photoUrl" :name="user.username" :size="28" />
            <span>{{ user.username }}</span>
          </div>
          <div v-if="filteredAllUsers(newChatFilter).length === 0" class="search-empty">No users found</div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showNewChat = false">Cancel</button>
        </div>
      </div>
    </div>

    <!-- New Group Modal: name + members -->
    <div v-if="showNewGroup" class="modal-overlay" @click.self="showNewGroup = false">
      <div class="modal">
        <h3>Create Group</h3>
        <input v-model="newGroupName" placeholder="Group name" maxlength="50" />
        <label class="field-label">Members ({{ selectedMembers.length }} selected)</label>
        <div class="chips" v-if="selectedMembers.length > 0">
          <span v-for="u in selectedMembers" :key="u.id" class="chip">
            {{ u.username }} <button class="chip-remove" @click="toggleMember(u)">✕</button>
          </span>
        </div>
        <input v-model="groupMemberFilter" placeholder="Filter users..." />
        <div class="search-results modal-list">
          <label
            v-for="user in filteredAllUsers(groupMemberFilter)"
            :key="user.id"
            class="search-result-item user-row"
          >
            <input type="checkbox" class="member-checkbox" :checked="isSelected(user)" @change="toggleMember(user)" />
            <UserAvatar :url="user.photoUrl" :name="user.username" :size="28" />
            <span>{{ user.username }}</span>
          </label>
        </div>
        <ErrorMsg :msg="groupError" @close="groupError = ''" />
        <div class="modal-actions">
          <button class="btn-secondary" @click="showNewGroup = false">Cancel</button>
          <button class="btn-primary" @click="createGroup" :disabled="!newGroupName.trim() || creatingGroup">
            {{ creatingGroup ? 'Creating...' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js'
import ErrorMsg from '../components/ErrorMsg.vue'
import UserAvatar from '../components/UserAvatar.vue'
import { formatDateTime, errorText } from '../services/format.js'

const REFRESH_ERROR = 'Could not load conversations. Retrying...'

export default {
  name: 'ConversationsView',
  components: { ErrorMsg, UserAvatar },
  data() {
    return {
      conversations: [],
      loaded: false,
      error: '',
      userId: localStorage.getItem('wasatext_user_id'),
      username: localStorage.getItem('wasatext_username') || '',
      searchQuery: '',
      searchResults: [],
      searchDone: false,
      allUsers: [],
      showNewChat: false,
      newChatFilter: '',
      showNewGroup: false,
      newGroupName: '',
      groupMemberFilter: '',
      selectedMembers: [],
      groupError: '',
      creatingGroup: false
    }
  },
  computed: {
    selectedConversation() {
      return this.$route.params.id || null
    }
  },
  mounted() {
    this.loadConversations()
    this.pollInterval = setInterval(this.loadConversations, 3000)
  },
  beforeUnmount() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval)
    }
  },
  methods: {
    formatDateTime,
    async loadConversations() {
      try {
        const response = await axios.get(`/users/${this.userId}/conversations`)
        this.conversations = response.data || []
        this.loaded = true
        if (this.error === REFRESH_ERROR) this.error = ''
        this.username = localStorage.getItem('wasatext_username') || this.username
        if (this.searchQuery) {
          await this.searchUsers()
        }
      } catch (err) {
        if (err.response?.status !== 401) {
          this.error = REFRESH_ERROR
        }
      }
    },
    async searchUsers() {
      if (this.searchQuery.length < 1) {
        this.searchResults = []
        this.searchDone = false
        return
      }
      try {
        const response = await axios.get('/users', { params: { q: this.searchQuery } })
        this.searchResults = (response.data || []).filter(u => u.id !== this.userId)
        this.searchDone = true
      } catch (err) {
        this.error = errorText(err, 'User search failed')
      }
    },
    async loadAllUsers() {
      try {
        const response = await axios.get('/users', { params: { q: '' } })
        this.allUsers = (response.data || []).filter(u => u.id !== this.userId)
      } catch (err) {
        this.error = errorText(err, 'Could not load users')
      }
    },
    filteredAllUsers(filter) {
      const f = (filter || '').toLowerCase()
      return this.allUsers.filter(u => u.username.toLowerCase().includes(f))
    },
    openNewChat() {
      this.newChatFilter = ''
      this.showNewChat = true
      this.loadAllUsers()
    },
    openNewGroup() {
      this.newGroupName = ''
      this.groupMemberFilter = ''
      this.selectedMembers = []
      this.groupError = ''
      this.showNewGroup = true
      this.loadAllUsers()
    },
    isSelected(user) {
      return this.selectedMembers.some(u => u.id === user.id)
    },
    toggleMember(user) {
      if (this.isSelected(user)) {
        this.selectedMembers = this.selectedMembers.filter(u => u.id !== user.id)
      } else {
        this.selectedMembers.push(user)
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
        this.error = errorText(err, 'Could not start the conversation')
      }
    },
    async startConversationFromModal(user) {
      this.showNewChat = false
      await this.startConversation(user)
    },
    async createGroup() {
      this.creatingGroup = true
      this.groupError = ''
      try {
        const response = await axios.post('/groups', { name: this.newGroupName.trim() })
        const groupId = response.data.id
        for (const member of this.selectedMembers) {
          await axios.post(`/groups/${groupId}/members`, { userId: member.id })
        }
        this.showNewGroup = false
        await this.loadConversations()
        this.$router.push(`/conversations/${groupId}`)
      } catch (err) {
        this.groupError = errorText(err, 'Could not create the group')
      } finally {
        this.creatingGroup = false
      }
    },
    openConversation(conv) {
      this.$router.push(`/conversations/${conv.id}`)
    },
    goToProfile() {
      this.$router.push('/profile')
    }
  }
}
</script>
