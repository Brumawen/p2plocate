package p2plocate

import (
	"fmt"
	"log"
	"net"
	"time"
)

// P2PServer ...
// Peer discovery server.
// Used to locate other devices and determine their functions.
type P2PServer struct {
	PortNo   int
	ClientID string

	IsRunning      bool
	LastDiscoverOK bool
	LastDiscover   time.Time
	LastError      error

	Functions []string
	Devices   []Device

	startFlag chan bool
	stopFlag  chan bool

	localAddresses []string
}

// Discover ...
// Broadcasts a discovery message and rebuilds the list of devices from the
// replies recieved from the other devices on the local network
func (s *P2PServer) Discover() error {
	req := Msg{
		MsgType:   "Discover",
		ClientID:  s.ClientID,
		Functions: s.Functions,
		Data:      "",
	}
	log.Println("Sending Discover Message.")
	return s.sendMsg(req)
}

// Start ...
// Starts the Server Listening for broadcasts from clients
func (s *P2PServer) Start() error {
	if s.IsRunning {
		return nil
	}

	if s.ClientID == "" {
		s.ClientID = GetClientID()
	}
	l, err := GetLocalIPAddresses()
	if err != nil {
		return err
	}
	s.localAddresses = l

	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", s.PortNo))
	if err != nil {
		s.LastError = err
		return err
	}

	s.IsRunning = true
	s.stopFlag = make(chan bool)
	s.startFlag = make(chan bool)

	go func() {

		// Start Listening to the port
		log.Println("Listening for device messages on UDP port", s.PortNo)
		serverCon, err := net.ListenUDP("udp", serverAddr)
		if err != nil {
			s.LastError = err
			s.IsRunning = false
			<-s.startFlag
			return
		}
		defer serverCon.Close()

		go func() {
			// Send a signal to all devices to reveal themselves
			time.Sleep(1 * time.Second)
			s.Discover()
		}()

		go func() {
			s.startFlag <- true
			for {
				// Get the next message
				buf := make([]byte, 1024)
				n, addr, err := serverCon.ReadFromUDP(buf)
				if err != nil {
					if s.IsRunning {
						log.Println("Error whilst listening for messages.", err)
					} else {
						// We are no longer running
						break
					}
				} else {
					// Handle the message
					go s.handleMsg(buf[:n], addr)
				}
			}
		}()

		// Block until we get a signal to stop
		select {
		case c := <-s.stopFlag:
			{
				if c {
					s.IsRunning = false
					serverCon.Close()
				}
			}
		}
		// Send the signal that we are done
		s.stopFlag <- true
	}()

	// Wait until we have started, or failed
	<-s.startFlag

	return s.LastError
}

// Stop ...
// Stops the server
func (s *P2PServer) Stop() {
	if !s.IsRunning {
		return
	}
	if s.stopFlag != nil {
		log.Println("Stopped listening for device messages.")
		// Set the stop flag to stop the listener
		s.stopFlag <- true
		// Wait for the listener to stop
		<-s.stopFlag
	}
}

// GetDevicesForFunction ...
// Returns a list of devices that offer the specified function
func (s *P2PServer) GetDevicesForFunction(function string) []Device {
	dl := []Device{}

	for _, d := range s.Devices {
		if d.HasFunction(function) {
			dl = append(dl, d)
		}
	}

	return dl
}

func (s *P2PServer) sendMsg(m Msg) error {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", s.PortNo))
	if err != nil {
		return err
	}
	localAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	j, err := m.ToJSON()
	if err != nil {
		return err
	}
	_, err = conn.Write(j)
	return err
}

func (s *P2PServer) handleMsg(buf []byte, addr *net.UDPAddr) {
	m := Msg{}
	err := m.FromJSON(buf)
	if err != nil {
		log.Println("Failed to deserialize device message. ", string(buf))
	} else {
		switch m.MsgType {
		case "Discover":
			log.Println("Discover message received from", m.ClientID, addr.String())
			s.LastDiscover = time.Now()
			s.LastDiscoverOK = true
			if s.ClientID == m.ClientID {
				// This message came from us, so we can ignore it
			} else {
				d := Device{
					ClientID:   m.ClientID,
					Functions:  m.Functions,
					IPAddress:  addr.String(),
					Discovered: time.Now(),
				}
				s.Devices = append(s.Devices, d)
			}
			break
		default:
			log.Println("Unknown message type", m.MsgType, "received from", m.ClientID, addr.String())
			break
		}
	}
}
