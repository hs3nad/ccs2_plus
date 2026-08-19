class Room {
  Room({
    required this.roomId,
    required this.status,
    required this.stayMode,
    required this.caseCode,
    required this.floor,
    required this.controllerOnline,
    required this.remainingMinutes,
  });

  final String roomId;
  final String status;
  final String stayMode;
  final String caseCode;
  final int floor;
  final bool controllerOnline;
  final int? remainingMinutes;

  factory Room.fromJson(Map<String, dynamic> json) {
    return Room(
      roomId: json['room_id']?.toString() ?? '',
      status: json['status']?.toString() ?? 'unknown',
      stayMode: json['stay_mode']?.toString() ?? '',
      caseCode: json['case']?.toString() ?? '',
      floor: (json['floor'] as num?)?.toInt() ?? 0,
      controllerOnline: json['controller_online'] as bool? ?? true,
      remainingMinutes: (json['remaining_minutes'] as num?)?.toInt(),
    );
  }

  Room copyWith({
    String? status,
    String? stayMode,
    String? caseCode,
    bool? controllerOnline,
    int? remainingMinutes,
  }) {
    return Room(
      roomId: roomId,
      status: status ?? this.status,
      stayMode: stayMode ?? this.stayMode,
      caseCode: caseCode ?? this.caseCode,
      floor: floor,
      controllerOnline: controllerOnline ?? this.controllerOnline,
      remainingMinutes: remainingMinutes ?? this.remainingMinutes,
    );
  }
}
