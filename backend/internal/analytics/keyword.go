package analytics

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/go-ego/gse"
)

// KeywordExtractor 关键词提取器
type KeywordExtractor struct {
	seg              gse.Segmenter
	stopWords        map[string]bool
	meaninglessWords map[string]bool
}

// KeywordFreq 关键词频率
type KeywordFreq struct {
	Word  string
	Count int
}

// NewKeywordExtractor 创建关键词提取器
func NewKeywordExtractor() *KeywordExtractor {
	// 初始化 gse 分词器
	var seg gse.Segmenter
	seg.LoadDict() // 加载默认词典

	// 加载停用词表
	stopWords := loadStopWords()

	// 加载无意义词表
	meaninglessWords := loadMeaninglessWords()

	return &KeywordExtractor{
		seg:              seg,
		stopWords:        stopWords,
		meaninglessWords: meaninglessWords,
	}
}

// Close 关闭分词器
func (ke *KeywordExtractor) Close() {
	// gse 不需要释放资源
}

// ExtractTopKeywords 提取 Top N 关键词（使用 TF-IDF 权重）
func (ke *KeywordExtractor) ExtractTopKeywords(texts []string, topN int) []KeywordFreq {
	if len(texts) == 0 {
		return []KeywordFreq{}
	}

	// 第一步：统计每篇文档的词频
	docWordFreqs := make([]map[string]int, len(texts))
	totalWordCount := make(map[string]int) // 全局词频
	docCount := make(map[string]int)       // 包含该词的文档数

	for i, text := range texts {
		docWordFreqs[i] = make(map[string]int)
		words := ke.seg.Cut(text, true)
		seenInDoc := make(map[string]bool)

		for _, word := range words {
			word = strings.TrimSpace(word)

			// 过滤无效词
			if !ke.isValidKeyword(word) {
				continue
			}

			docWordFreqs[i][word]++
			totalWordCount[word]++

			// 记录文档频率（每篇文档只计数一次）
			if !seenInDoc[word] {
				docCount[word]++
				seenInDoc[word] = true
			}
		}
	}

	// 第二步：计算 TF-IDF 分数
	tfidfScores := make(map[string]float64)
	numDocs := float64(len(texts))

	for word, totalFreq := range totalWordCount {
		// TF: 词频 / 总词数
		tf := float64(totalFreq)

		// IDF: log(总文档数 / 包含该词的文档数)
		idf := math.Log(numDocs / float64(docCount[word]))

		// TF-IDF 分数
		tfidfScores[word] = tf * idf
	}

	// 第三步：按 TF-IDF 分数排序
	keywords := make([]KeywordFreq, 0, len(tfidfScores))
	for word, score := range tfidfScores {
		keywords = append(keywords, KeywordFreq{
			Word:  word,
			Count: int(score * 100), // 转换为整数便于显示
		})
	}

	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Count > keywords[j].Count
	})

	// 返回 Top N
	if len(keywords) > topN {
		keywords = keywords[:topN]
	}

	return keywords
}

// isValidKeyword 判断是否为有效关键词
func (ke *KeywordExtractor) isValidKeyword(word string) bool {
	// 1. 长度过滤：至少2个字符
	if len([]rune(word)) < 2 {
		return false
	}

	// 2. 长度过滤：不超过10个字符（避免长句子）
	if len([]rune(word)) > 10 {
		return false
	}

	// 3. 停用词过滤
	if ke.stopWords[word] {
		return false
	}

	// 4. 无意义词过滤
	if ke.meaninglessWords[word] {
		return false
	}

	// 5. 纯数字过滤
	if isNumber(word) {
		return false
	}

	// 6. 纯标点符号过滤
	if isPunctuation(word) {
		return false
	}

	// 7. 纯英文字母过滤（单个英文单词）
	if isPureEnglish(word) {
		return false
	}

	// 8. 包含特殊字符过滤
	if containsSpecialChars(word) {
		return false
	}

	// 9. 必须包含至少一个中文字符
	if !containsChinese(word) {
		return false
	}

	return true
}

// loadStopWords 加载停用词表（扩展版）
func loadStopWords() map[string]bool {
	// 常用中文停用词（大幅扩展）
	stopWords := []string{
		// 基础停用词
		"的", "了", "在", "是", "我", "有", "和", "就", "不", "人", "都", "一", "一个",
		"上", "也", "很", "到", "说", "要", "去", "你", "会", "着", "没有", "看", "好",
		"自己", "这", "那", "里", "为", "以", "个", "用", "来", "时", "大", "地", "可以",
		"这个", "中", "么", "出", "而", "能", "对", "多", "然后", "她", "他", "但是",

		// 连词介词
		"与", "及", "等", "被", "从", "由", "于", "将", "或", "把", "让", "给", "向",
		"如", "若", "则", "且", "又", "之", "所", "其", "些", "某", "该", "每",
		"各", "另", "别", "只", "仅", "还", "更", "最", "非", "无", "未", "已", "曾",

		// 时间词
		"正在", "正", "在于", "今天", "昨天", "明天", "今年", "去年", "明年",
		"现在", "过去", "将来", "以前", "以后", "之前", "之后", "当时", "此时",

		// 关联词
		"关于", "对于", "根据", "按照", "通过", "经过", "为了",
		"因为", "所以", "如果", "虽然", "但是", "然而", "因此", "于是", "接着", "然后",
		"首先", "其次", "最后", "总之", "总的来说", "一般来说", "换句话说", "也就是说",

		// 疑问词
		"比如", "例如", "譬如", "诸如", "等等", "之类", "什么", "怎么", "怎样", "如何",
		"哪里", "哪儿", "哪些", "多少", "几", "为什么", "为何",

		// 指示词
		"第", "当", "此", "彼", "今", "昨", "明", "这些", "那些", "这样", "那样",
		"如此", "这里", "那里", "这儿", "那儿",

		// 程度副词
		"非常", "十分", "特别", "尤其", "格外", "更加", "越来越", "比较", "相当",
		"稍微", "略微", "有点", "一点", "一些",

		// 助词
		"啊", "呀", "吗", "吧", "呢", "哦", "哈", "嘛", "啦", "哪", "呐",

		// 代词
		"我们", "你们", "他们", "她们", "它们", "咱们", "大家", "别人", "其他",
		"自己", "本人", "彼此", "相互",

		// 动词（无实际意义）
		"进行", "实现", "完成", "做", "搞", "弄", "整", "成为", "变成", "得到",
		"获得", "取得", "达到", "产生", "形成", "造成", "引起", "导致",

		// 形容词（过于宽泛）
		"新", "旧", "大", "小", "多", "少", "高", "低", "长", "短", "好", "坏",
		"美", "丑", "强", "弱", "快", "慢", "远", "近",

		// 量词
		"个", "位", "名", "条", "项", "件", "次", "遍", "回", "趟", "场", "番",
		"种", "类", "样", "份", "批", "群", "队", "组", "套", "副", "幅",

		// 其他无意义词
		"方面", "情况", "问题", "工作", "事情", "东西", "地方", "时候", "方式",
		"方法", "过程", "结果", "原因", "目的", "作用", "意义", "价值", "影响",
		"内容", "形式", "性质", "特点", "特征", "状态", "程度", "范围", "领域",
	}

	stopWordsMap := make(map[string]bool)
	for _, word := range stopWords {
		stopWordsMap[word] = true
	}

	return stopWordsMap
}

// loadMeaninglessWords 加载无意义词表（新增）
func loadMeaninglessWords() map[string]bool {
	// 这些词虽然不是停用词，但在关键词提取中没有实际意义
	meaninglessWords := []string{
		// 网络用语
		"哈哈", "呵呵", "嘿嘿", "嘻嘻", "哎呀", "哇塞", "天哪",

		// 口语化表达
		"知道", "觉得", "认为", "以为", "感觉", "发现", "看到", "听到",
		"想要", "希望", "打算", "准备", "开始", "继续", "结束", "停止",

		// 模糊表达
		"可能", "也许", "大概", "差不多", "左右", "上下", "前后", "左右",
		"基本", "主要", "重要", "关键", "核心", "根本", "本质", "实质",

		// 空洞词汇
		"东西", "事物", "物品", "物质", "材料", "资料", "信息", "数据",
		"系统", "平台", "渠道", "途径", "手段", "措施", "办法", "方案",

		// 时间相关（过于宽泛）
		"时间", "日期", "年份", "月份", "星期", "周末", "假期", "节日",

		// 空间相关（过于宽泛）
		"空间", "位置", "地点", "场所", "区域", "范围", "地区", "地域",

		// 数量相关（过于宽泛）
		"数量", "数字", "数目", "总数", "总计", "合计", "共计",

		// 程度相关
		"程度", "水平", "层次", "等级", "级别", "档次",

		// 关系相关
		"关系", "联系", "关联", "相关", "有关", "无关",

		// 状态相关
		"状况", "状态", "情形", "境况", "局面", "形势",

		// 其他
		"部分", "全部", "整体", "局部", "总体", "个体",
		"一般", "特殊", "普通", "常见", "罕见", "稀有",
		"正常", "异常", "标准", "规范", "要求", "条件",
	}

	meaninglessWordsMap := make(map[string]bool)
	for _, word := range meaninglessWords {
		meaninglessWordsMap[word] = true
	}

	return meaninglessWordsMap
}

// isNumber 判断是否为纯数字
func isNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// isPunctuation 判断是否为标点符号
func isPunctuation(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsPunct(r) && !unicode.IsSymbol(r) {
			return false
		}
	}
	return true
}

// isPureEnglish 判断是否为纯英文
func isPureEnglish(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

// containsChinese 判断是否包含中文字符
func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// containsSpecialChars 判断是否包含特殊字符
func containsSpecialChars(s string) bool {
	// 匹配特殊字符（除了中文、英文、数字）
	pattern := regexp.MustCompile(`[^\p{Han}\p{L}\p{N}]`)
	return pattern.MatchString(s)
}
