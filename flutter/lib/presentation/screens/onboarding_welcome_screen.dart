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
              Text('Welcome to Spark',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold, color: const Color(0xFF6C5CE7))),
              const SizedBox(height: 12),
              Text('Find your spark through shared interests,\npersonality matching, and the stars.',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: Colors.grey[400], fontSize: 15, height: 1.5)),
              const Spacer(),
              _featureRow(Icons.interests_outlined, 'Pick your interests'),
              const SizedBox(height: 16),
              _featureRow(Icons.cake_outlined, 'Set your birthday for zodiac matching'),
              const SizedBox(height: 16),
              _featureRow(Icons.person_outline, 'Tell us about yourself'),
              const SizedBox(height: 16),
              _featureRow(Icons.face_outlined, 'Build your avatar'),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => context.push('/onboarding/interests'),
                  child: const Text('Get Started'),
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
