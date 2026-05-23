import 'package:flutter/material.dart';

import '../models/room.dart';
import '../state/room_store.dart';

class RoomBoardScreen extends StatefulWidget {
  const RoomBoardScreen({
    super.key,
    required this.store,
  });

  final RoomStore store;

  @override
  State<RoomBoardScreen> createState() => _RoomBoardScreenState();
}

class _RoomBoardScreenState extends State<RoomBoardScreen> {
  String _filter = 'all';
  String? _selectedRoomId;

  RoomStore get store => widget.store;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: store,
      builder: (BuildContext context, _) {
        final rooms = _filteredRooms(store.rooms, _filter);
        final selectedRoom = _selectedRoom(rooms);

        return Scaffold(
          body: Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: <Color>[
                  Color(0xFFFFF8EE),
                  Color(0xFFF8F4EC),
                ],
              ),
            ),
            child: SafeArea(
              child: Column(
                children: <Widget>[
                  if (store.error != null)
                    Padding(
                      padding: const EdgeInsets.fromLTRB(24, 16, 24, 0),
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
                      padding: const EdgeInsets.all(24),
                      child: LayoutBuilder(
                        builder: (BuildContext context, BoxConstraints constraints) {
                          final compact = constraints.maxWidth < 1220;

                          return Column(
                            children: <Widget>[
                              _CashierHero(
                                totalRooms: store.rooms.length,
                                vacantCount: store.rooms.where((Room room) => room.status == 'vacant').length,
                                occupiedCount: store.rooms.where((Room room) => room.status == 'occupied').length,
                                onRefresh: store.refresh,
                              ),
                              const SizedBox(height: 20),
                              Expanded(
                                child: compact
                                    ? ListView(
                                        children: <Widget>[
                                          _FilterBar(
                                            currentFilter: _filter,
                                            onChanged: (String value) {
                                              setState(() {
                                                _filter = value;
                                              });
                                            },
                                          ),
                                          const SizedBox(height: 16),
                                          if (selectedRoom != null)
                                            _RoomDetailPanel(
                                              room: selectedRoom,
                                              pendingAction: store.pendingActionFor(selectedRoom.roomId),
                                              delayed: store.isRoomDelayed(selectedRoom.roomId),
                                              delayedAfterSeconds: store.pendingTimeoutSeconds,
                                            ),
                                          if (selectedRoom != null) const SizedBox(height: 16),
                                          if (selectedRoom != null)
                                            _BillingPanel(
                                              room: selectedRoom,
                                              pendingAction: store.pendingActionFor(selectedRoom.roomId),
                                              delayed: store.isRoomDelayed(selectedRoom.roomId),
                                              delayedAfterSeconds: store.pendingTimeoutSeconds,
                                              onCheckIn: () => _openCheckInSheet(context, selectedRoom),
                                              onCheckOut: () => store.checkOut(selectedRoom),
                                              onExtend: () => _openExtendSheet(context, selectedRoom),
                                              onPayment: () => _openPaymentSheet(context, selectedRoom),
                                            ),
                                          if (selectedRoom != null) const SizedBox(height: 16),
                                          _RoomGrid(
                                            rooms: rooms,
                                            columns: constraints.maxWidth < 760 ? 1 : 2,
                                            selectedRoomId: selectedRoom?.roomId,
                                            pendingActionFor: store.pendingActionFor,
                                            delayedFor: store.isRoomDelayed,
                                            delayedAfterSeconds: store.pendingTimeoutSeconds,
                                            shrinkWrap: true,
                                            onTap: _selectRoom,
                                            onCheckIn: (Room room) => _openCheckInSheet(context, room),
                                            onCheckOut: store.checkOut,
                                          ),
                                        ],
                                      )
                                    : Row(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: <Widget>[
                                          Expanded(
                                            flex: 7,
                                            child: Column(
                                              children: <Widget>[
                                                _FilterBar(
                                                  currentFilter: _filter,
                                                  onChanged: (String value) {
                                                    setState(() {
                                                      _filter = value;
                                                    });
                                                  },
                                                ),
                                                const SizedBox(height: 16),
                                                Expanded(
                                                  child: _RoomGrid(
                                                    rooms: rooms,
                                                    columns: constraints.maxWidth > 1520 ? 4 : 3,
                                                    selectedRoomId: selectedRoom?.roomId,
                                                    pendingActionFor: store.pendingActionFor,
                                                    delayedFor: store.isRoomDelayed,
                                                    delayedAfterSeconds: store.pendingTimeoutSeconds,
                                                    onTap: _selectRoom,
                                                    onCheckIn: (Room room) => _openCheckInSheet(context, room),
                                                    onCheckOut: store.checkOut,
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ),
                                          const SizedBox(width: 20),
                                          SizedBox(
                                            width: 380,
                                            child: selectedRoom == null
                                                ? const _EmptyPanel()
                                                : ListView(
                                                    children: <Widget>[
                                                      _RoomDetailPanel(
                                                        room: selectedRoom,
                                                        pendingAction: store.pendingActionFor(selectedRoom.roomId),
                                                        delayed: store.isRoomDelayed(selectedRoom.roomId),
                                                        delayedAfterSeconds: store.pendingTimeoutSeconds,
                                                      ),
                                                      const SizedBox(height: 16),
                                                      _BillingPanel(
                                                        room: selectedRoom,
                                                        pendingAction: store.pendingActionFor(selectedRoom.roomId),
                                                        delayed: store.isRoomDelayed(selectedRoom.roomId),
                                                        delayedAfterSeconds: store.pendingTimeoutSeconds,
                                                        onCheckIn: () => _openCheckInSheet(context, selectedRoom),
                                                        onCheckOut: () => store.checkOut(selectedRoom),
                                                        onExtend: () => _openExtendSheet(context, selectedRoom),
                                                        onPayment: () => _openPaymentSheet(context, selectedRoom),
                                                      ),
                                                    ],
                                                  ),
                                          ),
                                        ],
                                      ),
                              ),
                            ],
                          );
                        },
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

  Room? _selectedRoom(List<Room> rooms) {
    if (rooms.isEmpty) {
      return null;
    }

    final match = rooms.where((Room room) => room.roomId == _selectedRoomId);
    return match.isNotEmpty ? match.first : rooms.first;
  }

  void _selectRoom(Room room) {
    setState(() {
      _selectedRoomId = room.roomId;
    });
  }

  List<Room> _filteredRooms(List<Room> rooms, String filter) {
    switch (filter) {
      case 'vacant':
        return rooms.where((Room room) => room.status == 'vacant').toList();
      case 'occupied':
        return rooms.where((Room room) => room.status == 'occupied').toList();
      case 'overstay':
        return rooms.where((Room room) => room.status == 'overstay').toList();
      default:
        return rooms;
    }
  }

  Future<void> _openCheckInSheet(BuildContext context, Room room) async {
    final stayMode = await showModalBottomSheet<String>(
      context: context,
      showDragHandle: true,
      backgroundColor: const Color(0xFFFFFBF4),
      builder: (BuildContext context) {
        return _CheckInSheet(
          room: room,
          initialStayMode: room.stayMode.isEmpty ? 'overnight' : room.stayMode,
        );
      },
    );

    if (stayMode == null || !mounted) {
      return;
    }

    await store.checkInWithStayMode(room, stayMode);
  }

  Future<void> _openExtendSheet(BuildContext context, Room room) async {
    final result = await showModalBottomSheet<_CashierFlowResult>(
      context: context,
      showDragHandle: true,
      backgroundColor: const Color(0xFFFFFBF4),
      builder: (BuildContext context) => _ExtendSheet(room: room),
    );

    if (!mounted || result == null) {
      return;
    }

    await store.extendStay(
      room,
      stayMode: result.stayMode ?? (room.stayMode.isEmpty ? 'temporary' : room.stayMode),
      extendMinutes: result.extendMinutes ?? 60,
      paymentAmount: result.amount ?? 0,
      paymentRef: result.reference ?? 'EXT-${room.roomId}-${DateTime.now().millisecondsSinceEpoch}',
    );
  }

  Future<void> _openPaymentSheet(BuildContext context, Room room) async {
    final result = await showModalBottomSheet<_CashierFlowResult>(
      context: context,
      showDragHandle: true,
      backgroundColor: const Color(0xFFFFFBF4),
      builder: (BuildContext context) => _PaymentSheet(room: room),
    );

    if (!mounted || result == null) {
      return;
    }

    await store.receivePayment(
      room,
      amount: result.amount ?? 0,
      method: result.method ?? 'Cash',
      reference: result.reference ?? 'PAY-${room.roomId}-${DateTime.now().millisecondsSinceEpoch}',
    );
  }
}

class _CashierHero extends StatelessWidget {
  const _CashierHero({
    required this.totalRooms,
    required this.vacantCount,
    required this.occupiedCount,
    required this.onRefresh,
  });

  final int totalRooms;
  final int vacantCount;
  final int occupiedCount;
  final Future<void> Function() onRefresh;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: const Color(0xFF7C2D12),
        borderRadius: BorderRadius.circular(30),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x227C2D12),
            blurRadius: 32,
            offset: Offset(0, 18),
          ),
        ],
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final stacked = constraints.maxWidth < 820;
          final metrics = Wrap(
            spacing: 12,
            runSpacing: 12,
            children: <Widget>[
              _HeroChip(label: 'Rooms', value: '$totalRooms'),
              _HeroChip(label: 'Vacant', value: '$vacantCount'),
              _HeroChip(label: 'Occupied', value: '$occupiedCount'),
              FilledButton.icon(
                onPressed: onRefresh,
                style: FilledButton.styleFrom(
                  backgroundColor: const Color(0xFFFBBF24),
                  foregroundColor: const Color(0xFF7C2D12),
                  padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 18),
                ),
                icon: const Icon(Icons.refresh),
                label: const Text('Refresh'),
              ),
            ],
          );

          return stacked
              ? Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    const _HeroCopy(),
                    const SizedBox(height: 18),
                    metrics,
                  ],
                )
              : Row(
                  children: <Widget>[
                    const Expanded(child: _HeroCopy()),
                    const SizedBox(width: 16),
                    metrics,
                  ],
                );
        },
      ),
    );
  }
}

class _HeroCopy extends StatelessWidget {
  const _HeroCopy();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          'Cashier Front Office',
          style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.w800,
              ),
        ),
        const SizedBox(height: 8),
        Text(
          'Check-in, payment, extend time, and checkout from a focused cashier workflow.',
          style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                color: const Color(0xFFFDE7D7),
              ),
        ),
      ],
    );
  }
}

class _HeroChip extends StatelessWidget {
  const _HeroChip({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 110,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFF9A3412),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label.toUpperCase(),
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
                  color: const Color(0xFFFED7AA),
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

class _FilterBar extends StatelessWidget {
  const _FilterBar({
    required this.currentFilter,
    required this.onChanged,
  });

  final String currentFilter;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Wrap(
          spacing: 10,
          runSpacing: 10,
          children: <Widget>[
            _FilterChip(label: 'All Rooms', selected: currentFilter == 'all', onTap: () => onChanged('all')),
            _FilterChip(label: 'Vacant', selected: currentFilter == 'vacant', onTap: () => onChanged('vacant')),
            _FilterChip(
              label: 'Occupied',
              selected: currentFilter == 'occupied',
              onTap: () => onChanged('occupied'),
            ),
            _FilterChip(
              label: 'Overstay',
              selected: currentFilter == 'overstay',
              onTap: () => onChanged('overstay'),
            ),
          ],
        ),
      ),
    );
  }
}

class _FilterChip extends StatelessWidget {
  const _FilterChip({
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
          color: selected ? const Color(0xFF7C2D12) : const Color(0xFFFFFBF4),
          borderRadius: BorderRadius.circular(999),
          border: Border.all(
            color: selected ? const Color(0xFF7C2D12) : const Color(0xFFE8D8C5),
          ),
        ),
        child: Text(
          label,
          style: Theme.of(context).textTheme.labelLarge?.copyWith(
                color: selected ? Colors.white : const Color(0xFF7C2D12),
                fontWeight: FontWeight.w700,
              ),
        ),
      ),
    );
  }
}

class _RoomGrid extends StatelessWidget {
  const _RoomGrid({
    required this.rooms,
    required this.columns,
    required this.pendingActionFor,
    required this.delayedFor,
    required this.delayedAfterSeconds,
    required this.onTap,
    required this.onCheckIn,
    required this.onCheckOut,
    this.selectedRoomId,
    this.shrinkWrap = false,
  });

  final List<Room> rooms;
  final int columns;
  final String? Function(String roomId) pendingActionFor;
  final bool Function(String roomId) delayedFor;
  final int delayedAfterSeconds;
  final ValueChanged<Room> onTap;
  final ValueChanged<Room> onCheckIn;
  final ValueChanged<Room> onCheckOut;
  final String? selectedRoomId;
  final bool shrinkWrap;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              'Room Selection',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 6),
            Text(
              'Tap a room to open detail, billing, and extend actions.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
            ),
            const SizedBox(height: 16),
            if (rooms.isEmpty)
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(24),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFFBF4),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: const Color(0xFFE8D8C5)),
                ),
                child: const Text('No rooms match this filter.'),
              )
            else if (shrinkWrap)
              GridView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                itemCount: rooms.length,
                gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: columns,
                  crossAxisSpacing: 14,
                  mainAxisSpacing: 14,
                  childAspectRatio: 1.08,
                ),
                itemBuilder: (BuildContext context, int index) {
                  final room = rooms[index];
                  return _CashierRoomTile(
                    room: room,
                    pendingAction: pendingActionFor(room.roomId),
                    delayed: delayedFor(room.roomId),
                    delayedAfterSeconds: delayedAfterSeconds,
                    selected: room.roomId == selectedRoomId,
                    onTap: () => onTap(room),
                    onCheckIn: () => onCheckIn(room),
                    onCheckOut: () => onCheckOut(room),
                  );
                },
              )
            else
              Expanded(
                child: GridView.builder(
                  itemCount: rooms.length,
                  gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: columns,
                    crossAxisSpacing: 14,
                    mainAxisSpacing: 14,
                    childAspectRatio: 1.08,
                  ),
                  itemBuilder: (BuildContext context, int index) {
                    final room = rooms[index];
                    return _CashierRoomTile(
                      room: room,
                      pendingAction: pendingActionFor(room.roomId),
                      delayed: delayedFor(room.roomId),
                      delayedAfterSeconds: delayedAfterSeconds,
                      selected: room.roomId == selectedRoomId,
                      onTap: () => onTap(room),
                      onCheckIn: () => onCheckIn(room),
                      onCheckOut: () => onCheckOut(room),
                    );
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _CashierRoomTile extends StatelessWidget {
  const _CashierRoomTile({
    required this.room,
    required this.pendingAction,
    required this.delayed,
    required this.delayedAfterSeconds,
    required this.selected,
    required this.onTap,
    required this.onCheckIn,
    required this.onCheckOut,
  });

  final Room room;
  final String? pendingAction;
  final bool delayed;
  final int delayedAfterSeconds;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback onCheckIn;
  final VoidCallback onCheckOut;

  @override
  Widget build(BuildContext context) {
    final color = _statusColor(room.status);
    final isPending = pendingAction != null;
    final canRetry = isPending && delayed;

    return InkWell(
      onTap: isPending && !canRetry ? null : onTap,
      borderRadius: BorderRadius.circular(22),
      child: Container(
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF0E6) : Colors.white,
          borderRadius: BorderRadius.circular(22),
          border: Border.all(
            color: selected ? const Color(0xFFB45309) : color.withValues(alpha: 0.22),
            width: selected ? 2 : 1,
          ),
        ),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Text(
                    room.roomId,
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
                  ),
                  const Spacer(),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: <Widget>[
                      _StatusBadge(status: room.status),
                      if (isPending) ...<Widget>[
                        const SizedBox(height: 8),
                        _PendingBadge(
                          action: pendingAction!,
                          delayed: delayed,
                          delayedAfterSeconds: delayedAfterSeconds,
                        ),
                      ],
                    ],
                  ),
                ],
              ),
              const SizedBox(height: 12),
              _MetaRow(label: 'Stay', value: room.stayMode.isEmpty ? '-' : room.stayMode),
              _MetaRow(label: 'Case', value: room.caseCode.isEmpty ? '-' : room.caseCode),
              _MetaRow(
                label: 'Remain',
                value: room.remainingMinutes == null ? '-' : '${room.remainingMinutes} min',
              ),
              if (isPending) ...<Widget>[
                const SizedBox(height: 10),
                Text(
                  delayed
                      ? 'Delayed > ${delayedAfterSeconds}s for ${_titleCase(pendingAction!)}. You can retry from this tile.'
                      : 'Waiting for ${_titleCase(pendingAction!)} confirmation.',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: delayed ? const Color(0xFF991B1B) : const Color(0xFF92400E),
                        fontWeight: FontWeight.w700,
                      ),
                ),
              ],
              const Spacer(),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: <Widget>[
                  FilledButton(
                    onPressed: isPending && !canRetry ? null : onCheckIn,
                    style: FilledButton.styleFrom(backgroundColor: const Color(0xFF7C2D12)),
                    child: Text(
                      isPending && pendingAction == 'check_in'
                          ? (delayed ? 'Retry Check-In' : 'Sending...')
                          : 'Check-In',
                    ),
                  ),
                  OutlinedButton(
                    onPressed: isPending && !canRetry ? null : onCheckOut,
                    child: Text(
                      isPending && pendingAction == 'check_out'
                          ? (delayed ? 'Retry Out' : 'Sending...')
                          : 'Out',
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
}

class _RoomDetailPanel extends StatelessWidget {
  const _RoomDetailPanel({
    required this.room,
    required this.pendingAction,
    required this.delayed,
    required this.delayedAfterSeconds,
  });

  final Room room;
  final String? pendingAction;
  final bool delayed;
  final int delayedAfterSeconds;

  @override
  Widget build(BuildContext context) {
    final isPending = pendingAction != null;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              'Room Detail ${room.roomId}',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 8),
            Text(
              'Cashier view for current stay, billing state, and service summary.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
            ),
            if (isPending) ...<Widget>[
              const SizedBox(height: 12),
              _PendingBadge(
                action: pendingAction!,
                delayed: delayed,
                delayedAfterSeconds: delayedAfterSeconds,
              ),
            ],
            const SizedBox(height: 16),
            _MetaRow(label: 'Status', value: _titleCase(room.status)),
            _MetaRow(label: 'Stay Mode', value: room.stayMode.isEmpty ? '-' : _titleCase(room.stayMode)),
            _MetaRow(label: 'Case', value: room.caseCode.isEmpty ? '-' : room.caseCode),
            _MetaRow(label: 'Floor', value: '${room.floor}'),
            _MetaRow(
              label: 'Device',
              value: room.controllerOnline ? 'Online' : 'Offline',
              valueColor: room.controllerOnline ? const Color(0xFF16A34A) : const Color(0xFFDC2626),
            ),
            _MetaRow(
              label: 'Remain',
              value: room.remainingMinutes == null ? '-' : '${room.remainingMinutes} min',
            ),
          ],
        ),
      ),
    );
  }
}

class _BillingPanel extends StatelessWidget {
  const _BillingPanel({
    required this.room,
    required this.pendingAction,
    required this.delayed,
    required this.delayedAfterSeconds,
    required this.onCheckIn,
    required this.onCheckOut,
    required this.onExtend,
    required this.onPayment,
  });

  final Room room;
  final String? pendingAction;
  final bool delayed;
  final int delayedAfterSeconds;
  final VoidCallback onCheckIn;
  final VoidCallback onCheckOut;
  final VoidCallback onExtend;
  final VoidCallback onPayment;

  @override
  Widget build(BuildContext context) {
    final baseCharge = room.status == 'occupied' ? 390 : 0;
    final extendCharge = room.remainingMinutes != null && room.remainingMinutes! < 60 ? 120 : 0;
    final total = baseCharge + extendCharge;
    final isPending = pendingAction != null;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              'Payment / Extend',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 8),
            Text(
              'Draft cashier flow for settlement and extending the current stay.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
            ),
            if (isPending) ...<Widget>[
              const SizedBox(height: 12),
              _PendingBadge(
                action: pendingAction!,
                delayed: delayed,
                delayedAfterSeconds: delayedAfterSeconds,
              ),
            ],
            const SizedBox(height: 16),
            _PriceRow(label: 'Room charge', value: baseCharge),
            _PriceRow(label: 'Extend recommendation', value: extendCharge),
            const Divider(height: 28),
            _PriceRow(label: 'Current total', value: total, strong: true),
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                onPressed: isPending && !delayed ? null : onPayment,
                style: FilledButton.styleFrom(
                  backgroundColor: const Color(0xFFB45309),
                  padding: const EdgeInsets.symmetric(vertical: 15),
                ),
                child: Text(
                  isPending && pendingAction == 'receive_payment'
                      ? (delayed ? 'Retry Payment' : 'Sending...')
                      : 'Receive Payment',
                ),
              ),
            ),
            const SizedBox(height: 10),
            SizedBox(
              width: double.infinity,
              child: FilledButton.tonal(
                onPressed: isPending && !delayed ? null : onExtend,
                child: Text(
                  isPending && pendingAction == 'extend_stay'
                      ? (delayed ? 'Retry Extend' : 'Sending...')
                      : 'Extend Stay',
                ),
              ),
            ),
            const SizedBox(height: 10),
            Row(
              children: <Widget>[
                Expanded(
                  child: OutlinedButton(
                    onPressed: isPending && !delayed ? null : onCheckIn,
                    child: Text(
                      isPending && pendingAction == 'check_in'
                          ? (delayed ? 'Retry Check-In' : 'Sending...')
                          : 'Check-In',
                    ),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: OutlinedButton(
                    onPressed: isPending && !delayed ? null : onCheckOut,
                    child: Text(
                      isPending && pendingAction == 'check_out'
                          ? (delayed ? 'Retry Check-Out' : 'Sending...')
                          : 'Check-Out',
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _PriceRow extends StatelessWidget {
  const _PriceRow({
    required this.label,
    required this.value,
    this.strong = false,
  });

  final String label;
  final int value;
  final bool strong;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Text(
              label,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: strong ? const Color(0xFF111827) : const Color(0xFF6B7280),
                    fontWeight: strong ? FontWeight.w700 : FontWeight.w500,
                  ),
            ),
          ),
          Text(
            '$value THB',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w800,
                  color: strong ? const Color(0xFF7C2D12) : const Color(0xFF111827),
                ),
          ),
        ],
      ),
    );
  }
}

class _PendingBadge extends StatelessWidget {
  const _PendingBadge({
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
        color: delayed ? const Color(0xFFFEE2E2) : const Color(0xFFFFF3E0),
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
                  color: delayed ? const Color(0xFF991B1B) : const Color(0xFF9A3412),
                  fontWeight: FontWeight.w700,
                ),
          ),
        ],
      ),
    );
  }
}

class _CheckInSheet extends StatefulWidget {
  const _CheckInSheet({
    required this.room,
    required this.initialStayMode,
  });

  final Room room;
  final String initialStayMode;

  @override
  State<_CheckInSheet> createState() => _CheckInSheetState();
}

class _CheckInSheetState extends State<_CheckInSheet> {
  late String _stayMode = widget.initialStayMode;

  @override
  Widget build(BuildContext context) {
    const stayModes = <String>['overnight', 'temporary', 'monthly', 'yearly'];

    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            'Check-In Room ${widget.room.roomId}',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 8),
          Text(
            'Choose the stay mode before confirming cashier check-in.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
          ),
          const SizedBox(height: 18),
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: stayModes
                .map(
                  (String mode) => ChoiceChip(
                    label: Text(_titleCase(mode)),
                    selected: _stayMode == mode,
                    onSelected: (_) {
                      setState(() {
                        _stayMode = mode;
                      });
                    },
                  ),
                )
                .toList(),
          ),
          const SizedBox(height: 20),
          SizedBox(
            width: double.infinity,
            child: FilledButton(
              onPressed: () => Navigator.of(context).pop(_stayMode),
              style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFF7C2D12),
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: const Text('Confirm Check-In'),
            ),
          ),
        ],
      ),
    );
  }
}

class _ExtendSheet extends StatefulWidget {
  const _ExtendSheet({required this.room});

  final Room room;

  @override
  State<_ExtendSheet> createState() => _ExtendSheetState();
}

class _ExtendSheetState extends State<_ExtendSheet> {
  String _duration = '1 hour';
  String _stayMode = 'temporary';

  @override
  Widget build(BuildContext context) {
    const options = <String>['1 hour', '2 hours', 'overnight'];
    const stayModes = <String>['temporary', 'overnight'];

    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            'Extend Stay ${widget.room.roomId}',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 8),
          Text(
            'Cashier can propose extend time before taking payment.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
          ),
          const SizedBox(height: 18),
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: stayModes
                .map(
                  (String mode) => ChoiceChip(
                    label: Text(_titleCase(mode)),
                    selected: _stayMode == mode,
                    onSelected: (_) {
                      setState(() {
                        _stayMode = mode;
                      });
                    },
                  ),
                )
                .toList(),
          ),
          const SizedBox(height: 12),
          ...options.map(
            (String option) => Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: _SelectionTile(
                label: option,
                selected: _duration == option,
                onTap: () {
                  setState(() {
                    _duration = option;
                  });
                },
              ),
            ),
          ),
          const SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            child: FilledButton(
              onPressed: () => Navigator.of(context).pop(
                _CashierFlowResult(
                  label: _duration,
                  extendMinutes: _minutesForDuration(_duration),
                  amount: _amountForDuration(_duration),
                  stayMode: _stayMode,
                  reference: 'EXT-${widget.room.roomId}-${DateTime.now().millisecondsSinceEpoch}',
                ),
              ),
              style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFFB45309),
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: const Text('Confirm Extend'),
            ),
          ),
        ],
      ),
    );
  }

  int _minutesForDuration(String duration) {
    switch (duration) {
      case '2 hours':
        return 120;
      case 'overnight':
        return 720;
      case '1 hour':
      default:
        return 60;
    }
  }

  double _amountForDuration(String duration) {
    switch (duration) {
      case '2 hours':
        return 240;
      case 'overnight':
        return 390;
      case '1 hour':
      default:
        return 120;
    }
  }
}

class _PaymentSheet extends StatefulWidget {
  const _PaymentSheet({required this.room});

  final Room room;

  @override
  State<_PaymentSheet> createState() => _PaymentSheetState();
}

class _PaymentSheetState extends State<_PaymentSheet> {
  String _method = 'Cash';

  @override
  Widget build(BuildContext context) {
    const methods = <String>['Cash', 'QR PromptPay', 'Transfer'];

    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            'Receive Payment ${widget.room.roomId}',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 8),
          Text(
            'Payment screen draft for cashier settlement flow.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
          ),
          const SizedBox(height: 18),
          ...methods.map(
            (String method) => Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: _SelectionTile(
                label: method,
                selected: _method == method,
                onTap: () {
                  setState(() {
                    _method = method;
                  });
                },
              ),
            ),
          ),
          const SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            child: FilledButton(
              onPressed: () => Navigator.of(context).pop(
                _CashierFlowResult(
                  label: _method,
                  amount: _estimatedAmount(widget.room),
                  method: _method,
                  reference: 'PAY-${widget.room.roomId}-${DateTime.now().millisecondsSinceEpoch}',
                ),
              ),
              style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFFB45309),
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: const Text('Confirm Payment'),
            ),
          ),
        ],
      ),
    );
  }

  double _estimatedAmount(Room room) {
    if (room.status == 'occupied') {
      return 390;
    }
    if (room.status == 'overstay') {
      return 510;
    }
    return 120;
  }
}

class _StatusBadge extends StatelessWidget {
  const _StatusBadge({required this.status});

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

class _SelectionTile extends StatelessWidget {
  const _SelectionTile({
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
      borderRadius: BorderRadius.circular(18),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF0E6) : Colors.white,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(
            color: selected ? const Color(0xFFB45309) : const Color(0xFFE5E7EB),
            width: selected ? 2 : 1,
          ),
        ),
        child: Row(
          children: <Widget>[
            Expanded(
              child: Text(
                label,
                style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700),
              ),
            ),
            Icon(
              selected ? Icons.check_circle_rounded : Icons.radio_button_unchecked_rounded,
              color: selected ? const Color(0xFFB45309) : const Color(0xFF9CA3AF),
            ),
          ],
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
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: <Widget>[
          SizedBox(
            width: 88,
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
                    color: valueColor ?? const Color(0xFF111827),
                  ),
            ),
          ),
        ],
      ),
    );
  }
}

class _EmptyPanel extends StatelessWidget {
  const _EmptyPanel();

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(22),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              'No Room Selected',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 10),
            Text(
              'Select a room from the board to open detail, billing, and extend flow.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: const Color(0xFF6B7280)),
            ),
          ],
        ),
      ),
    );
  }
}

class _CashierFlowResult {
  const _CashierFlowResult({
    required this.label,
    this.amount,
    this.method,
    this.reference,
    this.extendMinutes,
    this.stayMode,
  });

  final String label;
  final double? amount;
  final String? method;
  final String? reference;
  final int? extendMinutes;
  final String? stayMode;
}

Color _statusColor(String status) {
  switch (status) {
    case 'occupied':
      return const Color(0xFF16A34A);
    case 'overstay':
      return const Color(0xFFEA580C);
    case 'cleaning':
      return const Color(0xFF0284C7);
    case 'vacant':
      return const Color(0xFF6B7280);
    default:
      return const Color(0xFF7C3AED);
  }
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
