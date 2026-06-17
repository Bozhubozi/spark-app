import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/providers/match_provider.dart';
import 'package:spark_app/data/providers/notification_provider.dart';
import 'package:spark_app/core/constants/zodiac_helper.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:uuid/uuid.dart';
import 'package:spark_app/presentation/widgets/shimmer_loading.dart';

class LikersScreen extends ConsumerWidget {
  const LikersScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final likers = ref.watch(likersProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Who Liked You')),
      body: likers.when(
        data: (list) {
          if (list.isEmpty) {
            return Center(
              child: Column(mainAxisSize: MainAxisSize.min, children: [
                Icon(Icons.favorite_border, size: 80, color: Colors.grey[600]),
                const SizedBox(height: 16),
                Text('No likes yet',
                    style: TextStyle(color: Colors.grey[500], fontSize: 16)),
                const SizedBox(height: 4),
                Text('Keep swiping to get noticed',
                    style: TextStyle(color: Colors.grey[600], fontSize: 13)),
              ]),
            );
          }

          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(likersProvider),
            child: ListView.builder(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: list.length,
              itemBuilder: (_, i) {
                final item = list[i];
                final u = item.user;
                final name = u.nickname;
                final zodiac = zodiacFromBirth(u.birthDate);
                final age = _age(u.birthDate);
                final initial = name.isNotEmpty ? name[0].toUpperCase() : '?';

                return Card(
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Row(
                      children: [
                        CircleAvatar(
                          radius: 28,
                          backgroundColor: const Color(0xFFFD79A8),
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
                                  if (age.isNotEmpty) ...[
                                    const SizedBox(width: 6),
                                    Text(age, style: TextStyle(color: Colors.grey[500], fontSize: 14)),
                                  ],
                                  if (zodiac.isNotEmpty) ...[
                                    const SizedBox(width: 6),
                                    Text(zodiac, style: const TextStyle(fontSize: 16)),
                                  ],
                                ],
                              ),
                              if (u.bio != null && u.bio!.isNotEmpty)
                                Padding(
                                  padding: const EdgeInsets.only(top: 4),
                                  child: Text(
                                    u.bio!,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: TextStyle(color: Colors.grey[500], fontSize: 13),
                                  ),
                                ),
                            ],
                          ),
                        ),
                        ElevatedButton(
                          onPressed: () => _likeBack(context, ref, item.matchId, u, name),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFFFD79A8),
                            shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(20)),
                            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                          ),
                          child: const Text('Like Back',
                              style: TextStyle(color: Colors.white, fontSize: 13)),
                        ),
                      ],
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
            Text('Error: $e', style: TextStyle(color: Colors.grey[500])),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => ref.invalidate(likersProvider),
              child: const Text('Retry'),
            ),
          ]),
        ),
      ),
    );
  }

  Future<void> _likeBack(BuildContext context, WidgetRef ref,
      String matchId, dynamic user, String name) async {
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.post('/api/v1/match/swipe', data: {
        'target_user_id': user.id,
        'direction': 'like',
      });
      ref.invalidate(likersProvider);
      ref.invalidate(likesCountProvider);
      if (resp.data['matched'] == true && context.mounted) {
        ref.read(notificationProvider.notifier).add(AppNotification(
          id: const Uuid().v4(),
          title: 'New Match!',
          body: 'You and $name liked each other',
        ));
        final roomResp = await api.post('/api/v1/chat/rooms',
            data: {'target_user_id': user.id});
        final roomId = roomResp.data['id'] ?? matchId;
        if (context.mounted) {
          context.push('/chat/$roomId', extra: {
            'otherName': name,
            'otherId': user.id,
          });
        }
      }
    } catch (_) {}
  }

  String _age(String? birthDate) {
    if (birthDate == null) return '';
    try {
      final bd = DateTime.tryParse(birthDate);
      if (bd == null) return '';
      final now = DateTime.now();
      var age = now.year - bd.year;
      if (now.month < bd.month || (now.month == bd.month && now.day < bd.day)) {
        age--;
      }
      return '$age';
    } catch (_) {
      return '';
    }
  }
}
