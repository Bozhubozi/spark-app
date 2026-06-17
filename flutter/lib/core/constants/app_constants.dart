class AppConstants {
  static const apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8080',
  );
  static const wsUrl = String.fromEnvironment(
    'WS_URL',
    defaultValue: 'ws://10.0.2.2:8080/ws',
  );
  static const connectTimeout = Duration(seconds: 10);
  static const receiveTimeout = Duration(seconds: 10);
}
