import 'package:flutter/material.dart';

import 'config/app_config.dart';
import 'screens/room_board_screen.dart';
import 'services/backend_client.dart';
import 'services/sse_client.dart';
import 'state/room_store.dart';

class CashierApp extends StatefulWidget {
  const CashierApp({super.key});

  @override
  State<CashierApp> createState() => _CashierAppState();
}

class _CashierAppState extends State<CashierApp> {
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
      pendingTimeout: Duration(seconds: config.pendingTimeoutSeconds),
    )..initialize();
  }

  @override
  void dispose() {
    _store.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    const seed = Color(0xFFB45309);

    return MaterialApp(
      title: 'CCS2 Cashier',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: seed),
        useMaterial3: true,
        scaffoldBackgroundColor: const Color(0xFFF8F4EC),
        cardTheme: CardThemeData(
          color: Colors.white,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(24),
            side: const BorderSide(color: Color(0xFFEADBC8)),
          ),
        ),
        textTheme: ThemeData.light().textTheme.apply(
              bodyColor: const Color(0xFF1F2937),
              displayColor: const Color(0xFF111827),
            ),
      ),
      home: RoomBoardScreen(store: _store),
    );
  }
}
