import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/providers/auth_provider.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _accountCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  final _nicknameCtrl = TextEditingController();
  bool _loading = false;
  bool _isEmail = false;

  @override
  void dispose() {
    _accountCtrl.dispose();
    _passwordCtrl.dispose();
    _nicknameCtrl.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    setState(() => _loading = true);
    try {
      await ref.read(authProvider.notifier).register(
        account: _accountCtrl.text.trim(),
        password: _passwordCtrl.text,
        nickname: _nicknameCtrl.text.trim(),
        isEmail: _isEmail,
      );
      if (mounted) context.go('/onboarding/welcome');
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Register failed: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Create Account')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            children: [
              const SizedBox(height: 32),
              TextField(
                controller: _nicknameCtrl,
                decoration: const InputDecoration(
                  hintText: 'Nickname',
                  prefixIcon: Icon(Icons.face),
                ),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: _accountCtrl,
                decoration: InputDecoration(
                  hintText: _isEmail ? 'Email' : 'Phone number',
                  prefixIcon: Icon(_isEmail ? Icons.email_outlined : Icons.phone_outlined),
                ),
                keyboardType: _isEmail ? TextInputType.emailAddress : TextInputType.phone,
              ),
              const SizedBox(height: 8),
              Align(
                alignment: Alignment.centerRight,
                child: TextButton(
                  onPressed: () => setState(() => _isEmail = !_isEmail),
                  child: Text(_isEmail ? 'Use phone instead' : 'Use email instead'),
                ),
              ),
              const SizedBox(height: 8),
              TextField(
                controller: _passwordCtrl,
                decoration: const InputDecoration(
                  hintText: 'Password (min 6 chars)',
                  prefixIcon: Icon(Icons.lock_outline),
                ),
                obscureText: true,
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _loading ? null : _register,
                  child: _loading
                      ? const SizedBox(height: 20, width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2))
                      : const Text('Sign Up'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
