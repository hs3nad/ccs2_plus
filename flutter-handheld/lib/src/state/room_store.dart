import 'dart:async';

import 'package:flutter/foundation.dart';

import '../models/room.dart';
import '../models/room_event.dart';
import '../services/backend_client.dart';
import '../services/sse_client.dart';

class RoomStore extends ChangeNotifier {
  RoomStore({
    required BackendClient backend,
    required SseClient sseClient,
    required Duration pendingTimeout,
  })  : _backend = backend,
        _sseClient = sseClient,
        _pendingTimeout = pendingTimeout;

  final BackendClient _backend;
  final SseClient _sseClient;
  final Duration _pendingTimeout;

  List<Room> rooms = <Room>[];
  final Map<String, _PendingCommand> _pendingCommands = <String, _PendingCommand>{};
  bool isLoading = false;
  String? error;
  StreamSubscription<RoomEvent>? _subscription;

  bool isRoomPending(String roomId) => _pendingCommands.containsKey(roomId);

  String? pendingActionFor(String roomId) => _pendingCommands[roomId]?.action;

  bool isRoomDelayed(String roomId) => _pendingCommands[roomId]?.delayed ?? false;

  int get pendingTimeoutSeconds => _pendingTimeout.inSeconds;

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
    } catch (err) {
      error = err.toString();
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  Future<void> checkIn(Room room) async {
    await _runAction(room.roomId, 'check_in', () async {
      await _backend.checkIn(room.roomId, stayMode: room.stayMode.isEmpty ? 'overnight' : room.stayMode);
    });
  }

  Future<void> checkOut(Room room) async {
    await _runAction(room.roomId, 'check_out', () async {
      await _backend.checkOut(room.roomId);
    });
  }

  Future<void> startCleaning(Room room) async {
    await _runAction(room.roomId, 'start_cleaning', () async {
      await _backend.startCleaning(room.roomId);
    });
  }

  Future<void> finishCleaning(Room room) async {
    await _runAction(room.roomId, 'finish_cleaning', () async {
      await _backend.finishCleaning(room.roomId);
    });
  }

  Future<void> openRoom(Room room) async {
    await _runAction(room.roomId, 'open_room', () async {
      await _backend.openRoom(room.roomId);
    });
  }

  Future<void> sendService(Room room, String serviceType) async {
    await _runAction(room.roomId, 'service', () async {
      await _backend.sendService(room.roomId, serviceType: serviceType);
    });
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

    if (event.type != 'room_update') {
      return;
    }

    final room = Room.fromJson(event.payload);
    _clearPending(room.roomId);
    final index = rooms.indexWhere((Room item) => item.roomId == room.roomId);
    if (index == -1) {
      rooms = <Room>[...rooms, room]..sort((Room a, Room b) => a.roomId.compareTo(b.roomId));
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

  Future<void> _runAction(String roomId, String actionName, Future<void> Function() action) async {
    error = null;
    _setPending(roomId, actionName);
    notifyListeners();
    try {
      await action();
    } catch (err) {
      _clearPending(roomId);
      error = err.toString();
      notifyListeners();
      rethrow;
    }
  }

  void _setPending(String roomId, String actionName) {
    _clearPending(roomId);
    final timer = Timer(_pendingTimeout, () {
      final pending = _pendingCommands[roomId];
      if (pending == null) {
        return;
      }
      _pendingCommands[roomId] = pending.copyWith(delayed: true);
      notifyListeners();
    });
    _pendingCommands[roomId] = _PendingCommand(action: actionName, timer: timer);
  }

  void _clearPending(String roomId) {
    final pending = _pendingCommands.remove(roomId);
    pending?.timer.cancel();
  }

  @override
  void dispose() {
    for (final pending in _pendingCommands.values) {
      pending.timer.cancel();
    }
    _subscription?.cancel();
    super.dispose();
  }
}

class _PendingCommand {
  const _PendingCommand({
    required this.action,
    required this.timer,
    this.delayed = false,
  });

  final String action;
  final Timer timer;
  final bool delayed;

  _PendingCommand copyWith({
    String? action,
    Timer? timer,
    bool? delayed,
  }) {
    return _PendingCommand(
      action: action ?? this.action,
      timer: timer ?? this.timer,
      delayed: delayed ?? this.delayed,
    );
  }
}
