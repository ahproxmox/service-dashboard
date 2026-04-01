<template>
  <div id="app" class="app">
    <header class="app-header">
      <h1 class="app-title">Service Dashboard</h1>
      <div class="header-meta">
        <span class="last-updated">{{ lastUpdated }}</span>
        <button @click="refreshServices" class="refresh-button" :disabled="loading">
          Refresh
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
  background: var(--color-background-page-start);
  display: flex;
  flex-direction: column;
}

.app-header {
  background: var(--color-background-surface);
  border-bottom: 1px solid var(--color-border-default);
  padding: 0.75rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  justify-content: space-between;
}

.app-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
}

.header-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.last-updated {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.refresh-button {
  background: none;
  color: var(--color-text-muted);
  border: 1px solid var(--color-border-default);
  padding: 0.3rem 0.75rem;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 0.8rem;
  font-weight: var(--font-weight-medium);
  font-family: var(--font-family-sans);
  transition: all var(--transition-fast);
}

.refresh-button:hover:not(:disabled) {
  color: var(--color-text-primary);
  border-color: var(--color-text-muted);
}

.refresh-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.app-main {
  flex: 1;
  padding: 1.5rem;
  overflow-y: auto;
}

.loading,
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  color: var(--color-text-muted);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--color-border-default);
  border-top-color: var(--color-primary);
  border-radius: var(--radius-full);
  animation: spin 1s linear infinite;
  margin-bottom: var(--spacing-md);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.app-footer {
  background: var(--color-background-surface);
  border-top: 1px solid var(--color-border-default);
  padding: var(--spacing-md) var(--spacing-xl);
  text-align: center;
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin-top: auto;
}

.app-footer p {
  margin: 0;
}

@media (max-width: 768px) {
  .app-header {
    padding: 0.75rem 1rem;
  }

  .app-main {
    padding: 1rem;
  }
}
</style>
