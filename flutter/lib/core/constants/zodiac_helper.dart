import 'package:flutter/material.dart';

String zodiacFromBirth(String? birthDate) {
  if (birthDate == null) return '';
  final bd = DateTime.tryParse(birthDate);
  if (bd == null) return '';
  final month = bd.month;
  final day = bd.day;
  if ((month == 1 && day >= 20) || (month == 2 && day <= 18)) return '♒';
  if ((month == 2 && day >= 19) || (month == 3 && day <= 20)) return '♓';
  if ((month == 3 && day >= 21) || (month == 4 && day <= 19)) return '♈';
  if ((month == 4 && day >= 20) || (month == 5 && day <= 20)) return '♉';
  if ((month == 5 && day >= 21) || (month == 6 && day <= 21)) return '♊';
  if ((month == 6 && day >= 22) || (month == 7 && day <= 22)) return '♋';
  if ((month == 7 && day >= 23) || (month == 8 && day <= 22)) return '♌';
  if ((month == 8 && day >= 23) || (month == 9 && day <= 22)) return '♍';
  if ((month == 9 && day >= 23) || (month == 10 && day <= 23)) return '♎';
  if ((month == 10 && day >= 24) || (month == 11 && day <= 22)) return '♏';
  if ((month == 11 && day >= 23) || (month == 12 && day <= 21)) return '♐';
  return '♑';
}

String zodiacEmojiToName(String emoji) {
  switch (emoji) {
    case '♒': return '水瓶座';
    case '♓': return '双鱼座';
    case '♈': return '白羊座';
    case '♉': return '金牛座';
    case '♊': return '双子座';
    case '♋': return '巨蟹座';
    case '♌': return '狮子座';
    case '♍': return '处女座';
    case '♎': return '天秤座';
    case '♏': return '天蝎座';
    case '♐': return '射手座';
    case '♑': return '摩羯座';
    default: return '';
  }
}

class ZodiacTrait {
  final String element;
  final String strength;
  final Color color;
  const ZodiacTrait({required this.element, required this.strength, required this.color});
}

const zodiacTraits = {
  '♒': ZodiacTrait(element: '风象', strength: '独立创新，思维超前', color: Color(0xFF74B9FF)),
  '♓': ZodiacTrait(element: '水象', strength: '温柔敏感，富有想象力', color: Color(0xFF0984E3)),
  '♈': ZodiacTrait(element: '火象', strength: '热情直接，行动力强', color: Color(0xFFE17055)),
  '♉': ZodiacTrait(element: '土象', strength: '稳重可靠，懂得享受', color: Color(0xFF00B894)),
  '♊': ZodiacTrait(element: '风象', strength: '机智灵活，好奇心强', color: Color(0xFFFDCB6E)),
  '♋': ZodiacTrait(element: '水象', strength: '细腻温暖，重视感情', color: Color(0xFF6C5CE7)),
  '♌': ZodiacTrait(element: '火象', strength: '自信大方，天生的主角', color: Color(0xFFD63031)),
  '♍': ZodiacTrait(element: '土象', strength: '细致入微，追求完美', color: Color(0xFF636E72)),
  '♎': ZodiacTrait(element: '风象', strength: '优雅平衡，品味出众', color: Color(0xFFE17055)),
  '♏': ZodiacTrait(element: '水象', strength: '专一深情，洞察力强', color: Color(0xFF2D1B69)),
  '♐': ZodiacTrait(element: '火象', strength: '乐观自由，热爱冒险', color: Color(0xFF00CEC9)),
  '♑': ZodiacTrait(element: '土象', strength: '踏实勤奋，目标明确', color: Color(0xFF636E72)),
};
