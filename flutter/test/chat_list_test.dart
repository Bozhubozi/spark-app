import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:spark_app/presentation/screens/login_screen.dart';
import 'package:spark_app/presentation/screens/register_screen.dart';

void main() {
  testWidgets('LoginScreen renders phone and password fields', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(child: MaterialApp(home: LoginScreen())),
    );
    await tester.pump();

    expect(find.text('火花'), findsOneWidget);
    expect(find.text('登录'), findsWidgets);
  });

  testWidgets('RegisterScreen renders form fields', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(child: MaterialApp(home: RegisterScreen())),
    );
    await tester.pump();

    expect(find.text('创建账号'), findsOneWidget);
    expect(find.text('注册'), findsWidgets);
  });
}
