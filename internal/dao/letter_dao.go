package dao

import (
	"people-page-backend/internal/model"
)

// GetPromptByType 获取指定类型的提示词
func GetPromptByType(promptType string) (*model.Prompt, error) {
	prompt := &model.Prompt{}
	err := DB.QueryRow(
		"SELECT id, prompt_type, content, created_at, updated_at FROM prompts WHERE prompt_type = ? LIMIT 1",
		promptType,
	).Scan(&prompt.ID, &prompt.PromptType, &prompt.Content, &prompt.CreatedAt, &prompt.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return prompt, nil
}

// GetAllUnits 获取所有单位信息
func GetAllUnits() ([]model.Unit, error) {
	rows, err := DB.Query("SELECT id, level1, level2, level3, system_code FROM units ORDER BY level1, level2, level3")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []model.Unit
	for rows.Next() {
		var u model.Unit
		if err := rows.Scan(&u.ID, &u.Level1, &u.Level2, &u.Level3, &u.SystemCode); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, rows.Err()
}

// GetDistinctClassifications 获取所有不同的分类（去重）
func GetDistinctClassifications() ([]model.Letter, error) {
	rows, err := DB.Query(`
		SELECT DISTINCT category_l1, category_l2, category_l3 
		FROM letters 
		WHERE category_l1 IS NOT NULL AND category_l1 != ''
		ORDER BY category_l1, category_l2, category_l3
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classifications []model.Letter
	for rows.Next() {
		var l model.Letter
		if err := rows.Scan(&l.Category1, &l.Category2, &l.Category3); err != nil {
			return nil, err
		}
		classifications = append(classifications, l)
	}
	return classifications, rows.Err()
}

// InsertLetter 插入信件表记录
func InsertLetter(letter *model.Letter) error {
	_, err := DB.Exec(`
		INSERT INTO letters (
			letter_no, citizen_name, phone, id_card, received_at, channel,
			category_l1, category_l2, category_l3, content, special_tags,
			current_status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		letter.LetterNo, letter.CitizenName, letter.Phone, letter.IDCard,
		letter.ReceivedAt, letter.Channel,
		letter.Category1, letter.Category2, letter.Category3, letter.Content,
		letter.SpecialTags, letter.CurrentStatus,
	)
	return err
}

// InsertLetterFlow 插入流转表记录
func InsertLetterFlow(flow *model.LetterFlow) error {
	_, err := DB.Exec(
		"INSERT INTO letter_flows (letter_no, flow_records) VALUES (?, ?)",
		flow.LetterNo, flow.FlowRecords,
	)
	return err
}

// InsertLetterAttachment 插入文件表记录
func InsertLetterAttachment(att *model.LetterAttachment) error {
	_, err := DB.Exec(`
		INSERT INTO letter_attachments (letter_no, city_dispatch_files, district_dispatch_files, handler_feedback_files, district_feedback_files, call_recordings)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		att.LetterNo, att.CityDispatchFiles, att.DistrictDispatchFiles,
		att.HandlerFeedbackFiles, att.DistrictFeedbackFiles, att.CallRecordings,
	)
	return err
}

// GetAllCategories 获取所有分类（含三级）
func GetAllCategories() ([]model.Category, error) {
	rows, err := DB.Query("SELECT id, level1, level2, level3 FROM categories ORDER BY level1, level2, level3")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Level1, &c.Level2, &c.Level3); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}
