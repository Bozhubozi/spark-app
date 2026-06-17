import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/data/providers/notification_provider.dart';

class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifications = ref.watch(notificationProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Notifications'),
        actions: [
          if (notifications.isNotEmpty)
            PopupMenuButton<String>(
              onSelected: (action) {
                if (action == 'mark_read') {
                  ref.read(notificationProvider.notifier).markAllRead();
                } else if (action == 'clear') {
                  ref.read(notificationProvider.notifier).clear();
                }
              },
              itemBuilder: (_) => const [
                PopupMenuItem(value: 'mark_read', child: Text('Mark All Read')),
                PopupMenuItem(value: 'clear', child: Text('Clear All')),
              ],
            ),
        ],
      ),
      body: notifications.isEmpty
          ? Center(
              child: Column(mainAxisSize: MainAxisSize.min, children: [
                Icon(Icons.notifications_none, size: 80, color: Colors.grey[600]),
                const SizedBox(height: 16),
                Text('No notifications yet',
                    style: TextStyle(color: Colors.grey[500], fontSize: 16)),
              ]),
            )
          : ListView.builder(
              itemCount: notifications.length,
              itemBuilder: (_, i) {
                final n = notifications[i];
                return ListTile(
                  leading: CircleAvatar(
                    backgroundColor: n.read
                        ? Colors.grey[800]
                        : const Color(0xFF6C5CE7),
                    child: Icon(
                      n.title.contains('Match') ? Icons.favorite : Icons.notifications,
                      color: Colors.white,
                      size: 18,
                    ),
                  ),
                  title: Text(
                    n.title,
                    style: TextStyle(
                      fontWeight: n.read ? FontWeight.normal : FontWeight.w600,
                    ),
                  ),
                  subtitle: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(n.body, style: TextStyle(color: Colors.grey[500], fontSize: 13)),
                      const SizedBox(height: 2),
                      Text(_formatTime(n.createdAt),
                          style: TextStyle(color: Colors.grey[600], fontSize: 11)),
                    ],
                  ),
                  onTap: () {
                    ref.read(notificationProvider.notifier).markAllRead();
                  },
                );
              },
            ),
    );
  }

  String _formatTime(DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inMinutes < 1) return 'just now';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    if (diff.inDays < 7) return '${diff.inDays}d ago';
    return '${dt.month}/${dt.day}';
  }
}
