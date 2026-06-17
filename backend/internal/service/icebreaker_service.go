package service

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/spark-app/backend/internal/model"
)

// IcebreakerService generates contextual opening messages for new matches.
type IcebreakerService struct {
	zodiacSvc *ZodiacService
}

func NewIcebreakerService(zs *ZodiacService) *IcebreakerService {
	return &IcebreakerService{zodiacSvc: zs}
}

func (s *IcebreakerService) Generate(
	userZodiac, targetZodiac string,
	userInterests, targetInterests []model.InterestTag,
	userPersonality []model.PersonalityDimension,
) []string {
	var topics []string

	// 1. Find shared interests
	shared := sharedInterests(userInterests, targetInterests)
	if len(shared) > 0 {
		topics = append(topics, s.interestIcebreaker(shared))
	}

	// 2. Zodiac compatibility
	score := s.zodiacSvc.Compatibility(userZodiac, targetZodiac)
	topics = append(topics, s.zodiacIcebreaker(userZodiac, targetZodiac, score))

	// 3. Personality-based opener
	if len(userPersonality) > 0 {
		topics = append(topics, s.personalityIcebreaker(userPersonality))
	}

	// 4. Generic warm opener
	topics = append(topics, s.genericIcebreaker())

	return topics
}

func sharedInterests(a, b []model.InterestTag) []model.InterestTag {
	set := make(map[int]model.InterestTag, len(a))
	for _, t := range a {
		set[t.ID] = t
	}
	var shared []model.InterestTag
	for _, t := range b {
		if _, ok := set[t.ID]; ok {
			shared = append(shared, t)
		}
	}
	return shared
}

func (s *IcebreakerService) interestIcebreaker(shared []model.InterestTag) string {
	if len(shared) == 0 {
		return ""
	}
	var names []string
	for _, t := range shared[:min(3, len(shared))] {
		name := t.Name
		if t.Icon != "" {
			name = t.Icon + " " + name
		}
		names = append(names, name)
	}

	templates := []string{
		fmt.Sprintf("发现我们都喜欢%s！你最喜欢的%s是哪一部/哪个？", names[0], names[0]),
		fmt.Sprintf("%s爱好者握手！最近有什么推荐的%s相关的东西吗？", strings.Join(names, "、"), names[0]),
		fmt.Sprintf("看到你也喜欢%s，忍不住想问问你是从什么时候开始的？", names[0]),
	}
	hash := md5.Sum([]byte(strings.Join(names, ",")))
	idx := int(binary.BigEndian.Uint32(hash[:4])) % len(templates)
	return templates[idx]
}

func (s *IcebreakerService) zodiacIcebreaker(a, b string, score int) string {
	var templates []string
	if score >= 85 {
		templates = []string{
			fmt.Sprintf("星象说%s和%s是绝配✨ 认识一下吧～", a, b),
			fmt.Sprintf("%s遇到%s，星座书上说是天造地设的一对！", a, b),
		}
	} else if score >= 65 {
		templates = []string{
			fmt.Sprintf("%s和%s，听说我们相处会很舒服～聊聊看？", a, b),
			fmt.Sprintf("你是%s？我是%s，感觉应该挺合得来的！", b, a),
		}
	} else {
		templates = []string{
			fmt.Sprintf("星座不是一切～%s和%s也想认识一下", a, b),
			fmt.Sprintf("我是%s，不管星座合不合，先聊聊天吧？", a),
		}
	}
	seed := fmt.Sprintf("%s:%s:%s", a, b, time.Now().Format("2006-01-02"))
	hash := md5.Sum([]byte(seed))
	idx := int(binary.BigEndian.Uint32(hash[:4])) % len(templates)
	return templates[idx]
}

func (s *IcebreakerService) personalityIcebreaker(dims []model.PersonalityDimension) string {
	scores := map[string]float64{}
	for _, d := range dims {
		scores[d.Dimension] = d.Score
	}

	var trait string
	if scores["extraversion"] >= 4 {
		trait = "外向开朗"
	} else if scores["openness"] >= 4 {
		trait = "好奇心强"
	} else if scores["agreeableness"] >= 4 {
		trait = "温柔友善"
	} else {
		trait = "随和自在"
	}

	templates := []string{
		fmt.Sprintf("据人格测试说我是个「%s」的人，你觉得准吗？", trait),
		fmt.Sprintf("我的性格标签是「%s」，你呢？", trait),
		fmt.Sprintf("看资料觉得你很有趣，作为「%s」的我向你发起聊天邀请～", trait),
	}
	hash := md5.Sum([]byte(trait + time.Now().Format("2006-01-02")))
	idx := int(binary.BigEndian.Uint32(hash[:4])) % len(templates)
	return templates[idx]
}

func (s *IcebreakerService) genericIcebreaker() string {
	pool := []string{
		"你平时周末一般都会做什么？",
		"用了这个App之后有没有什么有趣的事？",
		"如果你可以瞬间学会一项技能，会选什么？",
		"最近有没有看到什么让你笑出声的东西？",
		"如果能穿越回五年前，你会对自己说什么？",
		"用一种食物形容自己，你会选什么？",
	}
	hash := md5.Sum([]byte(time.Now().Format("2006-01-02 15:04")))
	idx := int(binary.BigEndian.Uint32(hash[:4])) % len(pool)
	return pool[idx]
}
