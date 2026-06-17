import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

final discoveryGenderProvider = StateProvider<int>((ref) => 0);
final discoveryMinAgeProvider = StateProvider<int>((ref) => 18);
final discoveryMaxAgeProvider = StateProvider<int>((ref) => 45);

class DiscoveryPrefsScreen extends ConsumerWidget {
  const DiscoveryPrefsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final gender = ref.watch(discoveryGenderProvider);
    final minAge = ref.watch(discoveryMinAgeProvider);
    final maxAge = ref.watch(discoveryMaxAgeProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Discovery Preferences')),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          // Gender
          const Text('Show me', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          SegmentedButton<int>(
            segments: const [
              ButtonSegment(value: 0, label: Text('Everyone')),
              ButtonSegment(value: 1, label: Text('Male')),
              ButtonSegment(value: 2, label: Text('Female')),
            ],
            selected: {gender},
            onSelectionChanged: (s) =>
                ref.read(discoveryGenderProvider.notifier).state = s.first,
          ),

          const SizedBox(height: 32),

          // Age range
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Age range', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
              Text('$minAge - $maxAge',
                  style: TextStyle(color: Colors.grey[400], fontSize: 16)),
            ],
          ),
          const SizedBox(height: 12),
          RangeSlider(
            values: RangeValues(minAge.toDouble(), maxAge.toDouble()),
            min: 18,
            max: 65,
            divisions: 47,
            activeColor: const Color(0xFF6C5CE7),
            labels: RangeLabels('$minAge', '$maxAge'),
            onChanged: (v) {
              ref.read(discoveryMinAgeProvider.notifier).state = v.start.round();
              ref.read(discoveryMaxAgeProvider.notifier).state = v.end.round();
            },
          ),

          const SizedBox(height: 48),

          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: () => context.pop(),
              child: const Text('Apply'),
            ),
          ),
        ],
      ),
    );
  }
}
