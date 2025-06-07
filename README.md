![image.png](https://cdn.violet.vin/v2/S8CVLHD.png)

# Build

```ps1
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o go-ip-info-linux-amd64 main.go
```

```cmd
set GOOS=linux
set GOARCH=amd64
go build -o go-ip-info-linux-amd64 main.go
```
# Run

```service
[Unit]
Description=Go IP Info Web Service
After=network.target

[Service]
ExecStart=/usr/local/bin/go-ip-info
Restart=always
RestartSec=5
WorkingDirectory=/usr/local/bin
StandardOutput=append:/var/log/go-ip-info.log
StandardError=append:/var/log/go-ip-info-error.log

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable go-ip-info.service
sudo systemctl start go-ip-info.service
sudo systemctl status go-ip-info.service
```