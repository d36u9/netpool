# netpool
It's a tiny light weight conneciton pool, written in golang.
It's useful for some load test.
still in early stage, not completed yet.

## usage
```
tcppool, err :=GetTCPPool("192.168.1.33", "80", 20)

conn := tcpool.Get()
```

or provide your own connection creation help func

```
 pool, err :=  CreatePool(poolInit, func() (net.Conn, error) {
		return net.Dial("tcp", host+":"+port)
        })
```