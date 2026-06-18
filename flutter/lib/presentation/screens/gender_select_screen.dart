import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/presentation/widgets/step_indicator.dart';

class GenderSelectScreen extends ConsumerStatefulWidget {
  const GenderSelectScreen({super.key});

  @override
  ConsumerState<GenderSelectScreen> createState() => _GenderSelectScreenState();
}

class _GenderSelectScreenState extends ConsumerState<GenderSelectScreen> {
  int _selected = 0;
  final _cityCtrl = TextEditingController();
  bool _saving = false;

  @override
  void dispose() {
    _cityCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    try {
      final data = <String, dynamic>{'gender': _selected};
      final city = _cityCtrl.text.trim();
      if (city.isNotEmpty) data['city'] = city;
      await ref.read(apiClientProvider).put('/api/v1/user/profile', data: data);
      if (mounted) context.push('/onboarding/avatar');
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('保存失败')),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('我是...'),
        actions: [
          TextButton(
            onPressed: () => context.push('/onboarding/avatar'),
            child: const Text('跳过'),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            const SizedBox(height: 32),
            Text(
              '你的性别是？',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(
              '这将帮助我们为你找到更合适的匹配',
              style: TextStyle(color: Colors.grey[500], fontSize: 14),
            ),
            const SizedBox(height: 16),
            const StepIndicator(total: 4, current: 3),
            const SizedBox(height: 32),
            _genderCard(0, '保密', Icons.visibility_off, '不愿透露'),
            const SizedBox(height: 16),
            _genderCard(1, '男', Icons.male, '向我展示女性资料'),
            const SizedBox(height: 16),
            _genderCard(2, '女', Icons.female, '向我展示男性资料'),
            const SizedBox(height: 32),
            TextField(
              controller: _cityCtrl,
              decoration: const InputDecoration(
                hintText: '你的城市（选填）',
                prefixIcon: Icon(Icons.location_on_outlined),
              ),
            ),
            const Spacer(),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _saving ? null : _save,
                child: _saving
                    ? const SizedBox(height: 20, width: 20,
                        child: CircularProgressIndicator(strokeWidth: 2))
                    : const Text('继续'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _genderCard(int value, String label, IconData icon, String subtitle) {
    final selected = _selected == value;
    return GestureDetector(
      onTap: () => setState(() => _selected = value),
      child: Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: selected ? const Color(0xFF6C5CE7) : Colors.grey[800]!,
            width: selected ? 2 : 1,
          ),
          color: selected
              ? const Color(0xFF6C5CE7).withValues(alpha: 0.1)
              : Colors.transparent,
        ),
        child: Row(
          children: [
            Icon(icon,
                color: selected ? const Color(0xFF6C5CE7) : Colors.grey,
                size: 28),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(label,
                      style: TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.w600,
                        color: selected ? Colors.white : Colors.grey[300],
                      )),
                  const SizedBox(height: 2),
                  Text(subtitle,
                      style: TextStyle(
                          color: Colors.grey[500], fontSize: 13)),
                ],
              ),
            ),
            if (selected)
              const Icon(Icons.check_circle, color: Color(0xFF6C5CE7)),
          ],
        ),
      ),
    );
  }
}
