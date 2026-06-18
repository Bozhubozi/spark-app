import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class OnboardingWelcomeScreen extends StatelessWidget {
  const OnboardingWelcomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            children: [
              const Spacer(flex: 2),
              const Icon(Icons.auto_awesome, size: 72, color: Color(0xFF6C5CE7)),
              const SizedBox(height: 24),
              Text('欢迎来到火花',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold, color: const Color(0xFF6C5CE7))),
              const SizedBox(height: 12),
              Text('通过共同兴趣、性格匹配和星座\n发现你的火花。',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: Colors.grey[400], fontSize: 15, height: 1.5)),
              const Spacer(),
              _featureRow(Icons.interests_outlined, '选择你的兴趣'),
              const SizedBox(height: 16),
              _featureRow(Icons.cake_outlined, '设置生日进行星座匹配'),
              const SizedBox(height: 16),
              _featureRow(Icons.person_outline, '介绍一下你自己'),
              const SizedBox(height: 16),
              _featureRow(Icons.face_outlined, '打造你的形象'),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => context.push('/onboarding/interests'),
                  child: const Text('立即开始'),
                ),
              ),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }

  Widget _featureRow(IconData icon, String text) {
    return Row(
      children: [
        Icon(icon, color: const Color(0xFF6C5CE7), size: 24),
        const SizedBox(width: 16),
        Text(text, style: const TextStyle(fontSize: 16)),
      ],
    );
  }
}
