package roommap

type RoomMapping struct {
	RoomID       string `json:"room_id"`
	ControllerID string `json:"controller_id"`
	Address      byte   `json:"address"`
	CardIndex    byte   `json:"card_index"`
	PortIndex    int    `json:"port_index"`
}

type Repository interface {
	GetByRoomID(roomID string) (RoomMapping, bool)
	List() []RoomMapping
}

type MemoryRepository struct {
	byRoom map[string]RoomMapping
	list   []RoomMapping
}

func NewMemoryRepository(mappings []RoomMapping) *MemoryRepository {
	byRoom := make(map[string]RoomMapping, len(mappings))
	for _, m := range mappings {
		byRoom[m.RoomID] = m
	}
	return &MemoryRepository{
		byRoom: byRoom,
		list:   append([]RoomMapping(nil), mappings...),
	}
}

func (r *MemoryRepository) GetByRoomID(roomID string) (RoomMapping, bool) {
	m, ok := r.byRoom[roomID]
	return m, ok
}

func (r *MemoryRepository) List() []RoomMapping {
	return append([]RoomMapping(nil), r.list...)
}
