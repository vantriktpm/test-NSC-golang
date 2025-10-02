import axios from 'axios'

// API base configuration
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    console.log(`🚀 API Request: ${config.method?.toUpperCase()} ${config.url}`)
    return config
  },
  (error) => {
    console.error('❌ Request Error:', error)
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => {
    console.log(`✅ API Response: ${response.status} ${response.config.url}`)
    return response
  },
  (error) => {
    console.error('❌ Response Error:', error.response?.data || error.message)
    return Promise.reject(error)
  }
)

// Types
export interface ShortenRequest {
  url: string
}

export interface ShortenResponse {
  shortCode: string
  shortUrl: string
  originalUrl: string
  createdAt: string
}

export interface AnalyticsResponse {
  shortCode: string
  originalUrl: string
  totalClicks: number
  uniqueIPs: number
  topReferers: Array<{
    referer: string
    count: number
  }>
  recentClicks: Array<{
    id: number
    urlId: number
    ipAddress: string
    userAgent: string
    referer: string
    clickedAt: string
  }>
}

export interface HealthResponse {
  status: string
  timestamp: string
  version: string
  uptime: number
  database: {
    status: string
    responseTime: number
  }
  redis: {
    status: string
    responseTime: number
  }
}

export interface LoadTestResult {
  totalRequests: number
  successfulRequests: number
  failedRequests: number
  averageResponseTime: number
  minResponseTime: number
  maxResponseTime: number
  requestsPerSecond: number
  errors: Array<{
    error: string
    count: number
  }>
}

// API functions
export const urlApi = {
  // Shorten URL
  async shortenUrl(data: ShortenRequest): Promise<ShortenResponse> {
    const response = await api.post('/api/v1/shorten', data)
    return response.data
  },

  // Get analytics
  async getAnalytics(shortCode: string): Promise<AnalyticsResponse> {
    const response = await api.get(`/api/v1/analytics/${shortCode}`)
    return response.data
  },

  // Health check
  async getHealth(): Promise<HealthResponse> {
    const response = await api.get('/api/v1/health')
    return response.data
  },

  // Redirect (for testing)
  async redirectUrl(shortCode: string): Promise<void> {
    const response = await api.get(`/${shortCode}`, {
      maxRedirects: 0,
      validateStatus: (status) => status === 302 || status === 301
    })
    return response.data
  }
}

// Load testing utilities
export const loadTestApi = {
  async performLoadTest(
    url: string,
    requests: number,
    concurrency: number
  ): Promise<LoadTestResult> {
    const results: Array<{
      success: boolean
      responseTime: number
      error?: string
    }> = []

    const startTime = Date.now()
    
    // Create batches for concurrency
    const batches = Math.ceil(requests / concurrency)
    
    for (let batch = 0; batch < batches; batch++) {
      const batchPromises = []
      const batchSize = Math.min(concurrency, requests - batch * concurrency)
      
      for (let i = 0; i < batchSize; i++) {
        batchPromises.push(
          this.makeRequest(url).catch(error => ({
            success: false,
            responseTime: 0,
            error: error.message
          }))
        )
      }
      
      const batchResults = await Promise.all(batchPromises)
      results.push(...batchResults)
    }
    
    const endTime = Date.now()
    const totalTime = endTime - startTime
    
    // Calculate statistics
    const successfulRequests = results.filter(r => r.success).length
    const failedRequests = results.length - successfulRequests
    const responseTimes = results.filter(r => r.success).map(r => r.responseTime)
    
    const averageResponseTime = responseTimes.length > 0 
      ? responseTimes.reduce((a, b) => a + b, 0) / responseTimes.length 
      : 0
    
    const minResponseTime = responseTimes.length > 0 ? Math.min(...responseTimes) : 0
    const maxResponseTime = responseTimes.length > 0 ? Math.max(...responseTimes) : 0
    const requestsPerSecond = (results.length / totalTime) * 1000
    
    // Count errors
    const errorCounts: { [key: string]: number } = {}
    results.filter(r => !r.success).forEach(r => {
      if (r.error) {
        errorCounts[r.error] = (errorCounts[r.error] || 0) + 1
      }
    })
    
    const errors = Object.entries(errorCounts).map(([error, count]) => ({
      error,
      count
    }))
    
    return {
      totalRequests: results.length,
      successfulRequests,
      failedRequests,
      averageResponseTime,
      minResponseTime,
      maxResponseTime,
      requestsPerSecond,
      errors
    }
  },

  async makeRequest(url: string): Promise<{ success: boolean; responseTime: number }> {
    const startTime = Date.now()
    
    try {
      await api.get(url, { timeout: 5000 })
      const responseTime = Date.now() - startTime
      return { success: true, responseTime }
    } catch (error) {
      const responseTime = Date.now() - startTime
      throw { success: false, responseTime, error: error }
    }
  }
}

// Bulk operations
export const bulkApi = {
  async shortenMultipleUrls(urls: string[]): Promise<ShortenResponse[]> {
    const promises = urls.map(url => 
      urlApi.shortenUrl({ url }).catch(error => ({
        shortCode: 'ERROR',
        shortUrl: 'ERROR',
        originalUrl: url,
        createdAt: new Date().toISOString(),
        error: error.message
      }))
    )
    
    return Promise.all(promises)
  },

  async getMultipleAnalytics(shortCodes: string[]): Promise<AnalyticsResponse[]> {
    const promises = shortCodes.map(code => 
      urlApi.getAnalytics(code).catch(error => ({
        shortCode: code,
        originalUrl: 'ERROR',
        totalClicks: 0,
        uniqueIPs: 0,
        topReferers: [],
        recentClicks: [],
        error: error.message
      }))
    )
    
    return Promise.all(promises)
  }
}

export default api
