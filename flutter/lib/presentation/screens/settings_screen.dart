import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/providers/auth_provider.dart';

final notifyNewMatchProvider = StateProvider<bool>((ref) => true);
final notifyNewMessageProvider = StateProvider<bool>((ref) => true);

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(title: const Text('设置')),
      body: ListView(
        children: [
          const SizedBox(height: 8),
          _sectionHeader('通知设置'),
          SwitchListTile(
            title: const Text('新匹配'),
            subtitle: Text('当有人也喜欢你时',
                style: TextStyle(color: Colors.grey[500], fontSize: 13)),
            value: ref.watch(notifyNewMatchProvider),
            onChanged: (v) =>
                ref.read(notifyNewMatchProvider.notifier).state = v,
            activeThumbColor: const Color(0xFF6C5CE7),
          ),
          SwitchListTile(
            title: const Text('新消息'),
            subtitle: Text('收到聊天消息时',
                style: TextStyle(color: Colors.grey[500], fontSize: 13)),
            value: ref.watch(notifyNewMessageProvider),
            onChanged: (v) =>
                ref.read(notifyNewMessageProvider.notifier).state = v,
            activeThumbColor: const Color(0xFF6C5CE7),
          ),
          const Divider(height: 32),
          _sectionHeader('隐私与法务'),
          ListTile(
            leading: const Icon(Icons.shield_outlined),
            title: const Text('Privacy Policy'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => _showPrivacyDialog(context),
          ),
          ListTile(
            leading: const Icon(Icons.description_outlined),
            title: const Text('Terms of Service'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => _showTermsDialog(context),
          ),
          ListTile(
            leading: const Icon(Icons.block_outlined),
            title: const Text('已拉黑用户'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/blocked-users'),
          ),
          ListTile(
            leading: const Icon(Icons.download_outlined),
            title: const Text('导出我的数据'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('数据导出请求已提交')),
              );
            },
          ),
          const Divider(height: 32),
          _sectionHeader('关于'),
          const ListTile(
            leading: Icon(Icons.info_outline),
            title: Text('版本'),
            trailing: Text('1.0.0', style: TextStyle(color: Colors.grey)),
          ),
          const Divider(height: 32),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: OutlinedButton(
              onPressed: () async {
                await ref.read(authProvider.notifier).logout();
                if (context.mounted) context.go('/login');
              },
              style: OutlinedButton.styleFrom(foregroundColor: Colors.redAccent),
              child: const Text('退出登录'),
            ),
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }

  Widget _sectionHeader(String text) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
      child: Text(text,
          style: const TextStyle(
              color: Color(0xFF6C5CE7),
              fontSize: 13,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.5)),
    );
  }

  void _showPrivacyDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Privacy Policy'),
        content: const SingleChildScrollView(
          child: Text(
            'We collect minimal data to provide our matching service:\n\n'
            '• Your phone/email for authentication\n'
            '• Profile info (nickname, bio, interests, birth date, gender, city)\n'
            '• Personality quiz results\n'
            '• Swipe and match data\n'
            '• Chat messages\n\n'
            'We do NOT share your personal data with third parties.\n'
            'All data is encrypted in transit and at rest.\n'
            'You can delete your account at any time.\n\n'
            'For questions, contact: privacy@spark.app',
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('关闭')),
        ],
      ),
    );
  }

  void _showTermsDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Terms of Service'),
        content: const SingleChildScrollView(
          child: Text(
            'By using Spark, you agree to:\n\n'
            '• Be at least 18 years old\n'
            '• Provide accurate information\n'
            '• Respect other users\n'
            '• Not post inappropriate content\n\n'
            'Violations may result in account suspension.\n'
            'We reserve the right to modify these terms.\n\n'
            'Last updated: June 2026',
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('关闭')),
        ],
      ),
    );
  }
}
