import 'dart:async';

import 'package:flutter/foundation.dart';

import '../models/alert_item.dart';
import '../models/audit_item.dart';
import '../models/google_drive_config.dart';
import '../models/room.dart';
import '../models/room_event.dart';
import '../services/backend_client.dart';
import '../services/sse_client.dart';

class RoomStore extends ChangeNotifier {
  RoomStore({
    required BackendClient backend,
    required SseClient sseClient,
  })  : _backend = backend,
        _sseClient = sseClient;

  final BackendClient _backend;
  final SseClient _sseClient;

  List<Room> rooms = <Room>[];
  List<AuditItem> auditItems = <AuditItem>[];
  List<AlertItem> alertItems = <AlertItem>[];
  bool isLoading = false;
  String? error;
  StreamSubscription<RoomEvent>? _subscription;

  Future<void> initialize() async {
    await refresh();
    _subscription = _sseClient.connect().listen(
      _handleEvent,
      onError: (Object err) {
        error = err.toString();
        notifyListeners();
      },
    );
  }

  Future<void> refresh() async {
    isLoading = true;
    error = null;
    notifyListeners();
    try {
      rooms = await _backend.fetchRooms();
      auditItems = await _backend.fetchAuditEvents();
      alertItems = await _backend.fetchAlerts();
    } catch (err) {
      error = err.toString();
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  int get vacantCount =>
      rooms.where((Room room) => room.status == 'vacant').length;
  int get occupiedCount =>
      rooms.where((Room room) => room.status == 'occupied').length;
  int get cleaningCount =>
      rooms.where((Room room) => room.status == 'cleaning').length;
  int get overstayCount =>
      rooms.where((Room room) => room.status == 'overstay').length;

  Future<void> checkIn(Room room) async {
    await _backend.checkIn(room.roomId,
        stayMode: room.stayMode.isEmpty ? 'overnight' : room.stayMode);
    await refresh();
  }

  Future<void> checkOut(Room room) async {
    await _backend.checkOut(room.roomId);
    await refresh();
  }

  Future<void> startCleaning(Room room) async {
    await _backend.startCleaning(room.roomId);
    await refresh();
  }

  Future<void> finishCleaning(Room room) async {
    await _backend.finishCleaning(room.roomId);
    await refresh();
  }

  Future<List<AuditItem>> fetchAuditEvents({String? roomId}) {
    if (roomId == null || roomId.isEmpty) {
      return Future<List<AuditItem>>.value(auditItems);
    }
    return Future<List<AuditItem>>.value(
      auditItems.where((AuditItem item) => item.roomId == roomId).toList(),
    );
  }

  Future<List<AlertItem>> fetchAlerts() {
    return Future<List<AlertItem>>.value(alertItems);
  }

  Future<GoogleDriveConfig> fetchGoogleDriveConfig() {
    return _backend.fetchGoogleDriveConfig();
  }

  Future<GoogleDriveConfig> saveGoogleDriveConfig(GoogleDriveConfig config) {
    return _backend.saveGoogleDriveConfig(config);
  }

  void _handleEvent(RoomEvent event) {
    if (event.type == 'full_state') {
      final items = (event.payload['rooms'] as List<dynamic>? ?? <dynamic>[])
          .whereType<Map<String, dynamic>>()
          .map(Room.fromJson)
          .toList();
      rooms = items;
      notifyListeners();
      return;
    }

    if (event.type == 'audit_event') {
      final item = AuditItem.fromJson(event.payload);
      auditItems = <AuditItem>[
        item,
        ...auditItems.where((AuditItem existing) => existing.id != item.id)
      ];
      if (auditItems.length > 200) {
        auditItems = auditItems.take(200).toList();
      }
      notifyListeners();
      return;
    }

    if (event.type == 'fraud_alert') {
      final item = AlertItem.fromJson(event.payload);
      alertItems = <AlertItem>[
        item,
        ...alertItems.where((AlertItem existing) => existing.id != item.id)
      ];
      if (alertItems.length > 200) {
        alertItems = alertItems.take(200).toList();
      }
      notifyListeners();
      return;
    }

    if (event.type != 'room_update') {
      return;
    }

    final room = Room.fromJson(event.payload);
    final index = rooms.indexWhere((Room item) => item.roomId == room.roomId);
    if (index == -1) {
      rooms = <Room>[...rooms, room]
        ..sort((Room a, Room b) => a.roomId.compareTo(b.roomId));
    } else {
      rooms[index] = rooms[index].copyWith(
        status: room.status,
        stayMode: room.stayMode,
        caseCode: room.caseCode,
        controllerOnline: room.controllerOnline,
        remainingMinutes: room.remainingMinutes,
      );
    }
    notifyListeners();
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }
}
