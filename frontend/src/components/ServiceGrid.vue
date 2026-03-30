<template>
  <div class="grid">
    <ServiceCard
      v-for="service in services"
      :key="service.id"
      :service="service"
      :icon="getIcon(service.name)"
    />
  </div>
</template>

<script setup>
import ServiceCard from './ServiceCard.vue'
import { getServiceIcon } from '../utils/icons'

defineProps({
  services: Array
})

const getIcon = (serviceName) => {
  return getServiceIcon(serviceName)
}
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

@media (max-width: 1200px) {
  .grid {
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  }
}

@media (max-width: 768px) {
  .grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  }
}

@media (max-width: 480px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
