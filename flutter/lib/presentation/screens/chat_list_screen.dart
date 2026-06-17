import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/providers/chat_provider.dart';
import 'package:spark_app/data/providers/auth_provider.dart';
import 'package:spark_app/presentation/widgets/shimmer_loading.dart';

class ChatListScreen extends ConsumerWidget {
  const ChatListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final rooms = ref.watch(chatRoomsProvider);
    final user = ref.watch(authProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Messages')),
      body: rooms.when(
        data: (list) {
          if (list.isEmpty) {
            return RefreshIndicator(
              onRefresh: () async => ref.invalidate(chatRoomsProvider),
              child: ListView(children: [
                SizedBox(height: MediaQuery.of(context).size.height * 0.3),
                Center(
                  child: Column(mainAxisSize: MainAxisSize.min, children: [
                    Icon(Icons.chat_bubble_outline, size: 80, color: Colors.grey[600]),
                    const SizedBox(height: 16),
                    Text('No messages yet', style: TextStyle(color: Colors.grey[500], fontSize: 16)),
                    const SizedBox(height: 4),
                    Text('Match with someone to start chatting',
                        style: TextStyle(color: Colors.grey[600], fontSize: 13)),
                  ]),
                ),
              ]),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(chatRoomsProvider),
            child: ListView.builder(
              itemCount: list.length,
              itemBuilder: (_, i) {
                final room = list[i];
                final myId = user.valueOrNull?.id ?? '';
                final initial = room.otherNickname.isNotEmpty
                    ? room.otherNickname[0].toUpperCase()
                    : room.otherUserId(myId).substring(0, 1).toUpperCase();
                return ListTile(
                  leading: CircleAvatar(
                    backgroundColor: const Color(0xFF6C5CE7),
                    child: Text(initial),
                  ),
                  title: Text(room.otherNickname.isNotEmpty ? room.otherNickname : 'Chat'),
                  subtitle: Text(
                    room.lastContent ?? 'Tap to chat',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(color: Colors.grey[500]),
                  ),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (room.unreadCount > 0)
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                          decoration: BoxDecoration(
                            color: const Color(0xFF6C5CE7),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: Text('${room.unreadCount}',
                              style: const TextStyle(color: Colors.white, fontSize: 11)),
                        ),
                      const SizedBox(width: 8),
                      Text(_formatTime(room.lastMessageAt),
                          style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                    ],
                  ),
                  onTap: () => context.push('/chat/${room.id}', extra: {
                    'otherName': room.otherNickname,
                    'otherId': room.otherUserId(myId),
                  }),
                );
              },
            ),
          );
        },
        loading: () => ListView(children: const [
          SkeletonListTile(), SkeletonListTile(), SkeletonListTile(),
          SkeletonListTile(), SkeletonListTile(),
        ]),
        error: (e, _) => Center(
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            Text('Error: $e', style: TextStyle(color: Colors.grey[500])),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => ref.invalidate(chatRoomsProvider),
              child: const Text('Retry'),
            ),
          ]),
        ),
      ),
    );
  }

  String _formatTime(DateTime dt) {
    final now = DateTime.now();
    if (dt.day == now.day) return '${dt.hour}:${dt.minute.toString().padLeft(2, '0')}';
    if (dt.year == now.year) return '${dt.month}/${dt.day}';
    return '${dt.year}/${dt.month}';
  }
}
