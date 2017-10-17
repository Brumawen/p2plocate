// Package p2plocate is a peer to peer discovery tool that uses UDP broadcasts on the local network to
// discover services running on other computers within that network and to determine what
// functions those services support.
package p2plocate

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
)

// P2PServer is used to locate services running on other devices and determine their functions.
type P2PServer struct {
	PortNo        int
	ClientID      string
	BroadcastAddr string

	IsRunning      bool
	LastDiscoverOK bool
	LastDiscover   time.Time
	LastError      error

	Functions []string
	Devices   []Device

	startFlag chan bool
	stopFlag  chan bool

	discoverFunc func()
	t            *time.Timer
}

// Discover broadcasts a discovery message and rebuilds the list of devices from the
// replies received from the other devices on the local network
func (s *P2PServer) Discover() error {
	req := Msg{
		MsgType:   "Discover",
		ClientID:  s.ClientID,
		Functions: s.Functions,
		Data:      "",
	}
	log.Println("Sending Discover Message.")
	s.LastDiscover = time.Now()
	return s.sendMsg(req)
}

// Start starts listening for broadcasted messages from other devices.
func (s *P2PServer) Start() error {
	if s.IsRunning {
		return nil
	}

	if s.ClientID == "" {
		s.ClientID = GetClientID()
	}
	s.BroadcastAddr = GetLocalBroadcastAddress()

	s.IsRunning = true
	s.stopFlag = make(chan bool)
	s.startFlag = make(chan bool)

	go func() {
		log.Println("Listening for device messages on UDP port", s.PortNo)
		serverCon, err := listenUDP(s.PortNo)
		if err != nil {
			s.LastError = err
			s.IsRunning = false
			s.startFlag <- true
			return
		}

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
		<-s.stopFlag
		s.IsRunning = false
		serverCon.Close()

		// Send the signal that we are done
		s.stopFlag <- true
	}()

	// Wait until we have started, or failed
	<-s.startFlag

	return s.LastError
}

// Stop stops listening for broadcasted messages from other devices.
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

// GetDevicesForFunction returns a list of devices that offer the specified function.
func (s *P2PServer) GetDevicesForFunction(function string) []Device {
	dl := []Device{}

	for _, d := range s.Devices {
		if d.HasFunction(function) {
			dl = append(dl, d)
		}
	}

	return dl
}

// GetDevice returns the specified device, if it has been discovered.
func (s *P2PServer) GetDevice(clientID string) (bool, Device) {
	for _, d := range s.Devices {
		if d.ClientID == clientID {
			return true, d
		}
	}
	return false, Device{}
}

// OnDiscover allows a func to be set that will be called when a device is discovered or updated
func (s *P2PServer) OnDiscover(f func()) {
	s.discoverFunc = f
}

// sendMsg broadcasts the message to all the devices listening on the network
func (s *P2PServer) sendMsg(m Msg) error {
	if s.BroadcastAddr == "" {
		s.BroadcastAddr = GetLocalBroadcastAddress()
	}
	conn, err := dialUDP(s.BroadcastAddr, s.PortNo)
	if err != nil {
		return errors.Wrap(err, "dialUDP")
	}
	defer conn.Close()

	j, err := m.ToJSON()
	if err != nil {
		return errors.Wrap(err, "ToJSON")
	}
	_, err = conn.Write(j)
	return errors.Wrap(err, "Write")
}

// listenUDP returns a UDP connection that is set up to listen for messages on the specified port number.
func listenUDP(portno int) (*net.UDPConn, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", portno))
	if err != nil {
		return nil, errors.Wrap(err, "net.ResolveUDPAddr")
	}
	// Start Listening to the port
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		return nil, errors.Wrap(err, "net.ListenUDP")
	}
	return conn, nil
}

// dialUDP returns a UDP connection that can be used to broadcast messages.
func dialUDP(addr string, portno int) (*net.UDPConn, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, portno))
	if err != nil {
		return nil, errors.Wrap(err, "net.ResolveUDPAddr")
	}
	log.Println("ServerAddr is", serverAddr.String())
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return nil, errors.Wrap(err, "net.DialUDP")
	}
	return conn, nil
}

// handleMsg handles a message that has been received.
func (s *P2PServer) handleMsg(buf []byte, addr *net.UDPAddr) {
	m := Msg{}
	err := m.FromJSON(buf)
	if err != nil {
		log.Println("Failed to deserialize device message. ", string(buf))
	} else {
		switch m.MsgType {
		case "Discover":
			log.Println("Discover message received from", m.ClientID, addr.String())
			if s.ClientID == m.ClientID {
				// This message came from us, so we can ignore it
				s.LastDiscoverOK = true

				// Remove old devices

			} else {
				exists, d := s.GetDevice(m.ClientID)
				if exists {
					d.Functions = m.Functions
					d.IPAddress = addr.String()
					d.Discovered = time.Now()
				} else {
					d := Device{
						ClientID:   m.ClientID,
						Functions:  m.Functions,
						IPAddress:  addr.String(),
						Discovered: time.Now(),
					}
					s.Devices = append(s.Devices, d)

					if s.t == nil {
						s.t = time.AfterFunc(500*time.Millisecond, func() {
							s.Discover()
							s.discoverFunc()
						})
					} else {
						s.t.Reset(500 * time.Millisecond)
					}
				}

			}
			break
		default:
			log.Println("Unknown message type", m.MsgType, "received from", m.ClientID, addr.String())
			break
		}
	}
}
