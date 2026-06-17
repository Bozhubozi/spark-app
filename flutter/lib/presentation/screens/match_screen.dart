import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/data/models/user_model.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/data/providers/match_provider.dart';
import 'package:spark_app/data/providers/notification_provider.dart';
import 'package:spark_app/presentation/widgets/match_card.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/presentation/widgets/horoscope_banner.dart';
import 'package:spark_app/presentation/screens/discovery_prefs_screen.dart';
import 'package:spark_app/presentation/widgets/shimmer_loading.dart';

class MatchScreen extends ConsumerStatefulWidget {
  const MatchScreen({super.key});

  @override
  ConsumerState<MatchScreen> createState() => _MatchScreenState();
}

class _MatchScreenState extends ConsumerState<MatchScreen> {
  final List<UserModel> _cards = [];
  bool _fetching = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _fetchCandidates());
  }

  Future<void> _fetchCandidates() async {
    if (_fetching) return;
    setState(() => _fetching = true);
    try {
      final api = ref.read(apiClientProvider);
      final gender = ref.read(discoveryGenderProvider);
      final minAge = ref.read(discoveryMinAgeProvider);
      final maxAge = ref.read(discoveryMaxAgeProvider);
      final params = <String, dynamic>{};
      if (gender > 0) params['gender'] = '$gender';
      if (minAge > 18) params['min_age'] = '$minAge';
      if (maxAge < 45) params['max_age'] = '$maxAge';
      final queryString = params.entries.map((e) => '${e.key}=${e.value}').join('&');
      final path = '/api/v1/match/candidates${queryString.isNotEmpty ? '?$queryString' : ''}';
      final resp = await api.get(path);
      final list = (resp.data as List<dynamic>?)
          ?.map((e) => UserModel.fromJson(e))
          .toList() ?? [];
      setState(() {
        final existing = _cards.map((c) => c.id).toSet();
        for (final u in list) {
          if (!existing.contains(u.id)) _cards.add(u);
        }
        _fetching = false;
      });
    } catch (_) {
      setState(() => _fetching = false);
    }
    ref.invalidate(likesCountProvider);
    ref.invalidate(remainingSwipesProvider);
  }

  void _onCardSwiped(UserModel user) {
    setState(() => _cards.removeWhere((c) => c.id == user.id));
    if (_cards.length < 3) _fetchCandidates();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Discover'),
        actions: [
          Builder(
            builder: (context) {
              final unread = ref.watch(unreadNotificationCountProvider);
              return IconButton(
                icon: unread > 0
                    ? Badge(
                        label: Text(unread > 9 ? '9+' : '$unread'),
                        child: const Icon(Icons.notifications_outlined),
                      )
                    : const Icon(Icons.notifications_outlined),
                onPressed: () {
                  ref.read(notificationProvider.notifier).markAllRead();
                  context.push('/notifications');
                },
              );
            },
          ),
          IconButton(
            icon: const Icon(Icons.tune),
            onPressed: () async {
              await context.push('/discovery-prefs');
              if (mounted) {
                setState(() => _cards.clear());
                _fetchCandidates();
              }
            },
          ),
          IconButton(
            icon: const Icon(Icons.people),
            onPressed: () => context.push('/matches'),
          ),
        ],
      ),
      body: _buildBody(context),
    );
  }

  Widget _buildBody(BuildContext context) {
    final likesCount = ref.watch(likesCountProvider).valueOrNull ?? 0;
    final remaining = ref.watch(remainingSwipesProvider).valueOrNull ?? 0;

    if (_cards.isEmpty && _fetching) {
      return ListView(children: const [
        SizedBox(height: 32),
        SkeletonCard(), SkeletonCard(), SkeletonCard(),
      ]);
    }
    if (_cards.isEmpty) {
      return RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(candidatesProvider);
          ref.invalidate(likesCountProvider);
          ref.invalidate(remainingSwipesProvider);
          await _fetchCandidates();
        },
        child: ListView(children: [
          SizedBox(height: MediaQuery.of(context).size.height * 0.25),
          Center(
            child: Column(mainAxisSize: MainAxisSize.min, children: [
              Icon(Icons.nightlight_outlined, size: 80, color: Colors.grey[600]),
              const SizedBox(height: 16),
              Text('All caught up!',
                  style: Theme.of(context).textTheme.titleLarge?.copyWith(color: Colors.grey)),
              const SizedBox(height: 8),
              Text('Check back tomorrow for new people',
                  style: TextStyle(color: Colors.grey[600], fontSize: 14)),
              const SizedBox(height: 16),
              OutlinedButton(
                onPressed: _fetchCandidates,
                child: const Text('Refresh'),
              ),
            ]),
          ),
        ]),
      );
    }

    final visible = _cards.take(3).toList();

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(candidatesProvider);
        await _fetchCandidates();
      },
      child: Column(
        children: [
          const HoroscopeBanner(),
          if (likesCount > 0) _likesBanner(likesCount),
          if (remaining < 20)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.flash_on, size: 16, color: Colors.grey[500]),
                  const SizedBox(width: 4),
                  Text('$remaining swipes remaining today',
                      style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                ],
              ),
            ),
          Expanded(
            child: Stack(
              fit: StackFit.expand,
              children: [
                // Background cards
                for (int i = visible.length - 1; i > 0; i--)
                  Positioned(
                    top: 16.0 * i,
                    left: 8.0 * i,
                    right: 8.0 * i,
                    bottom: 16.0 * i,
                    child: _staticCard(visible[i]),
                  ),
                // Top card (interactive)
                Positioned.fill(
                  child: MatchCard(
                    user: visible[0],
                    onSwiped: () => _onCardSwiped(visible[0]),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _likesBanner(int count) {
    return GestureDetector(
      onTap: () => context.push('/likers'),
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [
              const Color(0xFFFD79A8).withValues(alpha: 0.2),
              const Color(0xFF6C5CE7).withValues(alpha: 0.2),
            ],
          ),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xFFFD79A8).withValues(alpha: 0.3)),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.favorite, color: Color(0xFFFD79A8), size: 18),
            const SizedBox(width: 8),
            Text(
              '$count ${count == 1 ? 'person' : 'people'} liked you — keep swiping!',
              style: const TextStyle(color: Color(0xFFFD79A8), fontSize: 13),
            ),
          ],
        ),
      ),
    );
  }

  Widget _staticCard(UserModel user) {
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      child: Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(24),
          gradient: const LinearGradient(
            colors: [Color(0xFF2D1B69), Color(0xFF1A1A2E)],
            begin: Alignment.bottomCenter,
            end: Alignment.topCenter,
          ),
        ),
        child: Center(
          child: CircleAvatar(
            radius: 50,
            backgroundColor: Colors.white12,
            child: Text(user.nickname[0].toUpperCase(),
                style: const TextStyle(fontSize: 32, color: Colors.white38)),
          ),
        ),
      ),
    );
  }
}
