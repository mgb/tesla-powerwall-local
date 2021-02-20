# tesla-powerwall-local

Since Tesla requires authentication to get statistics about your powerwall on your network, created a basic proxy that will automatically log in for you.

## How to use

```
go install cmd/tesla-powerwall-proxy/tesla-powerwall-proxy.go
tesla-powerwall-proxy --username "your@email.address" --password "gateway password" --host "192.168.0.200" --listen "localhost:8043"
```

You should see "Successfully logged in" message. Verify your system is operating correctly via executing:
```
curl http://localhost:8043/api/system_status/soe
```

You should see something like `{"percentage":100}`, showing your current battery's state of charge.
