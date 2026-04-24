package model

// 信件表 - 民意诉求信件主表
type Letter struct {
	ID                     int    `json:"序号" db:"序号"`
	LetterNo              string `json:"信件编号" db:"信件编号"`
	CitizenName           string `json:"群众姓名" db:"群众姓名"`
	Phone                 string `json:"手机号" db:"手机号"`
	IDCard                string `json:"身份证号" db:"身份证号"`
	ReceivedAt            string `json:"来信时间" db:"来信时间"`
	Channel               string `json:"来信渠道" db:"来信渠道"`
	Category1             string `json:"信件一级分类" db:"信件一级分类"`
	Category2             string `json:"信件二级分类" db:"信件二级分类"`
	Category3             string `json:"信件三级分类" db:"信件三级分类"`
	Content               string `json:"诉求内容" db:"诉求内容"`
	SpecialTags           string `json:"专项关注标签" db:"专项关注标签"` // JSON array
	CurrentUnit           string `json:"当前信件处理单位" db:"当前信件处理单位"`
	CurrentStatus         string `json:"当前信件状态" db:"当前信件状态"`
}

// 流转表 - 信件流转记录
type LetterFlow struct {
	ID          int    `json:"序号" db:"序号"`
	LetterNo    string `json:"信件编号" db:"信件编号"`
	FlowRecords string `json:"流转记录" db:"流转记录"` // JSON array
	CreatedAt   string `json:"创建时间" db:"创建时间"`
	UpdatedAt   string `json:"更新时间" db:"更新时间"`
}

// 文件表 - 信件附件
type LetterAttachment struct {
	ID                  int    `json:"序号" db:"序号"`
	LetterNo            string `json:"信件编号" db:"信件编号"`
	CityDispatchAttach  string `json:"市局下发附件" db:"市局下发附件"`   // JSON
	CountyDispatchAttach string `json:"区县局下发附件" db:"区县局下发附件"` // JSON
	UnitFeedbackAttach  string `json:"办案单位反馈附件" db:"办案单位反馈附件"` // JSON
	CountyFeedbackAttach string `json:"区县局反馈附件" db:"区县局反馈附件"` // JSON
	CallRecordAttach    string `json:"通话录音附件" db:"通话录音附件"`  // JSON
}

// 单位表 - 组织机构
type Unit struct {
	ID       int    `json:"id" db:"id"`
	Level1   string `json:"一级" db:"一级"`
	Level2   string `json:"二级" db:"二级"`
	Level3   string `json:"三级" db:"三级"`
	SysCode  string `json:"系统编码" db:"系统编码"`
}

// 提示词表
type Prompt struct {
	ID        int    `json:"id" db:"id"`
	Type      string `json:"类型" db:"类型"`
	Content   string `json:"内容" db:"内容"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

// police_users 表
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

// user_sessions 表
type UserSession struct {
	ID         int    `json:"id" db:"id"`
	SessionKey string `json:"session_key" db:"session_key"`
	IPAddress  string `json:"ip_address" db:"ip_address"`
	UserAgent  string `json:"user_agent" db:"user_agent"`
	CreatedAt  string `json:"created_at" db:"created_at"`
	ExpiresAt  string `json:"expires_at" db:"expires_at"`
	UserID     int    `json:"user_id" db:"user_id"`
}
