import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/data/models/user_model.dart';
import 'package:spark_app/core/tracking/tracker.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/constants/zodiac_helper.dart';
import 'package:spark_app/data/providers/notification_provider.dart';
import 'package:spark_app/presentation/screens/user_detail_screen.dart';
import 'package:uuid/uuid.dart';

class MatchCard extends ConsumerStatefulWidget {
  final UserModel user;
  final VoidCallback? onSwiped;
  const MatchCard({super.key, required this.user, this.onSwiped});

  @override
  ConsumerState<MatchCard> createState() => _MatchCardState();
}

class _MatchCardState extends ConsumerState<MatchCard> with SingleTickerProviderStateMixin {
  Offset _dragOffset = Offset.zero;
  late AnimationController _springCtrl;
  late Animation<Offset> _springAnim;
  Offset _flyTarget = Offset.zero;

  @override
  void initState() {
    super.initState();
    _springCtrl = AnimationController(vsync: this, duration: const Duration(milliseconds: 300));
    _springAnim = Tween<Offset>(begin: Offset.zero, end: Offset.zero).animate(
      CurvedAnimation(parent: _springCtrl, curve: Curves.easeOutBack),
    );
    _springCtrl.addListener(() {
      if (!_springCtrl.isAnimating) return;
      setState(() => _dragOffset = _springAnim.value);
    });
  }

  @override
  void dispose() {
    _springCtrl.dispose();
    super.dispose();
  }

  void _onPanUpdate(DragUpdateDetails details) {
    setState(() {
      _dragOffset += details.delta;
    });
  }

  void _onPanEnd(DragEndDetails details) {
    final dx = _dragOffset.dx;
    final vx = details.velocity.pixelsPerSecond.dx;
    const threshold = 100.0;

    if (dx > threshold || vx > 800) {
      _flyTarget = Offset(MediaQuery.of(context).size.width * 1.5, _dragOffset.dy);
      _animateFlyAway().then((_) => _swipe('like'));
    } else if (dx < -threshold || vx < -800) {
      _flyTarget = Offset(-MediaQuery.of(context).size.width * 1.5, _dragOffset.dy);
      _animateFlyAway().then((_) => _swipe('pass'));
    } else {
      _springBack();
    }
  }

  Future<void> _animateFlyAway() async {
    _springAnim = Tween<Offset>(begin: _dragOffset, end: _flyTarget).animate(
      CurvedAnimation(parent: _springCtrl, curve: Curves.easeIn),
    );
    _springCtrl.reset();
    _springCtrl.forward();
    await Future.delayed(const Duration(milliseconds: 300));
  }

  void _springBack() {
    _springAnim = Tween<Offset>(begin: _dragOffset, end: Offset.zero).animate(
      CurvedAnimation(parent: _springCtrl, curve: Curves.elasticOut),
    );
    _springCtrl.reset();
    _springCtrl.forward();
  }

  void _reportUser() {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF1A1A2E),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Report User', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              Text('Blocking will prevent this user from appearing again.',
                  style: TextStyle(color: Colors.grey[500], fontSize: 13)),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton(
                  onPressed: () {
                    Navigator.of(context).pop();
                    _swipe('pass');
                  },
                  style: OutlinedButton.styleFrom(foregroundColor: Colors.orangeAccent),
                  child: const Text('Block User'),
                ),
              ),
              const SizedBox(height: 8),
              SizedBox(
                width: double.infinity,
                child: TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Cancel', style: TextStyle(color: Colors.grey)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _swipe(String direction) async {
    Tracker().track(
      direction == 'like' ? 'card_swipe_right' : 'card_swipe_left',
      properties: {'target_user_id': widget.user.id},
    );
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.post('/api/v1/match/swipe', data: {
        'target_user_id': widget.user.id,
        'direction': direction,
      });
      final matched = resp.data['matched'] == true;
      if (matched && mounted) {
        ref.read(notificationProvider.notifier).add(AppNotification(
          id: const Uuid().v4(),
          title: 'New Match!',
          body: 'You and ${widget.user.nickname} liked each other',
        ));
        final icebreakers = (resp.data['icebreakers'] as List<dynamic>?)
                ?.map((e) => e.toString())
                .toList() ??
            [];
        final matchId = resp.data['match_id']?.toString() ?? '';
        _showMatchDialog(matchId, icebreakers);
      }
      widget.onSwiped?.call();
    } catch (_) {}
  }

  void _showMatchDialog(String matchId, List<String> icebreakers) {
    showDialog(
      context: context,
      builder: (_) => _MatchCelebrationDialog(
        icebreakers: icebreakers,
        onChat: () async {
          final router = GoRouter.of(context);
          Navigator.of(context).pop();
          try {
            final api = ref.read(apiClientProvider);
            final resp = await api.post('/api/v1/chat/rooms',
                data: {'target_user_id': widget.user.id});
            final roomId = resp.data['id'] ?? matchId;
            router.push('/chat/$roomId', extra: {
              'otherName': widget.user.nickname,
              'otherId': widget.user.id,
            });
          } catch (_) {}
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final dx = _dragOffset.dx;
    final rotation = (dx / 400.0).clamp(-0.4, 0.4);
    final likeOpacity = (dx / 80.0).clamp(0.0, 1.0);
    final nopeOpacity = (-dx / 80.0).clamp(0.0, 1.0);

    return GestureDetector(
      onTap: _dragOffset == Offset.zero ? () async {
        final swiped = await Navigator.of(context).push<bool>(
          MaterialPageRoute(
            builder: (_) => UserDetailScreen(user: widget.user),
          ),
        );
        if (swiped == true) widget.onSwiped?.call();
      } : null,
      onPanUpdate: _onPanUpdate,
      onPanEnd: _onPanEnd,
      child: Transform.translate(
        offset: _dragOffset,
        child: Transform.rotate(
          angle: rotation,
          child: Stack(
            children: [
              Card(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        child: Stack(
          fit: StackFit.expand,
          children: [
            // Avatar placeholder
            Container(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(24),
                gradient: const LinearGradient(
                  colors: [Color(0xFF6C5CE7), Color(0xFF00CEC9)],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
              ),
              child: Center(
                child: CircleAvatar(
                  radius: 60,
                  backgroundColor: Colors.white24,
                  child: Text(
                    widget.user.nickname[0].toUpperCase(),
                    style: const TextStyle(fontSize: 40, color: Colors.white),
                  ),
                ),
              ),
            ),
            // Info overlay
            Positioned(
              bottom: 0, left: 0, right: 0,
              child: Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  borderRadius: const BorderRadius.vertical(bottom: Radius.circular(24)),
                  gradient: LinearGradient(
                    colors: [Colors.black.withValues(alpha: 0.8), Colors.transparent],
                    begin: Alignment.bottomCenter,
                    end: Alignment.topCenter,
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Row(
                      children: [
                        Text('${widget.user.nickname}, ${_age(widget.user.birthDate)}',
                            style: const TextStyle(
                                color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold)),
                        const SizedBox(width: 8),
                        Text(zodiacFromBirth(widget.user.birthDate),
                            style: const TextStyle(fontSize: 22)),
                      ],
                    ),
                    if (widget.user.bio != null)
                      Padding(
                        padding: const EdgeInsets.only(top: 4),
                        child: Text(widget.user.bio!, style: const TextStyle(color: Colors.white70)),
                      ),
                    const SizedBox(height: 8),
                    if (widget.user.interests.isNotEmpty)
                      Wrap(
                        spacing: 6, runSpacing: 4,
                        children: widget.user.interests.take(5).map((tag) =>
                            Chip(
                              label: Text('${tag.icon ?? ''} ${tag.name}',
                                  style: const TextStyle(color: Colors.white, fontSize: 12)),
                              backgroundColor: Colors.white24,
                              padding: EdgeInsets.zero,
                              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                            ),
                        ).toList(),
                      ),
                    if (widget.user.lastActiveAt != null) ...[
                      const SizedBox(height: 6),
                      Text(_lastActive(widget.user.lastActiveAt!),
                          style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                    ],
                  ],
                ),
              ),
            ),
            // Action buttons
            Positioned(
              bottom: 20, right: 20,
              child: Column(
                children: [
                  _ActionButton(
                    icon: Icons.close, color: Colors.redAccent,
                    onTap: () => _swipe('pass'),
                  ),
                  const SizedBox(height: 12),
                  _ActionButton(
                    icon: Icons.favorite, color: const Color(0xFFFD79A8),
                    onTap: () => _swipe('like'),
                  ),
                  const SizedBox(height: 12),
                  _ActionButton(
                    icon: Icons.flag_outlined, color: Colors.grey[700]!,
                    onTap: _reportUser,
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
      if (likeOpacity > 0)
        Positioned(
          top: 40,
          left: 30,
          child: Transform.rotate(
            angle: -0.2,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              decoration: BoxDecoration(
                border: Border.all(color: const Color(0xFF00B894), width: 3),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Text('LIKE',
                  style: TextStyle(
                      color: Color(0xFF00B894),
                      fontSize: 28,
                      fontWeight: FontWeight.bold)),
            ),
          ),
        ),
      if (nopeOpacity > 0)
        Positioned(
          top: 40,
          right: 30,
          child: Transform.rotate(
            angle: 0.2,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              decoration: BoxDecoration(
                border: Border.all(color: Colors.redAccent, width: 3),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Text('NOPE',
                  style: TextStyle(
                      color: Colors.redAccent,
                      fontSize: 28,
                      fontWeight: FontWeight.bold)),
            ),
          ),
        ),
    ],
  ),
  ),
  ),
);
  }

  String _age(String? birthDate) {
    if (birthDate == null) return '';
    try {
      final bd = DateTime.tryParse(birthDate);
      if (bd == null) return '';
      final now = DateTime.now();
      var age = now.year - bd.year;
      if (now.month < bd.month || (now.month == bd.month && now.day < bd.day)) age--;
      return '$age';
    } catch (_) {
      return '';
    }
  }

  String _lastActive(DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inMinutes < 1) return 'Active just now';
    if (diff.inMinutes < 60) return 'Active ${diff.inMinutes}m ago';
    if (diff.inHours < 24) return 'Active ${diff.inHours}h ago';
    if (diff.inDays < 7) return 'Active ${diff.inDays}d ago';
    return 'Active ${diff.inDays ~/ 7}w ago';
  }
}

class _MatchCelebrationDialog extends StatefulWidget {
  final List<String> icebreakers;
  final VoidCallback onChat;
  const _MatchCelebrationDialog({required this.icebreakers, required this.onChat});

  @override
  State<_MatchCelebrationDialog> createState() => _MatchCelebrationDialogState();
}

class _MatchCelebrationDialogState extends State<_MatchCelebrationDialog>
    with SingleTickerProviderStateMixin {
  late AnimationController _ctrl;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(vsync: this, duration: const Duration(milliseconds: 1500));
    _ctrl.repeat();
  }

  @override
  void dispose() {
    _ctrl.stop();
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      backgroundColor: const Color(0xFF1A1A2E),
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            SizedBox(
              height: 80,
              child: Stack(
                children: [
                  _FloatingHeart(ctrl: _ctrl, startX: 0.2, delay: 0.0),
                  _FloatingHeart(ctrl: _ctrl, startX: 0.5, delay: 0.3),
                  _FloatingHeart(ctrl: _ctrl, startX: 0.8, delay: 0.6),
                  _FloatingHeart(ctrl: _ctrl, startX: 0.35, delay: 0.15),
                  _FloatingHeart(ctrl: _ctrl, startX: 0.65, delay: 0.45),
                  const Center(
                    child: Icon(Icons.favorite, color: Color(0xFFFD79A8), size: 40),
                  ),
                ],
              ),
            ),
            const Text('It\'s a Match!',
                style: TextStyle(
                    color: Colors.white,
                    fontSize: 24,
                    fontWeight: FontWeight.bold)),
            const SizedBox(height: 16),
            ...widget.icebreakers.take(3).map((msg) => Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: const Color(0xFF2D1B69),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(msg,
                        style:
                            const TextStyle(color: Colors.white70, fontSize: 13)),
                  ),
                )),
            const SizedBox(height: 20),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF6C5CE7),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12)),
                ),
                onPressed: widget.onChat,
                child: const Text('Start Chatting'),
              ),
            ),
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Keep Swiping',
                  style: TextStyle(color: Colors.grey)),
            ),
          ],
        ),
      ),
    );
  }
}

class _FloatingHeart extends StatelessWidget {
  final AnimationController ctrl;
  final double startX;
  final double delay;
  const _FloatingHeart({required this.ctrl, required this.startX, required this.delay});

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: ctrl,
      builder: (context, child) {
        final t = (ctrl.value + delay) % 1.0;
        return Positioned(
          left: MediaQuery.of(context).size.width * startX - 100,
          top: 60 - t * 60,
          child: Opacity(
            opacity: (1.0 - t).clamp(0.0, 1.0),
            child: Icon(Icons.favorite,
                color: const Color(0xFFFD79A8).withValues(alpha: 0.7),
                size: 14 + t * 10),
          ),
        );
      },
    );
  }
}

class _ActionButton extends StatelessWidget {
  final IconData icon;
  final Color color;
  final VoidCallback onTap;
  const _ActionButton({required this.icon, required this.color, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Material(
      color: color,
      shape: const CircleBorder(),
      child: InkWell(
        onTap: onTap,
        customBorder: const CircleBorder(),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Icon(icon, color: Colors.white, size: 24),
        ),
      ),
    );
  }
}
