<template>
  <div class="profile-container">
    <button class="btn-secondary" @click="goBack" style="margin-bottom: 20px;">← Back to Conversations</button>
    
    <h2>Profile Settings</h2>

    <div class="profile-section">
      <h3>Username</h3>
      <input v-model="username" placeholder="New username" />
      <button @click="updateUsername" :disabled="!username || loading">
        {{ loading ? 'Updating...' : 'Update Username' }}
      </button>
      <p v-if="usernameError" style="color: red; margin-top: 10px;">{{ usernameError }}</p>
      <p v-if="usernameSuccess" style="color: green; margin-top: 10px;">{{ usernameSuccess }}</p>
    </div>

    <div class="profile-section">
      <h3>Profile Photo</h3>
      <input type="file" ref="photoInput" @change="updatePhoto" accept="image/*" />
      <p v-if="photoError" style="color: red; margin-top: 10px;">{{ photoError }}</p>
      <p v-if="photoSuccess" style="color: green; margin-top: 10px;">{{ photoSuccess }}</p>
    </div>

    <div class="profile-section">
      <h3>Logout</h3>
      <button @click="logout" style="background: #e74c3c;">Logout</button>
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js'

export default {
  name: 'ProfileView',
  data() {
    return {
      username: localStorage.getItem('wasatext_username') || '',
      loading: false,
      usernameError: '',
      usernameSuccess: '',
      photoError: '',
      photoSuccess: ''
    }
  },
  methods: {
    async updateUsername() {
      this.loading = true
      this.usernameError = ''
      this.usernameSuccess = ''

      try {
        const userId = localStorage.getItem('wasatext_user_id')
        await axios.put(`/users/${userId}/username`, { username: this.username })
        localStorage.setItem('wasatext_username', this.username)
        this.usernameSuccess = 'Username updated successfully!'
      } catch (err) {
        if (err.response?.status === 409) {
          this.usernameError = 'Username already taken'
        } else if (err.response?.status === 400) {
          this.usernameError = 'Invalid username format (3-16 alphanumeric characters)'
        } else {
          this.usernameError = 'Failed to update username'
        }
      } finally {
        this.loading = false
      }
    },
    async updatePhoto(event) {
      const file = event.target.files[0]
      if (!file) return

      this.photoError = ''
      this.photoSuccess = ''

      try {
        const userId = localStorage.getItem('wasatext_user_id')
        await axios.put(`/users/${userId}/photo`, file, {
          headers: { 'Content-Type': file.type }
        })
        this.photoSuccess = 'Photo updated successfully!'
      } catch (err) {
        this.photoError = 'Failed to update photo'
      }
    },
    logout() {
      localStorage.removeItem('wasatext_token')
      localStorage.removeItem('wasatext_user_id')
      localStorage.removeItem('wasatext_username')
      this.$router.push('/')
    },
    goBack() {
      this.$router.push('/conversations')
    }
  }
}
</script>
