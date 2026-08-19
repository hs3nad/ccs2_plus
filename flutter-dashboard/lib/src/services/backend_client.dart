import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/alert_item.dart';
import '../models/audit_item.dart';
import '../models/google_drive_config.dart';
import '../models/room.dart';

class BackendClient {
  BackendClient(this.baseUri, {http.Client? client})
      : _client = client ?? http.Client();

  final Uri baseUri;
  final http.Client _client;

  Future<List<Room>> fetchRooms() async {
    final uri = baseUri.replace(path: '/api/state');
    final response = await _client.get(uri);
    if (response.statusCode != 200) {
      throw Exception('Failed to fetch state: ${response.statusCode}');
    }

    final decoded = jsonDecode(response.body) as Map<String, dynamic>;
    final rooms = (decoded['rooms'] as List<dynamic>? ?? <dynamic>[])
        .cast<Map<String, dynamic>>()
        .map(Room.fromJson)
        .toList();
    return rooms;
  }

  Future<void> checkIn(String roomId, {String stayMode = 'overnight'}) {
    return _post(
      '/api/rooms/$roomId/checkin',
      <String, dynamic>{
        'action': 'check_in',
        'stay_mode': stayMode,
        'case': stayMode,
      },
    );
  }

  Future<void> checkOut(String roomId) {
    return _post(
      '/api/rooms/$roomId/checkout',
      <String, dynamic>{'action': 'check_out'},
    );
  }

  Future<void> startCleaning(String roomId) {
    return _post(
      '/api/rooms/$roomId/command',
      <String, dynamic>{'action': 'start_cleaning'},
    );
  }

  Future<void> finishCleaning(String roomId) {
    return _post(
      '/api/rooms/$roomId/command',
      <String, dynamic>{'action': 'finish_cleaning'},
    );
  }

  Future<List<AuditItem>> fetchAuditEvents({String? roomId}) async {
    final uri = baseUri.replace(
      path: '/api/audit/events',
      queryParameters:
          roomId == null ? null : <String, String>{'room_id': roomId},
    );
    final response = await _client.get(uri);
    if (response.statusCode != 200) {
      throw Exception('Failed to fetch audit events: ${response.statusCode}');
    }
    final decoded = jsonDecode(response.body) as Map<String, dynamic>;
    return (decoded['items'] as List<dynamic>? ?? <dynamic>[])
        .cast<Map<String, dynamic>>()
        .map(AuditItem.fromJson)
        .toList();
  }

  Future<List<AlertItem>> fetchAlerts() async {
    final uri = baseUri.replace(path: '/api/alerts');
    final response = await _client.get(uri);
    if (response.statusCode != 200) {
      throw Exception('Failed to fetch alerts: ${response.statusCode}');
    }
    final decoded = jsonDecode(response.body) as Map<String, dynamic>;
    return (decoded['items'] as List<dynamic>? ?? <dynamic>[])
        .cast<Map<String, dynamic>>()
        .map(AlertItem.fromJson)
        .toList();
  }

  Future<GoogleDriveConfig> fetchGoogleDriveConfig() async {
    final uri = baseUri.replace(path: '/api/config/gdrive');
    final response = await _client.get(uri);
    if (response.statusCode != 200) {
      throw Exception(
          'Failed to fetch Google Drive config: ${response.statusCode}');
    }
    return GoogleDriveConfig.fromJson(
        jsonDecode(response.body) as Map<String, dynamic>);
  }

  Future<GoogleDriveConfig> saveGoogleDriveConfig(
      GoogleDriveConfig config) async {
    final uri = baseUri.replace(path: '/api/config/gdrive');
    final response = await _client.put(
      uri,
      headers: <String, String>{'Content-Type': 'application/json'},
      body: jsonEncode(config.toJson()),
    );
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception(
          'Failed to save Google Drive config: ${response.statusCode} ${response.body}');
    }
    return GoogleDriveConfig.fromJson(
        jsonDecode(response.body) as Map<String, dynamic>);
  }

  Future<void> _post(String path, Map<String, dynamic> payload) async {
    final uri = baseUri.replace(path: path);
    final response = await _client.post(
      uri,
      headers: <String, String>{'Content-Type': 'application/json'},
      body: jsonEncode(payload),
    );
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception(
          'Request failed: ${response.statusCode} ${response.body}');
    }
  }
}
