package legacy

func SemanticToLegacy(state SemanticState) (RoomStatus, DisplayMode, bool) {
	switch state {
	case SemanticVacant:
		return StatusVacant, DisplayOff, true
	case SemanticOccupied:
		return StatusOccupied, DisplayOn, true
	case SemanticCleaning:
		return StatusCleaning, DisplayOn, true
	case SemanticOverstay:
		return StatusOverstay, DisplayOn, true
	default:
		return 0, 0, false
	}
}
