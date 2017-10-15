package p2plocate

import (
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/satori/go.uuid"
)

// GetClientID ...
// Returns the unique client id (UUID) for this application.
func GetClientID() string {
	fn := "clientid"
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

// GetLocalIPAddresses ...
// Get a list of IP address for the local machine
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
			if ip != nil {
				l = append(l, ip.String())
			}
		}
	}
	return l, nil
}
