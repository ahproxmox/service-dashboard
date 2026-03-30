export const fetchServices = async () => {
  const response = await fetch('/api/services')
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  return response.json()
}
