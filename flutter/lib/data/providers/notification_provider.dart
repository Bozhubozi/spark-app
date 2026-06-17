import 'package:flutter_riverpod/flutter_riverpod.dart';

class AppNotification {
  final String id;
  final String title;
  final String body;
  final String? route;
  final Map<String, dynamic>? routeExtra;
  final DateTime createdAt;
  bool read;

  AppNotification({
    required this.id,
    required this.title,
    required this.body,
    this.route,
    this.routeExtra,
    this.read = false,
    DateTime? createdAt,
  }) : createdAt = createdAt ?? DateTime.now();
}

class NotificationNotifier extends StateNotifier<List<AppNotification>> {
  NotificationNotifier() : super([]);

  void add(AppNotification notification) {
    state = [notification, ...state];
  }

  void markAllRead() {
    state = state.map((n) {
      n.read = true;
      return n;
    }).toList();
  }

  void clear() {
    state = [];
  }

  int get unreadCount => state.where((n) => !n.read).length;
}

final notificationProvider =
    StateNotifierProvider<NotificationNotifier, List<AppNotification>>(
        (ref) => NotificationNotifier());

final unreadNotificationCountProvider = Provider<int>((ref) {
  return ref.watch(notificationProvider.notifier).unreadCount;
});
