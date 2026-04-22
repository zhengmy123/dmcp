import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import './styles/main.css'
import 'nprogress/nprogress.css'

let configLoaded = false
window.APP_CONFIG = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:17050'
}

fetch('/config.json')
  .then(res => {
    if (res.ok) return res.json()
    throw new Error('config not found')
  })
  .then(config => {
    if (config.apiBaseUrl) {
      window.APP_CONFIG.apiBaseUrl = config.apiBaseUrl
    }
  })
  .catch(() => {
    console.log('Using default API base URL')
  })
  .finally(() => {
    configLoaded = true
    window.dispatchEvent(new CustomEvent('config-loaded'))
  })

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')