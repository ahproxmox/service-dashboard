import axios from 'axios'

const API_BASE_URL = '/api'

const client = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000
})

export const fetchServices = async () => {
  try {
    const response = await client.get('/services')
    return response.data
  } catch (error) {
    throw new Error(error.response?.data?.message || error.message || 'Failed to fetch services')
  }
}

export const fetchHealth = async () => {
  try {
    const response = await client.get('/health')
    return response.data
  } catch (error) {
    throw new Error(error.response?.data?.message || error.message || 'Failed to fetch health')
  }
}

export default client
