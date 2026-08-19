class AppConfig {
  AppConfig({
    required this.baseUrl,
  });

  final String baseUrl;

  Uri get baseUri => Uri.parse(baseUrl);
  Uri get eventsUri => Uri.parse('$baseUrl/events');

  static AppConfig fromEnvironment() {
    const baseUrl = String.fromEnvironment(
      'BACKEND_BASE_URL',
      defaultValue: 'http://localhost:8080',
    );
    return AppConfig(baseUrl: baseUrl);
  }
}
