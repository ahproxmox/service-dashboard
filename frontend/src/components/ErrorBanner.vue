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
  background-color: #2e1a1a;
  border: 1px solid #5a2e2e;
  border-left: 4px solid #f87171;
  padding: 12px 16px;
  margin-bottom: 16px;
  border-radius: 8px;
}

.error-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.error-icon {
  font-size: 18px;
}

.error-message {
  flex: 1;
  color: #f87171;
  font-size: 14px;
}

.error-close {
  background: none;
  border: none;
  font-size: 20px;
  color: #f87171;
  cursor: pointer;
  padding: 0;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  font-family: var(--font-family-sans);
}

.error-close:hover {
  opacity: 0.7;
}
</style>
