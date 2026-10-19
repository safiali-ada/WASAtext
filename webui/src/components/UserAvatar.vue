<template>
  <div class="avatar" :style="{ width: size + 'px', height: size + 'px', fontSize: Math.round(size / 2.4) + 'px' }">
    <img v-if="url && !broken" :src="assetURL(url)" alt="" @error="broken = true" />
    <span v-else>{{ initial }}</span>
  </div>
</template>

<script>
import { apiAssetURL } from '../services/axios.js'

export default {
  name: 'UserAvatar',
  props: {
    url: { type: String, default: '' },
    name: { type: String, default: '' },
    size: { type: Number, default: 40 }
  },
  data() {
    return { broken: false }
  },
  computed: {
    initial() {
      return this.name ? this.name.charAt(0).toUpperCase() : '?'
    }
  },
  watch: {
    url() {
      this.broken = false
    }
  },
  methods: {
    assetURL: apiAssetURL
  }
}
</script>
