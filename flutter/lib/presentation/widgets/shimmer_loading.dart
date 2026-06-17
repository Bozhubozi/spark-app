import 'package:flutter/material.dart';

class ShimmerLoading extends StatefulWidget {
  final Widget child;
  const ShimmerLoading({super.key, required this.child});

  @override
  State<ShimmerLoading> createState() => _ShimmerLoadingState();
}

class _ShimmerLoadingState extends State<ShimmerLoading>
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
    return AnimatedBuilder(
      animation: _ctrl,
      builder: (context, child) {
        return ShaderMask(
          blendMode: BlendMode.srcATop,
          shaderCallback: (bounds) {
            return LinearGradient(
              begin: Alignment.centerLeft,
              end: Alignment.centerRight,
              colors: const [
                Colors.transparent,
                Colors.white12,
                Colors.white24,
                Colors.white12,
                Colors.transparent,
              ],
              stops: [
                (_ctrl.value - 0.3).clamp(0.0, 1.0),
                (_ctrl.value - 0.15).clamp(0.0, 1.0),
                _ctrl.value.clamp(0.0, 1.0),
                (_ctrl.value + 0.15).clamp(0.0, 1.0),
                (_ctrl.value + 0.3).clamp(0.0, 1.0),
              ],
            ).createShader(bounds);
          },
          child: child!,
        );
      },
      child: widget.child,
    );
  }
}

class SkeletonCard extends StatelessWidget {
  const SkeletonCard({super.key});

  @override
  Widget build(BuildContext context) {
    return ShimmerLoading(
      child: Card(
        margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Container(width: 56, height: 56, decoration: const BoxDecoration(
                shape: BoxShape.circle, color: Colors.white12,
              )),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(height: 14, width: 120, decoration: BoxDecoration(
                      borderRadius: BorderRadius.circular(7), color: Colors.white12,
                    )),
                    const SizedBox(height: 8),
                    Container(height: 10, width: 80, decoration: BoxDecoration(
                      borderRadius: BorderRadius.circular(5), color: Colors.white12,
                    )),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class SkeletonListTile extends StatelessWidget {
  const SkeletonListTile({super.key});

  @override
  Widget build(BuildContext context) {
    return ShimmerLoading(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Row(
          children: [
            Container(width: 48, height: 48, decoration: const BoxDecoration(
              shape: BoxShape.circle, color: Colors.white12,
            )),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(height: 12, width: 140, decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(6), color: Colors.white12,
                  )),
                  const SizedBox(height: 6),
                  Container(height: 10, width: 200, decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(5), color: Colors.white12,
                  )),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
