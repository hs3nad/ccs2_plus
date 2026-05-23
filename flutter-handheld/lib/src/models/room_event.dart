class RoomEvent {
  RoomEvent({
    required this.type,
    required this.payload,
  });

  final String type;
  final Map<String, dynamic> payload;
}
