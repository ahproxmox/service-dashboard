/**
 * Format CPU percentage for display
 */
export const formatCpu = (value) => {
  if (!value && value !== 0) return '—'
  return `${value.toFixed(1)}%`
}

/**
 * Format RAM usage for display
 */
export const formatRam = (value, total = null) => {
  if (!value && value !== 0) return '—'
  if (total) {
    return `${value.toFixed(1)}% (${(total / 1024).toFixed(0)} GB)`
  }
  return `${value.toFixed(1)}%`
}

/**
 * Format disk usage for display
 */
export const formatDisk = (value) => {
  if (!value && value !== 0) return '—'
  return `${value.toFixed(1)}%`
}

/**
 * Format network speed for display
 */
export const formatNetwork = (inMbps, outMbps) => {
  const inStr = inMbps ? `${inMbps.toFixed(2)} Mbps` : '—'
  const outStr = outMbps ? `${outMbps.toFixed(2)} Mbps` : '—'
  return { inStr, outStr }
}

/**
 * Get metric severity level (0 = green, 1 = yellow, 2 = red)
 */
export const getMetricSeverity = (type, value) => {
  if (value === null || value === undefined) return -1

  switch (type) {
    case 'cpu':
      if (value < 50) return 0
      if (value < 80) return 1
      return 2
    case 'ram':
      if (value < 50) return 0
      if (value < 80) return 1
      return 2
    case 'disk':
      if (value < 70) return 0
      if (value < 85) return 1
      return 2
    default:
      return 0
  }
}

/**
 * Format metric color based on severity
 */
export const getMetricColor = (severity) => {
  switch (severity) {
    case 0: return '#4caf50' // green
    case 1: return '#ff9800' // orange
    case 2: return '#d32f2f' // red
    default: return '#999'
  }
}
