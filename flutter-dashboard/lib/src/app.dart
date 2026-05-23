import 'package:flutter/material.dart';

import 'config/app_config.dart';
import 'screens/dashboard_home_screen.dart';
import 'services/backend_client.dart';
import 'services/sse_client.dart';
import 'state/room_store.dart';

class DashboardApp extends StatefulWidget {
  const DashboardApp({super.key});

  @override
  State<DashboardApp> createState() => _DashboardAppState();
}

class _DashboardAppState extends State<DashboardApp> {
  late final RoomStore _store;

  @override
  void initState() {
    super.initState();
    final config = AppConfig.fromEnvironment();
    final backend = BackendClient(config.baseUri);
    final sse = SseClient(config.eventsUri);
    _store = RoomStore(
      backend: backend,
      sseClient: sse,
    )..initialize();
  }

  @override
  void dispose() {
    _store.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    const seed = Color(0xFF0F766E);

    return MaterialApp(
      title: 'CCS2 Dashboard',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: seed),
        useMaterial3: true,
        scaffoldBackgroundColor: const Color(0xFFF3F0E8),
        appBarTheme: const AppBarTheme(
          backgroundColor: Colors.transparent,
          foregroundColor: Color(0xFF0F172A),
          elevation: 0,
        ),
        cardTheme: CardThemeData(
          color: Colors.white,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(24),
            side: const BorderSide(color: Color(0xFFE7E0D4)),
          ),
        ),
        textTheme: ThemeData.light().textTheme.apply(
              bodyColor: const Color(0xFF1E293B),
              displayColor: const Color(0xFF0F172A),
            ),
      ),
      home: DashboardHomeScreen(store: _store),
    );
  }
}
