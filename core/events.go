package gozcore

import (
	"fmt"

	"github.com/olebedev/emitter"
)

// Types of events
const (
	EventHandshakeSuccess string = "e:1"
	EventModelRecognized  string = "e:2"
	EventReadyForUpload   string = "e:5:1:1"
	EventEngravingDone    string = "e:6"
	EventBatteryStatus    string = "e:9"
	EventCarvingTime      string = "e:10"
	EventLaserPowerStatus string = "e:13"
	EventChargingStatus   string = "e:15"
)

// Emit events from received payloads
func translatePayloadsToEvents(payloads <-chan []byte, events *emitter.Emitter) {
	for payload := range payloads {
		if len(payload) >= 4 && int(payload[0]) == 255 {
			switch int(payload[1]) {
			case 1:
				log.Debugf("Handshake successful")
				go events.Emit(EventHandshakeSuccess)
			case 2:
				model, name, err := translateModelFromPayload(payload)
				if err != nil {
					log.Fatal(err)
					break
				}
				log.Infof("Model: %s/%s", model, name)
				go events.Emit(EventModelRecognized, model)
			case 5:
				if payload[2] == 1 && payload[3] == 1 {
					log.Debug("Ready for image upload")
					go events.Emit(EventReadyForUpload)
				}
			case 6:
				log.Debugf("Engraving done")
				go events.Emit(EventEngravingDone)
			case 9:
				// battery status in %
				status := int(payload[2]) * 25
				if status > 100 {
					status = 100
				}
				log.Tracef("Battery status: %d%%", status)
				go events.Emit(EventBatteryStatus, status)
			case 10:
				// carving time in ms
				time := int(payload[2])*100 + int(payload[3])
				log.Debugf("Carving time was: %dms", time)
				go events.Emit(EventCarvingTime, time)
			case 13:
				// laser power in %
				status := int(payload[2])*100 + int(payload[3])
				log.Debugf("Laser power status: %d%%", status)
				go events.Emit(EventLaserPowerStatus, status)
			case 15:
				// charging current in mA
				status := int(payload[2])*100 + int(payload[3])
				log.Tracef("Charging current: %dmA", status)
				go events.Emit(EventChargingStatus, status)
			case 16:
				if payload[2] != 1 || payload[3] != 0 {
					log.Warnf("Control over laser power seems supported by the graver but is not implemented yet.")
				}
			}
		}
	}
}

// Find the corresponding model from the payload
func translateModelFromPayload(payload []byte) (Model, string, error) {
	switch int(payload[2]) {
	case 1:
		switch int(payload[3]) {
		case 0:
			return ModelOldBle, "NEJE-BL", nil
		case 10:
			return ModelNewBle, "NEJE-BL", nil
		}
	case 10:
		switch int(payload[3]) {
		case 1:
			return ModelOldKbot, "K-Bot V3S", nil
		}
	case 11:
		switch int(payload[3]) {
		case 1:
			return ModelOldNor, "DK-8-KZ", nil
		case 2:
			return ModelNewNor, "DK-8-KZ", nil
		}
	case 13:
		switch int(payload[3]) {
		case 1:
			return ModelOldLit, "DK-8-FKZ", nil
		case 2:
			return ModelNewLit, "DK-8-FKZ", nil
		}
	}

	return "", "", fmt.Errorf("Unknown model: %x", payload)
}
