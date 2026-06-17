import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/network/api_client.dart';

final horoscopeProvider = FutureProvider<Map<String, dynamic>?>((ref) async {
  try {
    final api = ref.read(apiClientProvider);
    final resp = await api.get('/api/v1/user/horoscope');
    return resp.data;
  } catch (_) {
    return null;
  }
});

class HoroscopeBanner extends ConsumerWidget {
  const HoroscopeBanner({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final data = ref.watch(horoscopeProvider);

    return data.when(
      data: (d) {
        if (d == null) return const SizedBox.shrink();
        return GestureDetector(
          onTap: () => context.push('/horoscope'),
          child: Container(
            margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [Color(0xFF2D1B69), Color(0xFF1A1A2E)],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: const Color(0xFF6C5CE7).withValues(alpha: 0.3)),
            ),
            child: Row(
              children: [
                Text(d['zodiac'] ?? '', style: const TextStyle(fontSize: 28)),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.auto_awesome, size: 14, color: Color(0xFFFDCB6E)),
                          const SizedBox(width: 4),
                          Text('Daily Horoscope',
                              style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                        ],
                      ),
                      const SizedBox(height: 4),
                      Text(
                        d['horoscope'] ?? '',
                        style: TextStyle(color: Colors.grey[300], height: 1.4, fontSize: 13),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
                const Icon(Icons.chevron_right, color: Colors.white24),
              ],
            ),
          ),
        );
      },
      loading: () => const SizedBox.shrink(),
      error: (_, __) => const SizedBox.shrink(),
    );
  }
}
