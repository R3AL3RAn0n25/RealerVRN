listen_port = 443
health_port = 8080
rotation_interval = "15m"
mac_interface = "eth0"          # change to your real interface (ip link show)
fallback_only_ss = false
adaptive_rotation = true

[v2ray]
inbound_tag = "vrn-in"
outbound_tag = "freedom"

[xray]
enable_quic = true