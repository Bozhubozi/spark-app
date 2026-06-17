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

    expect(find.text('Spark'), findsOneWidget);
    expect(find.text('Log In'), findsWidgets);
  });

  testWidgets('RegisterScreen renders form fields', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(child: MaterialApp(home: RegisterScreen())),
    );
    await tester.pump();

    expect(find.text('Create Account'), findsOneWidget);
    expect(find.text('Sign Up'), findsWidgets);
  });
}
