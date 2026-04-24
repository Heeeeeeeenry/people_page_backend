package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		CurrentUnit:   unitStr,
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
		LetterNo:             letterNo,
		CityDispatchAttach:   "[]",
		CountyDispatchAttach: "[]",
		UnitFeedbackAttach:   "[]",
		CountyFeedbackAttach: "[]",
		CallRecordAttach:     "[]",
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
