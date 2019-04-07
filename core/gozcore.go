package gozcore

import (
	"fmt"
	"image"
	"time"

	"github.com/olebedev/emitter"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

var log = logrus.New()

// GetLogger returns the logger instance
func GetLogger() *logrus.Logger {
	return log
}

// Model of graver
type Model string

// Models of graver
const (
	ModelNull    Model = "NULL"
	ModelOldKbot Model = "OLD_KBOT"
	ModelOldNor  Model = "OLD_NOR"
	ModelOldLit  Model = "OLD_LIT"
	ModelOldBle  Model = "OLD_BLE"
	ModelNewNor  Model = "NEW_NOR"
	ModelNewLit  Model = "NEW_LIT"
	ModelNewBle  Model = "NEW_BLE"
)

// Graver connected
type Graver struct {
	Model      Model
	Connection *Connection
	Events     *emitter.Emitter

	protocol protocol
}

// Connect to the graver with a serial connection
func Connect(com string) (*Graver, error) {
	sconf := &serial.Config{
		Name:     com,
		Baud:     57600,
		Size:     8,
		StopBits: serial.Stop1,
		Parity:   serial.ParityNone,
	}

	port, err := serial.OpenPort(sconf)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Infof("Connected to %s", com)

	connection := &Connection{
		port: port,
	}

	// Open a channel for receiving the payloads
	payloads := make(chan []byte)

	// Start listening on the connection
	go connection.listen(payloads)

	// Create an event emitter
	events := &emitter.Emitter{}

	// Start emitting events for received payloads
	go translatePayloadsToEvents(payloads, events)

	// Wait so the connection listener can be ready before the handshake
	time.Sleep(time.Millisecond * 20)

	// Send a handshake
	connection.Send([]byte{255, 9, 0, 0})

	// Wait for the response
	e := <-events.Once(EventModelRecognized)
	model := e.Args[0].(Model)

	p, err := openProtocol(model)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Sleep a few milliseconds before returning the graver connection,
	// so we can be sure we received every handshake payload before sending requests
	time.Sleep(time.Millisecond * 30)

	graver := &Graver{
		Model:      model,
		Connection: connection,
		Events:     events,
		protocol:   p,
	}

	return graver, nil
}

func openProtocol(model Model) (protocol, error) {
	switch model {
	case ModelNewNor:
		return protocolv4{width: 490, height: 490}, nil
	case ModelNewLit:
		return protocolv4{width: 490, height: 490}, nil
	case ModelNewBle:
		return protocolv4{width: 550, height: 550}, nil
	}

	return nil, fmt.Errorf("The protocol for %s has not been implemented", model)
}

// SetBurnTime sets the burn time, in ms, for upcoming engravings
func (graver *Graver) SetBurnTime(burn int) error {
	return graver.protocol.SetBurnTime(graver, burn)
}

// SetLaserPower sets the laser power, in %, for upcoming engravings
func (graver *Graver) SetLaserPower(power int) error {
	return graver.protocol.SetLaserPower(graver, power)
}

// Reset requests the graver to reload the default settings
func (graver *Graver) Reset() {
	graver.protocol.Reset(graver)
}

// Engrave requests the engraving of an image
func (graver *Graver) Engrave(img image.Image, times int) error {
	return graver.protocol.Engrave(graver, img, times)
}
