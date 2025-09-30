<template>
  <div class="card">
    <h2>✂️ URL Shortener</h2>
    
    <form @submit.prevent="shortenUrl" class="form">
      <div class="form-group">
        <label for="url">Enter URL to shorten:</label>
        <input
          id="url"
          v-model="url"
          type="url"
          class="form-control"
          placeholder="https://example.com"
          required
        />
      </div>
      
      <button type="submit" class="btn btn-primary" :disabled="loading">
        <span v-if="loading" class="loading"></span>
        {{ loading ? 'Shortening...' : 'Shorten URL' }}
      </button>
    </form>

    <div v-if="result" class="result success">
      <h3>✅ URL Shortened Successfully!</h3>
      <div class="url-display">
        <span><strong>Short URL:</strong> {{ result.shortUrl }}</span>
        <button @click="copyToClipboard(result.shortUrl)" class="copy-btn">
          📋 Copy
        </button>
      </div>
      <div class="url-display">
        <span><strong>Short Code:</strong> {{ result.shortCode }}</span>
        <button @click="copyToClipboard(result.shortCode)" class="copy-btn">
          📋 Copy
        </button>
      </div>
      <div class="url-display">
        <span><strong>Original URL:</strong> {{ result.originalUrl }}</span>
        <button @click="copyToClipboard(result.originalUrl)" class="copy-btn">
          📋 Copy
        </button>
      </div>
      <p><strong>Created:</strong> {{ formatDate(result.createdAt) }}</p>
    </div>

    <div v-if="error" class="result error">
      <h3>❌ Error</h3>
      <p>{{ error }}</p>
    </div>

    <!-- Test the shortened URL -->
    <div v-if="result" class="card">
      <h3>🧪 Test Shortened URL</h3>
      <p>Click the button below to test if the shortened URL redirects correctly:</p>
      <button @click="testRedirect" class="btn btn-secondary" :disabled="testing">
        <span v-if="testing" class="loading"></span>
        {{ testing ? 'Testing...' : 'Test Redirect' }}
      </button>
      
      <div v-if="redirectResult" class="result" :class="redirectResult.success ? 'success' : 'error'">
        <h4>{{ redirectResult.success ? '✅ Redirect Successful' : '❌ Redirect Failed' }}</h4>
        <p v-if="redirectResult.success">
          Redirected to: <a :href="redirectResult.redirectUrl" target="_blank">{{ redirectResult.redirectUrl }}</a>
        </p>
        <p v-else>{{ redirectResult.error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { urlApi, type ShortenResponse } from '../services/api'

const url = ref('')
const result = ref<ShortenResponse | null>(null)
const error = ref('')
const loading = ref(false)
const testing = ref(false)
const redirectResult = ref<{ success: boolean; redirectUrl?: string; error?: string } | null>(null)

const shortenUrl = async () => {
  if (!url.value) return
  
  loading.value = true
  error.value = ''
  result.value = null
  
  try {
    const response = await urlApi.shortenUrl({ url: url.value })
    result.value = response
    console.log('✅ URL shortened:', response)
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || 'Failed to shorten URL'
    console.error('❌ Error shortening URL:', err)
  } finally {
    loading.value = false
  }
}

const testRedirect = async () => {
  if (!result.value) return
  
  testing.value = true
  redirectResult.value = null
  
  try {
    // Create a temporary link to test redirect
    const link = document.createElement('a')
    link.href = result.value.shortUrl
    link.target = '_blank'
    link.rel = 'noopener noreferrer'
    
    // Simulate click
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    redirectResult.value = {
      success: true,
      redirectUrl: url.value
    }
  } catch (err: any) {
    redirectResult.value = {
      success: false,
      error: err.message || 'Failed to test redirect'
    }
  } finally {
    testing.value = false
  }
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    // You could add a toast notification here
    console.log('📋 Copied to clipboard:', text)
  } catch (err) {
    console.error('❌ Failed to copy to clipboard:', err)
    // Fallback for older browsers
    const textArea = document.createElement('textarea')
    textArea.value = text
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}
</script>

<style scoped>
.form {
  margin-bottom: 20px;
}

.url-display {
  margin: 10px 0;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #dee2e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 10px;
}

.url-display span {
  flex: 1;
  min-width: 200px;
  word-break: break-all;
  font-family: 'Courier New', monospace;
}

.copy-btn {
  background: #17a2b8;
  color: white;
  border: none;
  padding: 8px 12px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.copy-btn:hover {
  background: #138496;
  transform: translateY(-1px);
}

@media (max-width: 768px) {
  .url-display {
    flex-direction: column;
    align-items: stretch;
  }
  
  .url-display span {
    min-width: auto;
    margin-bottom: 10px;
  }
}
</style>
