package p2plocate

import (
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/satori/go.uuid"
)

// GetClientID returns the unique client id (UUID) for this application.
func GetClientID() string {
	fn := "clientid" // File Name
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		// File does not exists, create a new uuid
		uuid := uuid.NewV4().String()
		log.Println("Created new Client ID.", uuid)
		err = ioutil.WriteFile(fn, []byte(uuid), 0666)
		if err != nil {
			log.Println("Failed to create Client ID file.", err)
		}
		return uuid
	}
	// Read the uuid from the file
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Println("Failed to read the Client ID file. Attempting to recreate it.", err)
		uuid := uuid.NewV4().String()
		log.Println("Created new Client ID.", uuid)
		ioutil.WriteFile(fn, []byte(uuid), 0666)
		return uuid
	}
	return string(data)
}

// GetLocalIPAddresses gets a list of valid IPv4 addresses for the local machine
func GetLocalIPAddresses() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	l := []string{}
	for _, i := range ifaces {
		adds, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range adds {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// Only select valid IPv4 addresses that are not loopbacks
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				l = append(l, ip.String())
			}
		}
	}
	return l, nil
}

// GetLocalBroadcastAddress returns the broadcast address for the local network
func GetLocalBroadcastAddress() string {
	adds, err := GetLocalIPAddresses()
	if err != nil {
		return ""
	}

	if len(adds) != 0 {
		return GetBroadcastAddress(adds[0])
	}
	return ""
}

// GetBroadcastAddress returns the broadcast address for the specified IP address's network.
func GetBroadcastAddress(ip string) string {
	a := net.ParseIP(ip).To4()
	if a == nil {
		return ""
	}

	m := a.DefaultMask()
	b := a.Mask(m)

	l := net.IPv4(
		a[0]|^m[0],
		a[1]|^m[1],
		a[2]|^m[2],
		a[3]|^m[3],
	).To4()

	return net.IPv4(
		l[0]|b[0],
		l[1]|b[1],
		l[2]|b[2],
		l[3]|b[3],
	).String()
}
