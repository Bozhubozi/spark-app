import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/data/models/user_model.dart';
import 'package:spark_app/presentation/widgets/step_indicator.dart';

class AvatarSelectScreen extends ConsumerStatefulWidget {
  const AvatarSelectScreen({super.key});

  @override
  ConsumerState<AvatarSelectScreen> createState() => _AvatarSelectScreenState();
}

class _AvatarSelectScreenState extends ConsumerState<AvatarSelectScreen> {
  List<AvatarComponent> _components = [];
  String _selectedCategory = 'face';
  final Map<String, int> _selected = {};
  bool _loading = true;

  static const _order = ['face', 'hair', 'eyes', 'clothes', 'accessory', 'background'];

  @override
  void initState() {
    super.initState();
    _loadComponents();
  }

  Future<void> _loadComponents() async {
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.get('/api/v1/user/avatars');
      setState(() {
        _components =
            (resp.data as List<dynamic>).map((e) => AvatarComponent.fromJson(e)).toList();
        _loading = false;
      });
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  Map<String, List<AvatarComponent>> get _grouped {
    final map = <String, List<AvatarComponent>>{};
    for (final c in _components) {
      map.putIfAbsent(c.category, () => []).add(c);
      if (!_selected.containsKey(c.category) && map[c.category]!.isNotEmpty) {
        _selected[c.category] = map[c.category]!.first.id;
      }
    }
    return map;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Build Your Avatar'),
        actions: [
          TextButton(
            onPressed: () => context.push('/personality-quiz'),
            child: const Text('Skip'),
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                // Preview
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(vertical: 32),
                  decoration: BoxDecoration(
                    color: Colors.grey[900],
                    border: Border(bottom: BorderSide(color: Colors.grey[800]!)),
                  ),
                  child: Column(
                    children: [
                      _avatarPreview(),
                      const SizedBox(height: 12),
                      Text(
                        _selectedName() ?? 'Select components',
                        style: TextStyle(color: Colors.grey[400]),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 8),
                const StepIndicator(total: 4, current: 4),
                const SizedBox(height: 8),
                // Category tabs
                SizedBox(
                  height: 48,
                  child: ListView(
                    scrollDirection: Axis.horizontal,
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                    children: _order.where((c) => _grouped.containsKey(c)).map((cat) {
                      final selected = _selectedCategory == cat;
                      return Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 4),
                        child: ChoiceChip(
                          label: Text(_catLabel(cat)),
                          selected: selected,
                          onSelected: (_) => setState(() => _selectedCategory = cat),
                          selectedColor: const Color(0xFF6C5CE7).withValues(alpha: 0.3),
                        ),
                      );
                    }).toList(),
                  ),
                ),
                const Divider(height: 1),
                // Component grid
                Expanded(
                  child: GridView.builder(
                    padding: const EdgeInsets.all(12),
                    gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 3,
                      mainAxisSpacing: 8,
                      crossAxisSpacing: 8,
                      childAspectRatio: 0.8,
                    ),
                    itemCount: _grouped[_selectedCategory]?.length ?? 0,
                    itemBuilder: (ctx, i) {
                      final comp = _grouped[_selectedCategory]![i];
                      final isSelected = _selected[_selectedCategory] == comp.id;
                      return GestureDetector(
                        onTap: () => setState(() => _selected[_selectedCategory] = comp.id),
                        child: Container(
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(12),
                            color: isSelected
                                ? const Color(0xFF6C5CE7).withValues(alpha: 0.2)
                                : Colors.grey[850],
                            border: isSelected
                                ? Border.all(color: const Color(0xFF6C5CE7), width: 2)
                                : null,
                          ),
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(_catIcon(comp.category),
                                  size: 32, color: Colors.grey[400]),
                              const SizedBox(height: 6),
                              Text(comp.name,
                                  textAlign: TextAlign.center,
                                  style: TextStyle(fontSize: 11, color: Colors.grey[400])),
                              if (comp.rarity > 1)
                                Padding(
                                  padding: const EdgeInsets.only(top: 4),
                                  child: Icon(Icons.star,
                                      size: 12,
                                      color: comp.rarity == 3
                                          ? const Color(0xFFE17055)
                                          : const Color(0xFFFDCB6E)),
                                ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
                SafeArea(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: () => context.push('/personality-quiz'),
                        child: const Text('Continue'),
                      ),
                    ),
                  ),
                ),
              ],
            ),
    );
  }

  Widget _avatarPreview() {
    return Container(
      width: 100,
      height: 100,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: Colors.grey[800],
        border: Border.all(color: const Color(0xFF6C5CE7), width: 2),
      ),
      child: const Icon(Icons.person, size: 48, color: Colors.grey),
    );
  }

  String? _selectedName() {
    for (final cat in _order) {
      final id = _selected[cat];
      if (id != null) {
        for (final c in _components) {
          if (c.id == id) return c.name;
        }
      }
    }
    return null;
  }

  String _catLabel(String cat) {
    switch (cat) {
      case 'face':
        return 'Face';
      case 'hair':
        return 'Hair';
      case 'eyes':
        return 'Eyes';
      case 'clothes':
        return 'Clothes';
      case 'accessory':
        return 'Accessory';
      case 'background':
        return 'BG';
      default:
        return cat;
    }
  }

  IconData _catIcon(String cat) {
    switch (cat) {
      case 'face':
        return Icons.face;
      case 'hair':
        return Icons.face_3;
      case 'eyes':
        return Icons.remove_red_eye_outlined;
      case 'clothes':
        return Icons.checkroom;
      case 'accessory':
        return Icons.auto_awesome;
      case 'background':
        return Icons.landscape;
      default:
        return Icons.category;
    }
  }
}
