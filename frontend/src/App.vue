<template>
  <div id="app" class="app">
    <header class="app-header">
      <h1 class="app-title">Service Dashboard</h1>
      <div class="header-meta">
        <span class="last-updated">Updated: {{ lastUpdated }}</span>
        <button @click="refreshServices" class="refresh-button" :disabled="loading">
          🔄 Refresh
        </button>
      </div>
    </header>

    <main class="app-main">
      <ErrorBanner
        :error="error"
        @close="error = null"
      />

      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p>Loading services...</p>
      </div>

      <div v-else-if="services.length === 0" class="empty">
        <p>No services found</p>
      </div>

      <ServiceGrid v-else :services="services" />
    </main>

    <footer class="app-footer">
      <p>{{ services.length }} services • Last updated: {{ fullTimestamp }}</p>
    </footer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import ErrorBanner from './components/ErrorBanner.vue'
import ServiceGrid from './components/ServiceGrid.vue'
import { fetchServices } from './api'

const services = ref([])
const loading = ref(false)
const error = ref(null)
const lastFetchTime = ref(null)

const lastUpdated = computed(() => {
  if (!lastFetchTime.value) return 'Never'
  return new Date(lastFetchTime.value * 1000).toLocaleTimeString()
})

const fullTimestamp = computed(() => {
  if (!lastFetchTime.value) return 'Never'
  return new Date(lastFetchTime.value * 1000).toLocaleString()
})

const refreshServices = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await fetchServices()
    services.value = response.services || []
    lastFetchTime.value = response.timestamp
  } catch (err) {
    error.value = `Failed to load services: ${err.message}`
    services.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refreshServices()
  // Auto-refresh every 30 seconds
  setInterval(refreshServices, 30000)
})
</script>

<style scoped>
#app {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  display: flex;
  flex-direction: column;
}

.app-header {
  background: white;
  border-bottom: 1px solid #e0e0e0;
  padding: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.app-title {
  margin: 0 0 12px 0;
  font-size: 32px;
  font-weight: 700;
  color: #222;
}

.header-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.last-updated {
  font-size: 12px;
  color: #999;
}

.refresh-button {
  background: #667eea;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.refresh-button:hover:not(:disabled) {
  background: #5568d3;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.4);
}

.refresh-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.app-main {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.loading,
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  color: #999;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e0e0e0;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.app-footer {
  background: white;
  border-top: 1px solid #e0e0e0;
  padding: 16px 24px;
  text-align: center;
  font-size: 12px;
  color: #999;
  margin-top: auto;
}

.app-footer p {
  margin: 0;
}

@media (max-width: 768px) {
  #app {
    padding: 0;
  }

  .app-header {
    padding: 16px;
  }

  .app-title {
    font-size: 24px;
  }

  .app-main {
    padding: 16px;
  }
}
</style>
