class AppConfig {
  AppConfig({
    required this.baseUrl,
    required this.pendingTimeoutSeconds,
  });

  final String baseUrl;
  final int pendingTimeoutSeconds;

  Uri get baseUri => Uri.parse(baseUrl);
  Uri get eventsUri => Uri.parse('$baseUrl/events');

  static AppConfig fromEnvironment() {
    const baseUrl = String.fromEnvironment(
      'BACKEND_BASE_URL',
      defaultValue: 'http://localhost:8080',
    );
    const pendingTimeoutSeconds = int.fromEnvironment(
      'PENDING_TIMEOUT_SECONDS',
      defaultValue: 6,
    );
    return AppConfig(
      baseUrl: baseUrl,
      pendingTimeoutSeconds: pendingTimeoutSeconds,
    );
  }
}
