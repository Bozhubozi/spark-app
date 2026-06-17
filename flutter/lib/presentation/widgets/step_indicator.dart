import 'package:flutter/material.dart';

class StepIndicator extends StatelessWidget {
  final int total;
  final int current;
  const StepIndicator({super.key, required this.total, required this.current});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(total, (i) {
        final active = i < current;
        return AnimatedContainer(
          duration: const Duration(milliseconds: 300),
          margin: const EdgeInsets.symmetric(horizontal: 4),
          width: active ? 24 : 8,
          height: 8,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(4),
            color: active ? const Color(0xFF6C5CE7) : Colors.grey[700],
          ),
        );
      }),
    );
  }
}
