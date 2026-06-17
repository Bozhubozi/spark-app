import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/tracking/tracker.dart';
import 'package:spark_app/data/models/user_model.dart';
import 'package:spark_app/presentation/widgets/step_indicator.dart';

class InterestSelectScreen extends ConsumerStatefulWidget {
  final bool editMode;
  final List<int> initialTagIds;
  final VoidCallback? onSaved;
  const InterestSelectScreen({super.key, this.editMode = false, this.initialTagIds = const [], this.onSaved});

  @override
  ConsumerState<InterestSelectScreen> createState() => _InterestSelectScreenState();
}

class _InterestSelectScreenState extends ConsumerState<InterestSelectScreen> {
  List<InterestTag> _allTags = [];
  final Set<int> _selected = {};
  bool _loading = true;
  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    _selected.addAll(widget.initialTagIds);
    _loadTags();
  }

  Future<void> _loadTags() async {
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.get('/api/v1/user/tags');
      final list =
          (resp.data as List<dynamic>).map((e) => InterestTag.fromJson(e)).toList();
      setState(() {
        _allTags = list;
        _loading = false;
      });
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  Map<String, List<InterestTag>> get _grouped {
    final map = <String, List<InterestTag>>{};
    for (final t in _allTags) {
      map.putIfAbsent(t.category, () => []).add(t);
    }
    return map;
  }

  Future<void> _submit() async {
    setState(() => _submitting = true);
    try {
      final api = ref.read(apiClientProvider);
      await api.put('/api/v1/user/interests', data: {'tag_ids': _selected.toList()});
      Tracker().track('interest_tag_submit', properties: {
        'tag_ids': _selected.toList(),
        'total_count': _selected.length,
      });
      if (mounted) {
        if (widget.editMode) {
          widget.onSaved?.call();
          Navigator.of(context).pop();
        } else {
          context.push('/onboarding/birthdate');
        }
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Save failed, please retry')),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Pick Your Interests'),
        actions: [
          if (widget.editMode)
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Cancel'),
            )
          else
            TextButton(
              onPressed: () => context.push('/onboarding/birthdate'),
              child: const Text('Skip'),
            ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Text(
                    widget.editMode ? 'Update your interests' : 'Select at least 3 interests to find your spark',
                    style: TextStyle(color: Colors.grey[400], fontSize: 14),
                  ),
                ),
                if (!widget.editMode) const StepIndicator(total: 4, current: 1),
                Expanded(
                  child: ListView(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    children: _grouped.entries.map((entry) {
                      return _buildCategory(entry.key, entry.value);
                    }).toList(),
                  ),
                ),
                SafeArea(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: (widget.editMode || _selected.length >= 3) && !_submitting ? _submit : null,
                        child: _submitting
                            ? const SizedBox(
                                height: 20,
                                width: 20,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : Text(widget.editMode
                                  ? 'Save (${_selected.length} selected)'
                                  : 'Continue (${_selected.length} selected)'),
                      ),
                    ),
                  ),
                ),
              ],
            ),
    );
  }

  Widget _buildCategory(String category, List<InterestTag> tags) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 16),
        Text(
          _categoryLabel(category),
          style: TextStyle(
            color: Colors.grey[500],
            fontSize: 13,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.5,
          ),
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: tags.map((t) => _buildChip(t)).toList(),
        ),
      ],
    );
  }

  Widget _buildChip(InterestTag tag) {
    final selected = _selected.contains(tag.id);
    return FilterChip(
      label: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (tag.icon != null) Text('${tag.icon} '),
          Text(tag.name),
        ],
      ),
      selected: selected,
      onSelected: (val) {
        setState(() {
          if (val) {
            _selected.add(tag.id);
          } else {
            _selected.remove(tag.id);
          }
        });
      },
      selectedColor: const Color(0xFF6C5CE7).withValues(alpha: 0.3),
      checkmarkColor: const Color(0xFF6C5CE7),
      side: BorderSide(
        color: selected ? const Color(0xFF6C5CE7) : Colors.grey[700]!,
      ),
    );
  }

  String _categoryLabel(String cat) {
    switch (cat) {
      case 'entertainment':
        return 'Entertainment';
      case 'music':
        return 'Music';
      case 'gaming':
        return 'Gaming';
      case 'social':
        return 'Social';
      case 'outdoor':
        return 'Outdoor';
      case 'sports':
        return 'Sports';
      case 'art':
        return 'Art';
      case 'lifestyle':
        return 'Lifestyle';
      default:
        return cat[0].toUpperCase() + cat.substring(1);
    }
  }
}
