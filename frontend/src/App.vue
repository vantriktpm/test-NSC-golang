<template>
  <div class="container">
    <header class="header">
      <h1>🔗 URL Shortener</h1>
      <p>Testing Interface & Analytics Dashboard</p>
    </header>

    <nav class="nav">
      <button 
        v-for="tab in tabs" 
        :key="tab.id"
        @click="activeTab = tab.id"
        :class="['nav-button', { active: activeTab === tab.id }]"
      >
        {{ tab.icon }} {{ tab.name }}
      </button>
    </nav>

    <div class="content">
      <!-- URL Shortener Tab -->
      <div v-if="activeTab === 'shortener'" class="tab-content">
        <UrlShortener />
      </div>

      <!-- Analytics Tab -->
      <div v-if="activeTab === 'analytics'" class="tab-content">
        <Analytics />
      </div>

      <!-- Health Check Tab -->
      <div v-if="activeTab === 'health'" class="tab-content">
        <HealthCheck />
      </div>

      <!-- Load Testing Tab -->
      <div v-if="activeTab === 'loadtest'" class="tab-content">
        <LoadTesting />
      </div>

      <!-- Bulk Operations Tab -->
      <div v-if="activeTab === 'bulk'" class="tab-content">
        <BulkOperations />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import UrlShortener from './components/UrlShortener.vue'
import Analytics from './components/Analytics.vue'
import HealthCheck from './components/HealthCheck.vue'
import LoadTesting from './components/LoadTesting.vue'
import BulkOperations from './components/BulkOperations.vue'

const activeTab = ref('shortener')

const tabs = [
  { id: 'shortener', name: 'URL Shortener', icon: '✂️' },
  { id: 'analytics', name: 'Analytics', icon: '📊' },
  { id: 'health', name: 'Health Check', icon: '🏥' },
  { id: 'loadtest', name: 'Load Testing', icon: '⚡' },
  { id: 'bulk', name: 'Bulk Operations', icon: '📦' }
]
</script>

<style scoped>
.content {
  min-height: 400px;
}

.tab-content {
  animation: fadeIn 0.3s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
