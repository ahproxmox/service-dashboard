import servicesConfig from '../../public/config/services-config.json'

export const getServiceIcon = (serviceName) => {
  const config = servicesConfig.services[serviceName.toLowerCase().replace(/_/g, '-')]
  return config?.icon || null
}

export const getServiceColor = (serviceName) => {
  const config = servicesConfig.services[serviceName.toLowerCase().replace(/_/g, '-')]
  return config?.color || '#667eea'
}

export const getServiceDisplayName = (serviceName) => {
  const config = servicesConfig.services[serviceName.toLowerCase().replace(/_/g, '-')]
  return config?.displayName || serviceName
}

export const getThreshold = (metric, level) => {
  return servicesConfig.thresholds[metric]?.[level] || null
}
