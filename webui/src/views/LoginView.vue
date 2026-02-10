<template>
  <div class="login-container">
    <div class="login-box">
      <h1>WASAText</h1>
      <p style="margin-bottom: 20px; color: #666;">Enter your username to login</p>
      
      <div v-if="error" class="error-message">{{ error }}</div>
      
      <input 
        v-model="username" 
        type="text" 
        placeholder="Username (3-16 characters)"
        @keyup.enter="login"
        :disabled="loading"
      />
      
      <button @click="login" :disabled="loading || !username">
        {{ loading ? 'Logging in...' : 'Login' }}
      </button>
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js'

export default {
  name: 'LoginView',
  data() {
    return {
      username: '',
      error: '',
      loading: false
    }
  },
  methods: {
    async login() {
      if (!this.username || this.username.length < 3) {
        this.error = 'Username must be at least 3 characters'
        return
      }

      this.loading = true
      this.error = ''

      try {
        const response = await axios.post('/session', { name: this.username })
        const { identifier } = response.data

        localStorage.setItem('wasatext_token', identifier)
        localStorage.setItem('wasatext_user_id', identifier)
        localStorage.setItem('wasatext_username', this.username)

        this.$router.push('/conversations')
      } catch (err) {
        if (err.response && err.response.status === 400) {
          this.error = 'Invalid username format'
        } else {
          this.error = 'Login failed. Please try again.'
        }
      } finally {
        this.loading = false
      }
    }
  }
}
</script>
