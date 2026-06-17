import 'package:flutter/foundation.dart';

/// Lightweight tracking abstraction for 神策数据 (Sensors Analytics).
///
/// In production, replace the body of [track] with a Sensors Analytics SDK call.
/// All events follow the tracking plan defined in docs/tracking-plan.md.

class Tracker {
  static final Tracker _instance = Tracker._();
  factory Tracker() => _instance;
  Tracker._();

  bool _initialized = false;
  String? _userId;
  bool _isGuest = true;
  final String _platform = defaultTargetPlatform == TargetPlatform.iOS ? 'iOS' : 'Android';
  final String _appVersion = '1.0.0';
  String? _city;

  void init({String? userId, bool isGuest = true, String? city}) {
    _userId = userId;
    _isGuest = isGuest;
    _city = city;
    _initialized = true;
  }

  void setUser(String userId, {bool isGuest = false}) {
    _userId = userId;
    _isGuest = isGuest;
  }

  void setCity(String? city) => _city = city;

  void track(String event, {Map<String, dynamic>? properties}) {
    if (!_initialized) return;

    final data = <String, dynamic>{
      'event': event,
      'user_id': _userId ?? 'anonymous',
      'is_guest': _isGuest,
      'platform': _platform,
      'app_version': _appVersion,
      if (_city != null) 'city': _city,
      if (properties != null) ...properties,
      'timestamp': DateTime.now().toIso8601String(),
    };

    // Console fallback; swap with Sensors Analytics SDK in production:
    // SensorsAnalyticsFlutterPlugin.track(event, properties: data);
    debugPrint('[Tracker] $event ${data.toString()}');
  }

  void login(String userId) => setUser(userId, isGuest: false);

  void logout() {
    _userId = null;
    _isGuest = true;
  }
}
