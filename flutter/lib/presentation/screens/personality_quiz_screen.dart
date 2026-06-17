import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:spark_app/core/network/api_client.dart';
import 'package:spark_app/core/tracking/tracker.dart';
import 'package:spark_app/data/models/user_model.dart';

class PersonalityQuizScreen extends ConsumerStatefulWidget {
  const PersonalityQuizScreen({super.key});

  @override
  ConsumerState<PersonalityQuizScreen> createState() => _PersonalityQuizScreenState();
}

class _PersonalityQuizScreenState extends ConsumerState<PersonalityQuizScreen> {
  List<PersonalityQuestion>? _questions;
  int _current = 0;
  final Map<int, int> _answers = {};
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadQuestions();
  }

  Future<void> _loadQuestions() async {
    try {
      Tracker().track('personality_quiz_start');
      final api = ref.read(apiClientProvider);
      final resp = await api.get('/api/v1/user/personality/questions');
      final list = (resp.data as List<dynamic>)
          .map((e) => PersonalityQuestion.fromJson(e))
          .toList();
      setState(() { _questions = list; _loading = false; });
    } catch (_) {
      setState(() => _loading = false);
    }
  }

  Future<void> _submit() async {
    final api = ref.read(apiClientProvider);
    final answers = _answers.entries
        .map((e) => {'question_id': e.key, 'option_id': e.value})
        .toList();
    await api.post('/api/v1/user/personality', data: {'answers': answers});
    Tracker().track('personality_quiz_complete', properties: {
      'duration_ms': DateTime.now().millisecondsSinceEpoch,
    });
    if (mounted) context.go('/match');
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (_questions == null || _questions!.isEmpty) {
      return Scaffold(
        body: Center(
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            const Text('No questions available'),
            ElevatedButton(
              onPressed: () {
                Tracker().track('personality_quiz_skip', properties: {'current_question_index': 0});
                context.go('/match');
              },
              child: const Text('Skip')),
          ]),
        ),
      );
    }

    final q = _questions![_current];
    final progress = (_current + 1) / _questions!.length;

    return Scaffold(
      appBar: AppBar(title: const Text('Personality Quiz')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: LinearProgressIndicator(
                  value: progress,
                  minHeight: 6,
                ),
              ),
              const SizedBox(height: 8),
              Text('Question ${_current + 1} of ${_questions!.length}',
                  style: TextStyle(color: Colors.grey[500])),
              const Spacer(flex: 2),
              Text(q.questionText,
                  style: Theme.of(context).textTheme.headlineSmall,
                  textAlign: TextAlign.center),
              const Spacer(),
              ...q.options.map((opt) => Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: OutlinedButton(
                  onPressed: () {
                    Tracker().track('personality_quiz_answer', properties: {
                      'question_id': q.id,
                      'option_id': opt.id,
                      'question_index': _current,
                    });
                    setState(() {
                      _answers[q.id] = opt.id;
                      if (_current < _questions!.length - 1) {
                        _current++;
                      } else {
                        _submit();
                      }
                    });
                  },
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: Text(opt.optionText),
                ),
              )),
              const Spacer(flex: 2),
            ],
          ),
        ),
      ),
    );
  }
}
