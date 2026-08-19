class AlertItem {
  AlertItem({
    required this.id,
    required this.timestamp,
    required this.roomId,
    required this.severity,
    required this.ruleCode,
    required this.status,
    required this.title,
    required this.detail,
  });

  final String id;
  final String timestamp;
  final String roomId;
  final String severity;
  final String ruleCode;
  final String status;
  final String title;
  final String detail;

  factory AlertItem.fromJson(Map<String, dynamic> json) {
    return AlertItem(
      id: json['id']?.toString() ?? '',
      timestamp: json['timestamp']?.toString() ?? '',
      roomId: json['room_id']?.toString() ?? '',
      severity: json['severity']?.toString() ?? '',
      ruleCode: json['rule_code']?.toString() ?? '',
      status: json['status']?.toString() ?? '',
      title: json['title']?.toString() ?? '',
      detail: json['detail']?.toString() ?? '',
    );
  }
}
