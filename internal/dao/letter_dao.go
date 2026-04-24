package dao

import (
	"people-page-backend/internal/model"
)

// GetPromptByType 获取指定类型的提示词
func GetPromptByType(promptType string) (*model.Prompt, error) {
	prompt := &model.Prompt{}
	err := DB.QueryRow(
		"SELECT id, 类型, 内容, created_at, updated_at FROM 提示词 WHERE 类型 = ? LIMIT 1",
		promptType,
	).Scan(&prompt.ID, &prompt.Type, &prompt.Content, &prompt.CreatedAt, &prompt.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return prompt, nil
}

// GetAllUnits 获取所有单位信息
func GetAllUnits() ([]model.Unit, error) {
	rows, err := DB.Query("SELECT id, 一级, 二级, 三级, 系统编码 FROM 单位 ORDER BY 一级, 二级, 三级")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []model.Unit
	for rows.Next() {
		var u model.Unit
		if err := rows.Scan(&u.ID, &u.Level1, &u.Level2, &u.Level3, &u.SysCode); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, rows.Err()
}

// GetDistinctClassifications 获取所有不同的分类（去重）
func GetDistinctClassifications() ([]model.Letter, error) {
	rows, err := DB.Query(`
		SELECT DISTINCT 信件一级分类, 信件二级分类, 信件三级分类 
		FROM 信件表 
		WHERE 信件一级分类 IS NOT NULL AND 信件一级分类 != ''
		ORDER BY 信件一级分类, 信件二级分类, 信件三级分类
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
		INSERT INTO 信件表 (
			信件编号, 群众姓名, 手机号, 身份证号, 来信时间, 来信渠道,
			信件一级分类, 信件二级分类, 信件三级分类, 诉求内容, 专项关注标签,
			当前信件处理单位, 当前信件状态
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		letter.LetterNo, letter.CitizenName, letter.Phone, letter.IDCard,
		letter.ReceivedAt, letter.Channel,
		letter.Category1, letter.Category2, letter.Category3, letter.Content,
		letter.SpecialTags, letter.CurrentUnit, letter.CurrentStatus,
	)
	return err
}

// InsertLetterFlow 插入流转表记录
func InsertLetterFlow(flow *model.LetterFlow) error {
	_, err := DB.Exec(
		"INSERT INTO 流转表 (信件编号, 流转记录) VALUES (?, ?)",
		flow.LetterNo, flow.FlowRecords,
	)
	return err
}

// InsertLetterAttachment 插入文件表记录
func InsertLetterAttachment(att *model.LetterAttachment) error {
	_, err := DB.Exec(`
		INSERT INTO 文件表 (信件编号, 市局下发附件, 区县局下发附件, 办案单位反馈附件, 区县局反馈附件, 通话录音附件)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		att.LetterNo, att.CityDispatchAttach, att.CountyDispatchAttach,
		att.UnitFeedbackAttach, att.CountyFeedbackAttach, att.CallRecordAttach,
	)
	return err
}
