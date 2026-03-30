<template>
  <div v-if="error" class="error-banner">
    <div class="error-content">
      <span class="error-icon">⚠️</span>
      <span class="error-message">{{ error }}</span>
      <button class="error-close" @click="$emit('close')">×</button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  error: String
})

defineEmits(['close'])
</script>

<style scoped>
.error-banner {
  background-color: #fee;
  border-left: 4px solid #d32f2f;
  padding: 16px;
  margin-bottom: 16px;
  border-radius: 4px;
}

.error-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.error-icon {
  font-size: 20px;
}

.error-message {
  flex: 1;
  color: #d32f2f;
  font-size: 14px;
}

.error-close {
  background: none;
  border: none;
  font-size: 24px;
  color: #d32f2f;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.error-close:hover {
  opacity: 0.7;
}
</style>
