import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/data/providers/auth_provider.dart';
import 'package:spark_app/presentation/screens/interest_select_screen.dart';
import 'package:spark_app/presentation/widgets/shimmer_loading.dart';

class ProfileScreen extends ConsumerStatefulWidget {
  const ProfileScreen({super.key});

  @override
  ConsumerState<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends ConsumerState<ProfileScreen> {
  bool _editing = false;
  bool _saving = false;
  late TextEditingController _nicknameCtrl;
  late TextEditingController _bioCtrl;
  late TextEditingController _cityCtrl;
  DateTime? _birthDate;
  int _gender = 0;

  @override
  void initState() {
    super.initState();
    _nicknameCtrl = TextEditingController();
    _bioCtrl = TextEditingController();
    _cityCtrl = TextEditingController();
  }

  @override
  void dispose() {
    _nicknameCtrl.dispose();
    _bioCtrl.dispose();
    _cityCtrl.dispose();
    super.dispose();
  }

  void _enterEdit(u) {
    _nicknameCtrl.text = u.nickname;
    _bioCtrl.text = u.bio ?? '';
    _cityCtrl.text = u.city ?? '';
    _gender = u.gender;
    _birthDate = u.birthDate != null ? DateTime.tryParse(u.birthDate!) : null;
    setState(() => _editing = true);
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    try {
      final api = ref.read(apiClientProvider);
      final data = <String, dynamic>{
        'nickname': _nicknameCtrl.text.trim(),
        'bio': _bioCtrl.text.trim(),
        'city': _cityCtrl.text.trim(),
        'gender': _gender,
      };
      if (_birthDate != null) {
        data['birth_date'] =
            '${_birthDate!.year}-${_birthDate!.month.toString().padLeft(2, '0')}-${_birthDate!.day.toString().padLeft(2, '0')}';
      }
      await api.put('/api/v1/user/profile', data: data);
      ref.invalidate(authProvider);
      setState(() => _editing = false);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('保存失败：$e')),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  Future<void> _restoreAccount() async {
    try {
      final api = ref.read(apiClientProvider);
      await api.post('/api/v1/user/account/restore');
      ref.invalidate(authProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('账号已恢复！')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('恢复失败：$e')),
        );
      }
    }
  }

  Future<void> _confirmDelete() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('注销账号'),
        content: const Text(
          'Your account will be deactivated for 7 days.\n'
          '在此期间重新登录即可恢复.\n\n'
          '7 天后数据将被永久删除.',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('取消')),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: TextButton.styleFrom(foregroundColor: Colors.redAccent),
            child: const Text('删除'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    try {
      final api = ref.read(apiClientProvider);
      await api.post('/api/v1/user/account/cancel');
      await ref.read(authProvider.notifier).logout();
      if (mounted) context.go('/login');
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('失败：$e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('个人资料'),
        actions: [
          if (!_editing)
            IconButton(
              icon: const Icon(Icons.edit_outlined),
              onPressed: () {
                final u = user.valueOrNull;
                if (u != null) _enterEdit(u);
              },
            ),
        ],
      ),
      body: user.when(
        data: (u) {
          if (u == null) return const Center(child: Text('未登录'));
          if (_editing) return _buildEditForm(u);
          return _buildView(u);
        },
        loading: () => ListView(children: const [
          SizedBox(height: 32),
          SkeletonListTile(), SkeletonListTile(), SkeletonListTile(),
          SkeletonListTile(), SkeletonListTile(),
        ]),
        error: (e, _) => Center(child: Text('错误：$e')),
      ),
    );
  }

  Widget _buildView(u) {
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        const SizedBox(height: 24),
        Center(
          child: CircleAvatar(
            radius: 48,
            backgroundColor: const Color(0xFF6C5CE7),
            child: Text(u.nickname[0].toUpperCase(),
                style: const TextStyle(fontSize: 40, color: Colors.white)),
          ),
        ),
        const SizedBox(height: 16),
        Center(
          child: Text(u.nickname, style: Theme.of(context).textTheme.titleLarge),
        ),
        if (u.bio != null)
          Center(
            child: Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text(u.bio!, style: TextStyle(color: Colors.grey[500])),
            ),
          ),
        const SizedBox(height: 32),
        _infoTile(Icons.person_outline, '昵称', u.nickname),
        _infoTile(Icons.phone_outlined, '手机号', u.phone ?? '-'),
        _infoTile(Icons.email_outlined, '邮箱', u.email ?? '-'),
        _infoTile(Icons.location_on_outlined, '城市', u.city ?? '-'),
        _infoTile(Icons.male_outlined, '性别', _genderLabel(u.gender)),
        if (u.birthDate != null)
          _infoTile(Icons.cake_outlined, '生日', u.birthDate!),
        const Divider(height: 32),
        ListTile(
          leading: const Icon(Icons.interests_outlined),
          title: const Text('兴趣'),
          subtitle: Text(
            u.interests.isNotEmpty
                ? '${u.interests.length} 个'
                : '未设置',
            style: TextStyle(color: Colors.grey[600], fontSize: 13),
          ),
          trailing: const Icon(Icons.chevron_right),
          onTap: () async {
            await Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => InterestSelectScreen(
                  editMode: true,
                  initialTagIds: u.interests.map((t) => t.id).toList(),
                  onSaved: () => ref.invalidate(authProvider),
                ),
              ),
            );
            ref.invalidate(authProvider);
          },
        ),
        const Divider(height: 32),
        ListTile(
          leading: const Icon(Icons.favorite_border),
          title: const Text('我的匹配'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => context.push('/matches'),
        ),
        ListTile(
          leading: const Icon(Icons.auto_awesome),
          title: const Text('每日星座运势'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => context.push('/horoscope'),
        ),
        ListTile(
          leading: const Icon(Icons.psychology_outlined),
          title: const Text('人格档案'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => context.push('/profile/personality'),
        ),
        ListTile(
          leading: const Icon(Icons.edit_outlined),
          title: const Text('编辑资料'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => _enterEdit(u),
        ),
        ListTile(
          leading: const Icon(Icons.settings_outlined),
          title: const Text('设置'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => context.push('/settings'),
        ),
        const Divider(height: 32),
        if (u.deletedAt != null)
          Container(
            padding: const EdgeInsets.all(16),
            margin: const EdgeInsets.only(bottom: 12),
            decoration: BoxDecoration(
              color: Colors.orangeAccent.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.orangeAccent.withValues(alpha: 0.3)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Row(
                  children: [
                    Icon(Icons.warning_amber, color: Colors.orangeAccent, size: 20),
                    SizedBox(width: 8),
                    Text('账号已申请注销',
                        style: TextStyle(color: Colors.orangeAccent, fontWeight: FontWeight.w600)),
                  ],
                ),
                const SizedBox(height: 8),
                Text('你的账号即将被永久删除。恢复后可继续使用火花。',
                    style: TextStyle(color: Colors.grey[400], fontSize: 13)),
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _restoreAccount,
                    style: ElevatedButton.styleFrom(backgroundColor: Colors.orangeAccent),
                    child: const Text('恢复账号'),
                  ),
                ),
              ],
            ),
          )
        else
          SizedBox(
            width: double.infinity,
            child: OutlinedButton(
              onPressed: _confirmDelete,
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.redAccent,
                side: const BorderSide(color: Colors.redAccent),
              ),
              child: const Text('注销账号'),
            ),
          ),
        const SizedBox(height: 12),
        SizedBox(
          width: double.infinity,
          child: OutlinedButton(
            onPressed: () async {
              await ref.read(authProvider.notifier).logout();
              if (mounted) context.go('/login');
            },
            child: const Text('退出登录'),
          ),
        ),
      ],
    );
  }

  Widget _buildEditForm(u) {
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        const SizedBox(height: 16),
        TextField(
          controller: _nicknameCtrl,
          decoration: const InputDecoration(
            labelText: '昵称',
            prefixIcon: Icon(Icons.face),
          ),
        ),
        const SizedBox(height: 16),
        TextField(
          controller: _bioCtrl,
          decoration: const InputDecoration(
            labelText: '个人简介',
            prefixIcon: Icon(Icons.info_outline),
          ),
          maxLines: 3,
        ),
        const SizedBox(height: 16),
        TextField(
          controller: _cityCtrl,
          decoration: const InputDecoration(
            labelText: '城市',
            prefixIcon: Icon(Icons.location_on_outlined),
          ),
        ),
        const SizedBox(height: 16),
        GestureDetector(
          onTap: () async {
            final picked = await showDatePicker(
              context: context,
              initialDate: _birthDate ?? DateTime(2000, 1, 1),
              firstDate: DateTime(1970, 1, 1),
              lastDate: DateTime.now(),
              builder: (context, child) => Theme(
                data: Theme.of(context).copyWith(
                  colorScheme: const ColorScheme.dark(primary: Color(0xFF6C5CE7), surface: Color(0xFF1A1A2E)),
                ),
                child: child!,
              ),
            );
            if (picked != null) setState(() => _birthDate = picked);
          },
          child: Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 12),
            decoration: BoxDecoration(
              border: Border(bottom: BorderSide(color: Colors.grey[700]!)),
            ),
            child: Row(
              children: [
                const Icon(Icons.cake_outlined, size: 20, color: Colors.grey),
                const SizedBox(width: 12),
                Text(
                  _birthDate != null
                      ? '${_birthDate!.year}-${_birthDate!.month.toString().padLeft(2, '0')}-${_birthDate!.day.toString().padLeft(2, '0')}'
                      : '生日',
                  style: TextStyle(color: _birthDate != null ? Colors.white : Colors.grey[500]),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 16),
        Row(
          children: [
            const Text('性别：'),
            SegmentedButton<int>(
              segments: const [
                ButtonSegment(value: 0, label: Text('保密')),
                ButtonSegment(value: 1, label: Text('男')),
                ButtonSegment(value: 2, label: Text('女')),
              ],
              selected: {_gender},
              onSelectionChanged: (s) => setState(() => _gender = s.first),
            ),
          ],
        ),
        const SizedBox(height: 32),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: _saving ? null : _save,
            child: _saving
                ? const SizedBox(height: 20, width: 20,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('保存'),
          ),
        ),
        const SizedBox(height: 8),
        SizedBox(
          width: double.infinity,
          child: OutlinedButton(
            onPressed: () => setState(() => _editing = false),
            child: const Text('取消'),
          ),
        ),
      ],
    );
  }

  Widget _infoTile(IconData icon, String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        children: [
          Icon(icon, size: 20, color: Colors.grey[500]),
          const SizedBox(width: 12),
          Text(label, style: TextStyle(color: Colors.grey[500])),
          const Spacer(),
          Text(value, style: TextStyle(color: Colors.grey[300])),
        ],
      ),
    );
  }

  String _genderLabel(int g) {
    switch (g) {
      case 1:
        return '男';
      case 2:
        return '女';
      default:
        return '保密';
    }
  }
}
