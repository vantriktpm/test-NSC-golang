#!/usr/bin/env python3
"""
Script test 6000 requests đồng thời cho API URL Shortener
Sử dụng asyncio và aiohttp để test hiệu suất cao
"""

import asyncio
import aiohttp
import time
import json
from typing import Dict, List, Tuple
import statistics

# Configuration
BASE_URL = "http://localhost:8080"
TOTAL_REQUESTS = 6000
CONCURRENT_USERS = 100
TIMEOUT = 30

class LoadTestResult:
    def __init__(self, endpoint: str):
        self.endpoint = endpoint
        self.total_requests = 0
        self.success_count = 0
        self.error_count = 0
        self.response_times = []
        self.start_time = None
        self.end_time = None
    
    def add_result(self, success: bool, response_time: float):
        self.total_requests += 1
        if success:
            self.success_count += 1
        else:
            self.error_count += 1
        self.response_times.append(response_time)
    
    def get_stats(self) -> Dict:
        if not self.response_times:
            return {}
        
        total_time = (self.end_time - self.start_time).total_seconds()
        
        return {
            "endpoint": self.endpoint,
            "total_requests": self.total_requests,
            "success_count": self.success_count,
            "error_count": self.error_count,
            "success_rate": (self.success_count / self.total_requests) * 100,
            "total_time": total_time,
            "requests_per_second": self.total_requests / total_time,
            "min_response_time": min(self.response_times),
            "max_response_time": max(self.response_times),
            "avg_response_time": statistics.mean(self.response_times),
            "median_response_time": statistics.median(self.response_times),
            "p95_response_time": self._percentile(95),
            "p99_response_time": self._percentile(99),
        }
    
    def _percentile(self, p: int) -> float:
        if not self.response_times:
            return 0
        sorted_times = sorted(self.response_times)
        index = int(len(sorted_times) * p / 100)
        return sorted_times[min(index, len(sorted_times) - 1)]

async def make_request(session: aiohttp.ClientSession, url: str, method: str = "GET", data: dict = None) -> Tuple[bool, float]:
    """Thực hiện một request và trả về (success, response_time)"""
    start_time = time.time()
    
    try:
        if method == "POST" and data:
            async with session.post(url, json=data, timeout=TIMEOUT) as response:
                await response.text()
                success = 200 <= response.status < 400
        else:
            async with session.get(url, timeout=TIMEOUT) as response:
                await response.text()
                success = 200 <= response.status < 400
        
        response_time = time.time() - start_time
        return success, response_time
    
    except Exception as e:
        response_time = time.time() - start_time
        print(f"Request error: {e}")
        return False, response_time

async def run_load_test(endpoint: str, method: str = "GET", data: dict = None) -> LoadTestResult:
    """Chạy load test cho một endpoint"""
    url = BASE_URL + endpoint
    result = LoadTestResult(endpoint)
    result.start_time = time.time()
    
    # Tạo semaphore để giới hạn số concurrent requests
    semaphore = asyncio.Semaphore(CONCURRENT_USERS)
    
    async def worker():
        async with semaphore:
            success, response_time = await make_request(session, url, method, data)
            result.add_result(success, response_time)
    
    # Tạo session với connection pooling
    connector = aiohttp.TCPConnector(limit=CONCURRENT_USERS, limit_per_host=CONCURRENT_USERS)
    async with aiohttp.ClientSession(connector=connector) as session:
        # Tạo tasks cho tất cả requests
        tasks = [worker() for _ in range(TOTAL_REQUESTS)]
        
        # Chạy tất cả tasks đồng thời
        await asyncio.gather(*tasks)
    
    result.end_time = time.time()
    return result

def print_result(result: LoadTestResult):
    """In kết quả test"""
    stats = result.get_stats()
    if not stats:
        print("Không có dữ liệu để hiển thị")
        return
    
    print(f"Endpoint: {stats['endpoint']}")
    print(f"Total requests: {stats['total_requests']}")
    print(f"Success: {stats['success_count']} ({stats['success_rate']:.2f}%)")
    print(f"Errors: {stats['error_count']} ({100 - stats['success_rate']:.2f}%)")
    print(f"Total time: {stats['total_time']:.2f}s")
    print(f"Requests/sec: {stats['requests_per_second']:.2f}")
    print(f"Min response time: {stats['min_response_time']:.3f}s")
    print(f"Max response time: {stats['max_response_time']:.3f}s")
    print(f"Avg response time: {stats['avg_response_time']:.3f}s")
    print(f"Median response time: {stats['median_response_time']:.3f}s")
    print(f"95th percentile: {stats['p95_response_time']:.3f}s")
    print(f"99th percentile: {stats['p99_response_time']:.3f}s")
    print()

async def main():
    print("🚀 Bắt đầu test 6000 requests cho API URL Shortener")
    print(f"Base URL: {BASE_URL}")
    print(f"Total requests: {TOTAL_REQUESTS}")
    print(f"Concurrent users: {CONCURRENT_USERS}")
    print("==========================================")
    
    # Test 1: Health Check
    print("\n🔍 Test 1: Health Check")
    print("----------------------")
    health_result = await run_load_test("/api/v1/health")
    print_result(health_result)
    
    # Test 2: Shorten URL
    print("\n🔗 Test 2: Shorten URL")
    print("---------------------")
    shorten_data = {"url": "https://www.example.com/test-url-for-load-testing"}
    shorten_result = await run_load_test("/api/v1/shorten", "POST", shorten_data)
    print_result(shorten_result)
    
    # Test 3: Redirect URL (sử dụng short code giả định)
    print("\n↩️  Test 3: Redirect URL")
    print("----------------------")
    redirect_result = await run_load_test("/test123")
    print_result(redirect_result)
    
    # Test 4: Analytics
    print("\n📊 Test 4: Analytics")
    print("-------------------")
    analytics_result = await run_load_test("/api/v1/analytics/test123")
    print_result(analytics_result)
    
    print("\n✅ Hoàn thành test 6000 requests!")
    print("==========================================")

if __name__ == "__main__":
    # Cài đặt event loop policy cho Windows
    if hasattr(asyncio, 'WindowsProactorEventLoopPolicy'):
        asyncio.set_event_loop_policy(asyncio.WindowsProactorEventLoopPolicy())
    
    asyncio.run(main())
