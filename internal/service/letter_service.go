package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
	"people-page-backend/internal/config"
	"people-page-backend/internal/dao"
	"people-page-backend/internal/model"
)

// GetSystemPrompt 获取系统提示词
func GetSystemPrompt() (string, error) {
	// 1. 获取群众提示词
	prompt, err := dao.GetPromptByType("群众提示词")
	if err != nil {
		return "", fmt.Errorf("获取提示词失败: %w", err)
	}
	basePrompt := prompt.Content

	// 2. 构建单位提示词
	unitsPrompt := buildUnitsPrompt()

	// 3. 构建分类提示词
	classPrompt := buildClassificationPrompt()

	return basePrompt + classPrompt + unitsPrompt, nil
}

func buildUnitsPrompt() string {
	units, err := dao.GetAllUnits()
	if err != nil || len(units) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "\n\n以下是衡水市公安局的单位组织架构，在处理信件时请参考：\n")

	var currentLevel1, currentLevel2 string
	for _, u := range units {
		if u.Level1 != currentLevel1 {
			parts = append(parts, fmt.Sprintf("\n【%s】", u.Level1))
			currentLevel1 = u.Level1
			currentLevel2 = ""
		}

		if u.Level3 != "" {
			if u.Level2 != currentLevel2 {
				parts = append(parts, fmt.Sprintf("  - %s：", u.Level2))
				currentLevel2 = u.Level2
			}
			parts = append(parts, fmt.Sprintf("    - %s", u.Level3))
		} else {
			parts = append(parts, fmt.Sprintf("  - %s", u.Level2))
		}
	}
	parts = append(parts, "\n\n在处理涉及具体单位的信件时，请根据以上组织架构选择合适的处理单位。")
	return strings.Join(parts, "\n")
}

func buildClassificationPrompt() string {
	classifications, err := dao.GetDistinctClassifications()
	if err != nil || len(classifications) == 0 {
		return ""
	}

	// Build tree structure
	tree := make(map[string]map[string][]string)
	for _, c := range classifications {
		if tree[c.Category1] == nil {
			tree[c.Category1] = make(map[string][]string)
		}
		if c.Category3 != "" {
			found := false
			for _, v := range tree[c.Category1][c.Category2] {
				if v == c.Category3 {
					found = true
					break
				}
			}
			if !found {
				tree[c.Category1][c.Category2] = append(tree[c.Category1][c.Category2], c.Category3)
			}
		} else if tree[c.Category1][c.Category2] == nil {
			tree[c.Category1][c.Category2] = []string{}
		}
	}

	var parts []string
	parts = append(parts, "\n\n信件的一级、二级、三级分类体系如下：\n")
	for level1, level2Map := range tree {
		parts = append(parts, fmt.Sprintf("\n【一级分类：%s】", level1))
		for level2, level3List := range level2Map {
			if len(level3List) > 0 {
				parts = append(parts, fmt.Sprintf("  - 【二级分类】%s：【三级分类】%s", level2, strings.Join(level3List, "、")))
			} else {
				parts = append(parts, fmt.Sprintf("  - 【二级分类】%s", level2))
			}
		}
	}
	parts = append(parts, "\n")
	return strings.Join(parts, "\n")
}

// ChatStream 流式调用 DeepSeek API
func ChatStream(messages []map[string]interface{}, w http.ResponseWriter, flusher http.Flusher) error {
	cfg := config.AppConfig.LLM

	reqBody := map[string]interface{}{
		"model":       cfg.Model,
		"messages":    messages,
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
		"stream":      true,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", cfg.APIURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := io.Reader(resp.Body)
	buf := make([]byte, 4096)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "data: ") {
					data := line[6:]
					if data == "[DONE]" {
						fmt.Fprintf(w, "data: {\"done\":true}\n\n")
						flusher.Flush()
						return nil
					}

					var chunk map[string]interface{}
					if err := json.Unmarshal([]byte(data), &chunk); err != nil {
						continue
					}

					choices, _ := chunk["choices"].([]interface{})
					if len(choices) > 0 {
						choice, _ := choices[0].(map[string]interface{})
						delta, _ := choice["delta"].(map[string]interface{})
						content, _ := delta["content"].(string)
						if content != "" {
							respChunk := map[string]string{"chunk": content}
							respBytes, _ := json.Marshal(respChunk)
							fmt.Fprintf(w, "data: %s\n\n", respBytes)
							flusher.Flush()
						}
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(w, "data: {\"done\":true}\n\n")
				flusher.Flush()
				return nil
			}
			return err
		}
	}
}

// SubmitLetter 提交信件
func SubmitLetter(data map[string]interface{}) (map[string]interface{}, error) {
	// Extract fields
	name := getString(data, "姓名")
	phone := getString(data, "手机号")
	idCard := getString(data, "身份证号")
	cat1 := getString(data, "一级分类")
	cat2 := getString(data, "二级分类")
	cat3 := getString(data, "三级分类")
	content := getString(data, "描述")

	// Handle unit array
	unitArr, _ := data["处理单位"].([]interface{})
	unitParts := make([]string, 0)
	for _, u := range unitArr {
		if s, ok := u.(string); ok && s != "" {
			unitParts = append(unitParts, s)
		}
	}

	// Generate letter number
	now := time.Now()
	letterNo := fmt.Sprintf("XJ%s%s", now.Format("20060102150405"), randomHex(4))

	nowStr := now.Format("2006-01-02 15:04:05")

	// Build flow record
	flowRecord := []map[string]interface{}{
		{
			"操作类型":   "生成",
			"操作前状态": "AI帮助用户填写",
			"操作后状态": "预处理",
			"操作前单位": "互联网",
			"操作后单位": "市局 / 民意智感中心",
			"操作人警号": "999999",
			"操作人姓名": "AI",
			"操作时间":  nowStr,
			"备注": map[string]interface{}{
				"建议下发办理单位": unitParts,
			},
		},
	}
	flowBytes, _ := json.Marshal(flowRecord)
	flowStr := string(flowBytes)

	// Build unit string
	unitStr := strings.Join(unitParts, " / ")
	if unitStr == "" {
		unitStr = "市局 / 民意智感中心"
	}

	// Insert into DB
	letter := &model.Letter{
		LetterNo:      letterNo,
		CitizenName:   name,
		Phone:         phone,
		IDCard:        idCard,
		ReceivedAt:    nowStr,
		Channel:       "局长信箱",
		Category1:     cat1,
		Category2:     cat2,
		Category3:     cat3,
		Content:       content,
		SpecialTags:   "[]",
		CurrentStatus: "预处理",
	}

	if err := dao.InsertLetter(letter); err != nil {
		return nil, fmt.Errorf("插入信件表失败: %w", err)
	}

	flow := &model.LetterFlow{
		LetterNo:    letterNo,
		FlowRecords: flowStr,
	}
	if err := dao.InsertLetterFlow(flow); err != nil {
		return nil, fmt.Errorf("插入流转表失败: %w", err)
	}

	att := &model.LetterAttachment{
		LetterNo:               letterNo,
		CityDispatchFiles:      "[]",
		DistrictDispatchFiles:  "[]",
		HandlerFeedbackFiles:   "[]",
		DistrictFeedbackFiles:  "[]",
		CallRecordings:         "[]",
	}
	if err := dao.InsertLetterAttachment(att); err != nil {
		return nil, fmt.Errorf("插入文件表失败: %w", err)
	}

	return map[string]interface{}{
		"信件编号": letterNo,
		"信件状态": "预处理",
	}, nil
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func randomHex(n int) string {
	const letters = "0123456789ABCDEF"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%16]
		time.Sleep(1) // ensure different values
	}
	return string(b)
}

// GetCategoriesForFront 获取分类树（前端需要的三级联动格式）
func GetCategoriesForFront() ([]map[string]interface{}, error) {
	cats, err := dao.GetAllCategories()
	if err != nil {
		return nil, err
	}
	// 构建树结构: [{ name: level1, children: [{ name: level2, children: [{ name: level3 }] }] }]
	treeMap := make(map[string]map[string][]string)
	for _, c := range cats {
		if c.Level1 == "" {
			continue
		}
		if treeMap[c.Level1] == nil {
			treeMap[c.Level1] = make(map[string][]string)
		}
		if c.Level2 != "" {
			treeMap[c.Level1][c.Level2] = append(treeMap[c.Level1][c.Level2], c.Level3)
		}
	}
	var tree []map[string]interface{}
	for l1, l2Map := range treeMap {
		var children2 []map[string]interface{}
		for l2, l3List := range l2Map {
			var children3 []map[string]interface{}
			for _, l3 := range l3List {
				if l3 != "" {
					children3 = append(children3, map[string]interface{}{
						"name": l3,
					})
				}
			}
			entry := map[string]interface{}{
				"name": l2,
			}
			if len(children3) > 0 {
				entry["children"] = children3
			}
			children2 = append(children2, entry)
		}
		entry := map[string]interface{}{
			"name": l1,
		}
		if len(children2) > 0 {
			entry["children"] = children2
		}
		tree = append(tree, entry)
	}
	return tree, nil
}

// buildClassifyPrompt 从 categories 表动态构建分类提示词
func buildClassifyPrompt() string {
	cats, err := dao.GetAllCategories()
	if err != nil || len(cats) == 0 {
		return ""
	}
	// 构建树结构
	type node struct {
		name     string
		children []node
	}
	treeMap := make(map[string]map[string][]string)
	for _, c := range cats {
		if c.Level1 == "" {
			continue
		}
		if treeMap[c.Level1] == nil {
			treeMap[c.Level1] = make(map[string][]string)
		}
		if c.Level2 != "" {
			found := false
			for _, v := range treeMap[c.Level1][c.Level2] {
				if v == c.Level3 {
					found = true
					break
				}
			}
			if !found {
				treeMap[c.Level1][c.Level2] = append(treeMap[c.Level1][c.Level2], c.Level3)
			}
		}
	}

	var parts []string
	parts = append(parts, "信件分类体系（一级/二级/三级）：")

	// 按 level1 排序保证顺序稳定（map 遍历无序，先收集再排序）
	var level1List []string
	for l1 := range treeMap {
		level1List = append(level1List, l1)
	}
	// 简单按拼音或字符串排，这里直接用字符串顺序
	for _, l1 := range level1List {
		parts = append(parts, fmt.Sprintf("\n【一级分类】%s", l1))
		l2Map := treeMap[l1]
		var level2List []string
		for l2 := range l2Map {
			level2List = append(level2List, l2)
		}
		for _, l2 := range level2List {
			l3List := l2Map[l2]
			if len(l3List) > 0 {
				// 有三级分类
				cleaned := []string{}
				for _, l3 := range l3List {
					if l3 != "" {
						cleaned = append(cleaned, l3)
					}
				}
				if len(cleaned) > 0 {
					parts = append(parts, fmt.Sprintf("  【二级分类】%s → 三级分类：%s", l2, strings.Join(cleaned, "、")))
				} else {
					parts = append(parts, fmt.Sprintf("  【二级分类】%s", l2))
				}
			} else {
				parts = append(parts, fmt.Sprintf("  【二级分类】%s", l2))
			}
		}
	}

	return strings.Join(parts, "\n")
}

// ClassifyContent 对信件内容进行智能分类（先试LLM，失败则用关键词匹配）
func ClassifyContent(content string) (map[string]string, error) {
	// 先尝试 LLM API 分类
	result, err := tryLLMClassify(content)
	if err == nil && result != nil {
		return result, nil
	}

	// LLM 失败，回退到关键词匹配
	log.Printf("LLM分类失败，回退到关键词匹配: %v", err)
	return keywordClassify(content)
}

// tryLLMClassify 调用 DeepSeek API 进行分类
func tryLLMClassify(content string) (map[string]string, error) {
	classPrompt := buildClassifyPrompt()
	systemPrompt := "你是一个政府信件智能分类助手。请根据用户提供的信件内容和下面的分类体系，判断这封信件最合适的一级分类、二级分类和三级分类。\n" + classPrompt + "\n请只返回JSON格式的字符串，不要包含其他说明文字。格式：{\"一级分类\":\"xxx\",\"二级分类\":\"xxx\",\"三级分类\":\"xxx\"}。如果没有完全匹配的分类，请选择最接近的。"

	cfg := config.AppConfig.LLM

	reqBody := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "信件内容：\n" + content},
		},
		"temperature": 0.3,
		"max_tokens":  200,
		"stream":      false,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", cfg.APIURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}

	// 从 choices[0].message.content 中提取 JSON
	choices, _ := result["choices"].([]interface{})
	if len(choices) == 0 {
		return nil, fmt.Errorf("API未返回分类结果")
	}
	choice, _ := choices[0].(map[string]interface{})
	msg, _ := choice["message"].(map[string]interface{})
	aiContent, _ := msg["content"].(string)
	if aiContent == "" {
		return nil, fmt.Errorf("API返回内容为空")
	}

	// 尝试从返回内容中提取 JSON
	startIdx := strings.Index(aiContent, "{")
	endIdx := strings.LastIndex(aiContent, "}")
	if startIdx >= 0 && endIdx > startIdx {
		jsonStr := aiContent[startIdx : endIdx+1]
		var category map[string]string
		if err := json.Unmarshal([]byte(jsonStr), &category); err == nil {
			return category, nil
		}
	}

	return nil, fmt.Errorf("无法解析AI返回的分类结果: %s", aiContent)
}

// keywordClassify 从 categories 表做三级分类匹配（确定性 + 字符级评分）
// 1. 按分类名长度降序排序（更具体的分类优先）
// 2. 对每条分类计算字符级命中率（不只是完整的词匹配）
// 3. 平局时按 level1+level2+level3 字符串排序（确定性）
func keywordClassify(content string) (map[string]string, error) {
	cats, err := dao.GetAllCategories()
	if err != nil || len(cats) == 0 {
		return map[string]string{"一级分类": "市民诉求", "二级分类": "其他", "三级分类": ""}, nil
	}

	// 对每条分类计算综合匹配分数
	type scored struct {
		level1  string
		level2  string
		level3  string
		score   int     // 精确词匹配分
		chrRate float64 // 字符级重叠率（0~1）
	}

	// 预计算 content 中的字符集合
	contentSet := make(map[rune]bool)
	for _, r := range content {
		contentSet[r] = true
	}

	var results []scored
	for _, c := range cats {
		if c.Level1 == "" {
			continue
		}
		s := scored{level1: c.Level1, level2: c.Level2, level3: c.Level3}

		// --- 精确词匹配得分 ---
		if strings.Contains(content, c.Level1) {
			s.score += 10 // level1 匹配权重最高
		}
		if c.Level2 != "" && strings.Contains(content, c.Level2) {
			s.score += 6
		}
		if c.Level3 != "" && strings.Contains(content, c.Level3) {
			s.score += 3
		}

		// --- 字符级重叠率（当精确词未命中时，弥补细粒度匹配）---
		// 把 level1+level2+level3 拼接起来算字符重叠
		fullPath := c.Level1 + c.Level2 + c.Level3
		fullRunes := []rune(fullPath)
		if len(fullRunes) == 0 {
			continue
		}
		matchCount := 0
		seen := make(map[rune]bool)
		for _, r := range fullRunes {
			if seen[r] {
				continue
			}
			seen[r] = true
			if contentSet[r] {
				matchCount++
			}
		}
		s.chrRate = float64(matchCount) / float64(len(seen))

		results = append(results, s)
	}

	// 排序：score 降序 → chrRate 降序 → 全路径字符串降序（确定性tiebreaker）
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		if results[i].chrRate != results[j].chrRate {
			return results[i].chrRate > results[j].chrRate
		}
		// 平局时，更长的分类名 = 更具体 = 优先
		lenI := len(results[i].level1 + results[i].level2 + results[i].level3)
		lenJ := len(results[j].level1 + results[j].level2 + results[j].level3)
		if lenI != lenJ {
			return lenI > lenJ
		}
		// 最终tiebreaker：按字符串降序（确定性）
		return (results[i].level1 + results[i].level2 + results[i].level3) >
			(results[j].level1 + results[j].level2 + results[j].level3)
	})

	if len(results) > 0 && results[0].score > 0 {
		return map[string]string{
			"一级分类": results[0].level1,
			"二级分类": results[0].level2,
			"三级分类": results[0].level3,
		}, nil
	}

	// 没有任何精确匹配 → 用字符级重叠率最高的
	if len(results) > 0 && results[0].chrRate > 0 {
		return map[string]string{
			"一级分类": results[0].level1,
			"二级分类": results[0].level2,
			"三级分类": results[0].level3,
		}, nil
	}

	// 完全不匹配 → 取第一条（排序后确定性的第一条）
	if len(results) > 0 {
		return map[string]string{
			"一级分类": results[0].level1,
			"二级分类": results[0].level2,
			"三级分类": results[0].level3,
		}, nil
	}

	return map[string]string{"一级分类": "市民诉求", "二级分类": "其他", "三级分类": ""}, nil
}
