import 'dart:async';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/tracking/tracker.dart';
import 'package:spark_app/data/models/user_model.dart';

final authProvider = AsyncNotifierProvider<AuthNotifier, UserModel?>(AuthNotifier.new);

class AuthNotifier extends AsyncNotifier<UserModel?> {
  @override
  FutureOr<UserModel?> build() async {
    final token = await ref.read(apiClientProvider).token;
    if (token == null) return null;
    try {
      final resp = await ref.read(apiClientProvider).get('/api/v1/user/profile');
      return UserModel.fromJson(resp.data);
    } catch (_) {
      return null;
    }
  }

  Future<void> login(String account, String password) async {
    final api = ref.read(apiClientProvider);
    final resp = await api.post('/api/v1/auth/login', data: {
      'account': account,
      'password': password,
    });
    await api.setToken(resp.data['token']);
    final user = UserModel.fromJson(resp.data['user']);
    Tracker().login(user.id);
    Tracker().track('login', properties: {'method': 'phone'});
    state = AsyncData(user);
  }

  Future<void> register({
    required String account,
    required String password,
    required String nickname,
    bool isEmail = false,
  }) async {
    final api = ref.read(apiClientProvider);
    final data = <String, dynamic>{
      'password': password,
      'nickname': nickname,
    };
    if (isEmail) {
      data['email'] = account;
    } else {
      data['phone'] = account;
    }
    final resp = await api.post('/api/v1/auth/register', data: data);
    await api.setToken(resp.data['token']);
    final user = UserModel.fromJson(resp.data['user']);
    Tracker().login(user.id);
    Tracker().track('register_complete', properties: {
      'method': isEmail ? 'email' : 'phone',
    });
    state = AsyncData(user);
  }

  Future<void> wechatLogin(String code) async {
    final api = ref.read(apiClientProvider);
    final resp = await api.post('/api/v1/auth/wechat-login', data: {'code': code});
    await api.setToken(resp.data['token']);
    final user = UserModel.fromJson(resp.data['user']);
    Tracker().login(user.id);
    Tracker().track('register_complete', properties: {'method': 'wechat'});
    state = AsyncData(user);
  }

  Future<void> logout() async {
    Tracker().track('logout');
    Tracker().logout();
    await ref.read(apiClientProvider).clearToken();
    state = const AsyncData(null);
  }
}
