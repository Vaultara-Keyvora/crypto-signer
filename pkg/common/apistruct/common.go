package apistruct

const (
	LBE_SIGN     = "lbeSign"
	LBE_TOKEN    = "lbeToken"
	LBE_JA3      = "lbeJa3"
	LBE_SESSION  = "lbeSession"
	LBE_IDENTITY = "lbeIdentity"
)

type RequestPagination struct {
	PageNumber int32 `json:"pageNumber"` // 页数
	ShowNumber int32 `json:"showNumber"` // 页大小
}

type MessageSeq struct {
	StartSeq int32 `json:"startSeq"` // 开始消息序列号
	EndSeq   int32 `json:"endSeq"`   // 结束消息序列号
}

func (x *RequestPagination) GetPageNumber() int32 {
	return x.PageNumber
}

func (x *RequestPagination) GetShowNumber() int32 {
	if x.ShowNumber > 1000 {
		return 1000
	}
	return x.ShowNumber
}

type UserInfo struct {
	Account     string   `json:"account"`            // 账号
	Password    string   `json:"password,omitempty"` // 密码
	IdentityID  string   `json:"identityID"`         // 租户id
	Username    string   `json:"username"`           // 名称
	Nickname    string   `json:"nickname"`           // 对外展示名称
	UserID      string   `json:"userID"`             // 用户ID
	PhoneNumber string   `json:"phoneNumber"`        // 手机号码
	AreaCode    string   `json:"areaCode"`           // 手机区号
	FaceUrl     string   `json:"faceUrl"`            // 头像地址
	Role        RoleInfo `json:"role"`               // 角色
	Group       string   `json:"group"`              // 客服组
	Email       string   `json:"email"`              // 邮箱地址
	Status      uint     `json:"status"`             // 0-正常 1-封号
	MangerLevel uint     `json:"mangerLevel"`        // 1-admin 2-普通用户
	CreateTime  int64    `json:"createTime"`         // 创建时间(毫秒)
}

type PermissionInfo struct {
	ID string `json:"id"`
}

type PermissionDetail struct {
	ID             string `json:"id"`                     // 唯一id
	ParentID       string `json:"parentID"`               // 父类id
	PermissionName string `json:"name"`                   // 名称
	PermissionType uint   `json:"permissionType"`         // 类型 1-模块 2-页面 3-功能 4-数据
	Router         string `json:"router,omitempty"`       // 后端路由
	ClientRouter   string `json:"clientRouter,omitempty"` // 前端路由
	FieldName      string `json:"fieldName,omitempty"`    // 字段名
	Icon           string `json:"icon"`                   // icon地址
	Status         uint   `json:"status,omitempty"`       // 0-启用 1-禁用
	CreateTime     int64  `json:"createTime,omitempty"`   // 毫秒时间戳
}

type FieldInfo struct {
	PagePermissionID string   `json:"pagePermissionID"` // 所属页面权限ID
	AccessLevel      uint     `json:"accessLevel"`      // 访问等级 1-全部可见 2-部分数据不可见
	FieldIDs         []string `json:"fieldIds"`         // 部分数据不可见时: [ 字段权限ID ]
}

// PermissionTreeNode 表示树节点
type PermissionTreeNode struct {
	ID             string               `json:"id"`                       // 唯一id
	PermissionName string               `json:"name"`                     // 名称
	PermissionType uint                 `json:"permissionType,omitempty"` // 类型 1-模块 2-页面 3-功能 4-数据
	FieldName      string               `json:"fieldName,omitempty"`      // 字段名
	IsOption       bool                 `json:"isOption"`                 // 是否勾选
	AccessLevel    uint                 `json:"accessLevel,omitempty"`    // 访问等级 1-全部可见 2-部分数据不可见
	Children       []PermissionTreeNode `json:"children"`
}

type RoleInfo struct {
	ID              string               `json:"id"`                        // 角色id
	RoleName        string               `json:"name"`                      // 角色名
	Describe        string               `json:"describe"`                  // 描述
	BasePermissions []PermissionTreeNode `json:"basePermissions,omitempty"` // 基础权限
	DataPermissions []PermissionTreeNode `json:"dataPermissions,omitempty"` // 数据权限
}

type GroupInfo struct {
	GroupID           int32              `json:"groupID"`           // 群id
	GroupName         string             `json:"groupName"`         // 群名称
	FaceUrl           string             `json:"faceUrl"`           // 头像地址
	TopFewMemberInfos []*GroupMemberInfo `json:"topFewMemberInfos"` // 群成员信息(top 5)
	MemberCount       int                `json:"memberCount"`       // 群成员数量
	Status            uint               `json:"status"`            // 0-正常 1-封禁
	CreateTime        int64              `json:"createTime"`        // 创建时间(ms)
}

type GroupMemberInfo struct {
	UserID   string `json:"userID"`   // 用户id
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
	FaceUrl  string `json:"faceUrl"`  // 头像地址
}

type GroupInfoForSet struct {
	GroupID       int32    `json:"groupID"`       // 唯一id
	GroupName     string   `json:"groupName"`     // 组名
	FaceUrl       string   `json:"faceUrl"`       // 头像地址
	MemberUserIDs []string `json:"memberUserIDs"` // 成员id
}
