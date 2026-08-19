import 'package:flutter/material.dart';

import '../models/room.dart';
import '../state/room_store.dart';

class RoomListScreen extends StatefulWidget {
  const RoomListScreen({
    super.key,
    required this.store,
  });

  final RoomStore store;

  @override
  State<RoomListScreen> createState() => _RoomListScreenState();
}

class _RoomListScreenState extends State<RoomListScreen> {
  String _tab = 'priority';

  RoomStore get store => widget.store;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: store,
      builder: (BuildContext context, _) {
        final rooms = _filteredRooms(store.rooms);

        return Scaffold(
          body: Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: <Color>[
                  Color(0xFFF7FBF9),
                  Color(0xFFF2F7F5),
                ],
              ),
            ),
            child: SafeArea(
              child: Column(
                children: <Widget>[
                  if (store.error != null)
                    Padding(
                      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                      child: MaterialBanner(
                        backgroundColor: const Color(0xFFFFF1F2),
                        content: Text(store.error!),
                        actions: <Widget>[
                          TextButton(
                            onPressed: store.refresh,
                            child: const Text('Retry'),
                          ),
                        ],
                      ),
                    ),
                  Expanded(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        children: <Widget>[
                          _HandheldHero(
                            cleaningCount: store.rooms.where((Room room) => room.status == 'cleaning').length,
                            readyCount: store.rooms.where((Room room) => room.status == 'vacant').length,
                            onRefresh: store.refresh,
                          ),
                          const SizedBox(height: 16),
                          _TaskTabs(
                            currentTab: _tab,
                            onChanged: (String value) {
                              setState(() {
                                _tab = value;
                              });
                            },
                          ),
                          const SizedBox(height: 16),
                          Expanded(
                            child: store.isLoading && store.rooms.isEmpty
                                ? const Center(child: CircularProgressIndicator())
                                : ListView.separated(
                                    itemCount: rooms.length,
                                    separatorBuilder: (_, __) => const SizedBox(height: 12),
                                    itemBuilder: (BuildContext context, int index) {
                                      final room = rooms[index];
                                      return _TaskCard(
                                        room: room,
                                        pendingAction: store.pendingActionFor(room.roomId),
                                        delayed: store.isRoomDelayed(room.roomId),
                                        delayedAfterSeconds: store.pendingTimeoutSeconds,
                                        onTap: () => _openRoomDetail(context, room),
                                        onOpenRoom: () => store.openRoom(room),
                                        onService: () => _openServiceSheet(context, room),
                                        onStartCleaning: () => store.startCleaning(room),
                                        onFinishCleaning: () => store.finishCleaning(room),
                                      );
                                    },
                                  ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  List<Room> _filteredRooms(List<Room> rooms) {
    final sorted = <Room>[...rooms];
    sorted.sort((Room a, Room b) => _priorityScore(b).compareTo(_priorityScore(a)));

    switch (_tab) {
      case 'cleaning':
        return sorted.where((Room room) => room.status == 'cleaning').toList();
      case 'service':
        return sorted.where((Room room) => room.status == 'occupied' || room.status == 'overstay').toList();
      case 'ready':
        return sorted.where((Room room) => room.status == 'vacant').toList();
      default:
        return sorted;
    }
  }

  int _priorityScore(Room room) {
    if (!room.controllerOnline) {
      return 100;
    }
    if (room.status == 'cleaning') {
      return 90;
    }
    if (room.status == 'overstay') {
      return 80;
    }
    if (room.status == 'occupied') {
      return 70;
    }
    if (room.status == 'vacant') {
      return 60;
    }
    return 40;
  }

  Future<void> _openServiceSheet(BuildContext context, Room room) async {
    final result = await showModalBottomSheet<String>(
      context: context,
      showDragHandle: true,
      backgroundColor: const Color(0xFFF8FFFC),
      builder: (BuildContext context) => _ServiceSheet(room: room),
    );

    if (!mounted || result == null) {
      return;
    }

    await store.sendService(room, result);
  }

  Future<void> _openRoomDetail(BuildContext context, Room room) async {
    await Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (BuildContext context) => _RoomDetailScreen(
          roomId: room.roomId,
          store: store,
          delayedAfterSeconds: store.pendingTimeoutSeconds,
          onOpenRoom: () => store.openRoom(room),
          onService: () => _openServiceSheet(context, room),
          onStartCleaning: () => store.startCleaning(room),
          onFinishCleaning: () => store.finishCleaning(room),
        ),
      ),
    );
  }
}

class _HandheldHero extends StatelessWidget {
  const _HandheldHero({
    required this.cleaningCount,
    required this.readyCount,
    required this.onRefresh,
  });

  final int cleaningCount;
  final int readyCount;
  final Future<void> Function() onRefresh;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: const Color(0xFF115E59),
        borderRadius: BorderRadius.circular(28),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x22115E59),
            blurRadius: 28,
            offset: Offset(0, 18),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  'Handheld Tasks',
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                        color: Colors.white,
                        fontWeight: FontWeight.w800,
                      ),
                ),
              ),
              IconButton(
                onPressed: onRefresh,
                style: IconButton.styleFrom(
                  backgroundColor: const Color(0xFF0F766E),
                  foregroundColor: Colors.white,
                ),
                icon: const Icon(Icons.refresh),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            'Open room, service flow, and housekeeping actions designed for one-hand use.',
            style: Theme.of(context).textTheme.bodyLarge?.copyWith(color: const Color(0xFFD0FAE5)),
          ),
          const SizedBox(height: 16),
          Row(
            children: <Widget>[
              Expanded(child: _HeroStat(label: 'Cleaning', value: '$cleaningCount')),
              const SizedBox(width: 10),
              Expanded(child: _HeroStat(label: 'Ready', value: '$readyCount')),
              const SizedBox(width: 10),
              const Expanded(child: _HeroStat(label: 'Mode', value: 'Live')),
            ],
          ),
        ],
      ),
    );
  }
}

class _HeroStat extends StatelessWidget {
  const _HeroStat({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFF0F766E),
        borderRadius: BorderRadius.circular(18),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label.toUpperCase(),
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
                  color: const Color(0xFFA7F3D0),
                  letterSpacing: 1.1,
                ),
          ),
          const SizedBox(height: 6),
          Text(
            value,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: Colors.white,
                  fontWeight: FontWeight.w800,
                ),
          ),
        ],
      ),
    );
  }
}

class _TaskTabs extends StatelessWidget {
  const _TaskTabs({
    required this.currentTab,
    required this.onChanged,
  });

  final String currentTab;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Wrap(
          spacing: 8,
          runSpacing: 8,
          children: <Widget>[
            _TaskTab(
              label: 'Priority',
              selected: currentTab == 'priority',
              onTap: () => onChanged('priority'),
            ),
            _TaskTab(
              label: 'Cleaning',
              selected: currentTab == 'cleaning',
              onTap: () => onChanged('cleaning'),
            ),
            _TaskTab(
              label: 'Service',
              selected: currentTab == 'service',
              onTap: () => onChanged('service'),
            ),
            _TaskTab(
              label: 'Ready',
              selected: currentTab == 'ready',
              onTap: () => onChanged('ready'),
            ),
          ],
        ),
      ),
    );
  }
}

class _TaskTab extends StatelessWidget {
  const _TaskTab({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(999),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFF115E59) : const Color(0xFFF7FBF9),
          borderRadius: BorderRadius.circular(999),
          border: Border.all(
            color: selected ? const Color(0xFF115E59) : const Color(0xFFD6EAE4),
          ),
        ),
        child: Text(
          label,
          style: Theme.of(context).textTheme.labelLarge?.copyWith(
                color: selected ? Colors.white : const Color(0xFF115E59),
                fontWeight: FontWeight.w700,
              ),
        ),
      ),
    );
  }
}

class _TaskCard extends StatelessWidget {
  const _TaskCard({
    required this.room,
    required this.pendingAction,
    required this.delayed,
    required this.delayedAfterSeconds,
    required this.onTap,
    required this.onOpenRoom,
    required this.onService,
    required this.onStartCleaning,
    required this.onFinishCleaning,
  });

  final Room room;
  final String? pendingAction;
  final bool delayed;
  final int delayedAfterSeconds;
  final VoidCallback onTap;
  final VoidCallback onOpenRoom;
  final VoidCallback onService;
  final VoidCallback onStartCleaning;
  final VoidCallback onFinishCleaning;

  @override
  Widget build(BuildContext context) {
    final color = _statusColor(room.status);
    final isCleaning = room.status == 'cleaning';
    final isPending = pendingAction != null;
    final canRetry = isPending && delayed;

    return InkWell(
      onTap: isPending && !canRetry ? null : onTap,
      borderRadius: BorderRadius.circular(24),
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: color.withValues(alpha: 0.24)),
        ),
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                    decoration: BoxDecoration(
                      color: color.withValues(alpha: 0.12),
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Text(
                      room.roomId,
                      style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                            color: color,
                            fontWeight: FontWeight.w800,
                          ),
                    ),
                  ),
                  const Spacer(),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: <Widget>[
                      _StatusTag(status: room.status),
                      if (isPending) ...<Widget>[
                        const SizedBox(height: 8),
                        _PendingChip(
                          action: pendingAction!,
                          delayed: delayed,
                          delayedAfterSeconds: delayedAfterSeconds,
                        ),
                      ],
                    ],
                  ),
                ],
              ),
              const SizedBox(height: 14),
              Text(
                _headline(room),
                style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 6),
              Text(
                isPending
                    ? (delayed
                        ? 'Delayed > ${delayedAfterSeconds}s. You can retry this room action if needed.'
                        : 'Command is being sent. Waiting for room update from backend.')
                    : _detail(room),
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
              ),
              const SizedBox(height: 14),
              _MetaRow(label: 'Stay', value: room.stayMode.isEmpty ? '-' : _titleCase(room.stayMode)),
              _MetaRow(label: 'Case', value: room.caseCode.isEmpty ? '-' : room.caseCode),
              _MetaRow(label: 'Floor', value: '${room.floor}'),
              _MetaRow(
                label: 'Device',
                value: room.controllerOnline ? 'Online' : 'Offline',
                valueColor: room.controllerOnline ? const Color(0xFF16A34A) : const Color(0xFFDC2626),
              ),
              const SizedBox(height: 14),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: <Widget>[
                  FilledButton(
                    onPressed: isPending && !canRetry ? null : onOpenRoom,
                    style: FilledButton.styleFrom(backgroundColor: const Color(0xFF115E59)),
                    child: Text(
                      isPending && pendingAction == 'open_room'
                          ? (delayed ? 'Retry Open' : 'Sending...')
                          : 'Open Room',
                    ),
                  ),
                  FilledButton.tonal(
                    onPressed: isPending && !canRetry ? null : onService,
                    child: Text(
                      isPending && pendingAction == 'service'
                          ? (delayed ? 'Retry Service' : 'Sending...')
                          : 'Service',
                    ),
                  ),
                  OutlinedButton(
                    onPressed: isPending && !canRetry ? null : (isCleaning ? onFinishCleaning : onStartCleaning),
                    child: Text(
                      isPending && (pendingAction == 'start_cleaning' || pendingAction == 'finish_cleaning')
                          ? (delayed ? 'Retry' : 'Sending...')
                          : (isCleaning ? 'Clean Done' : 'Start Cleaning'),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _headline(Room room) {
    if (room.status == 'cleaning') {
      return 'Cleaning in progress';
    }
    if (room.status == 'occupied') {
      return 'Guest room service';
    }
    if (room.status == 'overstay') {
      return 'Coordinate with cashier';
    }
    return 'Room ready for action';
  }

}


class _RoomDetailScreen extends StatelessWidget {
  const _RoomDetailScreen({
    required this.roomId,
    required this.store,
    required this.delayedAfterSeconds,
    required this.onOpenRoom,
    required this.onService,
    required this.onStartCleaning,
    required this.onFinishCleaning,
  });

  final String roomId;
  final RoomStore store;
  final int delayedAfterSeconds;
  final VoidCallback onOpenRoom;
  final VoidCallback onService;
  final VoidCallback onStartCleaning;
  final VoidCallback onFinishCleaning;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: store,
      builder: (BuildContext context, _) {
        final room = store.rooms.firstWhere((Room item) => item.roomId == roomId);
        final color = _statusColor(room.status);
        final isCleaning = room.status == 'cleaning';
        final pendingAction = store.pendingActionFor(room.roomId);
        final isPending = pendingAction != null;
        final delayed = store.isRoomDelayed(room.roomId);
        final canRetry = isPending && delayed;

        return Scaffold(
          appBar: AppBar(
            title: Text('Room ${room.roomId}'),
          ),
          body: Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: <Color>[
                  Color(0xFFF7FBF9),
                  Color(0xFFF2F7F5),
                ],
              ),
            ),
            child: SafeArea(
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: <Widget>[
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: color,
                      borderRadius: BorderRadius.circular(28),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          room.roomId,
                          style: Theme.of(context).textTheme.displaySmall?.copyWith(
                                color: Colors.white,
                                fontWeight: FontWeight.w800,
                              ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _titleCase(room.status),
                          style: Theme.of(context).textTheme.titleLarge?.copyWith(
                                color: Colors.white,
                                fontWeight: FontWeight.w700,
                              ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          isPending
                              ? (delayed
                                  ? 'Delayed > ${delayedAfterSeconds}s for ${_titleCase(pendingAction)}. You can retry from this screen.'
                                  : 'Waiting for ${_titleCase(pendingAction)} confirmation from backend.')
                              : _detail(room),
                          style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                                color: const Color(0xFFEFFCF8),
                              ),
                        ),
                        if (isPending) ...<Widget>[
                          const SizedBox(height: 12),
                          _PendingChip(
                            action: pendingAction,
                            delayed: delayed,
                            delayedAfterSeconds: delayedAfterSeconds,
                          ),
                        ],
                      ],
                    ),
                  ),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(18),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        'Room Detail',
                        style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
                      ),
                      const SizedBox(height: 14),
                      _MetaRow(label: 'Stay', value: room.stayMode.isEmpty ? '-' : _titleCase(room.stayMode)),
                      _MetaRow(label: 'Case', value: room.caseCode.isEmpty ? '-' : room.caseCode),
                      _MetaRow(label: 'Floor', value: '${room.floor}'),
                      _MetaRow(
                        label: 'Device',
                        value: room.controllerOnline ? 'Online' : 'Offline',
                        valueColor: room.controllerOnline ? const Color(0xFF16A34A) : const Color(0xFFDC2626),
                      ),
                      if (room.remainingMinutes != null)
                        _MetaRow(label: 'Remain', value: '${room.remainingMinutes} min'),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(18),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        'Actions',
                        style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
                      ),
                      const SizedBox(height: 14),
                      SizedBox(
                        width: double.infinity,
                        child: FilledButton(
                          onPressed: isPending && !canRetry ? null : onOpenRoom,
                          style: FilledButton.styleFrom(
                            backgroundColor: const Color(0xFF115E59),
                            padding: const EdgeInsets.symmetric(vertical: 16),
                          ),
                          child: Text(
                            isPending && pendingAction == 'open_room'
                                ? (delayed ? 'Retry Open Room' : 'Sending...')
                                : 'Open Room',
                          ),
                        ),
                      ),
                      const SizedBox(height: 10),
                      SizedBox(
                        width: double.infinity,
                        child: FilledButton.tonal(
                          onPressed: isPending && !canRetry ? null : onService,
                          child: Text(
                            isPending && pendingAction == 'service'
                                ? (delayed ? 'Retry Service Flow' : 'Sending...')
                                : 'Service Flow',
                          ),
                        ),
                      ),
                      const SizedBox(height: 10),
                      SizedBox(
                        width: double.infinity,
                        child: OutlinedButton(
                          onPressed: isPending && !canRetry ? null : (isCleaning ? onFinishCleaning : onStartCleaning),
                          child: Text(
                            isPending && (pendingAction == 'start_cleaning' || pendingAction == 'finish_cleaning')
                                ? (delayed ? 'Retry Cleaning' : 'Sending...')
                                : (isCleaning ? 'Finish Cleaning' : 'Start Cleaning'),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

class _PendingChip extends StatelessWidget {
  const _PendingChip({
    required this.action,
    required this.delayed,
    required this.delayedAfterSeconds,
  });

  final String action;
  final bool delayed;
  final int delayedAfterSeconds;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: delayed ? const Color(0xFFFEE2E2) : const Color(0xFFFEF3C7),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: delayed ? const Color(0xFFEF4444) : const Color(0xFFF59E0B)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          const SizedBox(
            width: 12,
            height: 12,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
          const SizedBox(width: 8),
          Text(
            delayed
                ? 'Delayed > ${delayedAfterSeconds}s: ${_titleCase(action)}'
                : 'Sending ${_titleCase(action)}',
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
                  color: delayed ? const Color(0xFF991B1B) : const Color(0xFF92400E),
                  fontWeight: FontWeight.w700,
                ),
          ),
        ],
      ),
    );
  }
}

class _ServiceSheet extends StatefulWidget {
  const _ServiceSheet({required this.room});

  final Room room;

  @override
  State<_ServiceSheet> createState() => _ServiceSheetState();
}

class _ServiceSheetState extends State<_ServiceSheet> {
  String _service = 'Need towel';

  @override
  Widget build(BuildContext context) {
    const services = <String>['Need towel', 'Water refill', 'Mini bar check', 'Extend request'];

    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            'Service Flow ${widget.room.roomId}',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 8),
          Text(
            'Draft handheld service flow for room boy / housekeeping coordination.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
          ),
          const SizedBox(height: 18),
          RadioGroup<String>(
            groupValue: _service,
            onChanged: (String? value) {
              if (value == null) {
                return;
              }
              setState(() {
                _service = value;
              });
            },
            child: Column(
              children: <Widget>[
                for (final service in services)
                  RadioListTile<String>(
                    value: service,
                    title: Text(service),
                  ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            child: FilledButton(
              onPressed: () => Navigator.of(context).pop(_service),
              style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFF115E59),
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: const Text('Confirm Service'),
            ),
          ),
        ],
      ),
    );
  }
}

class _StatusTag extends StatelessWidget {
  const _StatusTag({required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final color = _statusColor(status);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        _titleCase(status),
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              color: color,
              fontWeight: FontWeight.w700,
            ),
      ),
    );
  }
}

class _MetaRow extends StatelessWidget {
  const _MetaRow({
    required this.label,
    required this.value,
    this.valueColor,
  });

  final String label;
  final String value;
  final Color? valueColor;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: Row(
        children: <Widget>[
          SizedBox(
            width: 64,
            child: Text(
              label,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(color: const Color(0xFF6B7280)),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: valueColor ?? const Color(0xFF1F2937),
                  ),
            ),
          ),
        ],
      ),
    );
  }
}

Color _statusColor(String status) {
  switch (status) {
    case 'cleaning':
      return const Color(0xFF0284C7);
    case 'vacant':
      return const Color(0xFF16A34A);
    case 'overstay':
      return const Color(0xFFEA580C);
    case 'occupied':
      return const Color(0xFF6B7280);
    default:
      return const Color(0xFF7C3AED);
  }
}

String _detail(Room room) {
  if (room.status == 'cleaning') {
    return 'Finish cleaning when the room is ready for cashier visibility.';
  }
  if (room.status == 'occupied') {
    return 'Use handheld for open room and quick service coordination.';
  }
  if (room.status == 'overstay') {
    return 'Avoid direct change without cashier confirmation.';
  }
  return 'Vacant room can be opened, inspected, or started for cleaning.';
}

String _titleCase(String value) {
  if (value.isEmpty) {
    return '-';
  }

  final words = value.split('_');
  return words
      .map((String word) => '${word.substring(0, 1).toUpperCase()}${word.substring(1)}')
      .join(' ');
}
