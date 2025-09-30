<template>
  <div class="card">
    <h2>🏥 Health Check</h2>
    
    <div class="health-controls">
      <button @click="checkHealth" class="btn btn-primary" :disabled="loading">
        <span v-if="loading" class="loading"></span>
        {{ loading ? 'Checking...' : 'Check Health' }}
      </button>
      
      <button @click="startAutoCheck" class="btn btn-secondary" :disabled="autoCheck">
        {{ autoCheck ? 'Stop Auto Check' : 'Start Auto Check' }}
      </button>
    </div>

    <div v-if="error" class="result error">
      <h3>❌ Error</h3>
      <p>{{ error }}</p>
    </div>

    <div v-if="health" class="health-container">
      <!-- Overall Status -->
      <div class="card">
        <h3>📊 Overall Status</h3>
        <div class="status-grid">
          <div class="status-card" :class="getStatusClass(health.status)">
            <h3>{{ health.status.toUpperCase() }}</h3>
            <p>System Status</p>
          </div>
          <div class="status-card">
            <h3>{{ formatUptime(health.uptime) }}</h3>
            <p>Uptime</p>
          </div>
          <div class="status-card">
            <h3>{{ health.version }}</h3>
            <p>Version</p>
          </div>
          <div class="status-card">
            <h3>{{ formatDate(health.timestamp) }}</h3>
            <p>Last Check</p>
          </div>
        </div>
      </div>

      <!-- Database Status -->
      <div class="card">
        <h3>🗄️ Database Status</h3>
        <div class="service-status">
          <div class="service-info">
            <span class="service-name">PostgreSQL</span>
            <span class="service-status-badge" :class="getStatusClass(health.database.status)">
              {{ health.database.status.toUpperCase() }}
            </span>
          </div>
          <div class="response-time">
            <span>Response Time: {{ health.database.responseTime }}ms</span>
            <div class="progress">
              <div 
                class="progress-bar" 
                :style="{ width: getResponseTimePercentage(health.database.responseTime) + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Redis Status -->
      <div class="card">
        <h3>🔴 Redis Status</h3>
        <div class="service-status">
          <div class="service-info">
            <span class="service-name">Redis Cache</span>
            <span class="service-status-badge" :class="getStatusClass(health.redis.status)">
              {{ health.redis.status.toUpperCase() }}
            </span>
          </div>
          <div class="response-time">
            <span>Response Time: {{ health.redis.responseTime }}ms</span>
            <div class="progress">
              <div 
                class="progress-bar" 
                :style="{ width: getResponseTimePercentage(health.redis.responseTime) + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Performance Metrics -->
      <div class="card">
        <h3>⚡ Performance Metrics</h3>
        <div class="metrics-grid">
          <div class="metric-item">
            <label>Database Response Time:</label>
            <span :class="getResponseTimeClass(health.database.responseTime)">
              {{ health.database.responseTime }}ms
            </span>
          </div>
          <div class="metric-item">
            <label>Redis Response Time:</label>
            <span :class="getResponseTimeClass(health.redis.responseTime)">
              {{ health.redis.responseTime }}ms
            </span>
          </div>
          <div class="metric-item">
            <label>Total Response Time:</label>
            <span :class="getResponseTimeClass(health.database.responseTime + health.redis.responseTime)">
              {{ health.database.responseTime + health.redis.responseTime }}ms
            </span>
          </div>
        </div>
      </div>

      <!-- Health History -->
      <div v-if="healthHistory.length > 0" class="card">
        <h3>📈 Health History</h3>
        <div class="history-container">
          <div 
            v-for="(record, index) in healthHistory.slice(-10)" 
            :key="index"
            class="history-item"
            :class="getStatusClass(record.status)"
          >
            <span class="history-time">{{ formatTime(record.timestamp) }}</span>
            <span class="history-status">{{ record.status.toUpperCase() }}</span>
            <span class="history-db">{{ record.database.responseTime }}ms</span>
            <span class="history-redis">{{ record.redis.responseTime }}ms</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { urlApi, type HealthResponse } from '../services/api'

const health = ref<HealthResponse | null>(null)
const healthHistory = ref<HealthResponse[]>([])
const error = ref('')
const loading = ref(false)
const autoCheck = ref(false)
let autoCheckInterval: number | null = null

const checkHealth = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const response = await urlApi.getHealth()
    health.value = response
    healthHistory.value.push(response)
    
    // Keep only last 50 records
    if (healthHistory.value.length > 50) {
      healthHistory.value = healthHistory.value.slice(-50)
    }
    
    console.log('✅ Health check completed:', response)
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || 'Failed to check health'
    console.error('❌ Error checking health:', err)
  } finally {
    loading.value = false
  }
}

const startAutoCheck = () => {
  if (autoCheck.value) {
    // Stop auto check
    if (autoCheckInterval) {
      clearInterval(autoCheckInterval)
      autoCheckInterval = null
    }
    autoCheck.value = false
  } else {
    // Start auto check
    autoCheck.value = true
    checkHealth() // Initial check
    autoCheckInterval = setInterval(checkHealth, 30000) // Check every 30 seconds
  }
}

const getStatusClass = (status: string) => {
  switch (status.toLowerCase()) {
    case 'healthy':
    case 'ok':
      return 'status-healthy'
    case 'degraded':
      return 'status-degraded'
    case 'unhealthy':
    case 'error':
      return 'status-unhealthy'
    default:
      return 'status-unknown'
  }
}

const getResponseTimeClass = (responseTime: number) => {
  if (responseTime < 100) return 'response-fast'
  if (responseTime < 500) return 'response-medium'
  return 'response-slow'
}

const getResponseTimePercentage = (responseTime: number) => {
  // Convert response time to percentage (0-100ms = 0-100%)
  return Math.min((responseTime / 100) * 100, 100)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const formatTime = (dateString: string) => {
  return new Date(dateString).toLocaleTimeString()
}

const formatUptime = (uptime: number) => {
  const seconds = Math.floor(uptime / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  if (days > 0) return `${days}d ${hours % 24}h ${minutes % 60}m`
  if (hours > 0) return `${hours}h ${minutes % 60}m ${seconds % 60}s`
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`
  return `${seconds}s`
}

// Cleanup on component unmount
onUnmounted(() => {
  if (autoCheckInterval) {
    clearInterval(autoCheckInterval)
  }
})
</script>

<style scoped>
.health-controls {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.health-container {
  margin-top: 20px;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.status-card {
  text-align: center;
  padding: 20px;
  border-radius: 15px;
  color: white;
  background: linear-gradient(135deg, #6c757d 0%, #495057 100%);
}

.status-card.status-healthy {
  background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
}

.status-card.status-degraded {
  background: linear-gradient(135deg, #ffc107 0%, #fd7e14 100%);
}

.status-card.status-unhealthy {
  background: linear-gradient(135deg, #dc3545 0%, #e83e8c 100%);
}

.status-card h3 {
  font-size: 1.5rem;
  margin-bottom: 10px;
  font-weight: 700;
}

.status-card p {
  opacity: 0.9;
  font-size: 1rem;
}

.service-status {
  padding: 20px;
  background: #f8f9fa;
  border-radius: 10px;
  border: 1px solid #dee2e6;
}

.service-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.service-name {
  font-size: 1.2rem;
  font-weight: 600;
  color: #2c3e50;
}

.service-status-badge {
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
}

.service-status-badge.status-healthy {
  background: #d4edda;
  color: #155724;
}

.service-status-badge.status-degraded {
  background: #fff3cd;
  color: #856404;
}

.service-status-badge.status-unhealthy {
  background: #f8d7da;
  color: #721c24;
}

.response-time {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.response-time span {
  font-weight: 500;
  color: #495057;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #dee2e6;
}

.metric-item label {
  font-weight: 600;
  color: #495057;
}

.metric-item span {
  font-weight: 700;
  padding: 4px 8px;
  border-radius: 4px;
}

.response-fast {
  background: #d4edda;
  color: #155724;
}

.response-medium {
  background: #fff3cd;
  color: #856404;
}

.response-slow {
  background: #f8d7da;
  color: #721c24;
}

.history-container {
  max-height: 300px;
  overflow-y: auto;
}

.history-item {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 15px;
  padding: 10px 15px;
  border-radius: 8px;
  margin-bottom: 5px;
  font-size: 0.9rem;
  align-items: center;
}

.history-item.status-healthy {
  background: #d4edda;
  color: #155724;
}

.history-item.status-degraded {
  background: #fff3cd;
  color: #856404;
}

.history-item.status-unhealthy {
  background: #f8d7da;
  color: #721c24;
}

.history-time {
  font-family: 'Courier New', monospace;
}

.history-status {
  font-weight: 600;
  text-transform: uppercase;
}

.history-db,
.history-redis {
  text-align: center;
  font-weight: 500;
}

@media (max-width: 768px) {
  .status-grid {
    grid-template-columns: 1fr;
  }
  
  .metrics-grid {
    grid-template-columns: 1fr;
  }
  
  .history-item {
    grid-template-columns: 1fr;
    gap: 5px;
    text-align: center;
  }
  
  .service-info {
    flex-direction: column;
    gap: 10px;
    align-items: flex-start;
  }
}
</style>
