import 'package:flutter/material.dart';
import 'package:spark_app/data/models/message_model.dart';

class MessageBubble extends StatelessWidget {
  final MessageModel message;
  final bool isMine;

  const MessageBubble({super.key, required this.message, required this.isMine});

  @override
  Widget build(BuildContext context) {
    return Align(
      alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.7),
        decoration: BoxDecoration(
          color: isMine ? const Color(0xFF6C5CE7) : const Color(0xFF1A1A2E),
          borderRadius: BorderRadius.only(
            topLeft: const Radius.circular(16),
            topRight: const Radius.circular(16),
            bottomLeft: Radius.circular(isMine ? 16 : 4),
            bottomRight: Radius.circular(isMine ? 4 : 16),
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(message.content,
                style: const TextStyle(color: Colors.white, fontSize: 15)),
            const SizedBox(height: 4),
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  _formatTime(message.sentAt),
                  style: const TextStyle(color: Colors.white54, fontSize: 11),
                ),
                if (isMine) ...[
                  const SizedBox(width: 4),
                  if (message.isRead)
                    const Text('已读', style: TextStyle(color: Colors.white54, fontSize: 11))
                  else
                    Icon(
                      message.id != message.clientMsgId ? Icons.check : Icons.access_time,
                      size: 12,
                      color: Colors.white54,
                    ),
                ],
              ],
            ),
          ],
        ),
      ),
    );
  }

  String _formatTime(DateTime dt) =>
      '${dt.hour}:${dt.minute.toString().padLeft(2, '0')}';
}
