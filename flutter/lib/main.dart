import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/core/router/app_router.dart';
import 'package:spark_app/core/theme/app_theme.dart';
import 'package:spark_app/core/tracking/tracker.dart';

void main() {
  Tracker().init();
  runApp(const ProviderScope(child: SparkApp()));
}

class SparkApp extends StatelessWidget {
  const SparkApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Spark',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      themeMode: ThemeMode.dark,
      routerConfig: router,
    );
  }
}
