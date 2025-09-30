#!/bin/bash

# Script test 6000 requests đồng thời cho API URL Shortener
# Sử dụng Apache Bench (ab) hoặc wrk

BASE_URL="http://localhost:8080"
TOTAL_REQUESTS=6000
CONCURRENT_USERS=100

echo "🚀 Bắt đầu test 6000 requests cho API URL Shortener"
echo "Base URL: $BASE_URL"
echo "Total requests: $TOTAL_REQUESTS"
echo "Concurrent users: $CONCURRENT_USERS"
echo "=========================================="

# Kiểm tra xem ab có sẵn không
if command -v ab &> /dev/null; then
    echo "✅ Sử dụng Apache Bench (ab)"
    TOOL="ab"
elif command -v wrk &> /dev/null; then
    echo "✅ Sử dụng wrk"
    TOOL="wrk"
else
    echo "❌ Cần cài đặt Apache Bench hoặc wrk"
    echo "Ubuntu/Debian: sudo apt-get install apache2-utils"
    echo "macOS: brew install httpd"
    echo "Hoặc: brew install wrk"
    exit 1
fi

# Test 1: Health Check
echo ""
echo "🔍 Test 1: Health Check"
echo "----------------------"
if [ "$TOOL" = "ab" ]; then
    ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS "$BASE_URL/api/v1/health"
else
    wrk -t12 -c$CONCURRENT_USERS -d30s "$BASE_URL/api/v1/health"
fi

# Test 2: Shorten URL
echo ""
echo "🔗 Test 2: Shorten URL"
echo "---------------------"
if [ "$TOOL" = "ab" ]; then
    ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS -p shorten_data.json -T application/json "$BASE_URL/api/v1/shorten"
else
    wrk -t12 -c$CONCURRENT_USERS -d30s -s shorten_url.lua "$BASE_URL/api/v1/shorten"
fi

# Test 3: Redirect URL (cần short code hợp lệ)
echo ""
echo "↩️  Test 3: Redirect URL"
echo "----------------------"
# Tạo một short URL trước
SHORT_CODE=$(curl -s -X POST "$BASE_URL/api/v1/shorten" \
    -H "Content-Type: application/json" \
    -d '{"url": "https://www.google.com"}' | \
    grep -o '"short_code":"[^"]*"' | \
    cut -d'"' -f4)

if [ -n "$SHORT_CODE" ]; then
    echo "Short code được tạo: $SHORT_CODE"
    if [ "$TOOL" = "ab" ]; then
        ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS "$BASE_URL/$SHORT_CODE"
    else
        wrk -t12 -c$CONCURRENT_USERS -d30s "$BASE_URL/$SHORT_CODE"
    fi
else
    echo "❌ Không thể tạo short code để test redirect"
fi

# Test 4: Analytics
echo ""
echo "📊 Test 4: Analytics"
echo "-------------------"
if [ -n "$SHORT_CODE" ]; then
    if [ "$TOOL" = "ab" ]; then
        ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS "$BASE_URL/api/v1/analytics/$SHORT_CODE"
    else
        wrk -t12 -c$CONCURRENT_USERS -d30s "$BASE_URL/api/v1/analytics/$SHORT_CODE"
    fi
else
    echo "❌ Không thể test analytics vì thiếu short code"
fi

echo ""
echo "✅ Hoàn thành test 6000 requests!"
echo "=========================================="
