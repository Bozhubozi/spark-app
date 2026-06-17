import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/data/models/user_model.dart';
import 'package:spark_app/data/models/match_model.dart';

final candidatesProvider = FutureProvider<List<UserModel>>((ref) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.get('/api/v1/match/candidates');
  final list = resp.data as List<dynamic>?;
  if (list == null) return [];
  return list.map((e) => UserModel.fromJson(e)).toList();
});

final likesCountProvider = FutureProvider<int>((ref) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.get('/api/v1/match/likes-count');
  return resp.data['count'] ?? 0;
});

final likersProvider = FutureProvider<List<LikerItem>>((ref) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.get('/api/v1/match/likers');
  final list = resp.data as List<dynamic>?;
  if (list == null) return [];
  return list.map((e) => LikerItem.fromJson(e)).toList();
});

final remainingSwipesProvider = FutureProvider<int>((ref) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.get('/api/v1/match/remaining');
  return resp.data['remaining'] ?? 0;
});

final blockedUsersProvider = FutureProvider<List<LikerItem>>((ref) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.get('/api/v1/match/blocked');
  final list = resp.data as List<dynamic>?;
  if (list == null) return [];
  return list.map((e) => LikerItem.fromJson(e)).toList();
});

final matchesProvider = FutureProvider<List<MatchModel>>((ref) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.get('/api/v1/match/list');
  final list = resp.data as List<dynamic>?;
  if (list == null) return [];
  return list.map((e) => MatchModel.fromJson(e)).toList();
});
