import 'package:flutter_test/flutter_test.dart';
import 'package:spark_app/data/models/user_model.dart';
import 'package:spark_app/data/models/message_model.dart';
import 'package:spark_app/data/models/match_model.dart';
import 'package:spark_app/core/constants/zodiac_helper.dart';

void main() {
  group('UserModel', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 'abc-123',
        'nickname': 'TestUser',
        'gender': 1,
        'bio': 'Hello world',
        'city': '深圳',
        'birth_date': '1998-06-15T00:00:00Z',
      };
      final user = UserModel.fromJson(json);
      expect(user.id, 'abc-123');
      expect(user.nickname, 'TestUser');
      expect(user.gender, 1);
      expect(user.bio, 'Hello world');
      expect(user.city, '深圳');
    });

    test('fromJson handles missing fields', () {
      final user = UserModel.fromJson({'id': 'x', 'nickname': 'u'});
      expect(user.id, 'x');
      expect(user.gender, 0);
      expect(user.interests, isEmpty);
      expect(user.bio, isNull);
    });
  });

  group('MessageModel', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 'msg-1',
        'room_id': 'room-1',
        'sender_id': 'user-1',
        'client_msg_id': 'client-1',
        'content_type': 1,
        'content': 'Hello!',
        'is_read': false,
        'sent_at': '2026-06-17T12:00:00Z',
      };
      final msg = MessageModel.fromJson(json);
      expect(msg.id, 'msg-1');
      expect(msg.content, 'Hello!');
      expect(msg.isRead, false);
      expect(msg.contentType, 1);
    });

    test('default content type is text', () {
      final msg = MessageModel.fromJson({
        'id': 'x',
        'room_id': 'r',
        'sender_id': 's',
        'client_msg_id': 'c',
        'content': 'hi',
        'sent_at': '2026-06-17T12:00:00Z',
      });
      expect(msg.contentType, 1);
    });
  });

  group('ChatRoom', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 'room-1',
        'user_id_1': 'alice',
        'user_id_2': 'bob',
        'last_message_at': '2026-06-17T12:00:00Z',
        'unread_count': 3,
        'other_nickname': 'Bob',
      };
      final room = ChatRoom.fromJson(json);
      expect(room.id, 'room-1');
      expect(room.unreadCount, 3);
      expect(room.otherNickname, 'Bob');
    });

    test('otherUserId returns correct peer', () {
      final room = ChatRoom(
        id: 'r1',
        userId1: 'alice',
        userId2: 'bob',
        lastMessageAt: DateTime.now(),
      );
      expect(room.otherUserId('alice'), 'bob');
      expect(room.otherUserId('bob'), 'alice');
    });
  });

  group('MatchModel', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 'match-1',
        'user_id_1': 'alice',
        'user_id_2': 'bob',
        'score': 0.85,
        'status': 1,
      };
      final m = MatchModel.fromJson(json);
      expect(m.id, 'match-1');
      expect(m.score, 0.85);
      expect(m.status, 1);
    });
  });

  group('ZodiacHelper', () {
    test('zodiacFromBirth returns emoji for valid date', () {
      // Aug 15 = Leo ♌
      expect(zodiacFromBirth('1998-08-15'), '♌');
      // Jan 25 = Aquarius ♒
      expect(zodiacFromBirth('1998-01-25'), '♒');
      // Mar 21 = Aries ♈
      expect(zodiacFromBirth('1998-03-21'), '♈');
    });

    test('zodiacFromBirth returns empty for null', () {
      expect(zodiacFromBirth(null), '');
    });

    test('zodiacEmojiToName converts all 12 signs', () {
      expect(zodiacEmojiToName('♒'), '水瓶座');
      expect(zodiacEmojiToName('♌'), '狮子座');
      expect(zodiacEmojiToName('♏'), '天蝎座');
      expect(zodiacEmojiToName('♑'), '摩羯座');
    });

    test('zodiacTraits has all 12 signs', () {
      expect(zodiacTraits.length, 12);
    });
  });
}
