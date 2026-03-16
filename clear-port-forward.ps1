# 本地端口列表，需要清除的端口
$localPorts = @(3306, 6379, 2379, 22379, 32379, 9092, 18080)

foreach ($port in $localPorts) {
    Write-Host "delete port = $port forward rules..."
    netsh interface portproxy delete v4tov4 `
        listenport=$port `
        listenaddress=127.0.0.1
}

Write-Host "clear port forward done"