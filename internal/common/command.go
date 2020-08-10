package common

// CommandID command id type
type CommandID int

const (
	// LOGIN used when devices connect to our servers they send a 15-byte long message containing their IMEI code in decimal format
	LOGIN CommandID = iota
	// LOGOUT used to indicate the client to log itself out
	LOGOUT
	// KILL used to indicate a client to terminate its reading loop
	KILL
	// READING typically every 25ms each connected device sends a payload of 40 bytes reading message
	READING
	// READING sent by the server after a succesfull login
	WELCOME
)

// Command is used to send data between clients and server core
type Command struct {
	ID              CommandID
	Sender          uint64
	CallbackChannel chan Command
	Body            []byte
}
