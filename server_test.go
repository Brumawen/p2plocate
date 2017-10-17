package p2plocate

import "testing"

import "time"

func TestServerStarts(t *testing.T) {
	portNo := 20401
	s := P2PServer{
		PortNo: portNo,
	}
	err := s.Start()
	if err != nil {
		t.Error(err)
	}
	s.Stop()
	if s.LastError != nil {
		t.Error(s.LastError)
	}
}

func TestServerCanReceiveMessages(t *testing.T) {
	portNo := 20402
	s := P2PServer{
		PortNo: portNo,
	}
	s.Start()
	time.Sleep(3 * time.Second)
	s.Stop()
	if s.LastError != nil {
		t.Error(s.LastError)
	}
	if !s.LastDiscoverOK {
		t.Error("Did not discover ourselves.")
	}
}

func TestServerCanDetectOtherDevices(t *testing.T) {
	portNo := 20403
	s1 := P2PServer{
		PortNo:   portNo,
		ClientID: "1",
	}
	s2 := P2PServer{
		PortNo:    portNo,
		ClientID:  "2",
		Functions: []string{"Function1", "Function2"},
	}
	a := false
	s1.OnDiscover(func() {
		a = true
	})
	s1.Start()

	time.Sleep(1 * time.Second)

	s2.Discover()
	time.Sleep(1 * time.Second)

	if !a {
		t.Error("OnDiscover func was not called.")
	}
	if len(s1.Devices) != 1 {
		t.Error("S1 failed to discover S2")
	}
	if s1.Devices[0].ClientID != "2" {
		t.Error("S1 failed to discover S2")
	}
	if s1.Devices[0].Functions[0] != "Function1" || s1.Devices[0].Functions[1] != "Function2" {
		t.Error("S1 failed to discover S2's functions")
	}

	s1.Stop()
}
