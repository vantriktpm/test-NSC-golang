#!/usr/bin/env node

/**
 * Script test 6000 requests đồng thời cho API URL Shortener
 * Sử dụng Node.js với async/await và Promise.all
 */

const http = require('http');
const https = require('https');
const { URL } = require('url');

// Configuration
const BASE_URL = 'http://localhost:8080';
const TOTAL_REQUESTS = 6000;
const CONCURRENT_USERS = 100;
const TIMEOUT = 30000;

class LoadTestResult {
    constructor(endpoint) {
        this.endpoint = endpoint;
        this.totalRequests = 0;
        this.successCount = 0;
        this.errorCount = 0;
        this.responseTimes = [];
        this.startTime = null;
        this.endTime = null;
    }

    addResult(success, responseTime) {
        this.totalRequests++;
        if (success) {
            this.successCount++;
        } else {
            this.errorCount++;
        }
        this.responseTimes.push(responseTime);
    }

    getStats() {
        if (this.responseTimes.length === 0) {
            return {};
        }

        const totalTime = (this.endTime - this.startTime) / 1000;
        const sortedTimes = [...this.responseTimes].sort((a, b) => a - b);

        return {
            endpoint: this.endpoint,
            totalRequests: this.totalRequests,
            successCount: this.successCount,
            errorCount: this.errorCount,
            successRate: (this.successCount / this.totalRequests) * 100,
            totalTime: totalTime,
            requestsPerSecond: this.totalRequests / totalTime,
            minResponseTime: Math.min(...this.responseTimes),
            maxResponseTime: Math.max(...this.responseTimes),
            avgResponseTime: this.responseTimes.reduce((a, b) => a + b, 0) / this.responseTimes.length,
            medianResponseTime: this.percentile(50),
            p95ResponseTime: this.percentile(95),
            p99ResponseTime: this.percentile(99),
        };
    }

    percentile(p) {
        if (this.responseTimes.length === 0) return 0;
        const sortedTimes = [...this.responseTimes].sort((a, b) => a - b);
        const index = Math.ceil((sortedTimes.length * p) / 100) - 1;
        return sortedTimes[Math.min(index, sortedTimes.length - 1)];
    }
}

function makeRequest(url, method = 'GET', data = null) {
    return new Promise((resolve) => {
        const startTime = Date.now();
        const urlObj = new URL(url);
        const options = {
            hostname: urlObj.hostname,
            port: urlObj.port || (urlObj.protocol === 'https:' ? 443 : 80),
            path: urlObj.pathname + urlObj.search,
            method: method,
            timeout: TIMEOUT,
            headers: {
                'User-Agent': 'LoadTest/1.0',
            }
        };

        if (method === 'POST' && data) {
            const jsonData = JSON.stringify(data);
            options.headers['Content-Type'] = 'application/json';
            options.headers['Content-Length'] = Buffer.byteLength(jsonData);
        }

        const client = urlObj.protocol === 'https:' ? https : http;
        
        const req = client.request(options, (res) => {
            let responseData = '';
            res.on('data', (chunk) => {
                responseData += chunk;
            });
            res.on('end', () => {
                const responseTime = Date.now() - startTime;
                const success = res.statusCode >= 200 && res.statusCode < 400;
                resolve({ success, responseTime });
            });
        });

        req.on('error', (err) => {
            const responseTime = Date.now() - startTime;
            console.error(`Request error: ${err.message}`);
            resolve({ success: false, responseTime });
        });

        req.on('timeout', () => {
            const responseTime = Date.now() - startTime;
            req.destroy();
            resolve({ success: false, responseTime });
        });

        if (method === 'POST' && data) {
            req.write(JSON.stringify(data));
        }
        
        req.end();
    });
}

async function runLoadTest(endpoint, method = 'GET', data = null) {
    const url = BASE_URL + endpoint;
    const result = new LoadTestResult(endpoint);
    result.startTime = Date.now();

    console.log(`Đang chạy test cho ${endpoint}...`);

    // Tạo array của promises cho tất cả requests
    const promises = [];
    for (let i = 0; i < TOTAL_REQUESTS; i++) {
        promises.push(makeRequest(url, method, data));
    }

    // Chạy tất cả requests đồng thời với giới hạn concurrent
    const chunks = [];
    for (let i = 0; i < promises.length; i += CONCURRENT_USERS) {
        chunks.push(promises.slice(i, i + CONCURRENT_USERS));
    }

    for (const chunk of chunks) {
        const results = await Promise.all(chunk);
        results.forEach(({ success, responseTime }) => {
            result.addResult(success, responseTime);
        });
    }

    result.endTime = Date.now();
    return result;
}

function printResult(result) {
    const stats = result.getStats();
    if (Object.keys(stats).length === 0) {
        console.log('Không có dữ liệu để hiển thị');
        return;
    }

    console.log(`Endpoint: ${stats.endpoint}`);
    console.log(`Total requests: ${stats.totalRequests}`);
    console.log(`Success: ${stats.successCount} (${stats.successRate.toFixed(2)}%)`);
    console.log(`Errors: ${stats.errorCount} (${(100 - stats.successRate).toFixed(2)}%)`);
    console.log(`Total time: ${stats.totalTime.toFixed(2)}s`);
    console.log(`Requests/sec: ${stats.requestsPerSecond.toFixed(2)}`);
    console.log(`Min response time: ${stats.minResponseTime.toFixed(3)}s`);
    console.log(`Max response time: ${stats.maxResponseTime.toFixed(3)}s`);
    console.log(`Avg response time: ${stats.avgResponseTime.toFixed(3)}s`);
    console.log(`Median response time: ${stats.medianResponseTime.toFixed(3)}s`);
    console.log(`95th percentile: ${stats.p95ResponseTime.toFixed(3)}s`);
    console.log(`99th percentile: ${stats.p99ResponseTime.toFixed(3)}s`);
    console.log();
}

async function main() {
    console.log('🚀 Bắt đầu test 6000 requests cho API URL Shortener');
    console.log(`Base URL: ${BASE_URL}`);
    console.log(`Total requests: ${TOTAL_REQUESTS}`);
    console.log(`Concurrent users: ${CONCURRENT_USERS}`);
    console.log('==========================================');

    try {
        // Test 1: Health Check
        console.log('\n🔍 Test 1: Health Check');
        console.log('----------------------');
        const healthResult = await runLoadTest('/api/v1/health');
        printResult(healthResult);

        // Test 2: Shorten URL
        console.log('\n🔗 Test 2: Shorten URL');
        console.log('---------------------');
        const shortenData = { url: 'https://www.example.com/test-url-for-load-testing' };
        const shortenResult = await runLoadTest('/api/v1/shorten', 'POST', shortenData);
        printResult(shortenResult);

        // Test 3: Redirect URL
        console.log('\n↩️  Test 3: Redirect URL');
        console.log('----------------------');
        const redirectResult = await runLoadTest('/test123');
        printResult(redirectResult);

        // Test 4: Analytics
        console.log('\n📊 Test 4: Analytics');
        console.log('-------------------');
        const analyticsResult = await runLoadTest('/api/v1/analytics/test123');
        printResult(analyticsResult);

        console.log('\n✅ Hoàn thành test 6000 requests!');
        console.log('==========================================');

    } catch (error) {
        console.error('❌ Lỗi trong quá trình test:', error);
    }
}

// Chạy test
main().catch(console.error);
