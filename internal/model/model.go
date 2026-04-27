package model

// Letter 民意诉求信件主表 (对应 letters 表)
type Letter struct {
	ID             int    `json:"id" db:"id"`
	LetterNo       string `json:"letter_no" db:"letter_no"`
	CitizenName    string `json:"citizen_name" db:"citizen_name"`
	Phone          string `json:"phone" db:"phone"`
	IDCard         string `json:"id_card" db:"id_card"`
	ReceivedAt     string `json:"received_at" db:"received_at"`
	Channel        string `json:"channel" db:"channel"`
	Category1      string `json:"category_l1" db:"category_l1"`
	Category2      string `json:"category_l2" db:"category_l2"`
	Category3      string `json:"category_l3" db:"category_l3"`
	Content        string `json:"content" db:"content"`
	SpecialTags    string `json:"special_tags" db:"special_tags"`           // JSON array
	CurrentUnit    string `json:"current_unit" db:"current_unit"`
	CurrentStatus  string `json:"current_status" db:"current_status"`
}

// LetterFlow 信件流转记录 (对应 letter_flows 表)
type LetterFlow struct {
	ID          int    `json:"id" db:"id"`
	LetterNo    string `json:"letter_no" db:"letter_no"`
	FlowRecords string `json:"flow_records" db:"flow_records"` // JSON array
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

// LetterAttachment 信件附件 (对应 letter_attachments 表)
type LetterAttachment struct {
	ID                     int    `json:"id" db:"id"`
	LetterNo               string `json:"letter_no" db:"letter_no"`
	CityDispatchFiles      string `json:"city_dispatch_files" db:"city_dispatch_files"`         // JSON
	DistrictDispatchFiles  string `json:"district_dispatch_files" db:"district_dispatch_files"` // JSON
	HandlerFeedbackFiles   string `json:"handler_feedback_files" db:"handler_feedback_files"`   // JSON
	DistrictFeedbackFiles  string `json:"district_feedback_files" db:"district_feedback_files"` // JSON
	CallRecordings         string `json:"call_recordings" db:"call_recordings"`                 // JSON
}

// Unit 组织机构 (对应 units 表)
type Unit struct {
	ID         int    `json:"id" db:"id"`
	Level1     string `json:"level1" db:"level1"`
	Level2     string `json:"level2" db:"level2"`
	Level3     string `json:"level3" db:"level3"`
	SystemCode string `json:"system_code" db:"system_code"`
}

// Prompt 提示词 (对应 prompts 表)
type Prompt struct {
	ID          int    `json:"id" db:"id"`
	PromptType  string `json:"prompt_type" db:"prompt_type"`
	Content     string `json:"content" db:"content"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

// PoliceUser (对应 police_users 表)
type PoliceUser struct {
	ID              int    `json:"id" db:"id"`
	Password        string `json:"password" db:"password"`
	Name            string `json:"name" db:"name"`
	PermissionLevel string `json:"permission_level" db:"permission_level"`
	UnitName        string `json:"unit_name" db:"unit_name"`
	Phone           string `json:"phone" db:"phone"`
	IsActive        int    `json:"is_active" db:"is_active"`
	CreatedAt       string `json:"created_at" db:"created_at"`
	LastLogin       string `json:"last_login" db:"last_login"`
	Nickname        string `json:"nickname" db:"nickname"`
	PoliceNumber    string `json:"police_number" db:"police_number"`
}

// UserSession (对应 user_sessions 表)
type UserSession struct {
	ID         int    `json:"id" db:"id"`
	SessionKey string `json:"session_key" db:"session_key"`
	IPAddress  string `json:"ip_address" db:"ip_address"`
	UserAgent  string `json:"user_agent" db:"user_agent"`
	CreatedAt  string `json:"created_at" db:"created_at"`
	ExpiresAt  string `json:"expires_at" db:"expires_at"`
	UserID     int    `json:"user_id" db:"user_id"`
}

// Category 信件分类 (对应 categories 表)
type Category struct {
	ID     int    `json:"id" db:"id"`
	Level1 string `json:"level1" db:"level1"`
	Level2 string `json:"level2" db:"level2"`
	Level3 string `json:"level3" db:"level3"`
}
