package legacy

func BuildSlaveFrame(address byte, payload SlavePayload) []byte {
	frame := make([]byte, 0, 5)
	frame = append(frame, ':', address, payload.LowByte(), payload.HighByte())
	frame = append(frame, checksum(frame))
	return frame
}

func BuildDisplayFrame(cardIndex byte, portIndex int, mode DisplayMode) []byte {
	portRef := byte(int(cardIndex)*16 + portIndex)
	return []byte{':', 0xfe, portRef, byte(mode)}
}

func checksum(frame []byte) byte {
	var sum byte
	for _, b := range frame {
		sum += b
	}
	return ^sum + 1
}
