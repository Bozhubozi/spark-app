package service

import "strings"

// 12×12 zodiac compatibility matrix (0-100).
// Based on traditional Chinese & Western zodiac pairing wisdom.
var zodiacMatrix = [12][12]int{
	// 水瓶 双鱼 白羊 金牛 双子 巨蟹 狮子 处女 天秤 天蝎 射手 摩羯
	{90, 60, 85, 40, 95, 50, 85, 45, 95, 55, 90, 50}, // 水瓶
	{60, 90, 55, 85, 50, 95, 55, 85, 50, 90, 55, 80}, // 双鱼
	{85, 55, 90, 55, 95, 50, 95, 45, 85, 50, 95, 55}, // 白羊
	{40, 85, 55, 85, 50, 90, 55, 90, 50, 80, 50, 95}, // 金牛
	{95, 50, 95, 50, 90, 55, 95, 50, 95, 45, 90, 55}, // 双子
	{50, 95, 50, 90, 55, 85, 55, 85, 55, 85, 50, 85}, // 巨蟹
	{85, 55, 95, 55, 95, 55, 90, 55, 95, 50, 90, 55}, // 狮子
	{45, 85, 45, 90, 50, 85, 55, 80, 50, 80, 55, 85}, // 处女
	{95, 50, 85, 50, 95, 55, 95, 50, 85, 55, 90, 55}, // 天秤
	{55, 90, 50, 80, 45, 85, 50, 80, 55, 85, 55, 80}, // 天蝎
	{90, 55, 95, 50, 90, 50, 90, 55, 90, 55, 90, 55}, // 射手
	{50, 80, 55, 95, 55, 85, 55, 85, 55, 80, 55, 85}, // 摩羯
}

var zodiacNames = []string{
	"水瓶座", "双鱼座", "白羊座", "金牛座", "双子座", "巨蟹座",
	"狮子座", "处女座", "天秤座", "天蝎座", "射手座", "摩羯座",
}

type ZodiacService struct{}

func NewZodiacService() *ZodiacService { return &ZodiacService{} }

// Compatibility returns 0-100 score for two zodiac signs.
func (s *ZodiacService) Compatibility(a, b string) int {
	ai := s.indexOf(a)
	bi := s.indexOf(b)
	if ai < 0 || bi < 0 {
		return 50
	}
	return zodiacMatrix[ai][bi]
}

// Report returns a textual compatibility description.
func (s *ZodiacService) Report(a, b string) string {
	score := s.Compatibility(a, b)
	switch {
	case score >= 85:
		return s.pickMsg(a, b, highMsgs)
	case score >= 65:
		return s.pickMsg(a, b, midMsgs)
	default:
		return s.pickMsg(a, b, lowMsgs)
	}
}

func (s *ZodiacService) indexOf(name string) int {
	for i, n := range zodiacNames {
		if n == name {
			return i
		}
	}
	return -1
}

func (s *ZodiacService) pickMsg(a, b string, pool []string) string {
	msg := pool[hashStrings(a, b)%len(pool)]
	msg = strings.ReplaceAll(msg, "{A}", a)
	msg = strings.ReplaceAll(msg, "{B}", b)
	return msg
}

func hashStrings(a, b string) int {
	h := 0
	for _, r := range a + b {
		h = h*31 + int(r)
	}
	if h < 0 {
		h = -h
	}
	return h
}

var highMsgs = []string{
	"{A}和{B}是天生的灵魂搭档！一个眼神就能懂对方的默契，在一起总有聊不完的话题。",
	"{A}和{B}的匹配度超高～你们互补又理解，是朋友圈里公认的神仙组合。",
	"{A}和{B}的缘分指数爆表！性格互补，兴趣相投，在一起会很舒服。",
	"星象显示，{A}和{B}是绝佳拍档。你们容易一见如故，相处越久越有默契。",
	"{A}遇到{B}，就像咖啡遇到牛奶——单独已经很好了，但在一起更完美。",
}

var midMsgs = []string{
	"{A}和{B}之间有好感的基础，但需要多一些耐心和理解。慢慢来，风景在路上。",
	"{A}和{B}的星象没有天雷地火，但有细水长流的潜力。时间会是最好的证明。",
	"{A}和{B}各有各的世界，但恰好是两个世界的交界处最容易产生奇妙反应。",
	"{A}和{B}不算是传统意义上的绝配，但谁说爱情一定要按照星座来？",
	"{A}和{B}在一起需要一些磨合期，但磨合过后的感情往往更坚固。",
}

var lowMsgs = []string{
	"{A}和{B}的星象差距有点大，但差异也是一种吸引力。敢挑战吗？",
	"{A}和{B}是两个不同频率的人，但有时候，正是不同才让彼此吸引。",
	"星座说{A}和{B}不太合，但别太当真～真心比星象更重要。",
	"{A}和{B}就像甜粽子和咸粽子——理念不同，但都是好粽子。",
	"{A}和{B}的组合很少见，但最罕见的搭配往往最让人惊喜。",
}
