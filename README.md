# p2plocate
Package p2plocate is a peer to peer discovery tool that uses UDP broadcasts on the local network to discover services running on other computers within that network and to determine what functions those services support.

## Installation

`go get github.com/brumawen/p2plocate`

## Quick Start

```go
s := p2plocate.P2PServer{
    PortNo: 20401,
    Functions: []string{"Function1", "Function2"},
}
s.OnDiscover(func() {
    fmt.Println("Devices discovered on ", s.LastDiscover.String())
    fmt.Println("-----------------------------------------------------------------------")
    fmt.Println(s.Devices)
    fmt.Println("")
})
s.Start()
```




