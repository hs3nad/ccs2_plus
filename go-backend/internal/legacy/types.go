package legacy

type RoomStatus byte

const (
	StatusNone      RoomStatus = 0
	StatusVacant    RoomStatus = 1
	StatusOccupied  RoomStatus = 2
	StatusCleaning  RoomStatus = 3
	StatusOverstay  RoomStatus = 4
)

type DisplayMode byte

const (
	DisplayOff       DisplayMode = 0
	DisplaySlowBlink DisplayMode = 1
	DisplayFastBlink DisplayMode = 2
	DisplayOn        DisplayMode = 3
)

type SemanticState string

const (
	SemanticVacant   SemanticState = "vacant"
	SemanticOccupied SemanticState = "occupied"
	SemanticCleaning SemanticState = "cleaning"
	SemanticOverstay SemanticState = "overstay"
)
