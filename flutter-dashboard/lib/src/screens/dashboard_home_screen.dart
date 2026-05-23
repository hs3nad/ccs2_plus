import 'package:flutter/material.dart';

import '../models/alert_item.dart';
import '../models/audit_item.dart';
import '../models/google_drive_config.dart';
import '../models/room.dart';
import '../state/room_store.dart';

class DashboardHomeScreen extends StatefulWidget {
  const DashboardHomeScreen({
    super.key,
    required this.store,
  });

  final RoomStore store;

  @override
  State<DashboardHomeScreen> createState() => _DashboardHomeScreenState();
}

class _DashboardHomeScreenState extends State<DashboardHomeScreen> {
  String _section = 'overview';

  RoomStore get store => widget.store;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: store,
      builder: (BuildContext context, _) {
        final rooms = store.rooms;
        final onlineCount =
            rooms.where((Room room) => room.controllerOnline).length;
        final offlineCount = rooms.length - onlineCount;

        return Scaffold(
          body: Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: <Color>[
                  Color(0xFFF6F2EA),
                  Color(0xFFE8EFE9),
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
                        builder:
                            (BuildContext context, BoxConstraints constraints) {
                          final compactHeight = constraints.maxHeight < 640;
                          void onSectionChanged(String value) {
                            setState(() {
                              _section = value;
                            });
                          }

                          return Column(
                            children: <Widget>[
                              if (compactHeight)
                                _CompactDashboardHeader(
                                  totalRooms: rooms.length,
                                  onlineCount: onlineCount,
                                  offlineCount: offlineCount,
                                  onRefresh: store.refresh,
                                  section: _section,
                                  onSectionChanged: onSectionChanged,
                                )
                              else
                                _DashboardHero(
                                  totalRooms: rooms.length,
                                  onlineCount: onlineCount,
                                  offlineCount: offlineCount,
                                  onRefresh: store.refresh,
                                  section: _section,
                                  onSectionChanged: onSectionChanged,
                                ),
                              const SizedBox(height: 16),
                              Expanded(
                                child: _sectionView(),
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

  Widget _sectionView() {
    switch (_section) {
      case 'reports':
        return _ReportsView(store: store);
      case 'audit':
        return _AuditView(store: store);
      case 'fraud':
        return _FraudView(store: store);
      case 'settings':
        return _SettingsView(store: store);
      case 'overview':
      default:
        return _OverviewView(store: store);
    }
  }
}

class _CompactDashboardHeader extends StatelessWidget {
  const _CompactDashboardHeader({
    required this.totalRooms,
    required this.onlineCount,
    required this.offlineCount,
    required this.onRefresh,
    required this.section,
    required this.onSectionChanged,
  });

  final int totalRooms;
  final int onlineCount;
  final int offlineCount;
  final Future<void> Function() onRefresh;
  final String section;
  final ValueChanged<String> onSectionChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFF12372A),
        borderRadius: BorderRadius.circular(18),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x2212372A),
            blurRadius: 18,
            offset: Offset(0, 10),
          ),
        ],
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: _DashboardNav(
                section: section,
                onChanged: onSectionChanged,
              ),
            ),
          ),
          const SizedBox(width: 10),
          Text(
            '$onlineCount/$totalRooms',
            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                  color: Colors.white,
                  fontWeight: FontWeight.w800,
                ),
          ),
          const SizedBox(width: 8),
          IconButton.filled(
            onPressed: onRefresh,
            style: IconButton.styleFrom(
              backgroundColor: offlineCount > 0
                  ? const Color(0xFFEA580C)
                  : const Color(0xFFE9C46A),
              foregroundColor: const Color(0xFF12372A),
            ),
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
          ),
        ],
      ),
    );
  }
}

class _DashboardHero extends StatelessWidget {
  const _DashboardHero({
    required this.totalRooms,
    required this.onlineCount,
    required this.offlineCount,
    required this.onRefresh,
    required this.section,
    required this.onSectionChanged,
  });

  final int totalRooms;
  final int onlineCount;
  final int offlineCount;
  final Future<void> Function() onRefresh;
  final String section;
  final ValueChanged<String> onSectionChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: const Color(0xFF12372A),
        borderRadius: BorderRadius.circular(30),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x2212372A),
            blurRadius: 30,
            offset: Offset(0, 18),
          ),
        ],
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final stacked = constraints.maxWidth < 920;

          final metrics = Wrap(
            spacing: 12,
            runSpacing: 12,
            children: <Widget>[
              _HeroMetric(label: 'Rooms', value: '$totalRooms'),
              _HeroMetric(label: 'Online', value: '$onlineCount'),
              _HeroMetric(label: 'Offline', value: '$offlineCount'),
              FilledButton.icon(
                onPressed: onRefresh,
                style: FilledButton.styleFrom(
                  backgroundColor: const Color(0xFFE9C46A),
                  foregroundColor: const Color(0xFF12372A),
                  padding:
                      const EdgeInsets.symmetric(horizontal: 18, vertical: 18),
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
                    _DashboardNav(
                        section: section, onChanged: onSectionChanged),
                    const SizedBox(height: 18),
                    const _HeroCopy(),
                    const SizedBox(height: 18),
                    metrics,
                  ],
                )
              : Row(
                  children: <Widget>[
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          _DashboardNav(
                              section: section, onChanged: onSectionChanged),
                          const SizedBox(height: 18),
                          const _HeroCopy(),
                        ],
                      ),
                    ),
                    const SizedBox(width: 20),
                    metrics,
                  ],
                );
        },
      ),
    );
  }
}

class _DashboardNav extends StatelessWidget {
  const _DashboardNav({
    required this.section,
    required this.onChanged,
  });

  final String section;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 10,
      runSpacing: 10,
      children: <Widget>[
        _NavPill(
          icon: Icons.grid_view_rounded,
          label: 'Overview',
          selected: section == 'overview',
          onTap: () => onChanged('overview'),
        ),
        _NavPill(
          icon: Icons.insert_chart_outlined_rounded,
          label: 'Reports',
          selected: section == 'reports',
          onTap: () => onChanged('reports'),
        ),
        _NavPill(
          icon: Icons.fact_check_outlined,
          label: 'Audit',
          selected: section == 'audit',
          onTap: () => onChanged('audit'),
        ),
        _NavPill(
          icon: Icons.shield_outlined,
          label: 'Anti-Fraud',
          selected: section == 'fraud',
          onTap: () => onChanged('fraud'),
        ),
        _NavPill(
          icon: Icons.settings_outlined,
          label: 'Settings',
          selected: section == 'settings',
          onTap: () => onChanged('settings'),
        ),
      ],
    );
  }
}

class _NavPill extends StatelessWidget {
  const _NavPill({
    required this.icon,
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(999),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFE9C46A) : const Color(0xFF1A4A39),
          borderRadius: BorderRadius.circular(999),
          border: Border.all(
              color:
                  selected ? const Color(0xFFE9C46A) : const Color(0xFF2C6A52)),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(
              icon,
              size: 18,
              color:
                  selected ? const Color(0xFF12372A) : const Color(0xFFD5E7DE),
            ),
            const SizedBox(width: 8),
            Text(
              label,
              style: Theme.of(context).textTheme.labelLarge?.copyWith(
                    color: selected ? const Color(0xFF12372A) : Colors.white,
                    fontWeight: FontWeight.w700,
                  ),
            ),
          ],
        ),
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
          'Manager Dashboard',
          style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.w800,
              ),
        ),
        const SizedBox(height: 8),
        Text(
          'Realtime occupancy, reporting, audit review, and anti-fraud monitoring in one surface.',
          style: Theme.of(context)
              .textTheme
              .bodyLarge
              ?.copyWith(color: const Color(0xFFD5E7DE)),
        ),
      ],
    );
  }
}

class _HeroMetric extends StatelessWidget {
  const _HeroMetric({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 110,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: const Color(0xFF1A4A39),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label.toUpperCase(),
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
                  color: const Color(0xFFAED0C0),
                  letterSpacing: 1.2,
                ),
          ),
          const SizedBox(height: 8),
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

class _OverviewView extends StatelessWidget {
  const _OverviewView({required this.store});

  final RoomStore store;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final compact = constraints.maxWidth < 1180;

        if (compact) {
          return ListView(
            children: <Widget>[
              _SummaryStrip(store: store),
              const SizedBox(height: 20),
              _RoomBoard(
                  store: store,
                  columns: constraints.maxWidth < 760 ? 1 : 2,
                  shrinkWrap: true),
              const SizedBox(height: 20),
              _OperationsPanel(store: store, compact: true),
            ],
          );
        }

        return Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Expanded(
              flex: 7,
              child: Column(
                children: <Widget>[
                  _SummaryStrip(store: store),
                  const SizedBox(height: 20),
                  Expanded(
                    child: _RoomBoard(
                      store: store,
                      columns: constraints.maxWidth > 1500 ? 4 : 3,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 20),
            SizedBox(
              width: 320,
              child: _OperationsPanel(store: store),
            ),
          ],
        );
      },
    );
  }
}

class _ReportsView extends StatelessWidget {
  const _ReportsView({required this.store});

  final RoomStore store;

  @override
  Widget build(BuildContext context) {
    final occupied = store.occupiedCount;
    final cleaning = store.cleaningCount;
    final total = store.rooms.isEmpty ? 1 : store.rooms.length;
    final occupancyRate = ((occupied / total) * 100).round();

    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final cards = <Widget>[
          _ReportCard(
            title: 'Occupancy Rate',
            value: '$occupancyRate%',
            note: 'Live estimate from current room state',
            color: const Color(0xFF16A34A),
          ),
          _ReportCard(
            title: 'Rooms Cleaning',
            value: '$cleaning rooms',
            note: 'Housekeeping workload snapshot',
            color: const Color(0xFF0284C7),
          ),
          _ReportCard(
            title: 'Estimated Revenue',
            value: '${occupied * 390} THB',
            note: 'UI draft placeholder formula',
            color: const Color(0xFFB45309),
          ),
        ];

        return ListView(
          children: <Widget>[
            if (constraints.maxWidth < 980)
              ...cards.expand(
                  (Widget card) => <Widget>[card, const SizedBox(height: 16)])
            else
              Row(
                children: <Widget>[
                  Expanded(child: cards[0]),
                  const SizedBox(width: 16),
                  Expanded(child: cards[1]),
                  const SizedBox(width: 16),
                  Expanded(child: cards[2]),
                ],
              ),
            const SizedBox(height: 20),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      'Daily Trend',
                      style: Theme.of(context)
                          .textTheme
                          .titleLarge
                          ?.copyWith(fontWeight: FontWeight.w800),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Visual draft for period / record / summary reporting.',
                      style: Theme.of(context)
                          .textTheme
                          .bodyMedium
                          ?.copyWith(color: const Color(0xFF64748B)),
                    ),
                    const SizedBox(height: 20),
                    const _BarChart(),
                  ],
                ),
              ),
            ),
          ],
        );
      },
    );
  }
}

class _SettingsView extends StatefulWidget {
  const _SettingsView({required this.store});

  final RoomStore store;

  @override
  State<_SettingsView> createState() => _SettingsViewState();
}

class _SettingsViewState extends State<_SettingsView> {
  final TextEditingController _credentialsPathController =
      TextEditingController();
  final TextEditingController _folderIdController = TextEditingController();
  bool _enabled = false;
  bool _isLoading = true;
  bool _isSaving = false;
  String? _message;
  String? _error;
  String _updatedAt = '';

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _credentialsPathController.dispose();
    _folderIdController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _message = null;
    });
    try {
      final config = await widget.store.fetchGoogleDriveConfig();
      _applyConfig(config);
    } catch (err) {
      _error = err.toString();
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _save() async {
    setState(() {
      _isSaving = true;
      _error = null;
      _message = null;
    });
    try {
      final config = await widget.store.saveGoogleDriveConfig(
        GoogleDriveConfig(
          credentialsPath: _credentialsPathController.text.trim(),
          folderId: _folderIdController.text.trim(),
          enabled: _enabled,
          configured: _credentialsPathController.text.trim().isNotEmpty &&
              _folderIdController.text.trim().isNotEmpty,
          updatedAt: _updatedAt,
        ),
      );
      _applyConfig(config);
      _message = 'Google Drive export config saved.';
    } catch (err) {
      _error = err.toString();
    } finally {
      if (mounted) {
        setState(() {
          _isSaving = false;
        });
      }
    }
  }

  void _applyConfig(GoogleDriveConfig config) {
    _credentialsPathController.text = config.credentialsPath;
    _folderIdController.text = config.folderId;
    _enabled = config.enabled;
    _updatedAt = config.updatedAt;
  }

  @override
  Widget build(BuildContext context) {
    return ListView(
      children: <Widget>[
        Card(
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Row(
                  children: <Widget>[
                    Container(
                      width: 44,
                      height: 44,
                      decoration: BoxDecoration(
                        color: const Color(0xFFE8F3EF),
                        borderRadius: BorderRadius.circular(14),
                      ),
                      child: const Icon(Icons.cloud_upload_outlined,
                          color: Color(0xFF0F766E)),
                    ),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          Text(
                            'Google Drive Export',
                            style: Theme.of(context)
                                .textTheme
                                .titleLarge
                                ?.copyWith(fontWeight: FontWeight.w800),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            _updatedAt.isEmpty
                                ? 'Daily and monthly Excel export destination.'
                                : 'Last updated $_updatedAt',
                            style: Theme.of(context)
                                .textTheme
                                .bodyMedium
                                ?.copyWith(color: const Color(0xFF64748B)),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      onPressed: _isLoading ? null : _load,
                      icon: const Icon(Icons.refresh),
                      tooltip: 'Reload config',
                    ),
                  ],
                ),
                const SizedBox(height: 22),
                if (_isLoading)
                  const LinearProgressIndicator()
                else ...<Widget>[
                  SwitchListTile(
                    value: _enabled,
                    onChanged: (bool value) {
                      setState(() {
                        _enabled = value;
                      });
                    },
                    contentPadding: EdgeInsets.zero,
                    title: const Text('Enable Google Drive upload'),
                    subtitle: const Text(
                        'When disabled, Excel files are kept in the backend export directory.'),
                  ),
                  const SizedBox(height: 14),
                  TextField(
                    controller: _credentialsPathController,
                    decoration: const InputDecoration(
                      labelText: 'Service account JSON path',
                      hintText: '/secure/ccs2plus-drive.json',
                      border: OutlineInputBorder(),
                      prefixIcon: Icon(Icons.key_outlined),
                    ),
                  ),
                  const SizedBox(height: 14),
                  TextField(
                    controller: _folderIdController,
                    decoration: const InputDecoration(
                      labelText: 'Google Drive folder ID',
                      hintText: '1AbCDefGhIjK...',
                      border: OutlineInputBorder(),
                      prefixIcon: Icon(Icons.folder_outlined),
                    ),
                  ),
                  const SizedBox(height: 18),
                  if (_error != null)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Text(_error!,
                          style: const TextStyle(color: Color(0xFFDC2626))),
                    ),
                  if (_message != null)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Text(_message!,
                          style: const TextStyle(color: Color(0xFF047857))),
                    ),
                  Align(
                    alignment: Alignment.centerRight,
                    child: FilledButton.icon(
                      onPressed: _isSaving ? null : _save,
                      icon: _isSaving
                          ? const SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.save_outlined),
                      label: Text(_isSaving ? 'Saving' : 'Save Config'),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ],
    );
  }
}

class _AuditView extends StatelessWidget {
  const _AuditView({required this.store});

  final RoomStore store;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: CustomScrollView(
          slivers: <Widget>[
            SliverToBoxAdapter(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    'Audit Trail',
                    style: Theme.of(context)
                        .textTheme
                        .titleLarge
                        ?.copyWith(fontWeight: FontWeight.w800),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Draft event list for user action review and room timeline.',
                    style: Theme.of(context)
                        .textTheme
                        .bodyMedium
                        ?.copyWith(color: const Color(0xFF64748B)),
                  ),
                  const SizedBox(height: 20),
                ],
              ),
            ),
            if (store.auditItems.isEmpty)
              const SliverToBoxAdapter(
                child: _EmptyPanel(
                  title: 'No audit events yet',
                  detail:
                      'Cashier, housekeeping, and service actions will appear here.',
                ),
              )
            else
              SliverList.separated(
                itemCount: store.auditItems.length,
                separatorBuilder: (_, __) => const SizedBox(height: 12),
                itemBuilder: (BuildContext context, int index) {
                  final entry = store.auditItems[index];
                  return Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF8FAFC),
                      borderRadius: BorderRadius.circular(18),
                      border: Border.all(color: const Color(0xFFE2E8F0)),
                    ),
                    child: InkWell(
                      onTap: () => _showAuditDetail(context, entry),
                      borderRadius: BorderRadius.circular(18),
                      child: LayoutBuilder(
                        builder:
                            (BuildContext context, BoxConstraints constraints) {
                          if (constraints.maxWidth < 760) {
                            return Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: <Widget>[
                                Text(
                                  '${_shortTimestamp(entry.timestamp)}  •  Room ${entry.roomId}',
                                  style: const TextStyle(
                                      fontWeight: FontWeight.w700),
                                ),
                                const SizedBox(height: 6),
                                Text(
                                    'Actor: ${entry.actor.isEmpty ? '-' : entry.actor}'),
                                Text('Action: ${_titleCase(entry.action)}'),
                                Text('Result: ${_titleCase(entry.result)}'),
                              ],
                            );
                          }
                          return Row(
                            children: <Widget>[
                              Expanded(
                                  flex: 2,
                                  child:
                                      Text(_shortTimestamp(entry.timestamp))),
                              Expanded(flex: 2, child: Text(entry.roomId)),
                              Expanded(
                                  flex: 2,
                                  child: Text(entry.actor.isEmpty
                                      ? '-'
                                      : entry.actor)),
                              Expanded(
                                  flex: 3,
                                  child: Text(_titleCase(entry.action))),
                              Expanded(
                                flex: 2,
                                child: Text(
                                  _titleCase(entry.result),
                                  style: const TextStyle(
                                      fontWeight: FontWeight.w700),
                                ),
                              ),
                            ],
                          );
                        },
                      ),
                    ),
                  );
                },
              ),
          ],
        ),
      ),
    );
  }
}

class _FraudView extends StatelessWidget {
  const _FraudView({required this.store});

  final RoomStore store;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              'Command ON != Actual ON',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 8),
            Text(
              'UI draft for anti-fraud monitoring, suspicious cases, and operator review.',
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: const Color(0xFF64748B)),
            ),
            const SizedBox(height: 18),
            Expanded(
              child: store.alertItems.isEmpty
                  ? const _EmptyPanel(
                      title: 'No suspicious events in current snapshot',
                      detail:
                          'This section will surface anti-fraud rules from backend alerts.',
                    )
                  : ListView.separated(
                      itemCount: store.alertItems.length,
                      separatorBuilder: (_, __) => const SizedBox(height: 12),
                      itemBuilder: (BuildContext context, int index) {
                        return _AlertTile(alert: store.alertItems[index]);
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}


class _SummaryStrip extends StatelessWidget {
  const _SummaryStrip({required this.store});

  final RoomStore store;

  @override
  Widget build(BuildContext context) {
    final cards = <Widget>[
      _SummaryCard(
        label: 'Vacant',
        value: store.vacantCount,
        color: const Color(0xFF6B7280),
        note: 'Ready for new stay',
      ),
      _SummaryCard(
        label: 'Occupied',
        value: store.occupiedCount,
        color: const Color(0xFF16A34A),
        note: 'Active guest rooms',
      ),
      _SummaryCard(
        label: 'Cleaning',
        value: store.cleaningCount,
        color: const Color(0xFF0284C7),
        note: 'Housekeeping in progress',
      ),
      _SummaryCard(
        label: 'Overstay',
        value: store.overstayCount,
        color: const Color(0xFFEA580C),
        note: 'Needs cashier attention',
      ),
    ];

    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        if (constraints.maxWidth < 880) {
          return Wrap(
            spacing: 12,
            runSpacing: 12,
            children: cards
                .map(
                  (Widget card) => SizedBox(
                    width: constraints.maxWidth < 560
                        ? constraints.maxWidth
                        : (constraints.maxWidth - 12) / 2,
                    child: card,
                  ),
                )
                .toList(),
          );
        }

        return Row(
          children: cards
              .expand(
                (Widget card) => <Widget>[
                  Expanded(child: card),
                  if (card != cards.last) const SizedBox(width: 12),
                ],
              )
              .toList(),
        );
      },
    );
  }
}

class _SummaryCard extends StatelessWidget {
  const _SummaryCard({
    required this.label,
    required this.value,
    required this.color,
    required this.note,
  });

  final String label;
  final int value;
  final Color color;
  final String note;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Container(
              width: 42,
              height: 8,
              decoration: BoxDecoration(
                color: color,
                borderRadius: BorderRadius.circular(999),
              ),
            ),
            const SizedBox(height: 14),
            Text(label,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(fontWeight: FontWeight.w700)),
            const SizedBox(height: 6),
            Text(
              '$value',
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w800,
                    color: color,
                  ),
            ),
            const SizedBox(height: 4),
            Text(note,
                style: Theme.of(context)
                    .textTheme
                    .bodyMedium
                    ?.copyWith(color: const Color(0xFF64748B))),
          ],
        ),
      ),
    );
  }
}

class _RoomBoard extends StatelessWidget {
  const _RoomBoard({
    required this.store,
    required this.columns,
    this.shrinkWrap = false,
  });

  final RoomStore store;
  final int columns;
  final bool shrinkWrap;

  @override
  Widget build(BuildContext context) {
    if (store.isLoading && store.rooms.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              'Live Room Board',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 6),
            Text(
              'Realtime room status and manager monitoring view.',
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: const Color(0xFF64748B)),
            ),
            const SizedBox(height: 18),
            if (shrinkWrap)
              GridView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                itemCount: store.rooms.length,
                gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: columns,
                  crossAxisSpacing: 14,
                  mainAxisSpacing: 14,
                  childAspectRatio: 1.18,
                ),
                itemBuilder: (BuildContext context, int index) {
                  return _DashboardRoomTile(room: store.rooms[index]);
                },
              )
            else
              Expanded(
                child: GridView.builder(
                  itemCount: store.rooms.length,
                  gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: columns,
                    crossAxisSpacing: 14,
                    mainAxisSpacing: 14,
                    childAspectRatio: 1.18,
                  ),
                  itemBuilder: (BuildContext context, int index) {
                    return _DashboardRoomTile(room: store.rooms[index]);
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _DashboardRoomTile extends StatelessWidget {
  const _DashboardRoomTile({required this.room});

  final Room room;

  @override
  Widget build(BuildContext context) {
    final statusColor = _colorForStatus(room.status);

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: statusColor.withValues(alpha: 0.22)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  decoration: BoxDecoration(
                    color: statusColor.withValues(alpha: 0.12),
                    borderRadius: BorderRadius.circular(14),
                  ),
                  child: Text(
                    room.roomId,
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          color: statusColor,
                          fontWeight: FontWeight.w800,
                        ),
                  ),
                ),
                const Spacer(),
                _StatusPill(status: room.status),
              ],
            ),
            const SizedBox(height: 14),
            _MetaLine(
                label: 'Stay',
                value: room.stayMode.isEmpty ? '-' : room.stayMode),
            _MetaLine(
                label: 'Case',
                value: room.caseCode.isEmpty ? '-' : room.caseCode),
            _MetaLine(label: 'Floor', value: '${room.floor}'),
            _MetaLine(
              label: 'Device',
              value: room.controllerOnline ? 'Online' : 'Offline',
              valueColor: room.controllerOnline
                  ? const Color(0xFF16A34A)
                  : const Color(0xFFDC2626),
            ),
            _MetaLine(
              label: 'Remain',
              value: room.remainingMinutes == null
                  ? '-'
                  : '${room.remainingMinutes} min',
            ),
          ],
        ),
      ),
    );
  }
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final color = _colorForStatus(status);

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

class _OperationsPanel extends StatelessWidget {
  const _OperationsPanel({
    required this.store,
    this.compact = false,
  });

  final RoomStore store;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final notePadding = compact ? 14.0 : 16.0;
    final noteGap = compact ? 8.0 : 12.0;
    final noteTiles = <Widget>[
      const _NoteTile(
        title: 'Reports',
        detail:
            'Period / record / summary flow is now represented in the report view.',
      ),
      const _NoteTile(
        title: 'Audit',
        detail:
            'Each room can become an action trail entry for review and investigation.',
      ),
      const _NoteTile(
        title: 'Anti-Fraud',
        detail:
            'Offline and overstay patterns are surfaced as suspicious signals in the draft.',
      ),
    ];
    final notesCard = Card(
      child: Padding(
        padding: EdgeInsets.all(notePadding),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              'Manager Notes',
              style: Theme.of(context)
                  .textTheme
                  .titleMedium
                  ?.copyWith(fontWeight: FontWeight.w800),
            ),
            SizedBox(height: noteGap),
            if (compact)
              ...noteTiles
            else
              Expanded(
                child: Column(
                  children: <Widget>[
                    Expanded(child: noteTiles[0]),
                    const SizedBox(height: 10),
                    Expanded(child: noteTiles[1]),
                    const SizedBox(height: 10),
                    Expanded(child: noteTiles[2]),
                  ],
                ),
              ),
          ],
        ),
      ),
    );

    return Column(
      children: <Widget>[
        Card(
          child: Padding(
            padding: const EdgeInsets.all(18),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  'Manager Focus',
                  style: Theme.of(context)
                      .textTheme
                      .titleLarge
                      ?.copyWith(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 8),
                Text(
                  'Snapshot for today before diving into reports or audit.',
                  style: Theme.of(context)
                      .textTheme
                      .bodyMedium
                      ?.copyWith(color: const Color(0xFF64748B)),
                ),
                const SizedBox(height: 18),
                _FocusLine(
                    label: 'Temporary stays',
                    value:
                        '${store.rooms.where((Room room) => room.stayMode == 'temporary').length}'),
                _FocusLine(
                    label: 'Overnight stays',
                    value:
                        '${store.rooms.where((Room room) => room.stayMode == 'overnight').length}'),
                _FocusLine(
                    label: 'Overstay alerts', value: '${store.overstayCount}'),
                _FocusLine(
                    label: 'Rooms cleaning', value: '${store.cleaningCount}'),
              ],
            ),
          ),
        ),
        const SizedBox(height: 20),
        if (compact) notesCard else Expanded(child: notesCard),
      ],
    );
  }
}

class _FocusLine extends StatelessWidget {
  const _FocusLine({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Text(label,
                style: Theme.of(context)
                    .textTheme
                    .bodyMedium
                    ?.copyWith(color: const Color(0xFF475569))),
          ),
          Text(value,
              style: Theme.of(context)
                  .textTheme
                  .titleMedium
                  ?.copyWith(fontWeight: FontWeight.w800)),
        ],
      ),
    );
  }
}

class _ReportCard extends StatelessWidget {
  const _ReportCard({
    required this.title,
    required this.value,
    required this.note,
    required this.color,
  });

  final String title;
  final String value;
  final String note;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(title,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(fontWeight: FontWeight.w700)),
            const SizedBox(height: 10),
            Text(
              value,
              style: Theme.of(context)
                  .textTheme
                  .headlineMedium
                  ?.copyWith(color: color, fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 6),
            Text(note,
                style: Theme.of(context)
                    .textTheme
                    .bodyMedium
                    ?.copyWith(color: const Color(0xFF64748B))),
          ],
        ),
      ),
    );
  }
}

class _BarChart extends StatelessWidget {
  const _BarChart();

  @override
  Widget build(BuildContext context) {
    const values = <double>[0.42, 0.55, 0.68, 0.61, 0.74, 0.80, 0.72];
    const labels = <String>['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

    return SizedBox(
      height: 220,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: List<Widget>.generate(values.length, (int index) {
          return Expanded(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 6),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  Expanded(
                    child: Align(
                      alignment: Alignment.bottomCenter,
                      child: Container(
                        height: 180 * values[index],
                        decoration: BoxDecoration(
                          color: const Color(0xFF12372A),
                          borderRadius: BorderRadius.circular(16),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 10),
                  Text(labels[index]),
                ],
              ),
            ),
          );
        }),
      ),
    );
  }
}

class _AlertTile extends StatelessWidget {
  const _AlertTile({required this.alert});

  final AlertItem alert;

  @override
  Widget build(BuildContext context) {
    final color = alert.severity == 'high'
        ? const Color(0xFFDC2626)
        : const Color(0xFFEA580C);

    return Container(
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: color.withValues(alpha: 0.24)),
      ),
      child: InkWell(
        onTap: () => _showFraudDetail(context, alert),
        borderRadius: BorderRadius.circular(18),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                alert.title,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(color: color, fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 8),
              Text(alert.detail),
              const SizedBox(height: 10),
              Text(
                'Rule: ${alert.ruleCode}  •  Room ${alert.roomId}',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: const Color(0xFF64748B),
                      fontWeight: FontWeight.w600,
                    ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _MetaLine extends StatelessWidget {
  const _MetaLine({
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
            width: 52,
            child: Text(label,
                style: Theme.of(context)
                    .textTheme
                    .bodySmall
                    ?.copyWith(color: const Color(0xFF64748B))),
          ),
          Expanded(
            child: Text(
              value,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: valueColor ?? const Color(0xFF1E293B),
                  ),
            ),
          ),
        ],
      ),
    );
  }
}

class _NoteTile extends StatelessWidget {
  const _NoteTile({
    required this.title,
    required this.detail,
  });

  final String title;
  final String detail;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 0),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 10),
        decoration: BoxDecoration(
          color: const Color(0xFFF8FAFC),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: const Color(0xFFE2E8F0)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(title,
                style: Theme.of(context)
                    .textTheme
                    .labelLarge
                    ?.copyWith(fontWeight: FontWeight.w800)),
            const SizedBox(height: 3),
            Text(
              detail,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: const Color(0xFF475569), height: 1.18),
            ),
          ],
        ),
      ),
    );
  }
}

class _EmptyPanel extends StatelessWidget {
  const _EmptyPanel({
    required this.title,
    required this.detail,
  });

  final String title;
  final String detail;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            title,
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 8),
          Text(
            detail,
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(color: const Color(0xFF64748B)),
          ),
        ],
      ),
    );
  }
}

void _showAuditDetail(BuildContext context, AuditItem entry) {
  showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    backgroundColor: const Color(0xFFFAFBFC),
    builder: (BuildContext context) {
      return Padding(
        padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
        child: ListView(
          shrinkWrap: true,
          children: <Widget>[
            Text(
              'Audit Log Detail',
              style: Theme.of(context)
                  .textTheme
                  .headlineSmall
                  ?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 8),
            Text(
              'Closer to the event-detail mockup with identity, action, and location context.',
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: const Color(0xFF64748B)),
            ),
            const SizedBox(height: 18),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(18),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    _DetailLine(
                        label: 'Actor',
                        value: entry.actor.isEmpty ? '-' : entry.actor),
                    _DetailLine(
                        label: 'Action', value: _titleCase(entry.action)),
                    _DetailLine(
                        label: 'Time', value: _fullTimestamp(entry.timestamp)),
                    _DetailLine(label: 'Room', value: entry.roomId),
                    _DetailLine(
                        label: 'Stay Mode',
                        value: entry.stayMode.isEmpty
                            ? '-'
                            : _titleCase(entry.stayMode)),
                    _DetailLine(
                        label: 'Case',
                        value: entry.caseCode.isEmpty ? '-' : entry.caseCode),
                    _DetailLine(
                        label: 'Result', value: _titleCase(entry.result)),
                    _DetailLine(
                        label: 'Detail',
                        value: entry.detail.isEmpty ? '-' : entry.detail),
                    const _DetailLine(
                        label: 'Handheld ID', value: 'Captured by backend'),
                    const _DetailLine(
                        label: 'GPS', value: 'Not collected in current scope'),
                    const _DetailLine(
                        label: 'Device IP',
                        value: 'Reserved for future capture'),
                    const _DetailLine(label: 'Network', value: 'REST / Wi-Fi'),
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
                    Text('Location Snapshot',
                        style: Theme.of(context)
                            .textTheme
                            .titleMedium
                            ?.copyWith(fontWeight: FontWeight.w800)),
                    const SizedBox(height: 12),
                    Container(
                      height: 180,
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: <Color>[Color(0xFFE0F2FE), Color(0xFFDCFCE7)],
                        ),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: const Center(
                        child: Icon(Icons.location_on_outlined,
                            size: 48, color: Color(0xFF0F766E)),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      );
    },
  );
}

void _showFraudDetail(BuildContext context, AlertItem alert) {
  final color = alert.severity == 'high'
      ? const Color(0xFFDC2626)
      : const Color(0xFFEA580C);

  showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    backgroundColor: const Color(0xFFFAFBFC),
    builder: (BuildContext context) {
      return Padding(
        padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
        child: ListView(
          shrinkWrap: true,
          children: <Widget>[
            Text(
              'Anti-Fraud Event Detail',
              style: Theme.of(context)
                  .textTheme
                  .headlineSmall
                  ?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 8),
            Text(
              'Expanded event card for suspicious activity review and manager decision.',
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: const Color(0xFF64748B)),
            ),
            const SizedBox(height: 18),
            Container(
              padding: const EdgeInsets.all(18),
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.08),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: color.withValues(alpha: 0.24)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    alert.title,
                    style: Theme.of(context)
                        .textTheme
                        .titleLarge
                        ?.copyWith(color: color, fontWeight: FontWeight.w800),
                  ),
                  const SizedBox(height: 10),
                  Text(alert.detail),
                  const SizedBox(height: 16),
                  _DetailLine(label: 'Rule', value: alert.ruleCode),
                  _DetailLine(
                      label: 'Severity', value: _titleCase(alert.severity)),
                  _DetailLine(label: 'Status', value: _titleCase(alert.status)),
                  _DetailLine(label: 'Room', value: alert.roomId),
                  _DetailLine(
                      label: 'Detected At',
                      value: _fullTimestamp(alert.timestamp)),
                  const _DetailLine(
                      label: 'Detected By', value: 'Backend alert rules'),
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
                    Text('Recommended Review',
                        style: Theme.of(context)
                            .textTheme
                            .titleMedium
                            ?.copyWith(fontWeight: FontWeight.w800)),
                    const SizedBox(height: 12),
                    const Text('1. Check cashier timeline'),
                    const Text('2. Verify room activity'),
                    const Text(
                        '3. Confirm whether payment or checkout event is missing'),
                  ],
                ),
              ),
            ),
          ],
        ),
      );
    },
  );
}

class _DetailLine extends StatelessWidget {
  const _DetailLine({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        children: <Widget>[
          SizedBox(
            width: 110,
            child: Text(
              label,
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: const Color(0xFF64748B)),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(fontWeight: FontWeight.w700),
            ),
          ),
        ],
      ),
    );
  }
}

String _shortTimestamp(String value) {
  if (value.length >= 16) {
    return value.substring(0, 16).replaceFirst('T', ' ');
  }
  return value;
}

String _fullTimestamp(String value) {
  if (value.isEmpty) {
    return '-';
  }
  return value.replaceFirst('T', ' ');
}

Color _colorForStatus(String status) {
  switch (status) {
    case 'occupied':
      return const Color(0xFF16A34A);
    case 'cleaning':
      return const Color(0xFF0284C7);
    case 'overstay':
      return const Color(0xFFEA580C);
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
      .map((String word) => '${word[0].toUpperCase()}${word.substring(1)}')
      .join(' ');
}
