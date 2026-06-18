import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/constants/zodiac_helper.dart';
import 'package:spark_app/presentation/widgets/step_indicator.dart';

class BirthDateScreen extends ConsumerStatefulWidget {
  const BirthDateScreen({super.key});

  @override
  ConsumerState<BirthDateScreen> createState() => _BirthDateScreenState();
}

class _BirthDateScreenState extends ConsumerState<BirthDateScreen> {
  DateTime? _selected;
  bool _saving = false;

  String? get _zodiac {
    if (_selected == null) return null;
    return zodiacFromBirth(
      '${_selected!.year}-${_selected!.month.toString().padLeft(2, '0')}-${_selected!.day.toString().padLeft(2, '0')}',
    );
  }

  Future<void> _save() async {
    if (_selected == null) return;
    setState(() => _saving = true);
    try {
      final api = ref.read(apiClientProvider);
      final dateStr =
          '${_selected!.year}-${_selected!.month.toString().padLeft(2, '0')}-${_selected!.day.toString().padLeft(2, '0')}';
      await api.put('/api/v1/user/profile', data: {'birth_date': dateStr});
      if (mounted) context.push('/onboarding/gender');
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('保存失败，请重试')),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  Future<void> _pickDate() async {
    final now = DateTime.now();
    final picked = await showDatePicker(
      context: context,
      initialDate: _selected ?? DateTime(2000, 1, 1),
      firstDate: DateTime(1970, 1, 1),
      lastDate: DateTime(now.year - 16, now.month, now.day),
      builder: (context, child) => Theme(
        data: Theme.of(context).copyWith(
          colorScheme: const ColorScheme.dark(
            primary: Color(0xFF6C5CE7),
            surface: Color(0xFF1A1A2E),
          ),
        ),
        child: child!,
      ),
    );
    if (picked != null) setState(() => _selected = picked);
  }

  @override
  Widget build(BuildContext context) {
    final zodiac = _zodiac;
    final trait = zodiac != null ? zodiacTraits[zodiac] : null;

    return Scaffold(
      appBar: AppBar(
        title: const Text('你的生日'),
        actions: [
          TextButton(
            onPressed: () => context.push('/onboarding/gender'),
            child: const Text('跳过'),
          ),
        ],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            children: [
              const Spacer(flex: 2),
              const Icon(Icons.cake_outlined, size: 64, color: Color(0xFF6C5CE7)),
              const SizedBox(height: 24),
              Text('你的生日是？',
                  style: Theme.of(context).textTheme.titleLarge),
              const SizedBox(height: 8),
              Text('你的星座帮助我们找到更合适的匹配',
                  style: TextStyle(color: Colors.grey[500], fontSize: 14)),
              const SizedBox(height: 16),
              const StepIndicator(total: 4, current: 2),
              const SizedBox(height: 24),
              GestureDetector(
                onTap: _pickDate,
                child: Container(
                  padding: const EdgeInsets.symmetric(vertical: 20, horizontal: 32),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: _selected != null
                          ? const Color(0xFF6C5CE7)
                          : Colors.grey[700]!,
                    ),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.calendar_today, color: Color(0xFF6C5CE7)),
                      const SizedBox(width: 12),
                      Text(
                        _selected != null
                            ? '${_selected!.year}-${_selected!.month.toString().padLeft(2, '0')}-${_selected!.day.toString().padLeft(2, '0')}'
                            : '点击选择',
                        style: TextStyle(
                          fontSize: 18,
                          color: _selected != null ? Colors.white : Colors.grey[500],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              if (_selected != null && zodiac != null) ...[
                const SizedBox(height: 32),
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        (trait?.color ?? const Color(0xFF6C5CE7)).withValues(alpha: 0.2),
                        const Color(0xFF1A1A2E),
                      ],
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                    ),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(
                        color: (trait?.color ?? const Color(0xFF6C5CE7)).withValues(alpha: 0.3)),
                  ),
                  child: Column(
                    children: [
                      Text(zodiac, style: const TextStyle(fontSize: 48)),
                      const SizedBox(height: 8),
                      Text(zodiacEmojiToName(zodiac),
                          style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w600)),
                      if (trait != null) ...[
                        const SizedBox(height: 4),
                        Text(trait.strength,
                            style: TextStyle(color: Colors.grey[400], fontSize: 13)),
                      ],
                    ],
                  ),
                ),
              ],
              const Spacer(),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _selected != null && !_saving ? _save : null,
                  child: _saving
                      ? const SizedBox(height: 20, width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2))
                      : const Text('继续'),
                ),
              ),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }
}
