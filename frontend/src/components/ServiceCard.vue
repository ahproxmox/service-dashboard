<template>
  <div class="service-card" :class="{ 'service-stopped': service.status === 'stopped' }">
    <div class="service-header">
      <div class="service-icon-container">
        <img
          v-if="icon"
          :src="icon"
          :alt="service.name"
          class="service-icon"
        />
        <div v-else class="service-icon-placeholder">{{ initials }}</div>
        <div
          class="status-indicator"
          :class="service.status"
          :title="`Status: ${service.status}`"
        ></div>
      </div>
      <div class="service-info">
        <h3 class="service-name">{{ service.name }}</h3>
        <p class="service-status">{{ capitalizeStatus(service.status) }}</p>
      </div>
    </div>

    <div v-if="service.httpsUrl" class="service-url">
      <a :href="service.httpsUrl" target="_blank" rel="noopener noreferrer">
        {{ service.httpsUrl }}
      </a>
    </div>

    <div v-if="service.metrics && service.status === 'running'" class="metrics">
      <div class="metric-row">
        <span class="metric-label">CPU:</span>
        <span class="metric-value">{{ service.metrics.cpuPercent?.toFixed(1) || 0 }}%</span>
      </div>
      <div class="metric-row">
        <span class="metric-label">RAM:</span>
        <span class="metric-value">{{ service.metrics.ramPercent?.toFixed(1) || 0 }}%</span>
      </div>
      <div class="metric-row">
        <span class="metric-label">Disk:</span>
        <span class="metric-value">{{ service.metrics.diskPercent?.toFixed(1) || 0 }}%</span>
      </div>
      <div class="metric-row">
        <span class="metric-label">Network:</span>
        <span class="metric-value">
          ↓{{ service.metrics.networkInMbps?.toFixed(1) || 0 }} ↑{{ service.metrics.networkOutMbps?.toFixed(1) || 0 }} Mbps
        </span>
      </div>
    </div>

    <div v-else-if="service.status === 'stopped'" class="metrics-placeholder">
      Container is stopped
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  service: Object,
  icon: String
})

const initials = computed(() => {
  return props.service.name
    .split(/[\s-]/)
    .map(word => word[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)
})

const capitalizeStatus = (status) => {
  return status.charAt(0).toUpperCase() + status.slice(1)
}
</script>

<style scoped>
.service-card {
  background: var(--color-background-surface);
  border: 1px solid var(--color-border-default);
  border-radius: 12px;
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-md);
  transition: all var(--transition-fast);
  min-height: 200px;
  display: flex;
  flex-direction: column;
}

.service-card:hover {
  box-shadow: var(--shadow-lg);
  border-color: #3e4460;
  transform: translateY(-2px);
}

.service-card.service-stopped {
  opacity: 0.6;
  background-color: var(--color-background-surface-dim);
}

.service-header {
  display: flex;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
}

.service-icon-container {
  position: relative;
  width: 64px;
  height: 64px;
  flex-shrink: 0;
}

.service-icon {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.service-icon-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-gradient-end) 100%);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-weight: var(--font-weight-bold);
  font-size: var(--font-size-lg);
}

.status-indicator {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 20px;
  height: 20px;
  border-radius: var(--radius-full);
  border: 3px solid var(--color-background-surface);
  box-shadow: var(--shadow-md);
}

.status-indicator.running {
  background-color: var(--color-status-running);
  box-shadow: 0 0 8px var(--color-status-running-glow), var(--shadow-md);
}

.status-indicator.stopped {
  background-color: var(--color-status-stopped);
}

.service-info {
  flex: 1;
}

.service-name {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.service-status {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.service-url {
  margin-bottom: var(--spacing-md);
}

.service-url a {
  color: var(--color-primary);
  text-decoration: none;
  font-size: var(--font-size-xs);
  word-break: break-all;
}

.service-url a:hover {
  text-decoration: underline;
}

.metrics {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.metric-row {
  display: flex;
  justify-content: space-between;
  font-size: var(--font-size-xs);
}

.metric-label {
  color: var(--color-text-secondary);
  font-weight: var(--font-weight-medium);
}

.metric-value {
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-weight: var(--font-weight-semibold);
}

.metrics-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  font-style: italic;
}
</style>
