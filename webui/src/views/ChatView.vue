<template>
  <div class="chat-view">
    <div class="chat-header">
      <button class="btn-icon" @click="goBack">←</button>
      <div class="conversation-avatar" style="margin: 0 10px;">
        <img v-if="conversationPhotoUrl" :src="assetURL(conversationPhotoUrl)" style="width:40px;height:40px;border-radius:50%;object-fit:cover;" />
        <span v-else>{{ getInitial(conversation.name) }}</span>
      </div>
      <div>
        <h3>{{ conversation.name }}</h3>
        <small v-if="conversation.type === 'group'">{{ conversation.members?.length || 0 }} members</small>
      </div>
      <div style="margin-left: auto;">
        <button v-if="conversation.type === 'group'" class="btn-icon" @click="showGroupSettings = true">⚙️</button>
      </div>
    </div>

    <div class="messages-container" ref="messagesContainer">
      <div 
        v-for="msg in messages" 
        :key="msg.id" 
        class="message"
        :class="{ sent: msg.senderId === userId, received: msg.senderId !== userId }"
      >
        <div v-if="msg.forwarded" class="forwarded-label">↪ Forwarded</div>
        <div v-if="msg.replyTo" class="reply-preview">
          Replying to: <span v-if="msg.replyTo.type === 'photo'">📷 </span>{{ msg.replyTo.content || (msg.replyTo.type === 'photo' ? 'Photo' : '') }}
        </div>
        <div v-if="msg.senderId !== userId" class="message-sender">{{ msg.senderUsername }}</div>
        <div class="message-content">
          <span v-if="msg.content && msg.content.length > 0">{{ msg.content }}</span>
          <img v-if="msg.photoUrl" :src="assetURL(msg.photoUrl)" style="max-width: 200px; border-radius: 8px;" @error="$event.target.style.display='none'" />
        </div>
        <div class="message-meta">
          {{ formatTime(msg.timestamp) }}
          <span class="checkmarks" v-if="msg.senderId === userId">
            {{ msg.checkmarks === 2 ? '✓✓' : '✓' }}
          </span>
        </div>
        <div class="reactions" v-if="msg.comments && msg.comments.length > 0">
          <span v-for="c in msg.comments" :key="c.userId" class="reaction" :title="c.username">
            {{ c.comment }} <small class="reaction-author">{{ c.username }}</small>
          </span>
        </div>
        <div class="message-actions">
          <button class="action-btn" @click="setReplyTo(msg)">↩</button>
          <button class="action-btn" @click="toggleReaction(msg)">😀</button>
          <button class="action-btn" @click="forwardMessage(msg)">↪</button>
          <button v-if="msg.senderId === userId" class="action-btn" @click="deleteMessage(msg)">🗑</button>
        </div>
      </div>

      <div v-if="messages.length === 0" class="empty-state">
        <p>No messages yet</p>
        <p style="font-size: 12px;">Send a message to start the conversation</p>
      </div>
    </div>

    <div v-if="replyingTo" style="padding: 10px 20px; background: #f0f0f0; display: flex; justify-content: space-between;">
      <span>Replying to: <span v-if="replyingTo.type === 'photo'">📷 </span>{{ replyingTo.content?.substring(0, 50) || (replyingTo.type === 'photo' ? 'Photo' : '') }}</span>
      <button class="btn-icon" @click="replyingTo = null">✕</button>
    </div>

    <div class="message-input-container">
      <input 
        v-model="newMessage" 
        placeholder="Type a message..." 
        @keyup.enter="sendMessage"
      />
      <input type="file" ref="fileInput" style="display: none;" @change="sendPhoto" accept="image/*" />
      <button class="btn-icon" @click="$refs.fileInput.click()">📷</button>
      <button @click="sendMessage" :disabled="!newMessage.trim()">Send</button>
    </div>

    <!-- Group Settings Modal -->
    <div v-if="showGroupSettings" class="modal-overlay" @click.self="showGroupSettings = false">
      <div class="modal">
        <h3>Group Settings</h3>
        
        <div style="margin-bottom: 15px;">
          <label>Group Name</label>
          <input v-model="groupName" placeholder="Group name" />
          <button class="btn-primary" @click="updateGroupName" style="margin-top: 5px;">Update Name</button>
        </div>

        <div style="margin-bottom: 15px;">
          <label>Add Member</label>
          <input v-model="addMemberQuery" placeholder="Search username..." @input="searchMembers" />
          <div v-if="memberSearchResults.length > 0" class="search-results">
            <div 
              v-for="user in memberSearchResults" 
              :key="user.id" 
              class="search-result-item"
              @click="addMember(user)"
            >
              {{ user.username }}
            </div>
          </div>
        </div>

        <div style="margin-bottom: 15px;">
          <label>Group Photo</label>
          <div v-if="conversation.photoUrl" style="margin: 5px 0;">
            <img :src="assetURL(conversation.photoUrl)" style="width:60px;height:60px;border-radius:50%;object-fit:cover;" />
          </div>
          <input type="file" accept="image/*" @change="updateGroupPhoto" />
        </div>

        <div style="margin-bottom: 15px;">
          <label>Members</label>
          <div v-for="member in conversation.members" :key="member.id" style="padding: 5px 0;">
            {{ member.username }}
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-secondary" @click="leaveGroup" style="margin-right: auto; color: red;">Leave Group</button>
          <button class="btn-secondary" @click="showGroupSettings = false">Close</button>
        </div>
      </div>
    </div>

    <!-- Forward Modal -->
    <div v-if="forwardingMessage" class="modal-overlay" @click.self="forwardingMessage = null">
      <div class="modal">
        <h3>Forward Message</h3>
        <input v-model="forwardQuery" placeholder="Search conversation..." @input="searchForForward" />
        <div v-if="forwardResults.length > 0" class="search-results">
          <div 
            v-for="conv in forwardResults" 
            :key="conv.id" 
            class="search-result-item"
            @click="doForward(conv)"
          >
            {{ conv.name }}
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="forwardingMessage = null">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios, { apiAssetURL } from '../services/axios.js'

export default {
  name: 'ChatView',
  data() {
    return {
      conversation: {},
      messages: [],
      newMessage: '',
      replyingTo: null,
      userId: localStorage.getItem('wasatext_user_id'),
      showGroupSettings: false,
      groupName: '',
      addMemberQuery: '',
      memberSearchResults: [],
      forwardingMessage: null,
      forwardQuery: '',
      forwardResults: [],
      allConversationsForForward: []
    }
  },
  computed: {
    conversationPhotoUrl() {
      if (!this.conversation || !this.conversation.type) return null
      if (this.conversation.type === 'group' && this.conversation.photoUrl) {
        return this.conversation.photoUrl
      }
      if (this.conversation.type === 'private' && this.conversation.members) {
        const other = this.conversation.members.find(m => m.id !== this.userId)
        if (other && other.photoUrl) return other.photoUrl
      }
      return null
    }
  },
  mounted() {
    this.loadConversation()
    this.pollInterval = setInterval(this.loadConversation, 3000)
  },
  beforeUnmount() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval)
    }
  },
  watch: {
    '$route.params.id'() {
      this.loadConversation()
    }
  },
  methods: {
    assetURL: apiAssetURL,
    async loadConversation() {
      const id = this.$route.params.id
      try {
        const response = await axios.get(`/conversations/${id}`)
        this.conversation = response.data
        this.messages = (response.data.messages || []).slice().reverse()
        this.groupName = this.conversation.name
      } catch (err) {
        console.error('Error loading conversation:', err)
      }
    },
    async sendMessage() {
      if (!this.newMessage.trim()) return

      try {
        const payload = {
          type: 'text',
          content: this.newMessage
        }
        if (this.replyingTo) {
          payload.replyToId = this.replyingTo.id
        }

        await axios.post(`/conversations/${this.conversation.id}/messages`, payload)
        this.newMessage = ''
        this.replyingTo = null
        await this.loadConversation()
        this.scrollToBottom()
      } catch (err) {
        console.error('Error sending message:', err)
      }
    },
    async sendPhoto(event) {
      const file = event.target.files[0]
      if (!file) return

      const formData = new FormData()
      formData.append('photo', file)
      formData.append('type', 'photo')
      if (this.newMessage && this.newMessage.trim()) {
        formData.append('content', this.newMessage.trim())
      }
      if (this.replyingTo) {
        formData.append('replyToId', this.replyingTo.id)
      }

      try {
        await axios.post(`/conversations/${this.conversation.id}/messages`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        this.newMessage = ''
        this.replyingTo = null
        await this.loadConversation()
        this.scrollToBottom()
      } catch (err) {
        console.error('Error sending photo:', err)
      }
      this.$refs.fileInput.value = ''
    },
    setReplyTo(msg) {
      this.replyingTo = msg
    },
    async toggleReaction(msg) {
      const existingReaction = msg.comments?.find(c => c.userId === this.userId)
      try {
        if (existingReaction) {
          await axios.delete(`/messages/${msg.id}/comment`)
        } else {
          await axios.put(`/messages/${msg.id}/comment`, { comment: '👍' })
        }
        await this.loadConversation()
      } catch (err) {
        console.error('Error toggling reaction:', err)
      }
    },
    async forwardMessage(msg) {
      this.forwardingMessage = msg
      this.forwardQuery = ''
      try {
        const response = await axios.get(`/users/${this.userId}/conversations`)
        this.allConversationsForForward = (response.data || []).filter(c => c.id !== this.conversation.id)
        this.forwardResults = this.allConversationsForForward
      } catch (err) {
        console.error('Error loading conversations for forward:', err)
      }
    },
    searchForForward() {
      if (!this.allConversationsForForward) {
        this.forwardResults = []
        return
      }
      if (this.forwardQuery.length < 1) {
        this.forwardResults = this.allConversationsForForward
        return
      }
      this.forwardResults = this.allConversationsForForward.filter(c =>
        c.name.toLowerCase().includes(this.forwardQuery.toLowerCase())
      )
    },
    async doForward(conv) {
      try {
        await axios.post(`/conversations/${conv.id}/messages/forward`, {
          messageId: this.forwardingMessage.id
        })
        this.forwardingMessage = null
        this.forwardQuery = ''
        this.forwardResults = []
        alert('Message forwarded!')
      } catch (err) {
        console.error('Error forwarding:', err)
      }
    },
    async deleteMessage(msg) {
      if (!confirm('Delete this message?')) return
      try {
        await axios.delete(`/messages/${msg.id}`)
        await this.loadConversation()
      } catch (err) {
        console.error('Error deleting message:', err)
      }
    },
    async updateGroupName() {
      try {
        await axios.put(`/groups/${this.conversation.id}/name`, { name: this.groupName })
        await this.loadConversation()
      } catch (err) {
        console.error('Error updating group name:', err)
      }
    },
    async searchMembers() {
      if (this.addMemberQuery.length < 1) {
        this.memberSearchResults = []
        return
      }
      try {
        const response = await axios.get(`/users?q=${encodeURIComponent(this.addMemberQuery)}`)
        const memberIds = this.conversation.members?.map(m => m.id) || []
        this.memberSearchResults = (response.data || []).filter(u => !memberIds.includes(u.id))
      } catch (err) {
        console.error('Error searching:', err)
      }
    },
    async addMember(user) {
      try {
        await axios.post(`/groups/${this.conversation.id}/members`, { userId: user.id })
        this.addMemberQuery = ''
        this.memberSearchResults = []
        await this.loadConversation()
      } catch (err) {
        console.error('Error adding member:', err)
      }
    },
    async leaveGroup() {
      if (!confirm('Are you sure you want to leave this group?')) return
      try {
        await axios.delete(`/groups/${this.conversation.id}/members/${this.userId}`)
        this.$router.push('/conversations')
      } catch (err) {
        console.error('Error leaving group:', err)
      }
    },
    async updateGroupPhoto(event) {
      const file = event.target.files[0]
      if (!file) return
      try {
        await axios.put(`/groups/${this.conversation.id}/photo`, file, {
          headers: { 'Content-Type': file.type }
        })
        await this.loadConversation()
      } catch (err) {
        console.error('Error updating group photo:', err)
      }
    },
    goBack() {
      this.$router.push('/conversations')
    },
    getInitial(name) {
      return name ? name.charAt(0).toUpperCase() : '?'
    },
    formatTime(timestamp) {
      if (!timestamp) return ''
      return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    },
    scrollToBottom() {
      this.$nextTick(() => {
        const container = this.$refs.messagesContainer
        if (container) {
          container.scrollTop = container.scrollHeight
        }
      })
    }
  }
}
</script>

<style scoped>
.chat-view {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.reaction-author {
  font-size: 10px;
  color: #888;
  margin-left: 2px;
}
</style>
