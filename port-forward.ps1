# 本地端口 -> 远程IP:远程端口 映射表
$portMappings = @{
    3306  = "172.27.148.199:3307" # Mysql
    6379  = "172.27.148.199:6379" # Redis
    2379  = "172.27.148.199:2379" # Etcd Node1
    22379 = "172.27.148.199:22379" # Etcd Node2
    32379 = "172.27.148.199:32379" # Etcd Node3
    9092  = "172.27.148.199:9092" # Kafka
    18080 = "172.27.148.199:18080" # Kafka UI
}

foreach ($localPort in $portMappings.Keys) {
    $remote = $portMappings[$localPort]
    $remoteParts = $remote.Split(":")
    $remoteIP = $remoteParts[0]
    $remotePort = $remoteParts[1]

    # 查看是否已经有相同本地端口的规则
    $existing = netsh interface portproxy show v4tov4 | Select-String "Listen on IPv4:             127.0.0.1:$localPort"

    if ($existing) {
        Write-Host "local port $localPort has existed, delete..."
        netsh interface portproxy delete v4tov4 `
            listenport=$localPort `
            listenaddress=127.0.0.1
    }

    Write-Host "add port forward ${localPort} -> ${remoteIP}:${remotePort}"
    netsh interface portproxy add v4tov4 `
        listenport=$localPort `
        listenaddress=127.0.0.1 `
        connectport=$remotePort `
        connectaddress=$remoteIP
}

Write-Host "port forward success"