package system

// SysUserAuthority 是 sysUser 和 sysAuthority 的连接表
type SysUserAuthority struct {
	DingUserUserId          string `gorm:"column:ding_user_user_id"` // 两部分，前面的部分DingUser是表名，后面的UserId是表中的主键
	SysAuthorityAuthorityId uint   `gorm:"column:sys_authority_authority_id"`
}

func (s *SysUserAuthority) TableName() string {
	return "sys_user_authority"
}
