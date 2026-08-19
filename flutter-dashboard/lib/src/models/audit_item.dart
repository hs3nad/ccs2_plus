class AuditItem {
  AuditItem({
    required this.id,
    required this.timestamp,
    required this.roomId,
    required this.action,
    required this.actor,
    required this.result,
    required this.stayMode,
    required this.caseCode,
    required this.detail,
  });

  final String id;
  final String timestamp;
  final String roomId;
  final String action;
  final String actor;
  final String result;
  final String stayMode;
  final String caseCode;
  final String detail;

  factory AuditItem.fromJson(Map<String, dynamic> json) {
    return AuditItem(
      id: json['id']?.toString() ?? '',
      timestamp: json['timestamp']?.toString() ?? '',
      roomId: json['room_id']?.toString() ?? '',
      action: json['action']?.toString() ?? '',
      actor: json['actor']?.toString() ?? '',
      result: json['result']?.toString() ?? '',
      stayMode: json['stay_mode']?.toString() ?? '',
      caseCode: json['case']?.toString() ?? '',
      detail: json['detail']?.toString() ?? '',
    );
  }
}
