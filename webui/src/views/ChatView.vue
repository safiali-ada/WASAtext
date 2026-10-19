<template>
  <div class="chat-view">
    <div class="chat-header">
      <button class="btn-icon back-btn" @click="goBack" title="Back">←</button>
      <UserAvatar :url="conversationPhotoUrl" :name="conversation.name" :size="40" class="header-avatar" />
      <div class="chat-title">
        <h3>{{ conversation.name }}</h3>
        <small v-if="conversation.type === 'group'">{{ memberNames }}</small>
      </div>
      <div style="margin-left: auto;">
        <button v-if="conversation.type === 'group'" class="btn-icon" @click="openGroupSettings" title="Group settings">⚙️</button>
      </div>
    </div>

    <ErrorMsg :msg="error" @close="error = ''" />
    <div v-if="notice" class="notice-banner">{{ notice }}</div>

    <div class="messages-container" ref="messagesContainer" @scroll="onScroll">
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="message-row"
        :class="{ mine: msg.senderId === userId }"
      >
        <UserAvatar
          v-if="msg.senderId !== userId && conversation.type === 'group'"
          :url="memberPhoto(msg.senderId)"
          :name="msg.senderUsername"
          :size="28"
          class="sender-avatar"
        />
        <div class="message" :class="{ sent: msg.senderId === userId, received: msg.senderId !== userId }">
          <div v-if="msg.senderId !== userId" class="message-sender">{{ msg.senderUsername }}</div>
          <div v-if="msg.forwarded" class="forwarded-label">↪ Forwarded</div>
          <div v-if="msg.replyTo" class="reply-preview">
            <div class="reply-author">↩ {{ replyAuthor(msg.replyTo) }}</div>
            <span v-if="msg.replyTo.type === 'photo'">📷 </span>{{ snippet(msg.replyTo.content) || (msg.replyTo.type === 'photo' ? 'Photo' : '') }}
          </div>
          <div class="message-content">
            <img v-if="msg.photoUrl" :src="assetURL(msg.photoUrl)" class="message-photo" alt="photo" @load="stickToBottom && scrollToBottom()" />
            <div v-if="msg.content">{{ msg.content }}</div>
          </div>
          <div class="message-meta">
            {{ formatTime(msg.timestamp) }}
            <span
              v-if="msg.senderId === userId"
              class="checkmarks"
              :class="{ read: msg.checkmarks === 2 }"
              :title="msg.checkmarks === 2 ? 'Read' : 'Received'"
            >{{ msg.checkmarks === 2 ? '✓✓' : '✓' }}</span>
          </div>
          <div class="reactions" v-if="msg.comments && msg.comments.length > 0">
            <span
              v-for="c in msg.comments"
              :key="c.userId"
              class="reaction"
              :class="{ mine: c.userId === userId }"
              :title="c.userId === userId ? 'Your reaction (tap to remove)' : c.username + ' reacted'"
              @click="c.userId === userId && removeReaction(msg)"
            >
              {{ c.comment }} <small class="reaction-author">{{ c.userId === userId ? 'You' : c.username }}</small>
            </span>
          </div>
          <div class="message-actions">
            <button class="action-btn" @click="setReplyTo(msg)" title="Reply">↩ Reply</button>
            <button class="action-btn" @click="toggleEmojiPicker(msg)" title="React">😀 React</button>
            <button class="action-btn" @click="openForward(msg)" title="Forward">↪ Forward</button>
            <button v-if="msg.senderId === userId" class="action-btn" @click="deleteMessage(msg)" title="Delete">🗑</button>
          </div>
          <div v-if="emojiPickerFor === msg.id" class="emoji-picker">
            <button
              v-for="e in emojis"
              :key="e"
              class="emoji-btn"
              :class="{ selected: myReaction(msg) === e }"
              @click="react(msg, e)"
            >{{ e }}</button>
            <button v-if="myReaction(msg)" class="action-btn" @click="removeReaction(msg)">Remove</button>
          </div>
        </div>
      </div>

      <div v-if="loaded && messages.length === 0" class="empty-state">
        <p>No messages yet</p>
        <p style="font-size: 12px;">Send a message to start the conversation</p>
      </div>
    </div>

    <div v-if="replyingTo" class="composer-banner">
      <span>
        Replying to <b>{{ replyingTo.senderId === userId ? 'yourself' : replyingTo.senderUsername }}</b>:
        <span v-if="replyingTo.type === 'photo'">📷 </span>{{ snippet(replyingTo.content, 50) || (replyingTo.type === 'photo' ? 'Photo' : '') }}
      </span>
      <button class="btn-icon" @click="replyingTo = null">✕</button>
    </div>

    <div v-if="attachment" class="composer-banner">
      <span class="attachment-preview">
        <img :src="attachmentPreview" alt="attachment" />
        {{ attachment.name }}
      </span>
      <button class="btn-icon" @click="clearAttachment">✕</button>
    </div>

    <div class="message-input-container">
      <input type="file" ref="fileInput" style="display: none;" @change="onFileSelected" accept="image/*" />
      <button class="btn-icon attach-btn" @click="$refs.fileInput.click()" title="Attach image">📷</button>
      <input
        v-model="newMessage"
        :placeholder="attachment ? 'Add a caption...' : 'Type a message...'"
        @keyup.enter="sendMessage"
      />
      <button @click="sendMessage" :disabled="sending || (!newMessage.trim() && !attachment)">Send</button>
    </div>

    <!-- Group Settings Modal -->
    <div v-if="showGroupSettings" class="modal-overlay" @click.self="showGroupSettings = false">
      <div class="modal">
        <h3>Group Settings</h3>

        <div class="modal-section">
          <label class="field-label">Group Photo</label>
          <div class="group-photo-row">
            <UserAvatar :url="conversation.photoUrl" :name="conversation.name" :size="60" />
            <input type="file" accept="image/*" @change="updateGroupPhoto" />
          </div>
        </div>

        <div class="modal-section">
          <label class="field-label">Group Name</label>
          <input v-model="groupName" placeholder="Group name" maxlength="50" />
          <button class="btn-primary" @click="updateGroupName" :disabled="!groupName.trim()">Update Name</button>
        </div>

        <div class="modal-section">
          <label class="field-label">Add Member</label>
          <input v-model="addMemberQuery" placeholder="Search username..." @input="searchMembers" />
          <div v-if="memberSearchResults.length > 0" class="search-results">
            <div
              v-for="user in memberSearchResults"
              :key="user.id"
              class="search-result-item user-row"
              @click="addMember(user)"
            >
              <UserAvatar :url="user.photoUrl" :name="user.username" :size="28" />
              <span>{{ user.username }}</span>
            </div>
          </div>
        </div>

        <div class="modal-section">
          <label class="field-label">Members ({{ conversation.members?.length || 0 }})</label>
          <div class="modal-list">
            <div v-for="member in conversation.members" :key="member.id" class="user-row member-row">
              <UserAvatar :url="member.photoUrl" :name="member.username" :size="28" />
              <span>{{ member.username }}<span v-if="member.id === userId"> (you)</span></span>
            </div>
          </div>
        </div>

        <div v-if="settingsMessage" class="success-text">{{ settingsMessage }}</div>
        <ErrorMsg :msg="settingsError" @close="settingsError = ''" />

        <div class="modal-actions">
          <button class="btn-danger" @click="leaveGroup" style="margin-right: auto;">Leave Group</button>
          <button class="btn-secondary" @click="showGroupSettings = false">Close</button>
        </div>
      </div>
    </div>

    <!-- Forward Modal -->
    <div v-if="forwardingMessage" class="modal-overlay" @click.self="forwardingMessage = null">
      <div class="modal">
        <h3>Forward Message</h3>
        <div class="reply-preview forward-preview">
          <span v-if="forwardingMessage.type === 'photo'">📷 </span>{{ snippet(forwardingMessage.content, 50) || 'Photo' }}
        </div>
        <input v-model="forwardQuery" placeholder="Search chats or users..." />
        <div class="search-results modal-list">
          <div class="list-heading" v-if="forwardChats.length > 0">Chats</div>
          <div
            v-for="conv in forwardChats"
            :key="'c-' + conv.id"
            class="search-result-item user-row"
            @click="forwardToConversation(conv.id)"
          >
            <UserAvatar :url="conv.photoUrl" :name="conv.name" :size="28" />
            <span><span v-if="conv.type === 'group'">👥 </span>{{ conv.name }}</span>
          </div>
          <div class="list-heading" v-if="forwardUsers.length > 0">Other users</div>
          <div
            v-for="user in forwardUsers"
            :key="'u-' + user.id"
            class="search-result-item user-row"
            @click="forwardToUser(user)"
          >
            <UserAvatar :url="user.photoUrl" :name="user.username" :size="28" />
            <span>{{ user.username }}</span>
          </div>
          <div v-if="forwardChats.length === 0 && forwardUsers.length === 0" class="search-empty">No matches</div>
        </div>
        <ErrorMsg :msg="forwardError" @close="forwardError = ''" />
        <div class="modal-actions">
          <button class="btn-secondary" @click="forwardingMessage = null">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios, { apiAssetURL } from '../services/axios.js'
import ErrorMsg from '../components/ErrorMsg.vue'
import UserAvatar from '../components/UserAvatar.vue'
import { formatTime, snippet, errorText } from '../services/format.js'

const REFRESH_ERROR = 'Could not refresh the conversation. Retrying...'

export default {
  name: 'ChatView',
  components: { ErrorMsg, UserAvatar },
  emits: ['changed'],
  data() {
    return {
      conversation: {},
      messages: [],
      loaded: false,
      error: '',
      notice: '',
      newMessage: '',
      attachment: null,
      attachmentPreview: '',
      sending: false,
      replyingTo: null,
      userId: localStorage.getItem('wasatext_user_id'),
      emojis: ['👍', '❤️', '😂', '😮', '😢', '🙏'],
      emojiPickerFor: null,
      showGroupSettings: false,
      groupName: '',
      addMemberQuery: '',
      memberSearchResults: [],
      settingsMessage: '',
      settingsError: '',
      forwardingMessage: null,
      forwardQuery: '',
      forwardConversations: [],
      forwardAllUsers: [],
      forwardError: '',
      stickToBottom: true
    }
  },
  computed: {
    conversationPhotoUrl() {
      if (!this.conversation || !this.conversation.type) return ''
      if (this.conversation.type === 'group') {
        return this.conversation.photoUrl || ''
      }
      const other = (this.conversation.members || []).find(m => m.id !== this.userId)
      return other?.photoUrl || ''
    },
    memberNames() {
      return (this.conversation.members || [])
        .map(m => (m.id === this.userId ? 'You' : m.username))
        .join(', ')
    },
    forwardChats() {
      const q = this.forwardQuery.toLowerCase()
      return this.forwardConversations.filter(c => c.id !== this.conversation.id && c.name.toLowerCase().includes(q))
    },
    forwardUsers() {
      // Users without an existing private chat: forwarding creates the conversation
      const q = this.forwardQuery.toLowerCase()
      const privateNames = new Set(this.forwardConversations.filter(c => c.type === 'private').map(c => c.name))
      return this.forwardAllUsers.filter(u => !privateNames.has(u.username) && u.username.toLowerCase().includes(q))
    }
  },
  mounted() {
    this.loadConversation()
    this.pollInterval = setInterval(this.loadConversation, 2000)
  },
  beforeUnmount() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval)
    }
    if (this.attachmentPreview) {
      URL.revokeObjectURL(this.attachmentPreview)
    }
  },
  methods: {
    assetURL: apiAssetURL,
    formatTime,
    snippet,
    async loadConversation() {
      const id = this.$route.params.id
      if (!id) return
      try {
        const response = await axios.get(`/conversations/${id}`)
        const previousLast = this.messages.length ? this.messages[this.messages.length - 1].id : null
        this.conversation = response.data
        // The API returns messages newest-first; the chat shows them oldest at the top.
        this.messages = (response.data.messages || []).slice().reverse()
        if (!this.showGroupSettings) {
          this.groupName = this.conversation.name
        }
        const newLast = this.messages.length ? this.messages[this.messages.length - 1].id : null
        if (!this.loaded || (newLast !== previousLast && this.stickToBottom)) {
          this.scrollToBottom()
        }
        this.loaded = true
        if (this.error === REFRESH_ERROR) this.error = ''
      } catch (err) {
        const status = err.response?.status
        if (status === 403 || status === 404) {
          // No longer a member (e.g. left the group): stop showing its content.
          clearInterval(this.pollInterval)
          this.messages = []
          this.$emit('changed')
          this.$router.replace('/conversations')
        } else if (status !== 401) {
          this.error = REFRESH_ERROR
        }
      }
    },
    onScroll() {
      const c = this.$refs.messagesContainer
      if (c) {
        this.stickToBottom = c.scrollHeight - c.scrollTop - c.clientHeight < 80
      }
    },
    onFileSelected(event) {
      const file = event.target.files[0]
      event.target.value = ''
      if (!file) return
      if (!file.type.startsWith('image/')) {
        this.error = 'Only image files can be sent'
        return
      }
      if (file.size > 5 * 1024 * 1024) {
        this.error = 'Image too large (max 5 MB)'
        return
      }
      this.clearAttachment()
      this.attachment = file
      this.attachmentPreview = URL.createObjectURL(file)
    },
    clearAttachment() {
      if (this.attachmentPreview) {
        URL.revokeObjectURL(this.attachmentPreview)
      }
      this.attachment = null
      this.attachmentPreview = ''
    },
    async sendMessage() {
      const text = this.newMessage.trim()
      if ((!text && !this.attachment) || this.sending) return
      this.sending = true
      this.error = ''
      try {
        if (this.attachment) {
          // Text and image travel together in a single message.
          const formData = new FormData()
          formData.append('type', 'photo')
          formData.append('photo', this.attachment)
          if (text) formData.append('content', text)
          if (this.replyingTo) formData.append('replyToId', this.replyingTo.id)
          await axios.post(`/conversations/${this.conversation.id}/messages`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
          })
        } else {
          const payload = { type: 'text', content: text }
          if (this.replyingTo) payload.replyToId = this.replyingTo.id
          await axios.post(`/conversations/${this.conversation.id}/messages`, payload)
        }
        this.newMessage = ''
        this.replyingTo = null
        this.clearAttachment()
        this.stickToBottom = true
        await this.loadConversation()
        this.scrollToBottom()
        this.$emit('changed')
      } catch (err) {
        this.error = errorText(err, 'Message could not be sent')
      } finally {
        this.sending = false
      }
    },
    setReplyTo(msg) {
      this.replyingTo = msg
    },
    replyAuthor(reply) {
      if (reply.senderId === this.userId) return 'You'
      return reply.senderUsername || 'Message'
    },
    myReaction(msg) {
      return msg.comments?.find(c => c.userId === this.userId)?.comment || null
    },
    toggleEmojiPicker(msg) {
      this.emojiPickerFor = this.emojiPickerFor === msg.id ? null : msg.id
    },
    async react(msg, emoji) {
      this.emojiPickerFor = null
      try {
        if (this.myReaction(msg) === emoji) {
          await axios.delete(`/messages/${msg.id}/comment`)
        } else {
          // One reaction per user: a new emoji replaces the previous one.
          await axios.put(`/messages/${msg.id}/comment`, { comment: emoji })
        }
        await this.loadConversation()
      } catch (err) {
        this.error = errorText(err, 'Reaction failed')
      }
    },
    async removeReaction(msg) {
      this.emojiPickerFor = null
      try {
        await axios.delete(`/messages/${msg.id}/comment`)
        await this.loadConversation()
      } catch (err) {
        this.error = errorText(err, 'Could not remove the reaction')
      }
    },
    async openForward(msg) {
      this.forwardingMessage = msg
      this.forwardQuery = ''
      this.forwardError = ''
      try {
        const [convs, users] = await Promise.all([
          axios.get(`/users/${this.userId}/conversations`),
          axios.get('/users', { params: { q: '' } })
        ])
        this.forwardConversations = convs.data || []
        this.forwardAllUsers = (users.data || []).filter(u => u.id !== this.userId)
      } catch (err) {
        this.forwardError = errorText(err, 'Could not load recipients')
      }
    },
    async forwardToConversation(conversationId) {
      try {
        await axios.post(`/conversations/${conversationId}/messages/forward`, {
          messageId: this.forwardingMessage.id
        })
        this.forwardingMessage = null
        this.showNotice('Message forwarded')
        this.$emit('changed')
      } catch (err) {
        this.forwardError = errorText(err, 'Forward failed')
      }
    },
    async forwardToUser(user) {
      try {
        const response = await axios.post('/conversations', { userId: user.id })
        await this.forwardToConversation(response.data.id)
      } catch (err) {
        this.forwardError = errorText(err, 'Forward failed')
      }
    },
    async deleteMessage(msg) {
      if (!confirm('Delete this message?')) return
      try {
        await axios.delete(`/messages/${msg.id}`)
        await this.loadConversation()
        this.$emit('changed')
      } catch (err) {
        this.error = errorText(err, 'Could not delete the message')
      }
    },
    openGroupSettings() {
      this.groupName = this.conversation.name
      this.addMemberQuery = ''
      this.memberSearchResults = []
      this.settingsMessage = ''
      this.settingsError = ''
      this.showGroupSettings = true
    },
    async updateGroupName() {
      this.settingsMessage = ''
      this.settingsError = ''
      try {
        await axios.put(`/groups/${this.conversation.id}/name`, { name: this.groupName.trim() })
        await this.loadConversation()
        this.settingsMessage = 'Group name updated'
        this.$emit('changed')
      } catch (err) {
        this.settingsError = errorText(err, 'Could not update the group name')
      }
    },
    async searchMembers() {
      if (this.addMemberQuery.length < 1) {
        this.memberSearchResults = []
        return
      }
      try {
        const response = await axios.get('/users', { params: { q: this.addMemberQuery } })
        const memberIds = this.conversation.members?.map(m => m.id) || []
        this.memberSearchResults = (response.data || []).filter(u => !memberIds.includes(u.id))
      } catch (err) {
        this.settingsError = errorText(err, 'User search failed')
      }
    },
    async addMember(user) {
      this.settingsMessage = ''
      this.settingsError = ''
      try {
        await axios.post(`/groups/${this.conversation.id}/members`, { userId: user.id })
        this.addMemberQuery = ''
        this.memberSearchResults = []
        await this.loadConversation()
        this.settingsMessage = `${user.username} added to the group`
      } catch (err) {
        this.settingsError = errorText(err, 'Could not add the member')
      }
    },
    async leaveGroup() {
      if (!confirm('Are you sure you want to leave this group?')) return
      try {
        await axios.delete(`/groups/${this.conversation.id}/members/${this.userId}`)
        clearInterval(this.pollInterval)
        this.showGroupSettings = false
        this.$emit('changed')
        this.$router.push('/conversations')
      } catch (err) {
        this.settingsError = errorText(err, 'Could not leave the group')
      }
    },
    async updateGroupPhoto(event) {
      const file = event.target.files[0]
      if (!file) return
      this.settingsMessage = ''
      this.settingsError = ''
      try {
        await axios.put(`/groups/${this.conversation.id}/photo`, file, {
          headers: { 'Content-Type': file.type || 'application/octet-stream' }
        })
        await this.loadConversation()
        this.settingsMessage = 'Group photo updated'
        this.$emit('changed')
      } catch (err) {
        this.settingsError = errorText(err, 'Could not update the group photo')
      }
      event.target.value = ''
    },
    memberPhoto(userId) {
      return (this.conversation.members || []).find(m => m.id === userId)?.photoUrl || ''
    },
    showNotice(text) {
      this.notice = text
      setTimeout(() => { this.notice = '' }, 2500)
    },
    goBack() {
      this.$router.push('/conversations')
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
  min-height: 0;
}
</style>
