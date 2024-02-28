package request

type ParamSetUserAuthorities struct {
	UserId       string `json:"userId"`
	AuthorityIds []uint `json:"authorityIds"` // 角色ID
}
