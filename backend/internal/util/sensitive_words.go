package util

// SensitiveWords is the base word list for DFA filtering.
// In production, supplement with Aliyun Content Safety API.
var SensitiveWords = []string{
	// Profanity / slurs
	"傻逼", "弱智", "脑残", "智障",
	"他妈的", "妈了个", "操你", "日你",
	"你妈", "草泥马", "滚蛋", "废物",
	"贱人", "婊子", "骚货", "荡妇",
	"流氓", "人渣", "败类",

	// Harassment / threatening
	"去死", "不得好死", "弄死你", "砍死",
	"报复", "打死你", "杀了你", "灭了你",

	// Sexual harassment
	"约炮", "一夜情", "裸聊", "视频裸",
	"包养", "援交", "卖淫", "嫖娼",
	"色情", "黄片", "AV", "三级片",

	// Scam / fraud
	"兼职刷单", "刷单返利", "投资理财", "稳赚不赔",
	"加微信", "加我微信", "扫码加", "加QQ",
	"免费领", "点击链接", "下载APP", "注册送",
	"赌博", "博彩", "彩票", "时时彩",
	"区块链投资", "数字货币", "虚拟币",

	// Gambling
	"赌场", "下注", "押注", "赔率",
	"百家乐", "六合彩", "老虎机",

	// Drugs
	"毒品", "吸毒", "大麻", "海洛因",
	"冰毒", "摇头丸", "K粉", "可卡因",

	// Violence / extremism
	"恐怖主义", "极端组织", "武器", "枪支",
	"爆炸", "炸弹",

	// Spam / ad
	"加好友", "关注我", "互粉",
}

// BuildDFAFilter creates a singleton DFA filter from the default word list.
func BuildDFAFilter() *DFAFilter {
	return NewDFAFilter(SensitiveWords)
}
