import { createApp } from 'vue'
import './styles/tokens.css'
import './styles/global.css'
import App from './App.vue'

const app = createApp(App)
app.mount('#app')

// Register service worker for PWA support
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/service-worker.js')
      .then((registration) => {
        console.log('[PWA] Service Worker registered:', registration)
      })
      .catch((error) => {
        console.warn('[PWA] Service Worker registration failed:', error)
      })
  })
}
