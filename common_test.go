package p2plocate

import "testing"
import "fmt"

func TestCanGetClientID(t *testing.T) {
	cid := GetClientID()
	if cid == "" {
		t.Errorf("Expected non blank string.")
	}
	fmt.Println("Client ID returned as", cid)
}

func TestCanGetNetworkInterfaces(t *testing.T) {
	l, err := GetLocalIPAddresses()
	if err != nil {
		t.Error("Error: ", err)
	}
	if l == nil || len(l) == 0 {
		t.Error("Expected address list, received none.")
	} else {
		fmt.Println("LocalAddresses = ", l)
	}

}

func TestCanGetBroadcastAddress(t *testing.T) {
	b := GetBroadcastAddress("192.168.1.1")
	fmt.Println(b)
}
