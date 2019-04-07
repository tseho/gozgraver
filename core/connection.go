package gozcore

import (
	"github.com/tarm/serial"
)

// Connection connected
type Connection struct {
	port *serial.Port
}

// Listen to the port and pass the payloads to the given channel
func (connection *Connection) listen(payloads chan<- []byte) {
	for {
		buf := make([]byte, 4)
		length, err := connection.port.Read(buf)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Tracef("received %x", buf)
		payloads <- buf[:length]
	}
}

// Send the packet to the connection
func (connection *Connection) Send(data []byte) {
	log.Tracef("send %x", data)
	_, err := connection.port.Write(data)
	if err != nil {
		log.Error(err)
	}
}
