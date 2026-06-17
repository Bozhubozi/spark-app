import 'user_model.dart';

class ChatRoom {
  final String id;
  final String userId1;
  final String userId2;
  final DateTime lastMessageAt;
  final String? lastContent;
  final DateTime? lastSentAt;
  final int unreadCount;
  final String otherNickname;
  final String? otherAvatarUrl;

  ChatRoom({
    required this.id,
    required this.userId1,
    required this.userId2,
    required this.lastMessageAt,
    this.lastContent,
    this.lastSentAt,
    this.unreadCount = 0,
    this.otherNickname = '',
    this.otherAvatarUrl,
  });

  factory ChatRoom.fromJson(Map<String, dynamic> json) => ChatRoom(
    id: json['id'] ?? '',
    userId1: json['user_id_1'] ?? '',
    userId2: json['user_id_2'] ?? '',
    lastMessageAt: json['last_message_at'] != null
        ? DateTime.tryParse(json['last_message_at']) ?? DateTime.now()
        : DateTime.now(),
    lastContent: json['last_content'],
    lastSentAt: json['last_sent_at'] != null
        ? DateTime.tryParse(json['last_sent_at'])
        : null,
    unreadCount: json['unread_count'] ?? 0,
    otherNickname: json['other_nickname'] ?? '',
    otherAvatarUrl: json['other_avatar_url'],
  );

  String otherUserId(String myId) => userId1 == myId ? userId2 : userId1;
}

class MessageModel {
  final String id;
  final String roomId;
  final String senderId;
  final String clientMsgId;
  final int contentType;
  final String content;
  final bool isRead;
  final DateTime sentAt;
  final UserModel? sender;

  MessageModel({
    required this.id,
    required this.roomId,
    required this.senderId,
    required this.clientMsgId,
    required this.contentType,
    required this.content,
    required this.isRead,
    required this.sentAt,
    this.sender,
  });

  factory MessageModel.fromJson(Map<String, dynamic> json) => MessageModel(
    id: json['id'] ?? '',
    roomId: json['room_id'] ?? '',
    senderId: json['sender_id'] ?? '',
    clientMsgId: json['client_msg_id'] ?? '',
    contentType: json['content_type'] ?? 1,
    content: json['content'] ?? '',
    isRead: json['is_read'] ?? false,
    sentAt: json['sent_at'] != null
        ? DateTime.tryParse(json['sent_at']) ?? DateTime.now()
        : DateTime.now(),
    sender: json['Sender'] != null ? UserModel.fromJson(json['Sender']) : null,
  );
}
