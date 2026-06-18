import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/constants/zodiac_helper.dart';

class HoroscopeScreen extends ConsumerStatefulWidget {
  const HoroscopeScreen({super.key});

  @override
  ConsumerState<HoroscopeScreen> createState() => _HoroscopeScreenState();
}

class _HoroscopeScreenState extends ConsumerState<HoroscopeScreen> {
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
      final resp = await api.get('/api/v1/user/horoscope');
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
      appBar: AppBar(title: const Text('每日星座运势')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _data == null
              ? const Center(
                  child: Column(mainAxisSize: MainAxisSize.min, children: [
                    Icon(Icons.error_outline, size: 48, color: Colors.grey),
                    SizedBox(height: 12),
                    Text('设置生日查看你的星座运势',
                        style: TextStyle(color: Colors.grey)),
                  ]),
                )
              : _buildContent(),
    );
  }

  Widget _buildContent() {
    final zodiac = _data!['zodiac'] as String? ?? '';
    final horoscope = _data!['horoscope'] as String? ?? '';
    final trait = zodiacTraits[zodiac];
    final name = zodiacEmojiToName(zodiac);

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        // Main zodiac card
        Container(
          padding: const EdgeInsets.all(32),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [
                trait?.color ?? const Color(0xFF6C5CE7),
                (trait?.color ?? const Color(0xFF6C5CE7)).withValues(alpha: 0.3),
              ],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(24),
          ),
          child: Column(
            children: [
              Text(zodiac, style: const TextStyle(fontSize: 72)),
              const SizedBox(height: 8),
              Text(name,
                  style: const TextStyle(fontSize: 28, fontWeight: FontWeight.bold, color: Colors.white)),
              if (trait != null) ...[
                const SizedBox(height: 8),
                Text('${trait.element} · ${trait.strength}',
                    style: const TextStyle(color: Colors.white70, fontSize: 14),
                    textAlign: TextAlign.center),
              ],
            ],
          ),
        ),
        const SizedBox(height: 24),

        // Horoscope message
        Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: const Color(0xFF1A1A2E),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.white12),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: const Color(0xFF6C5CE7).withValues(alpha: 0.3),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(DateTime.now().toString().split(' ')[0],
                        style: const TextStyle(color: Color(0xFF6C5CE7), fontSize: 12)),
                  ),
                  const Spacer(),
                  const Icon(Icons.auto_awesome, size: 16, color: Color(0xFFFDCB6E)),
                  const SizedBox(width: 4),
                  const Text('火花星座运势',
                      style: TextStyle(color: Color(0xFFFDCB6E), fontSize: 12)),
                ],
              ),
              const SizedBox(height: 16),
              Text(horoscope,
                  style: TextStyle(color: Colors.grey[300], height: 1.8, fontSize: 15)),
            ],
          ),
        ),
        const SizedBox(height: 24),

        // Element info
        if (trait != null) _elementCard(trait),
      ],
    );
  }

  Widget _elementCard(ZodiacTrait trait) {
    final elementDesc = {
      '风象': '风象星座的人思维敏捷、善于沟通，是天生的社交家和创新者。',
      '水象': '水象星座的人情感丰富、直觉敏锐，拥有强大的共情能力和创造力。',
      '火象': '火象星座的人热情奔放、行动力强，是天生的领导者和冒险家。',
      '土象': '土象星座的人踏实稳重、注重实际，是最可靠的朋友和伴侣。',
    };

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1A2E),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.wb_sunny_outlined, size: 20, color: trait.color),
              const SizedBox(width: 8),
              Text('${trait.element}星座',
                  style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
            ],
          ),
          const SizedBox(height: 8),
          Text(elementDesc[trait.element] ?? '',
              style: TextStyle(color: Colors.grey[400], height: 1.6, fontSize: 14)),
        ],
      ),
    );
  }
}
