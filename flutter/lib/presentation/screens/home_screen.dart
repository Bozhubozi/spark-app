import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/providers/chat_provider.dart';

class HomeScreen extends ConsumerWidget {
  final Widget child;
  const HomeScreen({super.key, required this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final rooms = ref.watch(chatRoomsProvider);
    final unreadTotal = rooms.whenOrNull(data: (list) {
      var total = 0;
      for (final r in list) {
        total += r.unreadCount;
      }
      return total;
    }) ?? 0;

    return Scaffold(
      body: child,
      bottomNavigationBar: BottomNavigationBar(
        type: BottomNavigationBarType.fixed,
        backgroundColor: Theme.of(context).colorScheme.surface,
        selectedItemColor: Theme.of(context).primaryColor,
        unselectedItemColor: Colors.grey,
        currentIndex: _currentIndex(context),
        onTap: (i) => _onTap(context, i),
        items: [
          const BottomNavigationBarItem(icon: Icon(Icons.explore), label: '发现'),
          BottomNavigationBarItem(
            icon: unreadTotal > 0
                ? Badge(
                    label: Text(unreadTotal > 99 ? '99+' : '$unreadTotal'),
                    child: const Icon(Icons.chat_bubble_outline),
                  )
                : const Icon(Icons.chat_bubble_outline),
            label: '聊天',
          ),
          const BottomNavigationBarItem(icon: Icon(Icons.person_outline), label: '我的'),
        ],
      ),
    );
  }

  int _currentIndex(BuildContext context) {
    final loc = GoRouterState.of(context).uri.toString();
    if (loc.startsWith('/chat')) return 1;
    if (loc.startsWith('/profile')) return 2;
    return 0;
  }

  void _onTap(BuildContext context, int i) {
    switch (i) {
      case 0: context.go('/match');
      case 1: context.go('/chat');
      case 2: context.go('/profile');
    }
  }
}
