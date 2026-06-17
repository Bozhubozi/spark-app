package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/config"
	"github.com/spark-app/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 5 persona archetypes with personality dimension targets
type Persona struct {
	Name          string
	Extraversion  [2]float64 // [min, max]
	Agreeableness [2]float64
	Conscientious [2]float64
	Neuroticism   [2]float64
	Openness      [2]float64
	Nicknames     []string
	Bios          []string
}

var personas = []Persona{
	{
		Name:          "阳光外向型",
		Extraversion:  [2]float64{4.0, 5.0},
		Agreeableness: [2]float64{3.0, 5.0},
		Conscientious: [2]float64{2.0, 4.0},
		Neuroticism:   [2]float64{1.0, 3.0},
		Openness:      [2]float64{3.0, 5.0},
		Nicknames: []string{
			"夏日晴天", "阳光小熊", "快乐星球", "元气满满", "微风轻拂",
			"柠檬不酸", "糖炒栗子", "彩虹糖", "晴天娃娃", "甜甜圈",
			"跳跃音符", "草莓泡芙", "向日葵", "橘子汽水", "棉花糖",
			"薄荷糖", "小太阳", "云朵飘飘", "蜜桃乌龙", "奶昔",
			"果冻布丁", "棒棒糖", "泡泡糖", "蓝莓之夜", "春田花花",
			"波波奶茶", "芒果冰", "樱桃小丸子", "抹茶拿铁", "小熊饼干",
			"冰淇淋", "酸奶盖", "华夫饼", "巧克力棒", "太妃糖",
			"香草天空", "蜜糖", "曲奇饼", "柠檬茶", "布丁狗",
		},
		Bios: []string{
			"每一天都是新的冒险～想找个人一起看日落吃火锅",
			"笑声是最好听的语言，希望能遇到同样开朗的你",
			"周末不宅家主义者，密室/剧本杀/露营都爱",
			"生活已经很苦了，我来当你的糖",
			"朋友圈里的话痨，想找个人聊到天亮",
		},
	},
	{
		Name:          "文艺敏感型",
		Extraversion:  [2]float64{2.0, 3.5},
		Agreeableness: [2]float64{4.0, 5.0},
		Conscientious: [2]float64{3.0, 4.5},
		Neuroticism:   [2]float64{3.5, 5.0},
		Openness:      [2]float64{4.0, 5.0},
		Nicknames: []string{
			"深夜诗人", "墨染青衣", "云深不知", "月光如水", "落花时节",
			"纸上月光", "雨巷", "画中游", "雾里看花", "半夏微凉",
			"素手挽风", "一纸荒年", "浮生若梦", "北城以北", "南山之南",
			"梦里花落", "山水间", "月下独酌", "江上月", "落雪时分",
			"青衫湿", "月华如水", "落笔成诗", "云卷云舒", "听雨",
			"枫叶", "烟雨蒙蒙", "暮色", "西窗烛", "清风徐来",
			"秋水长天", "竹影", "雪落无声", "山川湖海", "归去来兮",
			"旧时光", "墨色", "风筝误", "山水一程", "花间一壶酒",
		},
		Bios: []string{
			"摄影和画画是表达自己的方式，用镜头和画笔记录生活的温度",
			"书和咖啡是我最好的朋友，想找一个人一起逛书店",
			"有时候敏感是一种超能力，能听见花开的声音",
			"不太会说漂亮话，但会记住你喜欢的每一件小事",
			"在喧嚣的城市里，想找一个能安静待在一起的人",
		},
	},
	{
		Name:          "理性务实型",
		Extraversion:  [2]float64{2.0, 4.0},
		Agreeableness: [2]float64{2.0, 4.0},
		Conscientious: [2]float64{4.0, 5.0},
		Neuroticism:   [2]float64{1.0, 2.5},
		Openness:      [2]float64{2.0, 4.0},
		Nicknames: []string{
			"码代码的猫", "逻辑控", "数据分析师", "理性先生", "冷静小姐",
			"算法人生", "工程师小明", "产品经理", "架构师", "Bug终结者",
			"效率至上", "数码控", "极客精神", "目标达成", "凡事有计划",
			"健身达人", "早起冠军", "书虫一号", "知识图谱", "终身学习",
			"投资自己", "清醒生活", "秩序感", "复利思维", "深度工作",
			"独立行走", "靠谱青年", "说干就干", "未雨绸缪", "边界感",
			"长期主义", "工匠精神", "极简生活", "知行合一", "自律即自由",
			"逻辑狂魔", "棋盘人生", "几何体", "方程式", "质点",
		},
		Bios: []string{
			"健身四年，每周雷打不动。希望对方也有自己的热爱和追求。",
			"早睡早起打卡人。生活可以简单，但要有质感。",
			"不太会说甜言蜜语，但答应你的事一定会做到。",
			"读书和思考是我的日常。找一个能深度对话的人。",
			"生活规划得明明白白，就差一个你来一起。",
		},
	},
	{
		Name:          "社交达人型",
		Extraversion:  [2]float64{4.0, 5.0},
		Agreeableness: [2]float64{3.0, 5.0},
		Conscientious: [2]float64{2.0, 3.5},
		Neuroticism:   [2]float64{1.5, 3.0},
		Openness:      [2]float64{3.0, 5.0},
		Nicknames: []string{
			"派对女王", "社交牛逼症", "K歌之王", "舞池中央", "组局达人",
			"鸡尾酒达人", "城市猎人", "派对策划师", "交际花", "朋友圈C位",
			"时尚买手", "Vlogger", "潮流前线", "穿搭博主", "探店达人",
			"音乐节常客", "新店收割机", "活动策划", "圈子中心", "潮人",
			"脱口秀爱好者", "Livehouse常驻", "街舞少年", "说唱新星", "DJ",
			"风靡全场", "话题制造机", "永不冷场", "气氛担当", "局王",
			"约会达人", "社交牛人", "嗨翻全场", "万人迷", "王牌主唱",
			"潮流教父", "娱乐圈观察员", "达人", "秀场常客", "焦点",
		},
		Bios: []string{
			"周末不是在组局就是在去局的路上，KTV/剧本杀/密室都约",
			"穿衣打扮是每天的仪式感，希望你也对自己的风格有追求",
			"认识新朋友是我最大的快乐，期待和你一起探店打卡",
			"Livehouse和音乐节是我的快乐源泉，找演搭子",
			"朋友都说我是个E人，但我也会给你百分百的专注",
		},
	},
	{
		Name:          "低调走心型",
		Extraversion:  [2]float64{1.5, 3.0},
		Agreeableness: [2]float64{3.5, 5.0},
		Conscientious: [2]float64{3.0, 4.5},
		Neuroticism:   [2]float64{2.0, 4.0},
		Openness:      [2]float64{2.5, 4.5},
		Nicknames: []string{
			"小透明", "深海鱼", "藏于心", "微光", "平凡之路",
			"静静", "且听风吟", "不言", "初雪微凉", "人间值得",
			"守候者", "岁月静好", "小满", "见山", "朴素",
			"小城故事", "简单的快乐", "散步达人", "独自美好", "细水长流",
			"温暖大叔", "木木", "默然", "如初", "本真",
			"初心不变", "安静的美男子", "倾听者", "内心的光", "沉淀",
			"柔软时光", "午后阳光", "淡淡的", "安然", "随心",
			"平常心", "知足常乐", "一念", "温润", "踏实",
		},
		Bios: []string{
			"不太擅长表达自己，但会用行动证明一切。",
			"生活节奏不快，想有个人一起慢慢发现这座城市的角落。",
			"养了两只猫，偶尔做饭，期待平淡日子里有你。",
			"话不多但很会照顾人，朋友说我像一杯温水。",
			"一个人的时候看书听歌做饭，两个人的时候想做什么都可以。",
		},
	},
}

var maleFirst = []string{
	"宇", "浩", "轩", "子涵", "一鸣", "天", "杰", "浩然", "博文", "俊杰",
	"子豪", "明哲", "宇航", "天宇", "嘉诚", "健", "伟", "晓明", "文博", "泽宇",
	"瑞", "晨", "铭", "逸飞", "志远", "浩然", "文轩", "一飞", "辰逸", "昊天",
}

var femaleFirst = []string{
	"雨桐", "诗涵", "欣怡", "梓涵", "梦琪", "佳怡", "语嫣", "思雨", "一诺", "梓萱",
	"若曦", "雨萱", "紫涵", "晓雪", "雅琴", "雯", "婷婷", "晓婷", "嘉怡", "琪",
	"雪莹", "梦瑶", "慧", "晓萌", "雨欣", "艺琳", "思彤", "欣妍", "梓涵", "悦",
}

var cities = []struct {
	Name string
	Pop  int // weighting
}{
	{"上海", 25}, {"北京", 25}, {"广州", 20}, {"深圳", 20},
	{"杭州", 15}, {"成都", 15}, {"武汉", 15}, {"南京", 15},
	{"重庆", 12}, {"西安", 12}, {"长沙", 10}, {"苏州", 10},
	{"天津", 8}, {"郑州", 8}, {"厦门", 8}, {"青岛", 8},
	{"合肥", 6}, {"福州", 6}, {"昆明", 5}, {"大连", 5},
}

var zodiacSigns = []struct {
	Name   string
	Emoji  string
	Start  time.Time
	End    time.Time
}{
	{"水瓶座", "♒", date(0, 1, 20), date(0, 2, 18)},
	{"双鱼座", "♓", date(0, 2, 19), date(0, 3, 20)},
	{"白羊座", "♈", date(0, 3, 21), date(0, 4, 19)},
	{"金牛座", "♉", date(0, 4, 20), date(0, 5, 20)},
	{"双子座", "♊", date(0, 5, 21), date(0, 6, 21)},
	{"巨蟹座", "♋", date(0, 6, 22), date(0, 7, 22)},
	{"狮子座", "♌", date(0, 7, 23), date(0, 8, 22)},
	{"处女座", "♍", date(0, 8, 23), date(0, 9, 22)},
	{"天秤座", "♎", date(0, 9, 23), date(0, 10, 23)},
	{"天蝎座", "♏", date(0, 10, 24), date(0, 11, 22)},
	{"射手座", "♐", date(0, 11, 23), date(0, 12, 21)},
	{"摩羯座", "♑", date(0, 12, 22), date(0, 1, 19)},
}

func date(year, month, day int) time.Time {
	return time.Date(2000, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func main() {
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Load interest tags
	var tags []model.InterestTag
	if err := db.Find(&tags).Error; err != nil {
		log.Fatalf("load tags: %v", err)
	}
	fmt.Printf("Loaded %d interest tags\n", len(tags))

	// Load avatar components
	var avatars []model.AvatarComponent
	if err := db.Find(&avatars).Error; err != nil {
		log.Fatalf("load avatars: %v", err)
	}
	fmt.Printf("Loaded %d avatar components\n", len(avatars))

	// Load personality questions
	var questions []model.PersonalityQuestion
	if err := db.Find(&questions).Error; err != nil {
		log.Fatalf("load questions: %v", err)
	}

	// Load options
	var options []model.PersonalityOption
	if err := db.Find(&options).Error; err != nil {
		log.Fatalf("load options: %v", err)
	}

	// Group options by question
	optsByQuestion := make(map[int][]model.PersonalityOption)
	for _, o := range options {
		optsByQuestion[o.QuestionID] = append(optsByQuestion[o.QuestionID], o)
	}

	// Group tags by category
	tagsByCategory := make(map[string][]model.InterestTag)
	for _, t := range tags {
		tagsByCategory[t.Category] = append(tagsByCategory[t.Category], t)
	}

	botsPerPersona := 40
	totalBots := len(personas) * botsPerPersona
	fmt.Printf("Generating %d bots...\n", totalBots)

	hash, _ := bcrypt.GenerateFromPassword([]byte("bot"), bcrypt.DefaultCost)

	for pi, persona := range personas {
		for i := 0; i < botsPerPersona; i++ {
			gender := int8(rng.Intn(2) + 1) // 1=male, 2=female
			var nickname string
			if gender == 1 {
				nickname = persona.Nicknames[i%len(persona.Nicknames)] + maleFirst[rng.Intn(len(maleFirst))]
			} else {
				nickname = persona.Nicknames[i%len(persona.Nicknames)] + femaleFirst[rng.Intn(len(femaleFirst))]
			}

			// Random age 20-30
			year := 2006 - rng.Intn(11)
			month := rng.Intn(12) + 1
			day := rng.Intn(28) + 1
			birthDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

			// Determine zodiac
			zodiac := "未知星座"
			for _, z := range zodiacSigns {
				if (month == int(z.Start.Month()) && day >= z.Start.Day()) ||
					(month == int(z.End.Month()) && day <= z.End.Day()) {
					zodiac = z.Name
					break
				}
			}
			_ = zodiac // stored as metadata on the user

			// Weighted city selection
			totalPop := 0
			for _, c := range cities {
				totalPop += c.Pop
			}
			pick := rng.Intn(totalPop)
			city := cities[0].Name
			acc := 0
			for _, c := range cities {
				acc += c.Pop
				if pick < acc {
					city = c.Name
					break
				}
			}

			bio := persona.Bios[rng.Intn(len(persona.Bios))]

			user := model.User{
				ID:           uuid.New(),
				Nickname:     nickname,
				PasswordHash: string(hash),
				Gender:       gender,
				BirthDate:    &birthDate,
				Bio:          &bio,
				City:         &city,
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
				LastActiveAt: time.Now().Add(-time.Duration(rng.Intn(24)) * time.Hour),
			}

			if err := db.Create(&user).Error; err != nil {
				log.Printf("create user %s: %v", nickname, err)
				continue
			}

			// Assign 5-10 interests spread across categories
			numInterests := 5 + rng.Intn(6)
			categories := make([]string, 0, len(tagsByCategory))
			for cat := range tagsByCategory {
				categories = append(categories, cat)
			}
			rng.Shuffle(len(categories), func(i, j int) {
				categories[i], categories[j] = categories[j], categories[i]
			})

			assigned := 0
			for ci := 0; ci < len(categories) && assigned < numInterests; ci++ {
				cat := categories[ci]
				available := tagsByCategory[cat]
				if len(available) == 0 {
					continue
				}
				tag := available[rng.Intn(len(available))]
				ui := model.UserInterest{UserID: user.ID, TagID: tag.ID, Weight: 1}
				if err := db.Create(&ui).Error; err != nil {
					continue
				}
				assigned++
			}

			// Assign personality answers matching the persona archetype
			for _, q := range questions {
				qOpts := optsByQuestion[q.ID]
				if len(qOpts) == 0 {
					continue
				}
				dim := q.Dimension
				var targetRange [2]float64
				switch dim {
				case "extraversion":
					targetRange = persona.Extraversion
				case "agreeableness":
					targetRange = persona.Agreeableness
				case "conscientiousness":
					targetRange = persona.Conscientious
				case "neuroticism":
					targetRange = persona.Neuroticism
				case "openness":
					targetRange = persona.Openness
				default:
					targetRange = [2]float64{2.0, 4.0}
				}

				// Pick option whose score is closest to the persona's target range midpoint
				target := (targetRange[0] + targetRange[1]) / 2.0
				best := qOpts[0]
				bestDiff := abs(float64(best.Score) - target)
				for _, o := range qOpts {
					diff := abs(float64(o.Score) - target)
					if diff < bestDiff {
						bestDiff = diff
						best = o
					}
				}
				// Add some randomness
				if rng.Float64() < 0.3 {
					best = qOpts[rng.Intn(len(qOpts))]
				}

				ans := model.UserPersonalityAnswer{
					UserID:     user.ID,
					QuestionID: q.ID,
					OptionID:   best.ID,
				}
				if err := db.Create(&ans).Error; err != nil {
					log.Printf("create answer: %v", err)
				}
			}
		}
		fmt.Printf("Persona %d (%s): %d bots created\n", pi+1, persona.Name, botsPerPersona)
	}

	fmt.Printf("Done! Generated %d bots with interests and personality\n", totalBots)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
