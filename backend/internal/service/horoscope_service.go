package service

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/spark-app/backend/internal/model"
)

// Daily horoscope using compositional rule-based templates.
// 12 zodiac × 5 personality archetypes = 60 unique combinations.

type archetype int

const (
	archOutgoing  archetype = iota // high extraversion
	archArtistic                   // high openness, high neuroticism
	archPragmatic                  // high conscientiousness, low neuroticism
	archSocial                     // high extraversion, high agreeableness
	archLowKey                     // low extraversion, high agreeableness
)

// Zodiac traits — one-line descriptions for each sign.
var zodiacTraits = map[string]struct {
	Element  string
	Strength string
	Advice   string
}{
	"水瓶座": {"风象", "独立创新，思维超前", "偶尔接接地气，会让你的魅力更可触及"},
	"双鱼座": {"水象", "温柔敏感，富有想象力", "浪漫是你最大的武器，不必藏起来"},
	"白羊座": {"火象", "热情直接，行动力强", "冲的时候别忘了看看身边的人"},
	"金牛座": {"土象", "稳重可靠，懂得享受", "你的慢热是筛选器，对的人会等你"},
	"双子座": {"风象", "机智灵活，好奇心强", "你的多变是魅力，但偶尔专注更迷人"},
	"巨蟹座": {"水象", "细腻温暖，重视感情", "你的温柔值得被好好珍惜"},
	"狮子座": {"火象", "自信大方，天生的主角", "偶尔让别人发光，你会收获更多"},
	"处女座": {"土象", "细致入微，追求完美", "完美不是必需品，真实才是"},
	"天秤座": {"风象", "优雅平衡，品味出众", "选择困难时，相信直觉"},
	"天蝎座": {"水象", "专一深情，洞察力强", "不用把心墙筑太高，真诚比防御更安全"},
	"射手座": {"火象", "乐观自由，热爱冒险", "自由固然好，但有人一起看风景更美"},
	"摩羯座": {"土象", "踏实勤奋，目标明确", "事业很重要，但也别忘了给感情留点时间"},
}

// Archetype templates — each has opening, body, closing parts.
var archetypeTemplates = map[archetype]struct {
	Opening []string
	Body    []string
	Closing []string
}{
	archOutgoing: {
		Opening: []string{
			"今天的社交能量满满，",
			"活力四射的一天，",
			"你今天的感染力特别强，",
		},
		Body: []string{
			"适合主动发起邀约或参加活动。",
			"不妨约朋友出去走走，可能会有意外的缘分。",
			"你的热情会吸引同频的人靠近。",
		},
		Closing: []string{
			"记住，你本身就是一道光，不需要刻意讨好任何人。",
			"做最真实的自己就是最好的策略。",
			"保持笑容，今天会很好。",
		},
	},
	archArtistic: {
		Opening: []string{
			"今天的审美感知力在线，",
			"灵感会在不经意间造访，",
			"你的细腻在今天特别动人，",
		},
		Body: []string{
			"适合去看展、拍照，或者只是安静地听一张专辑。",
			"把一闪而过的想法记下来，它可能比你想象的重要。",
			"独处的时光也能滋养你，不用强迫自己合群。",
		},
		Closing: []string{
			"你的敏感是一份天赋，而不是缺陷。",
			"懂你的人不需要解释，不懂的人解释了也没用。",
			"允许自己脆弱，那是另一种勇敢。",
		},
	},
	archPragmatic: {
		Opening: []string{
			"今天的你思路清晰，",
			"节奏把握得刚刚好，",
			"你会发现效率比平时更高，",
		},
		Body: []string{
			"适合完成一件拖了很久的小事，成就感会带来好运。",
			"把生活整理得井井有条，好心情自然就来了。",
			"按计划行事，但预留10%给意外的惊喜。",
		},
		Closing: []string{
			"踏实是最靠谱的魅力，有人正在默默欣赏你的稳重。",
			"慢一点没关系，长期主义的人笑到最后。",
			"你不需要改变自己来取悦别人，你的认真本身就足够迷人。",
		},
	},
	archSocial: {
		Opening: []string{
			"今天的社交运很旺，",
			"你是人群中的润滑剂，",
			"你的笑容今天特别有感染力，",
		},
		Body: []string{
			"适合当那个牵线搭桥的人，帮朋友组个局。",
			"和不同圈子的人聊聊，会有化学反应。",
			"你的朋友圈里可能藏着一段缘分。",
		},
		Closing: []string{
			"社交是你的天赋，但别忘了留点时间给自己。",
			"关系不在于多，而在于真。",
			"你让周围所有人都舒服，也请让自己舒服一次。",
		},
	},
	archLowKey: {
		Opening: []string{
			"今天适合慢下来，",
			"不需要太用力的一天，",
			"安静是你的力量源泉，",
		},
		Body: []string{
			"一个人的咖啡、一本好书，就是很好的下午。",
			"不必强迫自己合群，享受独处也是一种能力。",
			"深度交流比泛泛而谈更适合今天的你。",
		},
		Closing: []string{
			"内向不是缺陷，安静的人往往最懂人心。",
			"你的深度，是有人正在寻找的宝藏。",
			"不用急着发光，静静绽放也很好。",
		},
	},
}

type HoroscopeService struct{}

func NewHoroscopeService() *HoroscopeService { return &HoroscopeService{} }

// Daily generates a horoscope message for a given zodiac and personality dimensions.
func (s *HoroscopeService) Daily(zodiac string, dims []model.PersonalityDimension) string {
	arch := classifyArchetype(dims)
	zodiacInfo, ok := zodiacTraits[zodiac]
	if !ok {
		return s.fallback(zodiac)
	}
	templates := archetypeTemplates[arch]

	idx := dailySeed(zodiac, int(arch))

	opening := templates.Opening[idx%len(templates.Opening)]
	body := templates.Body[(idx/7)%len(templates.Body)]
	closing := templates.Closing[(idx/13)%len(templates.Closing)]
	advice := zodiacInfo.Advice

	return fmt.Sprintf("%s%s%s %s %s",
		opening, zodiacInfo.Strength+"的你，", body, advice+"。", closing)
}

func classifyArchetype(dims []model.PersonalityDimension) archetype {
	scores := map[string]float64{}
	for _, d := range dims {
		scores[d.Dimension] = d.Score
	}
	e := scores["extraversion"]
	a := scores["agreeableness"]
	c := scores["conscientiousness"]
	n := scores["neuroticism"]
	o := scores["openness"]

	if e >= 4 && a >= 3.5 {
		return archSocial
	}
	if e >= 4 {
		return archOutgoing
	}
	if o >= 4 && n >= 3 {
		return archArtistic
	}
	if c >= 4 && n < 3 {
		return archPragmatic
	}
	return archLowKey
}

func dailySeed(zodiac string, arch int) int {
	date := time.Now().Format("2006-01-02")
	hash := md5.Sum([]byte(date + ":" + zodiac + ":" + string(rune(arch+'0'))))
	return int(binary.BigEndian.Uint32(hash[:4]))
}

func (s *HoroscopeService) fallback(zodiac string) string {
	msgs := []string{
		fmt.Sprintf("今天是%s的幸运日，保持好奇和开放，美好的事正在发生。", zodiac),
		fmt.Sprintf("%s今天适合顺其自然——不刻意不强求，缘分会在不经意间出现。", zodiac),
		fmt.Sprintf("星象显示%s今天桃花运不错，注意身边的细节。", zodiac),
	}
	hash := md5.Sum([]byte(time.Now().Format("2006-01-02") + zodiac))
	idx := int(binary.BigEndian.Uint32(hash[:4]))
	return msgs[idx%len(msgs)]
}
