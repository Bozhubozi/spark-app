package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/spark-app/backend/internal/model"
)

const aiServiceKey = "ai_service_enabled"

type PersonalityReportService struct {
	rdb *redis.Client
}

func NewPersonalityReportService(rdb *redis.Client) *PersonalityReportService {
	return &PersonalityReportService{rdb: rdb}
}

// IsAIEnabled checks the Redis degradation switch.
func (s *PersonalityReportService) IsAIEnabled(ctx context.Context) bool {
	if s.rdb == nil {
		return false
	}
	val, err := s.rdb.Get(ctx, aiServiceKey).Result()
	if err != nil {
		return false
	}
	return val == "true"
}

func (s *PersonalityReportService) Generate(dims []model.PersonalityDimension) *model.PersonalityReport {
	labels := map[string]string{}
	for _, d := range dims {
		labels[d.Dimension] = s.describe(d.Dimension, d.Score)
	}

	title := s.title(labels)
	summary := s.summary(labels)
	traits := s.traits(labels)
	advice := s.advice(labels)
	extraversion := s.extraversionDetail(dims)

	return &model.PersonalityReport{
		Title:              title,
		Summary:            summary,
		Traits:             traits,
		Advice:             advice,
		ExtraversionDetail: extraversion,
	}
}

func (s *PersonalityReportService) describe(dim string, score float64) string {
	switch dim {
	case "extraversion":
		if score >= 4 {
			return "外向活跃"
		} else if score >= 3 {
			return "适度外向"
		} else if score >= 2 {
			return "偏内向"
		}
		return "内向沉静"
	case "agreeableness":
		if score >= 4 {
			return "高度共情"
		} else if score >= 3 {
			return "温和友善"
		} else if score >= 2 {
			return "独立自主"
		}
		return "理性务实"
	case "conscientiousness":
		if score >= 4 {
			return "高度自律"
		} else if score >= 3 {
			return "有条不紊"
		} else if score >= 2 {
			return "随性灵活"
		}
		return "自由奔放"
	case "neuroticism":
		if score >= 4 {
			return "敏感细腻"
		} else if score >= 3 {
			return "情绪适中"
		} else if score >= 2 {
			return "心态平和"
		}
		return "淡定从容"
	case "openness":
		if score >= 4 {
			return "好奇心强"
		} else if score >= 3 {
			return "开放适中"
		} else if score >= 2 {
			return "务实稳重"
		}
		return "追求确定"
	}
	return "均衡"
}

func (s *PersonalityReportService) title(labels map[string]string) string {
	o := labels["openness"]
	e := labels["extraversion"]

	if o == "好奇心强" && e == "外向活跃" {
		return "冒险家型"
	}
	if o == "好奇心强" && (e == "内向沉静" || e == "偏内向") {
		return "思想者型"
	}
	if labels["agreeableness"] == "高度共情" && labels["extraversion"] == "外向活跃" {
		return "社交达⼈型"
	}
	if labels["conscientiousness"] == "高度自律" && labels["neuroticism"] == "淡定从容" {
		return "沉稳领航者型"
	}
	if labels["openness"] == "好奇心强" && labels["agreeableness"] == "高度共情" {
		return "文艺创作者型"
	}
	return "自由灵魂型"
}

func (s *PersonalityReportService) summary(labels map[string]string) string {
	parts := []string{}
	for dim, label := range labels {
		switch dim {
		case "extraversion":
			parts = append(parts, fmt.Sprintf("在社交方面你倾向「%s」", label))
		case "agreeableness":
			parts = append(parts, fmt.Sprintf("与人相处时你表现「%s」", label))
		case "conscientiousness":
			parts = append(parts, fmt.Sprintf("做事风格上你「%s」", label))
		case "neuroticism":
			parts = append(parts, fmt.Sprintf("情绪层面你「%s」", label))
		case "openness":
			parts = append(parts, fmt.Sprintf("面对新事物你「%s」", label))
		}
	}
	return strings.Join(parts, "；") + "。"
}

func (s *PersonalityReportService) traits(labels map[string]string) []string {
	var traits []string
	if labels["extraversion"] == "外向活跃" {
		traits = append(traits, "聚会中你是气氛担当", "愿意主动认识新朋友")
	} else if labels["extraversion"] == "内向沉静" || labels["extraversion"] == "偏内向" {
		traits = append(traits, "享受独处的深度时光", "小圈子里的灵魂人物")
	} else {
		traits = append(traits, "合群但不依赖社交", "有自己的节奏")
	}

	if labels["agreeableness"] == "高度共情" {
		traits = append(traits, "朋友眼中的温暖树洞", "能快速感知他人情绪")
	} else if labels["agreeableness"] == "独立自主" || labels["agreeableness"] == "理性务实" {
		traits = append(traits, "重视边界感的清醒派", "讲道理而不讲情面")
	} else {
		traits = append(traits, "好相处有分寸", "在助人与自保间平衡")
	}

	if labels["openness"] == "好奇心强" {
		traits = append(traits, "对新事物永远跃跃欲试", "脑洞大开的创意发电机")
	} else if labels["openness"] == "务实稳重" || labels["openness"] == "追求确定" {
		traits = append(traits, "脚踏实地让人安心", "偏好深耕而非广度")
	} else {
		traits = append(traits, "在探索与稳定间找到平衡")
	}

	return traits
}

func (s *PersonalityReportService) advice(labels map[string]string) string {
	var lines []string

	if labels["extraversion"] == "内向沉静" {
		lines = append(lines, "不用强迫自己变成社交焦点，你的深度才是稀有的吸引力。")
	}
	if labels["agreeableness"] == "独立自主" || labels["agreeableness"] == "理性务实" {
		lines = append(lines, "偶尔放下逻辑，纯粹地感受一次，也许会带来意想不到的连接。")
	}
	if labels["conscientiousness"] == "随性灵活" || labels["conscientiousness"] == "自由奔放" {
		lines = append(lines, "自由是你的魅力，但偶尔给对方一个确定的约定会更让人安心。")
	}
	if labels["neuroticism"] == "敏感细腻" {
		lines = append(lines, "你的敏感是一份天赋，它能让你捕捉到别人错过的心动信号。")
	}
	if labels["openness"] == "好奇心强" {
		lines = append(lines, "你的好奇心会引领你遇见有趣的人和事，保持这份探索的热忱。")
	}
	lines = append(lines, "真实比完美更有吸引力。做自己，对的人会为你停留。")

	return strings.Join(lines, " ")
}

func (s *PersonalityReportService) extraversionDetail(dims []model.PersonalityDimension) string {
	for _, d := range dims {
		if d.Dimension == "extraversion" {
			if d.Score >= 4 {
				return "你从社交中汲取能量，像一块太阳能板——人群就是你的阳光。但别忘了偶尔也需要树荫。"
			} else if d.Score >= 3 {
				return "社交对你来说像调味品：太少会淡，太多会腻。掌握这个微妙的平衡是你的天赋。"
			} else {
				return "你的内心世界足够丰盛，不需要喧闹的外部来填补。这不是孤僻，是自洽。"
			}
		}
	}
	return ""
}
