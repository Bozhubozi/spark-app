import 'user_model.dart';

class MatchModel {
  final String id;
  final String userId1;
  final String userId2;
  final double score;
  final int status;
  final UserModel? user1;
  final UserModel? user2;

  MatchModel({
    required this.id,
    required this.userId1,
    required this.userId2,
    required this.score,
    required this.status,
    this.user1,
    this.user2,
  });

  factory MatchModel.fromJson(Map<String, dynamic> json) => MatchModel(
    id: json['id'] ?? '',
    userId1: json['user_id_1'] ?? '',
    userId2: json['user_id_2'] ?? '',
    score: (json['score'] as num?)?.toDouble() ?? 0,
    status: json['status'] ?? 0,
    user1: json['user1'] != null ? UserModel.fromJson(json['user1']) : null,
    user2: json['user2'] != null ? UserModel.fromJson(json['user2']) : null,
  );

  UserModel? otherUser(String myId) {
    if (user1?.id == myId) return user2;
    return user1;
  }
}

class LikerItem {
  final String matchId;
  final UserModel user;
  final DateTime createdAt;

  LikerItem({required this.matchId, required this.user, required this.createdAt});

  factory LikerItem.fromJson(Map<String, dynamic> json) => LikerItem(
    matchId: json['match_id'] ?? '',
    user: UserModel.fromJson(json['user']),
    createdAt: DateTime.tryParse(json['created_at'] ?? '') ?? DateTime.now(),
  );
}
