import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:spark_app/core/constants/app_constants.dart';
import 'package:spark_app/data/models/message_model.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/data/providers/auth_provider.dart';
import 'package:spark_app/data/providers/chat_provider.dart';
import 'package:spark_app/presentation/widgets/message_bubble.dart';
import 'package:uuid/uuid.dart';

class ChatScreen extends ConsumerStatefulWidget {
  final String roomId;
  final String? otherName;
  final String? otherId;
  const ChatScreen({super.key, required this.roomId, this.otherName, this.otherId});

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final _msgCtrl = TextEditingController();
  final _scrollCtrl = ScrollController();
  final _uuid = const Uuid();
  WebSocketChannel? _channel;
  final List<MessageModel> _messages = [];
  bool _connected = false;
  bool _reconnecting = false;
  int _retryCount = 0;
  Timer? _typingTimer;
  bool _partnerTyping = false;
  Timer? _typingTimeout;

  @override
  void initState() {
    super.initState();
    _connectWS();
    _loadHistory();
  }

  @override
  void dispose() {
    _msgCtrl.dispose();
    _scrollCtrl.dispose();
    _typingTimer?.cancel();
    _typingTimeout?.cancel();
    _reconnecting = false;
    _retryCount = 999;
    _channel?.sink.close();
    super.dispose();
  }

  Future<void> _connectWS() async {
    final token = await ref.read(apiClientProvider).token;
    if (token == null) return;
    final uri = Uri.parse('${AppConstants.wsUrl}?token=$token');
    try {
      _channel = WebSocketChannel.connect(uri);
      _channel!.stream.listen(
        _onWSMessage,
        onError: _onWSError,
        onDone: _onWSDone,
      );
      setState(() {
        _connected = true;
        _reconnecting = false;
        _retryCount = 0;
      });
    } catch (_) {
      _scheduleReconnect();
    }
  }

  void _onWSMessage(data) {
    final msg = jsonDecode(data);
    final type = msg['type'] as String?;
    final body = msg['data'];
    if (body == null) return;

    switch (type) {
      case 'chat.message.new':
        setState(() {
          _messages.insert(0, MessageModel(
            id: body['server_msg_id'] ?? '',
            roomId: body['room_id'] ?? '',
            senderId: body['sender_id'] ?? '',
            clientMsgId: body['client_msg_id'] ?? '',
            contentType: body['content_type'] ?? 1,
            content: body['content'] ?? '',
            isRead: false,
            sentAt: DateTime.tryParse(body['sent_at'] ?? '') ?? DateTime.now(),
          ));
        });
        ref.invalidate(chatRoomsProvider);
        break;
      case 'chat.message.ack':
        final ackClientId = body['client_msg_id'] ?? '';
        final serverId = body['server_msg_id'] ?? '';
        setState(() {
          for (final m in _messages) {
            if (m.clientMsgId == ackClientId) {
              if (serverId.isNotEmpty && m.id == m.clientMsgId) {
                final idx = _messages.indexOf(m);
                _messages[idx] = MessageModel(
                  id: serverId,
                  roomId: m.roomId,
                  senderId: m.senderId,
                  clientMsgId: m.clientMsgId,
                  contentType: m.contentType,
                  content: m.content,
                  isRead: m.isRead,
                  sentAt: m.sentAt,
                );
              }
              break;
            }
          }
        });
        break;
      case 'chat.message.read':
        final readerId = body['reader_id'] ?? '';
        final myId = ref.read(authProvider).valueOrNull?.id ?? '';
        if (readerId != myId) {
          setState(() {
            for (int i = 0; i < _messages.length; i++) {
              if (_messages[i].senderId == myId && !_messages[i].isRead) {
                _messages[i] = MessageModel(
                  id: _messages[i].id,
                  roomId: _messages[i].roomId,
                  senderId: _messages[i].senderId,
                  clientMsgId: _messages[i].clientMsgId,
                  contentType: _messages[i].contentType,
                  content: _messages[i].content,
                  isRead: true,
                  sentAt: _messages[i].sentAt,
                );
              }
            }
          });
        }
        break;
      case 'chat.typing':
        final isTyping = body['is_typing'] == true;
        final typerId = body['user_id'] ?? '';
        if (typerId != ref.read(authProvider).valueOrNull?.id) {
          _typingTimeout?.cancel();
          setState(() => _partnerTyping = isTyping);
          if (isTyping) {
            _typingTimeout = Timer(const Duration(seconds: 4), () {
              if (mounted) setState(() => _partnerTyping = false);
            });
          }
        }
        break;
      case 'system.heartbeat':
        _channel?.sink.add(jsonEncode({'type': 'system.heartbeat'}));
        break;
    }
  }

  void _onWSError(e) {
    debugPrint('WS error: $e');
    setState(() => _connected = false);
    _scheduleReconnect();
  }

  void _onWSDone() {
    setState(() => _connected = false);
    _scheduleReconnect();
  }

  void _scheduleReconnect() {
    if (_reconnecting || _retryCount >= 5) return;
    _reconnecting = true;
    if (mounted) setState(() {});
    final delay = Duration(seconds: [1, 2, 4, 8, 16][_retryCount]);
    _retryCount++;
    Future.delayed(delay, () {
      if (mounted && _reconnecting) _connectWS();
    });
  }

  Future<void> _loadHistory() async {
    final api = ref.read(apiClientProvider);
    try {
      final resp = await api.get('/api/v1/chat/rooms/${widget.roomId}/messages');
      final list = (resp.data as List<dynamic>?)
          ?.map((e) => MessageModel.fromJson(e))
          .toList() ?? [];
      setState(() => _messages.addAll(list.reversed));
      _markRead();
    } catch (_) {}
  }

  Future<void> _markRead() async {
    try {
      await ref.read(apiClientProvider).post('/api/v1/chat/rooms/${widget.roomId}/read');
    } catch (_) {}
  }

  Future<void> _blockUser() async {
    if (widget.otherId == null) return;
    try {
      await ref.read(apiClientProvider).post('/api/v1/match/swipe', data: {
        'target_user_id': widget.otherId,
        'direction': 'pass',
      });
      ref.invalidate(chatRoomsProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('User blocked')),
        );
      }
    } catch (_) {}
  }

  Future<void> _reportUser() async {
    if (widget.otherId == null) return;
    try {
      await ref.read(apiClientProvider).post('/api/v1/user/report', data: {
        'target_user_id': widget.otherId,
        'reason': 'reported from chat',
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Report submitted')),
        );
      }
    } catch (_) {}
  }

  void _onTyping(String text) {
    if (!_connected) return;
    _typingTimer?.cancel();
    _typingTimer = Timer(const Duration(milliseconds: 500), () {
      if (_connected && _channel != null) {
        _channel!.sink.add(jsonEncode({
          'type': 'chat.typing',
          'data': {
            'room_id': widget.roomId,
            'is_typing': text.isNotEmpty,
          },
        }));
      }
      _typingTimer = null;
    });
  }

  void _sendMessage() {
    final text = _msgCtrl.text.trim();
    if (text.isEmpty || !_connected) return;

    final myId = ref.read(authProvider).valueOrNull?.id ?? '';
    final clientMsgId = _uuid.v4();

    final msg = {
      'type': 'chat.message.send',
      'data': {
        'room_id': widget.roomId,
        'sender_id': myId,
        'client_msg_id': clientMsgId,
        'content_type': 1,
        'content': text,
      },
    };

    _channel!.sink.add(jsonEncode(msg));
    _typingTimer?.cancel();
    // Send typing stop
    _channel!.sink.add(jsonEncode({
      'type': 'chat.typing',
      'data': {'room_id': widget.roomId, 'is_typing': false},
    }));
    _msgCtrl.clear();
    ref.invalidate(chatRoomsProvider);

    setState(() {
      _messages.insert(0, MessageModel(
        id: clientMsgId,
        roomId: widget.roomId,
        senderId: myId,
        clientMsgId: clientMsgId,
        contentType: 1,
        content: text,
        isRead: false,
        sentAt: DateTime.now(),
      ));
    });
  }

  @override
  Widget build(BuildContext context) {
    final myId = ref.watch(authProvider).valueOrNull?.id ?? '';

    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(widget.otherName ?? 'Chat', style: const TextStyle(fontSize: 16)),
            Text(
              _partnerTyping ? 'typing...' : _connected ? 'online' : 'connecting...',
              style: TextStyle(
                color: _partnerTyping ? const Color(0xFF6C5CE7) : _connected ? const Color(0xFF00B894) : Colors.grey,
                fontSize: 12,
              ),
            ),
          ],
        ),
        actions: [
          if (widget.otherId != null)
            PopupMenuButton<String>(
              onSelected: (action) {
                if (action == 'block') {
                  _blockUser();
                } else if (action == 'report') {
                  _reportUser();
                }
              },
              itemBuilder: (_) => const [
                PopupMenuItem(value: 'block', child: Text('Block User')),
                PopupMenuItem(value: 'report', child: Text('Report User')),
              ],
            ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: _messages.isEmpty
                ? Center(
                    child: Column(mainAxisSize: MainAxisSize.min, children: [
                      const Icon(Icons.chat_bubble_outline, size: 56, color: Colors.white24),
                      const SizedBox(height: 16),
                      Text('No messages yet',
                          style: TextStyle(color: Colors.grey[600], fontSize: 15)),
                      const SizedBox(height: 8),
                      Text('Say hello to break the ice',
                          style: TextStyle(color: Colors.grey[700], fontSize: 13)),
                    ]),
                  )
                : ListView.builder(
                    controller: _scrollCtrl,
                    reverse: true,
                    itemCount: _messages.length,
                    itemBuilder: (_, i) => MessageBubble(
                      message: _messages[i],
                      isMine: _messages[i].senderId == myId,
                    ),
                  ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surface,
              border: Border(top: BorderSide(color: Colors.grey[800]!)),
            ),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _msgCtrl,
                    onChanged: _onTyping,
                    decoration: const InputDecoration(
                      hintText: 'Type a message...',
                      border: InputBorder.none,
                      contentPadding: EdgeInsets.symmetric(horizontal: 12),
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.send, color: Color(0xFF6C5CE7)),
                  onPressed: _sendMessage,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
