# tesla-powerwall-local

Since Tesla requires authentication to get statistics about your powerwall on your network, created a basic proxy that will automatically log in for you.

## How to use

[Install Go](https://golang.org/dl/). Then `go install` the proxy.
```
go get github.com/mgb/tesla-powerwall-local/cmd/tesla-powerwall-proxy
go install github.com/mgb/tesla-powerwall-local/cmd/tesla-powerwall-proxy
```

At this point, you should have `tesla-powerwall-proxy` installed to your go bin folder (usually `~/go/bin`). Execute it:
```
~/go/bin/tesla-powerwall-proxy --username "your@email.address" --password "gateway password" --host "192.168.0.200" --listen "localhost:8043"
```

You should see "Successfully logged in" message. Verify your system is operating correctly via executing:
```
curl http://localhost:8043/api/system_status/soe
```

You should see something like `{"percentage":100}`, showing your current battery's state of charge. If so, you have successfully started your proxy. Use whatever your system has for starting up in the background, and whatever tools you want to get json out of your powerwall.

## Raspberry Pi Proxy

Setup a Raspberry Pi to proxy your Powerwall ethernet or TEG wifi to your local network. Do this by picking your favor Raspberry OS (debian/ubuntu based are easiest) and connecting a second network adapter (USB ethernet or USB Wifi). Setup your Pi so it connects to your TEG network as well as your LAN. Install the `tesla-powerwall-proxy` command as well as setting up the service so it starts at boot.

```
sudo apt install git golang-go
go install github.com/mgb/tesla-powerwall-local/cmd/tesla-powerwall-proxy@latest

sudo nano /etc/systemd/system/tesla-powerwall-proxy.service
sudo systemctl enable tesla-powerwall-proxy.service
sudo systemctl restart tesla-powerwall-proxy.service
curl http://localhost:8043/api/system_status/soe
```

My `tesla-powerwall-proxy.service` looks like the following (change `YOUR@EMAIL.HERE` and `PASSWORD` to your login/password).
```
[Unit]
Description=Tesla Powerwall middleware, used to pull data from Tesla Gateway

[Service]
User=pi
WorkingDirectory=/home/pi/
ExecStart=/home/pi/go/bin/tesla-powerwall-proxy -h 192.168.91.1 -u YOUR@EMAIL.HERE -p PASSWORD -l=:8043
Type=simple
TimeoutStopSec=10
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```
