-- Lua script for wrk to test shorten URL endpoint
wrk.method = "POST"
wrk.body = '{"url": "https://www.example.com/test-url-for-load-testing"}'
wrk.headers["Content-Type"] = "application/json"

-- Function to handle response
function response(status, headers, body)
    if status ~= 201 then
        print("Error: " .. status .. " - " .. body)
    end
end
