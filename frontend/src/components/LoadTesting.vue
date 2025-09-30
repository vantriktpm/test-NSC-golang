<template>
  <div class="card">
    <h2>⚡ Load Testing</h2>
    
    <form @submit.prevent="startLoadTest" class="form">
      <div class="form-group">
        <label for="testUrl">Test URL:</label>
        <input
          id="testUrl"
          v-model="testUrl"
          type="url"
          class="form-control"
          placeholder="http://localhost:8080/api/v1/health"
          required
        />
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label for="requests">Number of Requests:</label>
          <input
            id="requests"
            v-model.number="requests"
            type="number"
            class="form-control"
            min="1"
            max="10000"
            required
          />
        </div>
        
        <div class="form-group">
          <label for="concurrency">Concurrency:</label>
          <input
            id="concurrency"
            v-model.number="concurrency"
            type="number"
            class="form-control"
            min="1"
            max="100"
            required
          />
        </div>
      </div>
      
      <button type="submit" class="btn btn-primary" :disabled="loading">
        <span v-if="loading" class="loading"></span>
        {{ loading ? 'Testing...' : 'Start Load Test' }}
      </button>
    </form>

    <div v-if="error" class="result error">
      <h3>❌ Error</h3>
      <p>{{ error }}</p>
    </div>

    <!-- Progress Bar -->
    <div v-if="loading && progress > 0" class="progress-container">
      <div class="progress-info">
        <span>Progress: {{ Math.round(progress) }}%</span>
        <span>{{ completedRequests }}/{{ requests }} requests</span>
      </div>
      <div class="progress">
        <div class="progress-bar" :style="{ width: progress + '%' }"></div>
      </div>
    </div>

    <div v-if="result" class="load-test-results">
      <!-- Summary Stats -->
      <div class="card">
        <h3>📊 Test Summary</h3>
        <div class="summary-grid">
          <div class="summary-card">
            <h3>{{ result.totalRequests }}</h3>
            <p>Total Requests</p>
          </div>
          <div class="summary-card success">
            <h3>{{ result.successfulRequests }}</h3>
            <p>Successful</p>
          </div>
          <div class="summary-card error">
            <h3>{{ result.failedRequests }}</h3>
            <p>Failed</p>
          </div>
          <div class="summary-card">
            <h3>{{ result.requestsPerSecond.toFixed(2) }}</h3>
            <p>Requests/sec</p>
          </div>
        </div>
      </div>

      <!-- Response Time Stats -->
      <div class="card">
        <h3>⏱️ Response Time Statistics</h3>
        <div class="response-time-grid">
          <div class="response-time-item">
            <label>Average:</label>
            <span :class="getResponseTimeClass(result.averageResponseTime)">
              {{ result.averageResponseTime.toFixed(2) }}ms
            </span>
          </div>
          <div class="response-time-item">
            <label>Minimum:</label>
            <span class="response-fast">
              {{ result.minResponseTime.toFixed(2) }}ms
            </span>
          </div>
          <div class="response-time-item">
            <label>Maximum:</label>
            <span :class="getResponseTimeClass(result.maxResponseTime)">
              {{ result.maxResponseTime.toFixed(2) }}ms
            </span>
          </div>
        </div>
      </div>

      <!-- Success Rate -->
      <div class="card">
        <h3>📈 Success Rate</h3>
        <div class="success-rate">
          <div class="success-rate-bar">
            <div 
              class="success-rate-fill" 
              :style="{ width: getSuccessRate() + '%' }"
            ></div>
          </div>
          <div class="success-rate-text">
            {{ getSuccessRate().toFixed(2) }}% ({{ result.successfulRequests }}/{{ result.totalRequests }})
          </div>
        </div>
      </div>

      <!-- Error Details -->
      <div v-if="result.errors.length > 0" class="card">
        <h3>❌ Error Details</h3>
        <div class="error-list">
          <div 
            v-for="(error, index) in result.errors" 
            :key="index"
            class="error-item"
          >
            <span class="error-message">{{ error.error }}</span>
            <span class="error-count">{{ error.count }} times</span>
          </div>
        </div>
      </div>

      <!-- Performance Chart -->
      <div class="card">
        <h3>📊 Performance Chart</h3>
        <div class="chart-container">
          <div class="chart">
            <div class="chart-bar success" :style="{ height: getSuccessRate() + '%' }">
              <span class="chart-label">Success</span>
              <span class="chart-value">{{ result.successfulRequests }}</span>
            </div>
            <div class="chart-bar error" :style="{ height: getFailureRate() + '%' }">
              <span class="chart-label">Failed</span>
              <span class="chart-value">{{ result.failedRequests }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Test Configuration -->
      <div class="card">
        <h3>⚙️ Test Configuration</h3>
        <div class="config-grid">
          <div class="config-item">
            <label>Test URL:</label>
            <span>{{ testUrl }}</span>
          </div>
          <div class="config-item">
            <label>Total Requests:</label>
            <span>{{ requests }}</span>
          </div>
          <div class="config-item">
            <label>Concurrency:</label>
            <span>{{ concurrency }}</span>
          </div>
          <div class="config-item">
            <label>Test Duration:</label>
            <span>{{ formatDuration(testDuration) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { loadTestApi, type LoadTestResult } from '../services/api'

const testUrl = ref('http://localhost:8080/api/v1/health')
const requests = ref(100)
const concurrency = ref(10)
const result = ref<LoadTestResult | null>(null)
const error = ref('')
const loading = ref(false)
const progress = ref(0)
const completedRequests = ref(0)
const testDuration = ref(0)

const startLoadTest = async () => {
  if (!testUrl.value) return
  
  loading.value = true
  error.value = ''
  result.value = null
  progress.value = 0
  completedRequests.value = 0
  testDuration.value = 0
  
  const startTime = Date.now()
  
  try {
    // Simulate progress updates
    const progressInterval = setInterval(() => {
      if (completedRequests.value < requests.value) {
        progress.value = (completedRequests.value / requests.value) * 100
      }
    }, 100)
    
    const response = await loadTestApi.performLoadTest(
      testUrl.value,
      requests.value,
      concurrency.value
    )
    
    clearInterval(progressInterval)
    result.value = response
    progress.value = 100
    completedRequests.value = requests.value
    testDuration.value = Date.now() - startTime
    
    console.log('✅ Load test completed:', response)
  } catch (err: any) {
    error.value = err.message || 'Failed to perform load test'
    console.error('❌ Error during load test:', err)
  } finally {
    loading.value = false
  }
}

const getSuccessRate = () => {
  if (!result.value) return 0
  return (result.value.successfulRequests / result.value.totalRequests) * 100
}

const getFailureRate = () => {
  if (!result.value) return 0
  return (result.value.failedRequests / result.value.totalRequests) * 100
}

const getResponseTimeClass = (responseTime: number) => {
  if (responseTime < 100) return 'response-fast'
  if (responseTime < 500) return 'response-medium'
  return 'response-slow'
}

const formatDuration = (milliseconds: number) => {
  const seconds = Math.floor(milliseconds / 1000)
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  
  if (minutes > 0) {
    return `${minutes}m ${remainingSeconds}s`
  }
  return `${remainingSeconds}s`
}
</script>

<style scoped>
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.progress-container {
  margin: 20px 0;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  font-weight: 600;
  color: #495057;
}

.load-test-results {
  margin-top: 20px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.summary-card {
  text-align: center;
  padding: 20px;
  border-radius: 15px;
  background: linear-gradient(135deg, #6c757d 0%, #495057 100%);
  color: white;
}

.summary-card.success {
  background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
}

.summary-card.error {
  background: linear-gradient(135deg, #dc3545 0%, #e83e8c 100%);
}

.summary-card h3 {
  font-size: 2rem;
  margin-bottom: 10px;
  font-weight: 700;
}

.summary-card p {
  opacity: 0.9;
  font-size: 1rem;
}

.response-time-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.response-time-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #dee2e6;
}

.response-time-item label {
  font-weight: 600;
  color: #495057;
}

.response-time-item span {
  font-weight: 700;
  padding: 4px 8px;
  border-radius: 4px;
}

.success-rate {
  text-align: center;
}

.success-rate-bar {
  width: 100%;
  height: 30px;
  background: #e9ecef;
  border-radius: 15px;
  overflow: hidden;
  margin-bottom: 10px;
}

.success-rate-fill {
  height: 100%;
  background: linear-gradient(45deg, #28a745, #20c997);
  transition: width 0.3s ease;
  border-radius: 15px;
}

.success-rate-text {
  font-size: 1.2rem;
  font-weight: 600;
  color: #495057;
}

.error-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.error-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  background: #f8d7da;
  border-radius: 8px;
  border: 1px solid #f5c6cb;
  color: #721c24;
}

.error-message {
  font-weight: 500;
  flex: 1;
}

.error-count {
  font-weight: 700;
  background: #dc3545;
  color: white;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.9rem;
}

.chart-container {
  margin-top: 20px;
}

.chart {
  display: flex;
  align-items: end;
  height: 200px;
  gap: 20px;
  padding: 20px 0;
  border-bottom: 2px solid #dee2e6;
  margin-bottom: 10px;
}

.chart-bar {
  flex: 1;
  border-radius: 8px 8px 0 0;
  position: relative;
  min-height: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: end;
  transition: all 0.3s ease;
}

.chart-bar.success {
  background: linear-gradient(45deg, #28a745, #20c997);
}

.chart-bar.error {
  background: linear-gradient(45deg, #dc3545, #e83e8c);
}

.chart-bar:hover {
  transform: scaleY(1.05);
}

.chart-label {
  position: absolute;
  top: -25px;
  font-size: 0.9rem;
  font-weight: 600;
  color: #333;
}

.chart-value {
  position: absolute;
  bottom: -25px;
  font-size: 0.8rem;
  font-weight: 600;
  color: #333;
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 15px;
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #dee2e6;
}

.config-item label {
  font-weight: 600;
  color: #666;
  font-size: 0.9rem;
}

.config-item span {
  font-size: 1rem;
  word-break: break-all;
  font-family: 'Courier New', monospace;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .summary-grid {
    grid-template-columns: 1fr;
  }
  
  .response-time-grid {
    grid-template-columns: 1fr;
  }
  
  .config-grid {
    grid-template-columns: 1fr;
  }
  
  .chart {
    height: 150px;
    gap: 10px;
  }
  
  .error-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
}
</style>
