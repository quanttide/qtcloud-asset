import 'package:flutter_test/flutter_test.dart';

import 'package:qtcloud_studio/main.dart';

void main() {
  test('provider URL accepts the production gateway build define', () {
    const expectedProviderBaseUrl = String.fromEnvironment(
      'EXPECTED_PROVIDER_BASE_URL',
      defaultValue: 'http://127.0.0.1:9000',
    );

    expect(providerBaseUrl, expectedProviderBaseUrl);
  });
}
