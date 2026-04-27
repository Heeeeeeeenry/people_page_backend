package service

import (
	"encoding/json"
	"fmt"
	"time"
	"people-page-backend/internal/dao"
	"people-page-backend/internal/model"
)

// SubmitLetterCitizen 市民提交信件（简化版，无分类/单位选择）
func SubmitLetterCitizen(data map[string]interface{}) (map[string]interface{}, error) {
	// 提取字段
	name := getString(data, "姓名")
	phone := getString(data, "手机号")
	idCard := getString(data, "身份证号")
	content := getString(data, "描述")

	// 位置信息（从高德地图获取）
	location := getString(data, "location")     // 经纬度 "lng,lat"
	address := getString(data, "address")       // 详细地址
	province := getString(data, "province")     // 省
	city := getString(data, "city")             // 市
	district := getString(data, "district")     // 区

	// 优先使用登录用户的手机号
	if loginPhone := getString(data, "登录手机号"); loginPhone != "" && phone == "" {
		phone = loginPhone
	}

	if name == "" {
		return nil, fmt.Errorf("请填写群众姓名")
	}
	if phone == "" {
		return nil, fmt.Errorf("请填写手机号")
	}
	if content == "" {
		return nil, fmt.Errorf("请填写诉求描述")
	}

	// 生成信件编号
	now := time.Now()
	letterNo := fmt.Sprintf("SM%s%s", now.Format("20060102150405"), randomHex(4))

	nowStr := now.Format("2006-01-02 15:04:05")

	// 构建流转记录 - 市民上报
	flowRecord := []map[string]interface{}{
		{
			"操作类型":   "市民上报",
			"操作前状态": "无",
			"操作后状态": "预处理",
			"操作前单位": "",
			"操作后单位": "市局 / 民意智感中心",
			"操作人":   name,
			"操作人角色": "市民",
			"操作时间":  nowStr,
			"备注": map[string]interface{}{
				"位置":    location,
				"地址":    address,
				"来源渠道": "市民上报",
			},
		},
	}
	flowBytes, _ := json.Marshal(flowRecord)
	flowStr := string(flowBytes)

	// 信件分类使用默认值 - 后续由AI或管理员处理
	defaultCat1 := getString(data, "一级分类")
	defaultCat2 := getString(data, "二级分类")
	defaultCat3 := getString(data, "三级分类")
	if defaultCat1 == "" {
		defaultCat1 = "市民诉求"
		defaultCat2 = "其他"
		defaultCat3 = ""
	}

	// 构建地址信息（存入备注或用其他字段，目前存入诉求内容尾部作为补充）
	addressInfo := ""
	if address != "" {
		addressInfo = fmt.Sprintf("\n[位置信息] 地址：%s", address)
		if province != "" {
			addressInfo += fmt.Sprintf("；%s %s %s", province, city, district)
		}
		if location != "" {
			addressInfo += fmt.Sprintf("（%s）", location)
		}
	}

	fullContent := content + addressInfo

	// 插入DB
	letter := &model.Letter{
		LetterNo:      letterNo,
		CitizenName:   name,
		Phone:         phone,
		IDCard:        idCard,
		ReceivedAt:    nowStr,
		Channel:       "市民上报",
		Category1:     defaultCat1,
		Category2:     defaultCat2,
		Category3:     defaultCat3,
		Content:       fullContent,
		SpecialTags:   "[]",
		CurrentUnit:   "市局 / 民意智感中心",
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

// GetCitizenLetters 获取市民自己的信件列表
func GetCitizenLetters(phone string) ([]map[string]interface{}, error) {
	rows, err := dao.DB.Query(`
		SELECT letter_no, citizen_name, received_at, current_status, content
		FROM letters
		WHERE phone = ?
		ORDER BY received_at DESC
	`, phone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var letters []map[string]interface{}
	for rows.Next() {
		var no, name, timeStr, status, content string
		if err := rows.Scan(&no, &name, &timeStr, &status, &content); err != nil {
			continue
		}
		letters = append(letters, map[string]interface{}{
			"信件编号": no,
			"群众姓名": name,
			"来信时间": timeStr,
			"当前状态": status,
			"诉求内容": truncateContent(content, 100),
		})
	}
	return letters, nil
}

func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return content
}
