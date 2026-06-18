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
        title: const Text('通知'),
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
                PopupMenuItem(value: 'mark_read', child: Text('全部标为已读')),
                PopupMenuItem(value: 'clear', child: Text('清空全部')),
              ],
            ),
        ],
      ),
      body: notifications.isEmpty
          ? Center(
              child: Column(mainAxisSize: MainAxisSize.min, children: [
                Icon(Icons.notifications_none, size: 80, color: Colors.grey[600]),
                const SizedBox(height: 16),
                Text('暂无通知',
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
    if (diff.inMinutes < 1) return '刚刚';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    if (diff.inDays < 7) return '${diff.inDays}d ago';
    return '${dt.month}/${dt.day}';
  }
}
