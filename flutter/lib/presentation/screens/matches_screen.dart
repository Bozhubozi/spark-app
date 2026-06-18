import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/providers/match_provider.dart';
import 'package:spark_app/data/providers/auth_provider.dart';
import 'package:spark_app/core/constants/zodiac_helper.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/presentation/widgets/shimmer_loading.dart';

class MatchesScreen extends ConsumerWidget {
  const MatchesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final matches = ref.watch(matchesProvider);
    final myId = ref.watch(authProvider).valueOrNull?.id ?? '';

    return Scaffold(
      appBar: AppBar(title: const Text('我的匹配')),
      body: matches.when(
        data: (list) {
          if (list.isEmpty) {
            return Center(
              child: Column(mainAxisSize: MainAxisSize.min, children: [
                Icon(Icons.favorite_border, size: 80, color: Colors.grey[600]),
                const SizedBox(height: 16),
                Text('还没有匹配',
                    style: TextStyle(color: Colors.grey[500], fontSize: 16)),
                const SizedBox(height: 4),
                Text('继续滑动寻找你的火花',
                    style: TextStyle(color: Colors.grey[600], fontSize: 13)),
              ]),
            );
          }

          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(matchesProvider),
            child: ListView.builder(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: list.length,
              itemBuilder: (_, i) {
                final match = list[i];
                final other = match.otherUser(myId);
                final name = other?.nickname ?? '用户';
                final zodiac = zodiacFromBirth(other?.birthDate);
                final initial = name.isNotEmpty ? name[0].toUpperCase() : '?';

                return Card(
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  child: InkWell(
                    borderRadius: BorderRadius.circular(16),
                    onTap: () async {
                      final api = ref.read(apiClientProvider);
                      try {
                        final resp = await api.post('/api/v1/chat/rooms',
                            data: {'target_user_id': other!.id});
                        final roomId = resp.data['id'] ?? match.id;
                        if (context.mounted) {
                          context.push('/chat/$roomId', extra: {
                            'otherName': name,
                            'otherId': other.id,
                          });
                        }
                      } catch (_) {}
                    },
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Row(
                        children: [
                          CircleAvatar(
                            radius: 28,
                            backgroundColor: const Color(0xFF6C5CE7),
                            child: Text(initial,
                                style: const TextStyle(fontSize: 22, color: Colors.white)),
                          ),
                          const SizedBox(width: 16),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  children: [
                                    Text(name,
                                        style: const TextStyle(
                                            fontSize: 17, fontWeight: FontWeight.w600)),
                                    if (zodiac.isNotEmpty) ...[
                                      const SizedBox(width: 6),
                                      Text(zodiac, style: const TextStyle(fontSize: 16)),
                                    ],
                                  ],
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  '匹配度 ${(match.score * 100).toInt()}%',
                                  style: TextStyle(
                                      color: _scoreColor(match.score), fontSize: 13),
                                ),
                              ],
                            ),
                          ),
                          if (other?.birthDate != null)
                            IconButton(
                              icon: const Icon(Icons.auto_awesome, size: 20, color: Color(0xFFFDCB6E)),
                              onPressed: () => context.push(
                                '/zodiac-compat/${other?.id}',
                                extra: {
                                  'targetName': name,
                                  'targetBirthDate': other?.birthDate,
                                },
                              ),
                            ),
                          const Icon(Icons.chevron_right, color: Colors.grey),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          );
        },
        loading: () => ListView(children: const [
          SkeletonCard(), SkeletonCard(), SkeletonCard(),
          SkeletonCard(), SkeletonCard(),
        ]),
        error: (e, _) => Center(
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            Text('错误：$e', style: TextStyle(color: Colors.grey[500])),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => ref.invalidate(matchesProvider),
              child: const Text('重试'),
            ),
          ]),
        ),
      ),
    );
  }

  Color _scoreColor(double score) {
    if (score >= 0.7) return const Color(0xFF00B894);
    if (score >= 0.4) return Colors.orangeAccent;
    return Colors.grey;
  }
}
