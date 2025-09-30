<template>
  <div class="card">
    <h2>📦 Bulk Operations</h2>
    
    <!-- Bulk URL Shortening -->
    <div class="operation-section">
      <h3>✂️ Bulk URL Shortening</h3>
      
      <form @submit.prevent="shortenBulkUrls" class="form">
        <div class="form-group">
          <label for="urls">Enter URLs (one per line):</label>
          <textarea
            id="urls"
            v-model="urlsText"
            class="form-control"
            rows="8"
            placeholder="https://example1.com&#10;https://example2.com&#10;https://example3.com"
            required
          ></textarea>
          <small class="form-help">Enter one URL per line. Maximum 100 URLs.</small>
        </div>
        
        <button type="submit" class="btn btn-primary" :disabled="loading || !urlsText.trim()">
          <span v-if="loading" class="loading"></span>
          {{ loading ? 'Processing...' : `Shorten ${getUrlCount()} URLs` }}
        </button>
      </form>

      <!-- Progress for bulk operations -->
      <div v-if="loading && bulkProgress > 0" class="progress-container">
        <div class="progress-info">
          <span>Progress: {{ Math.round(bulkProgress) }}%</span>
          <span>{{ completedBulkOperations }}/{{ getUrlCount() }} URLs</span>
        </div>
        <div class="progress">
          <div class="progress-bar" :style="{ width: bulkProgress + '%' }"></div>
        </div>
      </div>

      <!-- Bulk Results -->
      <div v-if="bulkResults.length > 0" class="bulk-results">
        <h4>📊 Bulk Shortening Results</h4>
        <div class="results-summary">
          <div class="summary-item success">
            <span class="count">{{ getSuccessfulCount() }}</span>
            <span class="label">Successful</span>
          </div>
          <div class="summary-item error">
            <span class="count">{{ getFailedCount() }}</span>
            <span class="label">Failed</span>
          </div>
        </div>
        
        <div class="results-list">
          <div 
            v-for="(result, index) in bulkResults" 
            :key="index"
            class="result-item"
            :class="result.error ? 'error' : 'success'"
          >
            <div class="result-header">
              <span class="result-index">#{{ index + 1 }}</span>
              <span class="result-status">
                {{ result.error ? '❌' : '✅' }}
              </span>
            </div>
            <div class="result-content">
              <div class="result-url">{{ result.originalUrl }}</div>
              <div v-if="!result.error" class="result-short">
                <span>Short URL: {{ result.shortUrl }}</span>
                <button @click="copyToClipboard(result.shortUrl)" class="copy-btn">
                  📋 Copy
                </button>
              </div>
              <div v-if="result.error" class="result-error">
                Error: {{ result.error }}
              </div>
            </div>
          </div>
        </div>
        
        <div class="bulk-actions">
          <button @click="downloadResults" class="btn btn-secondary">
            📥 Download Results
          </button>
          <button @click="clearResults" class="btn btn-danger">
            🗑️ Clear Results
          </button>
        </div>
      </div>
    </div>

    <!-- Bulk Analytics -->
    <div class="operation-section">
      <h3>📊 Bulk Analytics</h3>
      
      <form @submit.prevent="getBulkAnalytics" class="form">
        <div class="form-group">
          <label for="shortCodes">Enter Short Codes (one per line):</label>
          <textarea
            id="shortCodes"
            v-model="shortCodesText"
            class="form-control"
            rows="6"
            placeholder="abc123&#10;def456&#10;ghi789"
            required
          ></textarea>
          <small class="form-help">Enter one short code per line. Maximum 50 codes.</small>
        </div>
        
        <button type="submit" class="btn btn-primary" :disabled="analyticsLoading || !shortCodesText.trim()">
          <span v-if="analyticsLoading" class="loading"></span>
          {{ analyticsLoading ? 'Loading...' : `Get Analytics for ${getShortCodeCount()} Codes` }}
        </button>
      </form>

      <!-- Analytics Results -->
      <div v-if="analyticsResults.length > 0" class="analytics-results">
        <h4>📈 Bulk Analytics Results</h4>
        
        <div class="analytics-summary">
          <div class="summary-stats">
            <div class="stat-item">
              <span class="stat-value">{{ getTotalClicks() }}</span>
              <span class="stat-label">Total Clicks</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ getAverageClicks() }}</span>
              <span class="stat-label">Average Clicks</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ getActiveUrls() }}</span>
              <span class="stat-label">Active URLs</span>
            </div>
          </div>
        </div>
        
        <div class="analytics-list">
          <div 
            v-for="(analytics, index) in analyticsResults" 
            :key="index"
            class="analytics-item"
            :class="analytics.error ? 'error' : 'success'"
          >
            <div class="analytics-header">
              <span class="analytics-code">{{ analytics.shortCode }}</span>
              <span class="analytics-clicks">{{ analytics.clickCount || 0 }} clicks</span>
            </div>
            <div class="analytics-content">
              <div v-if="!analytics.error" class="analytics-details">
                <div class="analytics-url">{{ analytics.originalUrl }}</div>
                <div class="analytics-meta">
                  <span>Created: {{ formatDate(analytics.createdAt) }}</span>
                  <span v-if="analytics.lastAccessedAt">
                    Last accessed: {{ formatDate(analytics.lastAccessedAt) }}
                  </span>
                </div>
              </div>
              <div v-if="analytics.error" class="analytics-error">
                Error: {{ analytics.error }}
              </div>
            </div>
          </div>
        </div>
        
        <div class="bulk-actions">
          <button @click="downloadAnalytics" class="btn btn-secondary">
            📥 Download Analytics
          </button>
          <button @click="clearAnalytics" class="btn btn-danger">
            🗑️ Clear Analytics
          </button>
        </div>
      </div>
    </div>

    <!-- Error Display -->
    <div v-if="error" class="result error">
      <h3>❌ Error</h3>
      <p>{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { bulkApi, type ShortenResponse, type AnalyticsResponse } from '../services/api'

const urlsText = ref('')
const shortCodesText = ref('')
const bulkResults = ref<(ShortenResponse & { error?: string })[]>([])
const analyticsResults = ref<(AnalyticsResponse & { error?: string })[]>([])
const error = ref('')
const loading = ref(false)
const analyticsLoading = ref(false)
const bulkProgress = ref(0)
const completedBulkOperations = ref(0)

const getUrlCount = () => {
  return urlsText.value.trim().split('\n').filter(url => url.trim()).length
}

const getShortCodeCount = () => {
  return shortCodesText.value.trim().split('\n').filter(code => code.trim()).length
}

const shortenBulkUrls = async () => {
  const urls = urlsText.value.trim().split('\n').filter(url => url.trim())
  
  if (urls.length === 0) return
  if (urls.length > 100) {
    error.value = 'Maximum 100 URLs allowed'
    return
  }
  
  loading.value = true
  error.value = ''
  bulkResults.value = []
  bulkProgress.value = 0
  completedBulkOperations.value = 0
  
  try {
    // Simulate progress updates
    const progressInterval = setInterval(() => {
      if (completedBulkOperations.value < urls.length) {
        bulkProgress.value = (completedBulkOperations.value / urls.length) * 100
      }
    }, 100)
    
    const results = await bulkApi.shortenMultipleUrls(urls)
    
    clearInterval(progressInterval)
    bulkResults.value = results
    bulkProgress.value = 100
    completedBulkOperations.value = urls.length
    
    console.log('✅ Bulk shortening completed:', results)
  } catch (err: any) {
    error.value = err.message || 'Failed to shorten URLs'
    console.error('❌ Error during bulk shortening:', err)
  } finally {
    loading.value = false
  }
}

const getBulkAnalytics = async () => {
  const shortCodes = shortCodesText.value.trim().split('\n').filter(code => code.trim())
  
  if (shortCodes.length === 0) return
  if (shortCodes.length > 50) {
    error.value = 'Maximum 50 short codes allowed'
    return
  }
  
  analyticsLoading.value = true
  error.value = ''
  analyticsResults.value = []
  
  try {
    const results = await bulkApi.getMultipleAnalytics(shortCodes)
    analyticsResults.value = results
    
    console.log('✅ Bulk analytics completed:', results)
  } catch (err: any) {
    error.value = err.message || 'Failed to get analytics'
    console.error('❌ Error during bulk analytics:', err)
  } finally {
    analyticsLoading.value = false
  }
}

const getSuccessfulCount = () => {
  return bulkResults.value.filter(r => !r.error).length
}

const getFailedCount = () => {
  return bulkResults.value.filter(r => r.error).length
}

const getTotalClicks = () => {
  return analyticsResults.value
    .filter(a => !a.error)
    .reduce((total, a) => total + a.clickCount, 0)
}

const getAverageClicks = () => {
  const validResults = analyticsResults.value.filter(a => !a.error)
  if (validResults.length === 0) return 0
  return Math.round(getTotalClicks() / validResults.length)
}

const getActiveUrls = () => {
  return analyticsResults.value.filter(a => !a.error && a.clickCount > 0).length
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    console.log('📋 Copied to clipboard:', text)
  } catch (err) {
    console.error('❌ Failed to copy to clipboard:', err)
  }
}

const downloadResults = () => {
  const data = bulkResults.value.map((result, index) => ({
    index: index + 1,
    originalUrl: result.originalUrl,
    shortUrl: result.shortUrl || 'ERROR',
    shortCode: result.shortCode || 'ERROR',
    status: result.error ? 'FAILED' : 'SUCCESS',
    error: result.error || ''
  }))
  
  const csv = [
    'Index,Original URL,Short URL,Short Code,Status,Error',
    ...data.map(row => 
      `"${row.index}","${row.originalUrl}","${row.shortUrl}","${row.shortCode}","${row.status}","${row.error}"`
    )
  ].join('\n')
  
  downloadFile(csv, 'bulk-shortening-results.csv')
}

const downloadAnalytics = () => {
  const data = analyticsResults.value.map((result, index) => ({
    index: index + 1,
    shortCode: result.shortCode,
    originalUrl: result.originalUrl || 'ERROR',
    clickCount: result.clickCount || 0,
    createdAt: result.createdAt || '',
    lastAccessedAt: result.lastAccessedAt || '',
    status: result.error ? 'FAILED' : 'SUCCESS',
    error: result.error || ''
  }))
  
  const csv = [
    'Index,Short Code,Original URL,Click Count,Created At,Last Accessed At,Status,Error',
    ...data.map(row => 
      `"${row.index}","${row.shortCode}","${row.originalUrl}","${row.clickCount}","${row.createdAt}","${row.lastAccessedAt}","${row.status}","${row.error}"`
    )
  ].join('\n')
  
  downloadFile(csv, 'bulk-analytics-results.csv')
}

const downloadFile = (content: string, filename: string) => {
  const blob = new Blob([content], { type: 'text/csv' })
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(url)
}

const clearResults = () => {
  bulkResults.value = []
  urlsText.value = ''
}

const clearAnalytics = () => {
  analyticsResults.value = []
  shortCodesText.value = ''
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}
</script>

<style scoped>
.operation-section {
  margin-bottom: 40px;
  padding-bottom: 30px;
  border-bottom: 2px solid #e9ecef;
}

.operation-section:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.operation-section h3 {
  color: #2c3e50;
  margin-bottom: 20px;
  font-size: 1.5rem;
  font-weight: 600;
}

.form-help {
  color: #6c757d;
  font-size: 0.9rem;
  margin-top: 5px;
  display: block;
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

.bulk-results,
.analytics-results {
  margin-top: 30px;
}

.results-summary {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 15px;
  border-radius: 10px;
  min-width: 100px;
}

.summary-item.success {
  background: #d4edda;
  color: #155724;
}

.summary-item.error {
  background: #f8d7da;
  color: #721c24;
}

.summary-item .count {
  font-size: 1.5rem;
  font-weight: 700;
}

.summary-item .label {
  font-size: 0.9rem;
  opacity: 0.8;
}

.results-list {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #dee2e6;
  border-radius: 8px;
}

.result-item {
  padding: 15px;
  border-bottom: 1px solid #dee2e6;
  transition: all 0.3s ease;
}

.result-item:last-child {
  border-bottom: none;
}

.result-item:hover {
  background: #f8f9fa;
}

.result-item.success {
  border-left: 4px solid #28a745;
}

.result-item.error {
  border-left: 4px solid #dc3545;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.result-index {
  font-weight: 600;
  color: #495057;
}

.result-status {
  font-size: 1.2rem;
}

.result-content {
  margin-left: 20px;
}

.result-url {
  font-family: 'Courier New', monospace;
  background: #f8f9fa;
  padding: 8px;
  border-radius: 4px;
  margin-bottom: 8px;
  word-break: break-all;
}

.result-short {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.result-short span {
  font-family: 'Courier New', monospace;
  background: #e9ecef;
  padding: 4px 8px;
  border-radius: 4px;
  flex: 1;
  min-width: 200px;
}

.result-error {
  color: #dc3545;
  font-weight: 500;
}

.analytics-summary {
  margin-bottom: 20px;
}

.summary-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 20px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 15px;
}

.stat-value {
  display: block;
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 0.9rem;
  opacity: 0.9;
}

.analytics-list {
  max-height: 500px;
  overflow-y: auto;
  border: 1px solid #dee2e6;
  border-radius: 8px;
}

.analytics-item {
  padding: 15px;
  border-bottom: 1px solid #dee2e6;
  transition: all 0.3s ease;
}

.analytics-item:last-child {
  border-bottom: none;
}

.analytics-item:hover {
  background: #f8f9fa;
}

.analytics-item.success {
  border-left: 4px solid #28a745;
}

.analytics-item.error {
  border-left: 4px solid #dc3545;
}

.analytics-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.analytics-code {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  background: #e9ecef;
  padding: 4px 8px;
  border-radius: 4px;
}

.analytics-clicks {
  font-weight: 700;
  color: #28a745;
  background: #d4edda;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.9rem;
}

.analytics-content {
  margin-left: 20px;
}

.analytics-url {
  font-family: 'Courier New', monospace;
  background: #f8f9fa;
  padding: 8px;
  border-radius: 4px;
  margin-bottom: 8px;
  word-break: break-all;
}

.analytics-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 0.9rem;
  color: #6c757d;
}

.analytics-error {
  color: #dc3545;
  font-weight: 500;
}

.bulk-actions {
  display: flex;
  gap: 15px;
  margin-top: 20px;
  flex-wrap: wrap;
}

.copy-btn {
  background: #17a2b8;
  color: white;
  border: none;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.copy-btn:hover {
  background: #138496;
  transform: translateY(-1px);
}

@media (max-width: 768px) {
  .results-summary {
    flex-direction: column;
    align-items: center;
  }
  
  .summary-stats {
    grid-template-columns: 1fr;
  }
  
  .result-short {
    flex-direction: column;
    align-items: stretch;
  }
  
  .result-short span {
    min-width: auto;
  }
  
  .analytics-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .bulk-actions {
    flex-direction: column;
  }
}
</style>
