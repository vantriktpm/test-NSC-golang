<template>
  <div class="card">
    <h2>📊 Analytics Dashboard</h2>
    
    <form @submit.prevent="getAnalytics" class="form">
      <div class="form-group">
        <label for="shortCode">Enter Short Code:</label>
        <input
          id="shortCode"
          v-model="shortCode"
          type="text"
          class="form-control"
          placeholder="abc123"
          required
        />
      </div>
      
      <button type="submit" class="btn btn-primary" :disabled="loading">
        <span v-if="loading" class="loading"></span>
        {{ loading ? 'Loading...' : 'Get Analytics' }}
      </button>
    </form>

    <div v-if="error" class="result error">
      <h3>❌ Error</h3>
      <p>{{ error }}</p>
    </div>

    <div v-if="analytics" class="analytics-container">
      <!-- Basic Info -->
      <div class="card">
        <h3>📋 Basic Information</h3>
        <div class="info-grid">
          <div class="info-item">
            <label>Short Code:</label>
            <span>{{ analytics.shortCode }}</span>
          </div>
          <div class="info-item">
            <label>Original URL:</label>
            <span class="url-text">{{ analytics.originalUrl }}</span>
          </div>
          <div class="info-item">
            <label>Short URL:</label>
            <span class="url-text">{{ analytics.shortUrl }}</span>
          </div>
          <div class="info-item">
            <label>Created:</label>
            <span>{{ formatDate(analytics.createdAt) }}</span>
          </div>
          <div class="info-item">
            <label>Last Accessed:</label>
            <span>{{ analytics.lastAccessedAt ? formatDate(analytics.lastAccessedAt) : 'Never' }}</span>
          </div>
        </div>
      </div>

      <!-- Click Statistics -->
      <div class="card">
        <h3>📈 Click Statistics</h3>
        <div class="stats-grid">
          <div class="stats-card">
            <h3>{{ analytics.clickCount }}</h3>
            <p>Total Clicks</p>
          </div>
          <div class="stats-card">
            <h3>{{ analytics.clickCount > 0 ? 'Active' : 'Inactive' }}</h3>
            <p>Status</p>
          </div>
        </div>
      </div>

      <!-- Daily Stats Chart -->
      <div v-if="analytics.dailyStats.length > 0" class="card">
        <h3>📅 Daily Click Statistics</h3>
        <div class="chart-container">
          <div class="chart">
            <div 
              v-for="(stat, index) in analytics.dailyStats" 
              :key="index"
              class="chart-bar"
              :style="{ height: getBarHeight(stat.clicks) + '%' }"
              :title="`${stat.date}: ${stat.clicks} clicks`"
            >
              <span class="chart-value">{{ stat.clicks }}</span>
            </div>
          </div>
          <div class="chart-labels">
            <span 
              v-for="(stat, index) in analytics.dailyStats" 
              :key="index"
              class="chart-label"
            >
              {{ formatDateShort(stat.date) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Top Referrers -->
      <div v-if="analytics.topReferrers.length > 0" class="card">
        <h3>🔗 Top Referrers</h3>
        <div class="list-container">
          <div 
            v-for="(referrer, index) in analytics.topReferrers" 
            :key="index"
            class="list-item"
          >
            <span class="rank">#{{ index + 1 }}</span>
            <span class="name">{{ referrer.referrer || 'Direct' }}</span>
            <span class="count">{{ referrer.count }} clicks</span>
          </div>
        </div>
      </div>

      <!-- Top User Agents -->
      <div v-if="analytics.topUserAgents.length > 0" class="card">
        <h3>📱 Top User Agents</h3>
        <div class="list-container">
          <div 
            v-for="(ua, index) in analytics.topUserAgents" 
            :key="index"
            class="list-item"
          >
            <span class="rank">#{{ index + 1 }}</span>
            <span class="name">{{ formatUserAgent(ua.userAgent) }}</span>
            <span class="count">{{ ua.count }} clicks</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { urlApi, type AnalyticsResponse } from '../services/api'

const shortCode = ref('')
const analytics = ref<AnalyticsResponse | null>(null)
const error = ref('')
const loading = ref(false)

const getAnalytics = async () => {
  if (!shortCode.value) return
  
  loading.value = true
  error.value = ''
  analytics.value = null
  
  try {
    const response = await urlApi.getAnalytics(shortCode.value)
    analytics.value = response
    console.log('✅ Analytics loaded:', response)
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || 'Failed to get analytics'
    console.error('❌ Error getting analytics:', err)
  } finally {
    loading.value = false
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const formatDateShort = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const formatUserAgent = (userAgent: string) => {
  // Extract browser and OS info from user agent
  if (userAgent.includes('Chrome')) return 'Chrome'
  if (userAgent.includes('Firefox')) return 'Firefox'
  if (userAgent.includes('Safari')) return 'Safari'
  if (userAgent.includes('Edge')) return 'Edge'
  if (userAgent.includes('Mobile')) return 'Mobile Browser'
  return userAgent.substring(0, 30) + '...'
}

const getBarHeight = (clicks: number) => {
  if (!analytics.value) return 0
  const maxClicks = Math.max(...analytics.value.dailyStats.map(s => s.clicks))
  return maxClicks > 0 ? (clicks / maxClicks) * 100 : 0
}
</script>

<style scoped>
.analytics-container {
  margin-top: 20px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 15px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.info-item label {
  font-weight: 600;
  color: #666;
  font-size: 0.9rem;
}

.info-item span {
  font-size: 1rem;
  word-break: break-all;
}

.url-text {
  font-family: 'Courier New', monospace;
  background: #f8f9fa;
  padding: 5px 8px;
  border-radius: 4px;
  border: 1px solid #dee2e6;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.chart-container {
  margin-top: 20px;
}

.chart {
  display: flex;
  align-items: end;
  height: 200px;
  gap: 10px;
  padding: 20px 0;
  border-bottom: 2px solid #dee2e6;
  margin-bottom: 10px;
}

.chart-bar {
  flex: 1;
  background: linear-gradient(45deg, #667eea, #764ba2);
  border-radius: 4px 4px 0 0;
  position: relative;
  min-height: 20px;
  display: flex;
  align-items: end;
  justify-content: center;
  transition: all 0.3s ease;
}

.chart-bar:hover {
  background: linear-gradient(45deg, #764ba2, #667eea);
  transform: scaleY(1.05);
}

.chart-value {
  position: absolute;
  top: -25px;
  font-size: 0.8rem;
  font-weight: 600;
  color: #333;
}

.chart-labels {
  display: flex;
  gap: 10px;
}

.chart-label {
  flex: 1;
  text-align: center;
  font-size: 0.8rem;
  color: #666;
  transform: rotate(-45deg);
  transform-origin: center;
}

.list-container {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.list-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #dee2e6;
  transition: all 0.3s ease;
}

.list-item:hover {
  background: #e9ecef;
  transform: translateX(5px);
}

.rank {
  font-weight: 700;
  color: #667eea;
  min-width: 30px;
}

.name {
  flex: 1;
  font-weight: 500;
}

.count {
  font-weight: 600;
  color: #28a745;
  background: #d4edda;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.9rem;
}

@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .chart {
    height: 150px;
  }
  
  .list-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .rank {
    min-width: auto;
  }
}
</style>
