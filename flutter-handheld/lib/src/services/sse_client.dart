import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/room_event.dart';

class SseClient {
  SseClient(this.eventsUri, {http.Client? client}) : _client = client ?? http.Client();

  final Uri eventsUri;
  final http.Client _client;

  Stream<RoomEvent> connect() async* {
    final request = http.Request('GET', eventsUri);
    request.headers['Accept'] = 'text/event-stream';
    final response = await _client.send(request);
    if (response.statusCode != 200) {
      throw Exception('Failed to connect SSE: ${response.statusCode}');
    }

    String currentEvent = 'message';
    final lines = response.stream.transform(utf8.decoder).transform(const LineSplitter());
    await for (final line in lines) {
      if (line.startsWith('event:')) {
        currentEvent = line.substring(6).trim();
      } else if (line.startsWith('data:')) {
        final data = line.substring(5).trim();
        if (data.isEmpty) {
          continue;
        }
        final decoded = jsonDecode(data) as Map<String, dynamic>;
        yield RoomEvent(type: currentEvent, payload: decoded);
      }
    }
  }
}
