import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/constants/app_constants.dart';

void main() {
  test('ApiClient constructs cleanly', () {
    final client = ApiClient();
    expect(client, isNotNull);
  });

  test('apiClientProvider returns ApiClient', () {
    final container = ProviderContainer();
    final client = container.read(apiClientProvider);
    expect(client, isNotNull);
  });

  test('AppConstants has correct defaults', () {
    expect(AppConstants.apiBaseUrl, isNotEmpty);
    expect(AppConstants.wsUrl, startsWith('ws'));
    expect(AppConstants.connectTimeout, isNotNull);
    expect(AppConstants.receiveTimeout, isNotNull);
  });
}
