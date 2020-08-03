package common

// CommandID command id type
type CommandID int

const (
	// LOGIN used when devices connect to our servers they send a 15-byte long message containing their IMEI code in decimal format
	LOGIN CommandID = iota
	// READING typically every 25ms each connected device sends a payload of 40 bytes reading message
	READING
)

// Command is used to send data between clients and server core
type Command struct {
	ID     CommandID
	Sender uint64
	Body   []byte
}
