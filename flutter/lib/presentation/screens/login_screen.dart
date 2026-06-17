import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/providers/auth_provider.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _accountCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _accountCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    setState(() => _loading = true);
    try {
      await ref.read(authProvider.notifier).login(
        _accountCtrl.text.trim(),
        _passwordCtrl.text,
      );
      if (mounted) context.go('/match');
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Login failed: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Spacer(flex: 2),
              const Icon(Icons.auto_awesome, size: 64, color: Color(0xFF6C5CE7)),
              const SizedBox(height: 16),
              Text('Spark', style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                    fontWeight: FontWeight.bold, color: const Color(0xFF6C5CE7))),
              const SizedBox(height: 8),
              Text('Find your spark ✨',
                  style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      color: Colors.grey)),
              const Spacer(),
              TextField(
                controller: _accountCtrl,
                decoration: const InputDecoration(
                  hintText: 'Phone or Email',
                  prefixIcon: Icon(Icons.person_outline),
                ),
                keyboardType: TextInputType.emailAddress,
              ),
              const SizedBox(height: 16),
              TextField(
                controller: _passwordCtrl,
                decoration: const InputDecoration(
                  hintText: 'Password',
                  prefixIcon: Icon(Icons.lock_outline),
                ),
                obscureText: true,
                onSubmitted: (_) => _login(),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _loading ? null : _login,
                  child: _loading
                      ? const SizedBox(height: 20, width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2))
                      : const Text('Log In'),
                ),
              ),
              const SizedBox(height: 16),
              const SizedBox(height: 16),
              Text('or', style: TextStyle(color: Colors.grey[500])),
              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: () {
                    // WeChat SDK invocation goes here.
                    // For now, key in a test code for dev.
                  },
                  icon: const Icon(Icons.wechat, color: Color(0xFF07C160)),
                  label: const Text('Continue with WeChat'),
                  style: OutlinedButton.styleFrom(
                    side: const BorderSide(color: Color(0xFF07C160)),
                    foregroundColor: const Color(0xFF07C160),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              TextButton(
                onPressed: () => context.push('/register'),
                child: const Text("Don't have an account? Sign up"),
              ),
              const Spacer(),
            ],
          ),
        ),
      ),
    );
  }
}
