import 'package:go_router/go_router.dart';
import 'package:spark_app/presentation/screens/login_screen.dart';
import 'package:spark_app/presentation/screens/register_screen.dart';
import 'package:spark_app/presentation/screens/register_intercept_screen.dart';
import 'package:spark_app/presentation/screens/interest_select_screen.dart';
import 'package:spark_app/presentation/screens/personality_quiz_screen.dart';
import 'package:spark_app/presentation/screens/home_screen.dart';
import 'package:spark_app/presentation/screens/match_screen.dart';
import 'package:spark_app/presentation/screens/chat_list_screen.dart';
import 'package:spark_app/presentation/screens/chat_screen.dart';
import 'package:spark_app/presentation/screens/profile_screen.dart';
import 'package:spark_app/presentation/screens/personality_card_screen.dart';
import 'package:spark_app/presentation/screens/avatar_select_screen.dart';
import 'package:spark_app/presentation/screens/matches_screen.dart';
import 'package:spark_app/presentation/screens/horoscope_screen.dart';
import 'package:spark_app/presentation/screens/zodiac_compat_screen.dart';
import 'package:spark_app/presentation/screens/birth_date_screen.dart';
import 'package:spark_app/presentation/screens/gender_select_screen.dart';
import 'package:spark_app/presentation/screens/discovery_prefs_screen.dart';
import 'package:spark_app/presentation/screens/likers_screen.dart';
import 'package:spark_app/presentation/screens/notifications_screen.dart';
import 'package:spark_app/presentation/screens/settings_screen.dart';
import 'package:spark_app/presentation/screens/blocked_users_screen.dart';
import 'package:spark_app/presentation/screens/onboarding_welcome_screen.dart';

final router = GoRouter(
  initialLocation: '/welcome',
  routes: [
    GoRoute(path: '/welcome', builder: (_, __) => const RegisterInterceptScreen()),
    GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
    GoRoute(path: '/register', builder: (_, __) => const RegisterScreen()),
    GoRoute(path: '/onboarding/welcome', builder: (_, __) => const OnboardingWelcomeScreen()),
    GoRoute(path: '/onboarding/interests', builder: (_, __) => const InterestSelectScreen()),
    GoRoute(path: '/onboarding/birthdate', builder: (_, __) => const BirthDateScreen()),
    GoRoute(path: '/onboarding/gender', builder: (_, __) => const GenderSelectScreen()),
        GoRoute(path: '/onboarding/avatar', builder: (_, __) => const AvatarSelectScreen()),
    GoRoute(path: '/personality-quiz', builder: (_, __) => const PersonalityQuizScreen()),
    ShellRoute(
      builder: (_, __, child) => HomeScreen(child: child),
      routes: [
        GoRoute(path: '/match', builder: (_, __) => const MatchScreen()),
        GoRoute(path: '/chat', builder: (_, __) => const ChatListScreen()),
        GoRoute(path: '/chat/:roomId', builder: (_, state) {
            final extra = state.extra as Map<String, dynamic>?;
            return ChatScreen(
              roomId: state.pathParameters['roomId']!,
              otherName: extra?['otherName'] as String?,
              otherId: extra?['otherId'] as String?,
            );
          }),
        GoRoute(path: '/discovery-prefs', builder: (_, __) => const DiscoveryPrefsScreen()),
        GoRoute(path: '/matches', builder: (_, __) => const MatchesScreen()),
        GoRoute(path: '/likers', builder: (_, __) => const LikersScreen()),
        GoRoute(path: '/settings', builder: (_, __) => const SettingsScreen()),
	GoRoute(path: '/blocked-users', builder: (_, __) => const BlockedUsersScreen()),
	GoRoute(path: '/notifications', builder: (_, __) => const NotificationsScreen()),
        GoRoute(path: '/horoscope', builder: (_, __) => const HoroscopeScreen()),
        GoRoute(path: '/zodiac-compat/:targetUserId', builder: (_, state) {
            final extra = state.extra as Map<String, dynamic>?;
            return ZodiacCompatScreen(
              targetUserId: state.pathParameters['targetUserId']!,
              targetName: extra?['targetName'] as String? ?? 'User',
              targetBirthDate: extra?['targetBirthDate'] as String?,
            );
          }),
        GoRoute(path: '/profile', builder: (_, __) => const ProfileScreen()),
        GoRoute(path: '/profile/personality', builder: (_, __) => const PersonalityCardScreen()),
      ],
    ),
  ],
);
