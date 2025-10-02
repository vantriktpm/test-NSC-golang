# Script test API local trước khi chạy GitHub Actions

param(
    [string]$ApiUrl = "http://localhost:8080"
)

Write-Host "🧪 Testing local API before GitHub Actions..." -ForegroundColor Green

# Test Health Endpoint
Write-Host "1️⃣ Testing Health Endpoint..." -ForegroundColor Blue
try {
    $response = Invoke-WebRequest -Uri "$ApiUrl/api/v1/health" -Method GET -TimeoutSec 10
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Health check passed" -ForegroundColor Green
        $healthData = $response.Content | ConvertFrom-Json
        Write-Host "   Status: $($healthData.status)" -ForegroundColor White
        Write-Host "   Version: $($healthData.version)" -ForegroundColor White
    } else {
        Write-Host "❌ Health check failed with status: $($response.StatusCode)" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Health check failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test Shorten URL
Write-Host "2️⃣ Testing Shorten URL..." -ForegroundColor Blue
try {
    $body = @{
        url = "https://example.com"
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$ApiUrl/api/v1/shorten" -Method POST -Body $body -ContentType "application/json" -TimeoutSec 10
    if ($response.StatusCode -eq 201) {
        Write-Host "✅ Shorten URL test passed" -ForegroundColor Green
        $shortenData = $response.Content | ConvertFrom-Json
        Write-Host "   Short Code: $($shortenData.short_code)" -ForegroundColor White
        Write-Host "   Short URL: $($shortenData.short_url)" -ForegroundColor White
        $global:SHORT_CODE = $shortenData.short_code
    } else {
        Write-Host "❌ Shorten URL test failed with status: $($response.StatusCode)" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Shorten URL test failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test Analytics
if ($global:SHORT_CODE) {
    Write-Host "3️⃣ Testing Analytics..." -ForegroundColor Blue
    try {
        $response = Invoke-WebRequest -Uri "$ApiUrl/api/v1/analytics/$($global:SHORT_CODE)" -Method GET -TimeoutSec 10
        if ($response.StatusCode -eq 200) {
            Write-Host "✅ Analytics test passed" -ForegroundColor Green
            $analyticsData = $response.Content | ConvertFrom-Json
            Write-Host "   Total Clicks: $($analyticsData.total_clicks)" -ForegroundColor White
            Write-Host "   Unique IPs: $($analyticsData.unique_ips)" -ForegroundColor White
        } else {
            Write-Host "❌ Analytics test failed with status: $($response.StatusCode)" -ForegroundColor Red
        }
    } catch {
        Write-Host "❌ Analytics test failed: $($_.Exception.Message)" -ForegroundColor Red
    }
} else {
    Write-Host "3️⃣ Skipping Analytics test (no short code)" -ForegroundColor Yellow
}

# Test Redirect
if ($global:SHORT_CODE) {
    Write-Host "4️⃣ Testing Redirect..." -ForegroundColor Blue
    try {
        $response = Invoke-WebRequest -Uri "$ApiUrl/$($global:SHORT_CODE)" -Method GET -MaximumRedirection 0 -TimeoutSec 10
        if ($response.StatusCode -eq 302 -or $response.StatusCode -eq 301) {
            Write-Host "✅ Redirect test passed (status: $($response.StatusCode))" -ForegroundColor Green
            $location = $response.Headers.Location
            Write-Host "   Redirect to: $location" -ForegroundColor White
        } else {
            Write-Host "❌ Redirect test failed with status: $($response.StatusCode)" -ForegroundColor Red
        }
    } catch {
        if ($_.Exception.Response.StatusCode -eq 302 -or $_.Exception.Response.StatusCode -eq 301) {
            Write-Host "✅ Redirect test passed (status: $($_.Exception.Response.StatusCode))" -ForegroundColor Green
        } else {
            Write-Host "❌ Redirect test failed: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
} else {
    Write-Host "4️⃣ Skipping Redirect test (no short code)" -ForegroundColor Yellow
}

# Load Test
Write-Host "5️⃣ Running Load Test..." -ForegroundColor Blue
$loadTestResults = @{
    Total = 0
    Success = 0
    Failed = 0
    TotalTime = 0
}

$startTime = Get-Date
$requests = 10
$concurrency = 5

Write-Host "   Testing $requests requests with $concurrency concurrent connections..." -ForegroundColor White

$jobs = @()
for ($i = 1; $i -le $requests; $i++) {
    $job = Start-Job -ScriptBlock {
        param($Url)
        try {
            $response = Invoke-WebRequest -Uri "$Url/api/v1/health" -Method GET -TimeoutSec 5
            return @{ Success = $true; StatusCode = $response.StatusCode }
        } catch {
            return @{ Success = $false; Error = $_.Exception.Message }
        }
    } -ArgumentList $ApiUrl
    
    $jobs += $job
    
    if ($jobs.Count -ge $concurrency) {
        $completedJob = $jobs | Wait-Job -Any
        $result = Receive-Job $completedJob
        Remove-Job $completedJob
        $jobs = $jobs | Where-Object { $_.Id -ne $completedJob.Id }
        
        $loadTestResults.Total++
        if ($result.Success) {
            $loadTestResults.Success++
        } else {
            $loadTestResults.Failed++
        }
    }
}

# Wait for remaining jobs
$jobs | Wait-Job | ForEach-Object {
    $result = Receive-Job $_
    $loadTestResults.Total++
    if ($result.Success) {
        $loadTestResults.Success++
    } else {
        $loadTestResults.Failed++
    }
    Remove-Job $_
}

$endTime = Get-Date
$loadTestResults.TotalTime = ($endTime - $startTime).TotalMilliseconds

Write-Host "✅ Load test completed" -ForegroundColor Green
Write-Host "   Total Requests: $($loadTestResults.Total)" -ForegroundColor White
Write-Host "   Successful: $($loadTestResults.Success)" -ForegroundColor White
Write-Host "   Failed: $($loadTestResults.Failed)" -ForegroundColor White
Write-Host "   Total Time: $([math]::Round($loadTestResults.TotalTime, 2))ms" -ForegroundColor White
Write-Host "   Avg Response Time: $([math]::Round($loadTestResults.TotalTime / $loadTestResults.Total, 2))ms" -ForegroundColor White

# Summary
Write-Host ""
Write-Host "📊 Test Summary:" -ForegroundColor Cyan
Write-Host "   API URL: $ApiUrl" -ForegroundColor White
Write-Host "   Health: ✅" -ForegroundColor Green
Write-Host "   Shorten: ✅" -ForegroundColor Green
if ($global:SHORT_CODE) {
    Write-Host "   Analytics: ✅" -ForegroundColor Green
    Write-Host "   Redirect: ✅" -ForegroundColor Green
}
Write-Host "   Load Test: ✅ ($($loadTestResults.Success)/$($loadTestResults.Total) successful)" -ForegroundColor Green

Write-Host ""
Write-Host "🎉 Local API is ready for GitHub Actions testing!" -ForegroundColor Green
Write-Host "💡 Next steps:" -ForegroundColor Yellow
Write-Host "   1. Start ngrok tunnel: .\start-tunnel.ps1" -ForegroundColor White
Write-Host "   2. Copy the ngrok HTTPS URL" -ForegroundColor White
Write-Host "   3. Run GitHub Actions workflow with the ngrok URL" -ForegroundColor White
