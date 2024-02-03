package system

import (
	"ding/global"
	modelRequest "ding/model/common/request"
	"errors"
	"gorm.io/gorm"
	"time"
)

var ErrRoleExistence = errors.New("存在相同角色，请检查角色名称")

type SysAuthority struct {
	CreatedAt     time.Time     // 创建时间
	UpdatedAt     time.Time     // 更新时间
	DeletedAt     *time.Time    `sql:"index"`
	AuthorityId   uint          `json:"authorityId" gorm:"not null;unique;primary_key;comment:角色ID;size:90"` // 角色ID
	AuthorityName string        `json:"authorityName" gorm:"comment:角色名"`                                    // 角色名 	// 父角色ID
	SysBaseMenus  []SysBaseMenu `json:"menus" gorm:"many2many:sys_authority_menus;"`
	//DingUsers     []dingding.DingUser `json:"-" gorm:"many2many:sys_user_authority;"`
	DefaultRouter string `json:"defaultRouter" gorm:"comment:默认菜单;default:dashboard"` // 默认菜单(默认dashboard)
}

func DefaultMenu() []SysBaseMenu {
	return []SysBaseMenu{{
		Model:     gorm.Model{ID: 1},
		ParentId:  "0",
		Path:      "dashboard",
		Name:      "dashboard",
		Component: "view/dashboard/index.vue",
		Sort:      1,
		Meta: Meta{
			Title: "仪表盘",
			Icon:  "setting",
		},
	}}
}

func (auth *SysAuthority) CreateAuthority() (*SysAuthority, error) {

	if err := global.GLOAB_DB.Where("authority_name = ?", auth.AuthorityName).First(auth).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		return auth, ErrRoleExistence
	}

	e := global.GLOAB_DB.Transaction(func(tx *gorm.DB) (err error) {

		if err = tx.Create(&auth).Error; err != nil {
			return err
		}

		auth.SysBaseMenus = DefaultMenu()
		if err = tx.Model(&auth).Association("SysBaseMenus").Replace(&auth.SysBaseMenus); err != nil {
			return err
		}

		return
	})

	return auth, e
}

func (sysAuthority *SysAuthority) DeleteAuthority(auth *SysAuthority) error {
	if errors.Is(global.GLOAB_DB.Debug().First(&auth).Error, gorm.ErrRecordNotFound) {
		return errors.New("该角色不存在")
	}
	var count int64
	err := global.GLOAB_DB.Table("sys_user_authority").Where("sys_authority_authority_id = ?", auth.AuthorityId).Count(&count).Error
	if count != 0 {
		return errors.New("此角色有用户正在使用禁止删除")
	}
	if err != nil {
		return err
	}
	err = global.GLOAB_DB.Table("ding_users").Where("authority_id = ?", auth.AuthorityId).Count(&count).Error
	if count != 0 {
		return errors.New("此角色有用户正在使用禁止删除")
	}
	if err != nil {
		return err
	}
	return global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		var err error
		if err = tx.Preload("SysBaseMenus").Where("authority_id = ?", auth.AuthorityId).First(auth).Unscoped().Delete(auth).Error; err != nil {
			return err
		}

		if len(auth.SysBaseMenus) > 0 {
			if err = tx.Model(auth).Association("SysBaseMenus").Delete(auth.SysBaseMenus); err != nil {
				return err
			}
			// err = db.Association("SysBaseMenus").Delete(&auth)
		}

		if err = tx.Delete(&SysUserAuthority{}, "sys_authority_authority_id = ?", auth.AuthorityId).Error; err != nil {
			return err
		}
		if err = tx.Where("authority_id = ?", auth.AuthorityId).Delete(&[]SysAuthorityBtn{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (sysAuthority *SysAuthority) UpdateAuthority(auth SysAuthority) (authority SysAuthority, err error) {
	err = global.GLOAB_DB.Where("authority_id = ?", auth.AuthorityId).First(&SysAuthority{}).Updates(&auth).Error
	return auth, err
}
func (sysAuthority *SysAuthority) GetAuthorityInfoList(info modelRequest.PageInfo) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GLOAB_DB.Model(&SysAuthority{})
	if err = db.Count(&total).Error; total == 0 || err != nil {
		return
	}
	var authority []SysAuthority
	err = db.Limit(limit).Offset(offset).Find(&authority).Error
	return authority, total, err
}

func (SysAuthority *SysAuthority) SetUserAuthorities(id string, authorityIds []uint) (err error) {
	return global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		TxErr := tx.Delete(&[]SysUserAuthority{}, "ding_user_user_id = ?", id).Error
		if TxErr != nil {
			return TxErr
		}
		var useAuthority []SysUserAuthority
		for _, v := range authorityIds {
			useAuthority = append(useAuthority, SysUserAuthority{
				DingUserUserId: id, SysAuthorityAuthorityId: v,
			})
		}
		TxErr = tx.Create(&useAuthority).Error
		if TxErr != nil {
			return TxErr
		}
		TxErr = tx.Table("ding_users").Where("user_id = ?", id).Update("authority_id", authorityIds[0]).Error
		if TxErr != nil {
			return TxErr
		}
		// 返回 nil 提交事务
		return nil
	})
}
