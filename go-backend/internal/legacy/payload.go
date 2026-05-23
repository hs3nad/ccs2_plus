package legacy

import "fmt"

type SlavePayload struct {
	Bitmask uint16
}

func EmptyPayload() SlavePayload {
	return SlavePayload{}
}

func (p *SlavePayload) SetPortStatus(port int, status RoomStatus) error {
	if port < 0 || port > 15 {
		return fmt.Errorf("invalid port index %d", port)
	}
	if status > 0x0f {
		return fmt.Errorf("invalid room status %d", status)
	}

	mask := uint16(1) << port
	if status >= StatusOccupied && status <= StatusOverstay {
		p.Bitmask |= mask
		return nil
	}

	if status == StatusVacant || status == StatusNone {
		p.Bitmask &^= mask
		return nil
	}

	return fmt.Errorf("unsupported room status for AMP_V2D relay encoding: %d", status)
}

func (p SlavePayload) LowByte() byte {
	return byte(p.Bitmask & 0x00ff)
}

func (p SlavePayload) HighByte() byte {
	return byte((p.Bitmask >> 8) & 0x00ff)
}

func (p SlavePayload) PortRef(cardIndex byte, portIndex int) byte {
	return byte(int(cardIndex)*16 + portIndex)
}
