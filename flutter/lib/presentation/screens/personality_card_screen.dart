import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/network/api_client.dart';

class PersonalityCardScreen extends ConsumerStatefulWidget {
  const PersonalityCardScreen({super.key});

  @override
  ConsumerState<PersonalityCardScreen> createState() => _PersonalityCardScreenState();
}

class _PersonalityCardScreenState extends ConsumerState<PersonalityCardScreen> {
  Map<String, dynamic>? _report;
  List<Map<String, dynamic>>? _dimensions;
  bool _loading = true;
  bool _sharing = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.get('/api/v1/user/personality/report');
      final dimsResp = await api.get('/api/v1/user/personality');
      setState(() {
        _report = resp.data;
        _dimensions = (dimsResp.data as List<dynamic>?)?.cast<Map<String, dynamic>>();
        _loading = false;
      });
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  Future<void> _shareCard() async {
    setState(() => _sharing = true);
    try {
      final r = _report;
      if (r != null) {
        final text = '${r['title'] ?? ''}\n${r['summary'] ?? ''}\n#SparkApp';
        await Clipboard.setData(ClipboardData(text: text));
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Personality card copied to clipboard')),
          );
        }
      }
    } finally {
      if (mounted) setState(() => _sharing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Personality Card'),
        actions: [
          if (_report != null)
            IconButton(
              icon: _sharing
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Icon(Icons.share_outlined),
              onPressed: _sharing ? null : _shareCard,
            ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _report == null
              ? Center(
                  child: Column(mainAxisSize: MainAxisSize.min, children: [
                    const Icon(Icons.psychology_outlined, size: 64, color: Colors.grey),
                    const SizedBox(height: 16),
                    const Text('Complete the personality quiz first',
                        style: TextStyle(color: Colors.grey)),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: () => context.push('/personality-quiz'),
                      child: const Text('Take Personality Quiz'),
                    ),
                  ]),
                )
              : _buildCard(context),
    );
  }

  Widget _buildCard(BuildContext context) {
    final r = _report!;
    final traits = (r['traits'] as List<dynamic>?)?.cast<String>() ?? [];

    return ListView(
        padding: const EdgeInsets.all(24),
        children: [
        // Title card
        Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: [Color(0xFF6C5CE7), Color(0xFFE17055)],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Column(
            children: [
              const Icon(Icons.auto_awesome, size: 36, color: Colors.white),
              const SizedBox(height: 12),
              Text(
                r['title'] ?? '',
                style: const TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        // Summary
        Text(
          r['summary'] ?? '',
          style: TextStyle(color: Colors.grey[300], height: 1.6, fontSize: 15),
        ),
        const SizedBox(height: 24),
        // Dimension bars
        if (_dimensions != null)
          ..._dimensions!.map((d) => _dimensionBar(
                d['dimension'] ?? '',
                (d['score'] as num?)?.toDouble() ?? 0,
              )),
        const SizedBox(height: 24),
        // Traits
        Text('Your Traits', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: traits
              .map((t) => Chip(
                    label: Text(t),
                    backgroundColor: const Color(0xFF6C5CE7).withValues(alpha: 0.2),
                  ))
              .toList(),
        ),
        const SizedBox(height: 24),
        // Extraversion detail
        if (r['extraversion_detail'] != null && r['extraversion_detail'].toString().isNotEmpty)
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.grey[900],
              borderRadius: BorderRadius.circular(12),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.lightbulb_outline, color: Color(0xFFE17055)),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    r['extraversion_detail'] ?? '',
                    style: TextStyle(color: Colors.grey[300], height: 1.5),
                  ),
                ),
              ],
            ),
          ),
        const SizedBox(height: 24),
        // Advice
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: const Color(0xFF6C5CE7).withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const Color(0xFF6C5CE7).withValues(alpha: 0.3)),
          ),
          child: Text(
            r['advice'] ?? '',
            style: TextStyle(color: Colors.grey[300], height: 1.6),
          ),
        ),
      ],
  );
  }

  Widget _dimensionBar(String dim, double score) {
    final label = _dimLabel(dim);
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(label, style: TextStyle(color: Colors.grey[400], fontSize: 13)),
              Text(score.toStringAsFixed(1),
                  style: TextStyle(color: Colors.grey[500], fontSize: 13)),
            ],
          ),
          const SizedBox(height: 4),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: score / 5.0,
              minHeight: 8,
              backgroundColor: Colors.grey[800],
              valueColor: AlwaysStoppedAnimation<Color>(_dimColor(dim)),
            ),
          ),
        ],
      ),
    );
  }

  String _dimLabel(String dim) {
    switch (dim) {
      case 'extraversion':
        return 'Extraversion';
      case 'agreeableness':
        return 'Agreeableness';
      case 'conscientiousness':
        return 'Conscientiousness';
      case 'neuroticism':
        return 'Emotional Stability';
      case 'openness':
        return 'Openness';
      default:
        return dim;
    }
  }

  Color _dimColor(String dim) {
    switch (dim) {
      case 'extraversion':
        return const Color(0xFFE17055);
      case 'agreeableness':
        return const Color(0xFF00B894);
      case 'conscientiousness':
        return const Color(0xFF0984E3);
      case 'neuroticism':
        return const Color(0xFF6C5CE7);
      case 'openness':
        return const Color(0xFFFDCB6E);
      default:
        return Colors.grey;
    }
  }
}
