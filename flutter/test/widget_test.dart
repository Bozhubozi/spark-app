import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/main.dart';

void main() {
  testWidgets('Welcome screen renders Get Started button', (WidgetTester tester) async {
    await tester.pumpWidget(const ProviderScope(child: SparkApp()));
    await tester.pump();
    expect(find.text('立即开始'), findsOneWidget);
    expect(find.text('登录'), findsOneWidget);
  });
}
