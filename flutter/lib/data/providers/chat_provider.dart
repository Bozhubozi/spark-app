import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/data/models/message_model.dart';

final chatRoomsProvider = FutureProvider<List<ChatRoom>>((ref) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.get('/api/v1/chat/rooms');
  final list = resp.data as List<dynamic>?;
  if (list == null) return [];
  return list.map((e) => ChatRoom.fromJson(e)).toList();
});

final chatRoomProvider = FutureProvider.family<ChatRoom, String>((ref, targetUserId) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.post('/api/v1/chat/rooms', data: {'target_user_id': targetUserId});
  return ChatRoom.fromJson(resp.data);
});

final chatMessagesProvider = FutureProvider.family<List<MessageModel>, String>((ref, roomId) async {
  final api = ref.read(apiClientProvider);
  final resp = await api.get('/api/v1/chat/rooms/$roomId/messages');
  final list = resp.data as List<dynamic>?;
  if (list == null) return [];
  return list.map((e) => MessageModel.fromJson(e)).toList();
});
