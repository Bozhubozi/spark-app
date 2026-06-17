import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/main.dart';

void main() {
  testWidgets('App renders welcome screen', (WidgetTester tester) async {
    await tester.pumpWidget(const ProviderScope(child: SparkApp()));
    await tester.pumpAndSettle();
    expect(find.text('Spark'), findsOneWidget);
  });
}
