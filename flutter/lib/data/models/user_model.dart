class UserModel {
  final String id;
  final String? phone;
  final String? email;
  final String nickname;
  final String? avatarUrl;
  final int gender;
  final String? birthDate;
  final String? bio;
  final String? city;
  final DateTime? lastActiveAt;
  final List<InterestTag> interests;
  final List<PersonalityDimension>? personality;
  final DateTime? deletedAt;

  UserModel({
    required this.id,
    this.phone,
    this.email,
    required this.nickname,
    this.avatarUrl,
    this.gender = 0,
    this.birthDate,
    this.bio,
    this.city,
    this.lastActiveAt,
    this.interests = const [],
    this.personality,
    this.deletedAt,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) => UserModel(
    id: json['id'] ?? '',
    phone: json['phone'],
    email: json['email'],
    nickname: json['nickname'] ?? '',
    avatarUrl: json['avatar_url'],
    gender: json['gender'] ?? 0,
    birthDate: json['birth_date'],
    bio: json['bio'],
    city: json['city'],
    lastActiveAt: json['last_active_at'] != null
        ? DateTime.tryParse(json['last_active_at'])
        : null,
    interests: (json['interests'] as List<dynamic>?)
            ?.map((e) => InterestTag.fromJson(e))
            .toList() ??
        [],
    personality: (json['personality'] as List<dynamic>?)
            ?.map((e) => PersonalityDimension.fromJson(e))
            .toList(),
    deletedAt: json['deleted_at'] != null
        ? DateTime.tryParse(json['deleted_at'])
        : null,
  );
}

class InterestTag {
  final int id;
  final String name;
  final String category;
  final String? icon;

  InterestTag({required this.id, required this.name, required this.category, this.icon});

  factory InterestTag.fromJson(Map<String, dynamic> json) => InterestTag(
    id: json['id'] ?? 0,
    name: json['name'] ?? '',
    category: json['category'] ?? '',
    icon: json['icon'],
  );
}

class PersonalityDimension {
  final String dimension;
  final double score;

  PersonalityDimension({required this.dimension, required this.score});

  factory PersonalityDimension.fromJson(Map<String, dynamic> json) =>
      PersonalityDimension(
        dimension: json['dimension'] ?? '',
        score: (json['score'] as num?)?.toDouble() ?? 0,
      );
}

class PersonalityQuestion {
  final int id;
  final String dimension;
  final String questionText;
  final List<PersonalityOption> options;

  PersonalityQuestion({
    required this.id,
    required this.dimension,
    required this.questionText,
    required this.options,
  });

  factory PersonalityQuestion.fromJson(Map<String, dynamic> json) =>
      PersonalityQuestion(
        id: json['id'] ?? 0,
        dimension: json['dimension'] ?? '',
        questionText: json['question_text'] ?? '',
        options: (json['options'] as List<dynamic>?)
                ?.map((e) => PersonalityOption.fromJson(e))
                .toList() ??
            [],
      );
}

class PersonalityOption {
  final int id;
  final int questionId;
  final String optionText;
  final int score;

  PersonalityOption({
    required this.id,
    required this.questionId,
    required this.optionText,
    required this.score,
  });

  factory PersonalityOption.fromJson(Map<String, dynamic> json) =>
      PersonalityOption(
        id: json['id'] ?? 0,
        questionId: json['question_id'] ?? 0,
        optionText: json['option_text'] ?? '',
        score: json['score'] ?? 0,
      );
}

class AvatarComponent {
  final int id;
  final String category;
  final String name;
  final String imageUrl;
  final int rarity;

  AvatarComponent({
    required this.id,
    required this.category,
    required this.name,
    required this.imageUrl,
    required this.rarity,
  });

  factory AvatarComponent.fromJson(Map<String, dynamic> json) =>
      AvatarComponent(
        id: json['id'] ?? 0,
        category: json['category'] ?? '',
        name: json['name'] ?? '',
        imageUrl: json['image_url'] ?? '',
        rarity: json['rarity'] ?? 1,
      );
}
