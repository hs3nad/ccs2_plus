import 'package:flutter/material.dart';

import 'config/app_config.dart';
import 'screens/room_list_screen.dart';
import 'services/backend_client.dart';
import 'services/sse_client.dart';
import 'state/room_store.dart';

class HandheldApp extends StatefulWidget {
  const HandheldApp({super.key});

  @override
  State<HandheldApp> createState() => _HandheldAppState();
}

class _HandheldAppState extends State<HandheldApp> {
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
    const seed = Color(0xFF0F766E);

    return MaterialApp(
      title: 'CCS2 Handheld',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: seed),
        useMaterial3: true,
        scaffoldBackgroundColor: const Color(0xFFF2F7F5),
        cardTheme: CardThemeData(
          color: Colors.white,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(24),
            side: const BorderSide(color: Color(0xFFD6EAE4)),
          ),
        ),
        textTheme: ThemeData.light().textTheme.apply(
              bodyColor: const Color(0xFF1F2937),
              displayColor: const Color(0xFF0F172A),
            ),
      ),
      home: RoomListScreen(store: _store),
    );
  }
}
