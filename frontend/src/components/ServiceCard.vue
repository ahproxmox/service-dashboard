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
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
  min-height: 200px;
  display: flex;
  flex-direction: column;
}

.service-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.service-card.service-stopped {
  opacity: 0.7;
  background-color: #fafafa;
}

.service-header {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 24px;
}

.status-indicator {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 3px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.status-indicator.running {
  background-color: #4caf50;
  box-shadow: 0 0 8px rgba(76, 175, 80, 0.5), 0 2px 4px rgba(0, 0, 0, 0.2);
}

.status-indicator.stopped {
  background-color: #999;
}

.service-info {
  flex: 1;
}

.service-name {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: #222;
}

.service-status {
  margin: 0;
  font-size: 12px;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.service-url {
  margin-bottom: 16px;
}

.service-url a {
  color: #667eea;
  text-decoration: none;
  font-size: 12px;
  word-break: break-all;
}

.service-url a:hover {
  text-decoration: underline;
}

.metrics {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.metric-label {
  color: #666;
  font-weight: 500;
}

.metric-value {
  color: #222;
  font-family: monospace;
  font-weight: 600;
}

.metrics-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 14px;
  font-style: italic;
}
</style>
