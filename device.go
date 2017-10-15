package p2plocate

import "time"

// Device ...
// Holds data about a discovered device
type Device struct {
	ClientID   string
	Functions  []string
	IPAddress  string
	Discovered time.Time
}

// HasFunction ...
// Returns whether or not the device offers the specified function
func (d *Device) HasFunction(function string) bool {
	for _, f := range d.Functions {
		if f == function {
			return true
		}
	}
	return false
}
