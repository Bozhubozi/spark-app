import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/tracking/tracker.dart';

class RegisterInterceptScreen extends StatelessWidget {
  const RegisterInterceptScreen({super.key});

  @override
  Widget build(BuildContext context) {
    Tracker().track('register_intercept_show');
    return Scaffold(
      body: Stack(
        fit: StackFit.expand,
        children: [
          // Preview card mockups in background
          ..._buildPreviewCards(context),
          // Gradient overlay
          Positioned.fill(
            child: Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    Colors.transparent,
                    Color(0xBB1A1A2E),
                    Color(0xFF16213E),
                  ],
                ),
              ),
            ),
          ),
          // CTA content
          SafeArea(
            child: Column(
              children: [
                const Spacer(),
                const SizedBox(height: 24),
                const Icon(Icons.auto_awesome, size: 48, color: Color(0xFF6C5CE7)),
                const SizedBox(height: 16),
                Text(
                  '发现你的火花',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Meet people who share your vibe.\nNo swipe fatigue — just real connections.',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: Colors.grey[400], height: 1.5),
                ),
                const SizedBox(height: 32),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 32),
                  child: SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () => context.push('/register'),
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      child: const Text('立即开始', style: TextStyle(fontSize: 16)),
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text('已有账号？',
                        style: TextStyle(color: Colors.grey[500])),
                    TextButton(
                      onPressed: () => context.push('/login'),
                      child: const Text('登录'),
                    ),
                  ],
                ),
                const SizedBox(height: 48),
                Text(
                  '注册即表示同意我们的',
                  style: TextStyle(color: Colors.grey[600], fontSize: 12),
                ),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    TextButton(
                      onPressed: () => _showLegal(context, '隐私政策',
                          'We collect minimal data to provide our matching service.\n\n'
                          'Information we collect:\n'
                          '- Phone number or email for account creation\n'
                          '- Nickname and birth date for profile\n'
                          '- Interests and personality quiz results for matching\n'
                          '- Chat messages for communication\n'
                          '- Device token for push notifications\n\n'
                          'We do NOT sell your data to third parties.\n'
                          'Your data is stored securely and encrypted in transit.\n\n'
                          'You can delete your account at any time from Profile > Delete Account.\n'
                          'Account deletion is permanent after a 7-day grace period.'),
                      child: const Text('隐私政策', style: TextStyle(fontSize: 12)),
                    ),
                    Text(' 和 ', style: TextStyle(color: Colors.grey[600], fontSize: 12)),
                    TextButton(
                      onPressed: () => _showLegal(context, '用户协议',
                          'Welcome to Spark!\n\n'
                          'By using Spark, you agree to:\n'
                          '- Be at least 18 years old\n'
                          '- Provide accurate information\n'
                          '- Respect other users\n'
                          '- Not harass, spam, or impersonate others\n'
                          '- Not share inappropriate content\n\n'
                          'We reserve the right to suspend accounts that violate these terms.\n'
                          'Use Spark responsibly and enjoy meeting new people!'),
                      child: const Text('用户协议', style: TextStyle(fontSize: 12)),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  void _showLegal(BuildContext context, String title, String content) {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: Text(title),
        content: SingleChildScrollView(child: Text(content, style: const TextStyle(height: 1.6))),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('关闭')),
        ],
      ),
    );
  }

  List<Widget> _buildPreviewCards(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    return [
      Positioned(
        top: 60,
        left: screenWidth * 0.05,
        child: Transform.rotate(
          angle: -0.08,
          child: _previewCard(0xFF6C5CE7, 'ACG | K-pop | 咖啡'),
        ),
      ),
      Positioned(
        top: 120,
        right: screenWidth * 0.05,
        child: Transform.rotate(
          angle: 0.05,
          child: _previewCard(0xFFE17055, '露营 | 摄影 | 旅行'),
        ),
      ),
      Positioned(
        top: 200,
        left: screenWidth * 0.15,
        child: Transform.rotate(
          angle: -0.03,
          child: _previewCard(0xFF00B894, '健身 | 篮球 | 电竞'),
        ),
      ),
    ];
  }

  Widget _previewCard(int color, String tags) {
    return Container(
      width: 280,
      height: 180,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(20),
        color: Color(color).withValues(alpha: 0.3),
        border: Border.all(color: Color(color).withValues(alpha: 0.5)),
      ),
      child: Center(
        child: Text(
          tags,
          style: TextStyle(
            color: Color(color).withValues(alpha: 0.8),
            fontSize: 16,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
    );
  }
}
