import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/data/providers/match_provider.dart';
import 'package:spark_app/core/network/api_client.dart';

class BlockedUsersScreen extends ConsumerWidget {
  const BlockedUsersScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final blocked = ref.watch(blockedUsersProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('已拉黑用户')),
      body: blocked.when(
        data: (list) {
          if (list.isEmpty) {
            return Center(
              child: Column(mainAxisSize: MainAxisSize.min, children: [
                Icon(Icons.block_outlined, size: 80, color: Colors.grey[600]),
                const SizedBox(height: 16),
                Text('没有拉黑用户',
                    style: TextStyle(color: Colors.grey[500], fontSize: 16)),
                const SizedBox(height: 4),
                Text('你拉黑的用户会显示在这里',
                    style: TextStyle(color: Colors.grey[600], fontSize: 13)),
              ]),
            );
          }

          return ListView.builder(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: list.length,
            itemBuilder: (_, i) {
              final item = list[i];
              final u = item.user;
              final name = u.nickname;
              final initial = name.isNotEmpty ? name[0].toUpperCase() : '?';

              return Card(
                margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      CircleAvatar(
                        radius: 24,
                        backgroundColor: Colors.grey[700],
                        child: Text(initial,
                            style: const TextStyle(fontSize: 18, color: Colors.white)),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Text(name,
                            style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                      ),
                      TextButton(
                        onPressed: () => _unblock(context, ref, item.matchId, u.id, name),
                        child: const Text('解除拉黑'),
                      ),
                    ],
                  ),
                ),
              );
            },
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            Text('错误：$e', style: TextStyle(color: Colors.grey[500])),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => ref.invalidate(blockedUsersProvider),
              child: const Text('重试'),
            ),
          ]),
        ),
      ),
    );
  }

  Future<void> _unblock(BuildContext context, WidgetRef ref,
      String matchId, String userId, String name) async {
    try {
      await ref.read(apiClientProvider).post('/api/v1/match/unblock', data: {
        'target_user_id': userId,
      });
      ref.invalidate(blockedUsersProvider);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('已解除拉黑 $name')),
        );
      }
    } catch (_) {}
  }
}
