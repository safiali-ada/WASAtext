<template>
  <div v-if="msg" class="error-banner" role="alert">
    <span>{{ msg }}</span>
    <button class="btn-icon" @click="$emit('close')">✕</button>
  </div>
</template>

<script>
export default {
  name: 'ErrorMsg',
  props: {
    msg: { type: String, default: '' }
  },
  emits: ['close']
}
</script>
