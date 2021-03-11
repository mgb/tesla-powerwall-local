# tesla-powerwall-local

Since Tesla requires authentication to get statistics about your powerwall on your network, created a basic proxy that will automatically log in for you.

## How to use

[Install Go](https://golang.org/dl/). Then `go install` the proxy.
```
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
