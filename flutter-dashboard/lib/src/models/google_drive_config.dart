class GoogleDriveConfig {
  const GoogleDriveConfig({
    required this.credentialsPath,
    required this.folderId,
    required this.enabled,
    required this.configured,
    required this.updatedAt,
  });

  final String credentialsPath;
  final String folderId;
  final bool enabled;
  final bool configured;
  final String updatedAt;

  factory GoogleDriveConfig.fromJson(Map<String, dynamic> json) {
    return GoogleDriveConfig(
      credentialsPath: json['credentials_path'] as String? ?? '',
      folderId: json['folder_id'] as String? ?? '',
      enabled: json['enabled'] as bool? ?? false,
      configured: json['configured'] as bool? ?? false,
      updatedAt: json['updated_at'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'credentials_path': credentialsPath,
      'folder_id': folderId,
      'enabled': enabled,
    };
  }
}
