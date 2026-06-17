import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/constants/zodiac_helper.dart';
import 'package:spark_app/data/providers/auth_provider.dart';

class ZodiacCompatScreen extends ConsumerStatefulWidget {
  final String targetUserId;
  final String targetName;
  final String? targetBirthDate;
  const ZodiacCompatScreen({
    super.key,
    required this.targetUserId,
    required this.targetName,
    this.targetBirthDate,
  });

  @override
  ConsumerState<ZodiacCompatScreen> createState() => _ZodiacCompatScreenState();
}

class _ZodiacCompatScreenState extends ConsumerState<ZodiacCompatScreen> {
  Map<String, dynamic>? _data;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.get('/api/v1/match/zodiac-compat/${widget.targetUserId}');
      setState(() {
        _data = resp.data;
        _loading = false;
      });
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Zodiac Compatibility')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _data == null
              ? const Center(child: Text('Could not load compatibility data'))
              : _buildContent(),
    );
  }

  Widget _buildContent() {
    final userZodiac = _data!['user_zodiac'] as String? ?? '';
    final targetZodiac = _data!['target_zodiac'] as String? ?? '';
    final score = (_data!['score'] as num?)?.toDouble() ?? 50;
    final report = _data!['report'] as String? ?? '';
    final userTrait = zodiacTraits[userZodiac];
    final targetTrait = zodiacTraits[targetZodiac];
    final myName = ref.read(authProvider).valueOrNull?.nickname ?? 'You';

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        // Score circle
        Center(
          child: SizedBox(
            width: 160,
            height: 160,
            child: Stack(
              alignment: Alignment.center,
              children: [
                SizedBox(
                  width: 160,
                  height: 160,
                  child: CircularProgressIndicator(
                    value: score / 100,
                    strokeWidth: 10,
                    backgroundColor: Colors.grey[800],
                    valueColor: AlwaysStoppedAnimation<Color>(_scoreColor(score)),
                  ),
                ),
                Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text('${score.toInt()}%',
                        style: TextStyle(
                            fontSize: 36,
                            fontWeight: FontWeight.bold,
                            color: _scoreColor(score))),
                    Text('Match', style: TextStyle(color: Colors.grey[500], fontSize: 13)),
                  ],
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 32),

        // Two zodiac signs facing each other
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            _zodiacCard(myName, userZodiac, userTrait),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Icon(Icons.favorite, color: _scoreColor(score), size: 28),
            ),
            _zodiacCard(widget.targetName, targetZodiac, targetTrait),
          ],
        ),
        const SizedBox(height: 24),

        // Report
        Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [
                const Color(0xFF6C5CE7).withValues(alpha: 0.15),
                const Color(0xFFE17055).withValues(alpha: 0.15),
              ],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
                color: const Color(0xFF6C5CE7).withValues(alpha: 0.2)),
          ),
          child: Text(report,
              style: TextStyle(color: Colors.grey[300], height: 1.8, fontSize: 15)),
        ),
      ],
    );
  }

  Widget _zodiacCard(String name, String zodiac, ZodiacTrait? trait) {
    return Container(
      width: 120,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1A2E),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        children: [
          Text(zodiac, style: const TextStyle(fontSize: 40)),
          const SizedBox(height: 4),
          Text(name,
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis),
          if (trait != null) ...[
            const SizedBox(height: 4),
            Text(trait.element, style: TextStyle(color: trait.color, fontSize: 12)),
          ],
        ],
      ),
    );
  }

  Color _scoreColor(double score) {
    if (score >= 80) return const Color(0xFF00B894);
    if (score >= 60) return const Color(0xFFFDCB6E);
    if (score >= 40) return Colors.orangeAccent;
    return Colors.redAccent;
  }
}
