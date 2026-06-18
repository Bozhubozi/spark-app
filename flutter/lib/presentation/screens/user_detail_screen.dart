import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/models/user_model.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/constants/zodiac_helper.dart';

class UserDetailScreen extends ConsumerStatefulWidget {
  final UserModel user;
  const UserDetailScreen({super.key, required this.user});

  @override
  ConsumerState<UserDetailScreen> createState() => _UserDetailScreenState();
}

class _UserDetailScreenState extends ConsumerState<UserDetailScreen> {
  bool _swiping = false;

  Future<void> _swipe(String direction) async {
    setState(() => _swiping = true);
    try {
      final api = ref.read(apiClientProvider);
      await api.post('/api/v1/match/swipe', data: {
        'target_user_id': widget.user.id,
        'direction': direction,
      });
      if (mounted) context.pop(true);
    } catch (_) {
      if (mounted) setState(() => _swiping = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final user = widget.user;
    final zodiac = zodiacFromBirth(user.birthDate);
    final trait = zodiacTraits[zodiac];
    final age = _age(user.birthDate);

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 400,
            pinned: true,
            leading: IconButton(
              icon: const Icon(Icons.arrow_back),
              onPressed: () => context.pop(),
            ),
            flexibleSpace: FlexibleSpaceBar(
              background: Stack(
                fit: StackFit.expand,
                children: [
                  Container(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        colors: [
                          trait?.color ?? const Color(0xFF6C5CE7),
                          const Color(0xFF1A1A2E),
                        ],
                        begin: Alignment.topCenter,
                        end: Alignment.bottomCenter,
                      ),
                    ),
                    child: Center(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          CircleAvatar(
                            radius: 64,
                            backgroundColor: Colors.white24,
                            child: Text(
                              user.nickname[0].toUpperCase(),
                              style: const TextStyle(fontSize: 48, color: Colors.white),
                            ),
                          ),
                          const SizedBox(height: 16),
                          Text(zodiac, style: const TextStyle(fontSize: 56)),
                          if (trait != null) ...[
                            const SizedBox(height: 4),
                            Text(trait.element,
                                style: const TextStyle(color: Colors.white70, fontSize: 14)),
                          ],
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.all(24),
            sliver: SliverList(
              delegate: SliverChildListDelegate([
                // Name + age
                Row(
                  children: [
                    Text(user.nickname,
                        style: const TextStyle(fontSize: 28, fontWeight: FontWeight.bold)),
                    if (age.isNotEmpty) ...[
                      const SizedBox(width: 8),
                      Text(age,
                          style: TextStyle(fontSize: 24, color: Colors.grey[400])),
                    ],
                  ],
                ),
                if (zodiac.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Text('${zodiacEmojiToName(zodiac)} · ${trait?.strength ?? ""}',
                      style: TextStyle(color: Colors.grey[500], fontSize: 14)),
                ],

                // City & last active
                if (user.city != null) ...[
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      const Icon(Icons.location_on_outlined, size: 16, color: Colors.grey),
                      const SizedBox(width: 4),
                      Text(user.city!, style: TextStyle(color: Colors.grey[400])),
                      if (user.lastActiveAt != null) ...[
                        const SizedBox(width: 16),
                        const Icon(Icons.access_time, size: 16, color: Colors.grey),
                        const SizedBox(width: 4),
                        Text(_lastActive(user.lastActiveAt!),
                            style: TextStyle(color: Colors.grey[400])),
                      ],
                    ],
                  ),
                ],

                // Bio
                if (user.bio != null && user.bio!.isNotEmpty) ...[
                  const SizedBox(height: 24),
                  const Text('关于', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 8),
                  Text(user.bio!,
                      style: TextStyle(color: Colors.grey[300], height: 1.6, fontSize: 15)),
                ],

                // Interests
                if (user.interests.isNotEmpty) ...[
                  const SizedBox(height: 24),
                  const Text('兴趣', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8, runSpacing: 8,
                    children: user.interests.map((tag) => Chip(
                      label: Text('${tag.icon ?? ''} ${tag.name}',
                          style: const TextStyle(color: Colors.white, fontSize: 13)),
                      backgroundColor: const Color(0xFF6C5CE7).withValues(alpha: 0.2),
                      padding: EdgeInsets.zero,
                      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    )).toList(),
                  ),
                ],

                // Zodiac compat preview
                if (zodiac.isNotEmpty) ...[
                  const SizedBox(height: 24),
                  const Text('星座', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 12),
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: const Color(0xFF1A1A2E),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.white12),
                    ),
                    child: Row(
                      children: [
                        Column(
                          children: [
                            Text(zodiac, style: const TextStyle(fontSize: 32)),
                            Text(zodiacEmojiToName(zodiac),
                                style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                          ],
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(trait?.element ?? '',
                                  style: TextStyle(color: trait?.color, fontSize: 13)),
                              const SizedBox(height: 4),
                              Text(trait?.strength ?? '',
                                  style: TextStyle(color: Colors.grey[300], fontSize: 13)),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ],

                const SizedBox(height: 120),
              ]),
            ),
          ),
        ],
      ),
      // Bottom action bar
      bottomNavigationBar: SafeArea(
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 40, vertical: 16),
          decoration: BoxDecoration(
            color: const Color(0xFF1A1A2E),
            border: Border(top: BorderSide(color: Colors.grey[800]!)),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              Material(
                color: Colors.redAccent,
                shape: const CircleBorder(),
                child: InkWell(
                  onTap: _swiping ? null : () => _swipe('pass'),
                  customBorder: const CircleBorder(),
                  child: const Padding(
                    padding: EdgeInsets.all(20),
                    child: Icon(Icons.close, color: Colors.white, size: 32),
                  ),
                ),
              ),
              Material(
                color: const Color(0xFFFD79A8),
                shape: const CircleBorder(),
                child: InkWell(
                  onTap: _swiping ? null : () => _swipe('like'),
                  customBorder: const CircleBorder(),
                  child: const Padding(
                    padding: EdgeInsets.all(20),
                    child: Icon(Icons.favorite, color: Colors.white, size: 32),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _age(String? birthDate) {
    if (birthDate == null) return '';
    try {
      final bd = DateTime.tryParse(birthDate);
      if (bd == null) return '';
      final now = DateTime.now();
      var age = now.year - bd.year;
      if (now.month < bd.month || (now.month == bd.month && now.day < bd.day)) age--;
      return '$age';
    } catch (_) {
      return '';
    }
  }

  String _lastActive(DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inMinutes < 60) return '最近活跃';
    if (diff.inHours < 24) return 'Active ${diff.inHours}h ago';
    if (diff.inDays < 7) return 'Active ${diff.inDays}d ago';
    return 'Active ${diff.inDays ~/ 7}w ago';
  }
}
