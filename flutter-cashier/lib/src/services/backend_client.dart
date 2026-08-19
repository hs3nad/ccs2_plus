import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/room.dart';

class BackendClient {
  BackendClient(this.baseUri, {http.Client? client}) : _client = client ?? http.Client();

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

  Future<void> extendStay(
    String roomId, {
    required String stayMode,
    required int extendMinutes,
    required double paymentAmount,
    required String paymentRef,
  }) {
    return _post(
      '/api/rooms/$roomId/extend',
      <String, dynamic>{
        'action': 'extend_stay',
        'stay_mode': stayMode,
        'case': stayMode,
        'extend_minutes': extendMinutes,
        'payment_amount': paymentAmount,
        'payment_ref': paymentRef,
      },
    );
  }

  Future<void> receivePayment(
    String roomId, {
    required double amount,
    required String method,
    required String reference,
  }) {
    return _post(
      '/api/rooms/$roomId/payment',
      <String, dynamic>{
        'action': 'receive_payment',
        'amount': amount,
        'method': method,
        'reference': reference,
      },
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

  Future<void> _post(String path, Map<String, dynamic> payload) async {
    final uri = baseUri.replace(path: path);
    final response = await _client.post(
      uri,
      headers: <String, String>{'Content-Type': 'application/json'},
      body: jsonEncode(payload),
    );
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Request failed: ${response.statusCode} ${response.body}');
    }
  }
}
