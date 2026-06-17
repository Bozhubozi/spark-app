import 'package:flutter/material.dart';

class AppTheme {
  static const _primary = Color(0xFF6C5CE7);
  static const _secondary = Color(0xFF00CEC9);
  static const _bgDark = Color(0xFF0F0F1A);
  static const _surfaceDark = Color(0xFF1A1A2E);
  static const _bgLight = Color(0xFFF8F9FA);

  static ThemeData get dark => ThemeData(
    brightness: Brightness.dark,
    primaryColor: _primary,
    scaffoldBackgroundColor: _bgDark,
    colorScheme: const ColorScheme.dark(
      primary: _primary,
      secondary: _secondary,
      surface: _surfaceDark,
      error: Color(0xFFFF6B6B),
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: _surfaceDark,
      elevation: 0,
      centerTitle: true,
    ),
    cardTheme: CardThemeData(
      color: _surfaceDark,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: _primary,
        foregroundColor: Colors.white,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 14),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: _surfaceDark,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: BorderSide.none,
      ),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
    ),
    chipTheme: ChipThemeData(
      backgroundColor: _surfaceDark,
      selectedColor: _primary,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
    ),
  );

  static ThemeData get light => ThemeData(
    brightness: Brightness.light,
    primaryColor: _primary,
    scaffoldBackgroundColor: _bgLight,
    colorScheme: const ColorScheme.light(
      primary: _primary,
      secondary: _secondary,
      surface: Colors.white,
    ),
  );
}
