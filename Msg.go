package p2plocate

import (
	"encoding/json"
)

// Msg ...
// The structure containing the data that is sent with a broadcast message
type Msg struct {
	MsgType   string
	ClientID  string
	Functions []string
	Data      string
}

// ToJSON ...
// Converts the data into a JSON string
func (m *Msg) ToJSON() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FromJSON ...
// Loads the data from the specified JSON string
func (m *Msg) FromJSON(j []byte) error {
	err := json.Unmarshal(j, &m)
	return err
}
