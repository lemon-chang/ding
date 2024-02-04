package system

import (
	"ding/global"
	"ding/model/common/request"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

type SysBaseMenu struct {
	gorm.Model
	MenuLevel      uint                                       `json:"-"`
	ParentId       string                                     `json:"parentId" gorm:"comment:父菜单ID"`     // 父菜单ID
	Path           string                                     `json:"path" gorm:"comment:路由path"`        // 路由path
	Name           string                                     `json:"name" gorm:"comment:路由name"`        // 路由name
	Hidden         bool                                       `json:"hidden" gorm:"comment:是否在列表隐藏"`     // 是否在列表隐藏
	Component      string                                     `json:"component" gorm:"comment:对应前端文件路径"` // 对应前端文件路径
	Sort           int                                        `json:"sort" gorm:"comment:排序标记"`          // 排序标记
	Meta           `json:"meta" gorm:"embedded;comment:附加属性"` // 附加属性
	SysAuthorities []SysAuthority                             `json:"SysAuthorities" gorm:"many2many:sys_authority_menus;"`
	Children       []SysBaseMenu                              `json:"children" gorm:"-"`
	Parameters     []SysBaseMenuParameter                     `json:"parameters"`
	MenuBtn        []SysBaseMenuBtn                           `json:"menuBtn"`
}
type Meta struct {
	ActiveName  string `json:"activeName" gorm:"comment:高亮菜单"`
	KeepAlive   bool   `json:"keepAlive" gorm:"comment:是否缓存"`           // 是否缓存
	DefaultMenu bool   `json:"defaultMenu" gorm:"comment:是否是基础路由（开发中）"` // 是否是基础路由（开发中）
	Title       string `json:"title" gorm:"comment:菜单名"`                // 菜单名
	Icon        string `json:"icon" gorm:"comment:菜单图标"`                // 菜单图标
	CloseTab    bool   `json:"closeTab" gorm:"comment:自动关闭tab"`         // 自动关闭tab
}
type SysBaseMenuParameter struct {
	gorm.Model
	SysBaseMenuID uint
	Type          string `json:"type" gorm:"comment:地址栏携带参数为params还是query"` // 地址栏携带参数为params还是query
	Key           string `json:"key" gorm:"comment:地址栏携带参数的key"`            // 地址栏携带参数的key
	Value         string `json:"value" gorm:"comment:地址栏携带参数的值"`            // 地址栏携带参数的值
}

func (m *SysBaseMenu) GetMenuTree(authorityId uint) (menus []SysMenu, err error) {
	menuTree, err := m.getMenuTreeMap(authorityId)
	menus = menuTree["0"]
	for i := 0; i < len(menus); i++ {
		err = m.getChildrenList(&menus[i], menuTree)
	}
	return menus, err
}
func (m *SysBaseMenu) getChildrenList(menu *SysMenu, treeMap map[string][]SysMenu) (err error) {
	menu.Children = treeMap[menu.MenuId]
	for i := 0; i < len(menu.Children); i++ {
		err = m.getChildrenList(&menu.Children[i], treeMap)
	}
	return err
}
func (m *SysBaseMenu) getMenuTreeMap(authorityId uint) (treeMap map[string][]SysMenu, err error) {
	var allMenus []SysMenu
	var baseMenu []SysBaseMenu
	var btns []SysAuthorityBtn
	treeMap = make(map[string][]SysMenu)

	var SysAuthorityMenus []SysAuthorityMenu
	err = global.GLOAB_DB.Where("sys_authority_authority_id = ?", authorityId).Find(&SysAuthorityMenus).Error
	if err != nil {
		return
	}

	var MenuIds []string

	for i := range SysAuthorityMenus {
		MenuIds = append(MenuIds, SysAuthorityMenus[i].MenuId)
	}

	err = global.GLOAB_DB.Where("id in (?)", MenuIds).Order("sort").Preload("Parameters").Find(&baseMenu).Error
	if err != nil {
		return
	}

	for i := range baseMenu {
		allMenus = append(allMenus, SysMenu{
			SysBaseMenu: baseMenu[i],
			AuthorityId: authorityId,
			MenuId:      strconv.Itoa(int(baseMenu[i].ID)),
			//Parameters:  baseMenu[i].Parameters,
		})
	}

	err = global.GLOAB_DB.Where("authority_id = ?", authorityId).Preload("SysBaseMenuBtn").Find(&btns).Error
	if err != nil {
		return
	}
	var btnMap = make(map[uint]map[string]uint)
	for _, v := range btns {
		if btnMap[v.SysMenuID] == nil {
			btnMap[v.SysMenuID] = make(map[string]uint)
		}
		btnMap[v.SysMenuID][v.SysBaseMenuBtn.Name] = authorityId
	}
	for _, v := range allMenus {
		v.Btns = btnMap[v.ID]
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}
	return treeMap, err
}

func (m *SysBaseMenu) AddBaseMenu(menu SysBaseMenu) error {
	if !errors.Is(global.GLOAB_DB.Where("name = ?", menu.Name).First(&SysBaseMenu{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在重复name，请修改name")
	}
	err := global.GLOAB_DB.Create(&menu).Error
	if err != nil {
		return err
	}
	err = m.AddMenuAuthority([]SysBaseMenu{menu}, 888)
	return err
}

func (m *SysBaseMenu) AddMenuAuthority(menus []SysBaseMenu, authorityId uint) (err error) {
	var auth SysAuthority
	auth.AuthorityId = authorityId
	auth.SysBaseMenus = menus
	var s SysAuthority
	global.GLOAB_DB.Preload("SysBaseMenus").First(&s, "authority_id = ?", auth.AuthorityId)
	err = global.GLOAB_DB.Model(&s).Association("SysBaseMenus").Replace(&auth.SysBaseMenus)
	return err
}
func (m *SysBaseMenu) DeleteBaseMenu(id int) (err error) {
	err = global.GLOAB_DB.Preload("MenuBtn").Preload("Parameters").Where("parent_id = ?", id).First(&SysBaseMenu{}).Error
	if err != nil {
		var menu SysBaseMenu
		db := global.GLOAB_DB.Preload("SysAuthorities").Where("id = ?", id).First(&menu).Delete(&menu)
		err = global.GLOAB_DB.Delete(&SysBaseMenuParameter{}, "sys_base_menu_id = ?", id).Error
		err = global.GLOAB_DB.Delete(&SysBaseMenuBtn{}, "sys_base_menu_id = ?", id).Error
		err = global.GLOAB_DB.Delete(&SysAuthorityBtn{}, "sys_menu_id = ?", id).Error
		if err != nil {
			return err
		}
		if len(menu.SysAuthorities) > 0 {
			err = global.GLOAB_DB.Model(&menu).Association("SysAuthorities").Delete(&menu.SysAuthorities)
		} else {
			err = db.Error
			if err != nil {
				return
			}
		}
	} else {
		return errors.New("此菜单存在子菜单不可删除")
	}
	return err
}
func (m *SysBaseMenu) UpdateBaseMenu(menu SysBaseMenu) (err error) {
	var oldMenu SysBaseMenu
	upDateMap := make(map[string]interface{})
	upDateMap["keep_alive"] = menu.KeepAlive
	upDateMap["close_tab"] = menu.CloseTab
	upDateMap["default_menu"] = menu.DefaultMenu
	upDateMap["parent_id"] = menu.ParentId
	upDateMap["path"] = menu.Path
	upDateMap["name"] = menu.Name
	upDateMap["hidden"] = menu.Hidden
	upDateMap["component"] = menu.Component
	upDateMap["title"] = menu.Title
	upDateMap["active_name"] = menu.ActiveName
	upDateMap["icon"] = menu.Icon
	upDateMap["sort"] = menu.Sort

	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		db := tx.Where("id = ?", menu.ID).Find(&oldMenu)
		if oldMenu.Name != menu.Name {
			if !errors.Is(tx.Where("id <> ? AND name = ?", menu.ID, menu.Name).First(&SysBaseMenu{}).Error, gorm.ErrRecordNotFound) {
				zap.L().Debug("存在相同name修改失败")
				return errors.New("存在相同name修改失败")
			}
		}
		txErr := tx.Unscoped().Delete(&SysBaseMenuParameter{}, "sys_base_menu_id = ?", menu.ID).Error
		if txErr != nil {
			zap.L().Debug(txErr.Error())
			return txErr
		}
		txErr = tx.Unscoped().Delete(&SysBaseMenuBtn{}, "sys_base_menu_id = ?", menu.ID).Error
		if txErr != nil {
			zap.L().Debug(txErr.Error())
			return txErr
		}
		if len(menu.Parameters) > 0 {
			for k := range menu.Parameters {
				menu.Parameters[k].SysBaseMenuID = menu.ID
			}
			txErr = tx.Create(&menu.Parameters).Error
			if txErr != nil {
				zap.L().Debug(txErr.Error())
				return txErr
			}
		}

		if len(menu.MenuBtn) > 0 {
			for k := range menu.MenuBtn {
				menu.MenuBtn[k].SysBaseMenuID = menu.ID
			}
			txErr = tx.Create(&menu.MenuBtn).Error
			if txErr != nil {
				zap.L().Debug(txErr.Error())
				return txErr
			}
		}

		txErr = db.Updates(upDateMap).Error
		if txErr != nil {
			zap.L().Debug(txErr.Error())
			return txErr
		}
		return nil
	})
	return err
}
func (m *SysBaseMenu) GetBaseMenuById(id int) (menu SysBaseMenu, err error) {
	err = global.GLOAB_DB.Preload("MenuBtn").Preload("Parameters").Where("id = ?", id).First(&menu).Error
	return
}
func (m *SysBaseMenu) GetMenuAuthority(info *request.GetAuthorityId) (menus []SysMenu, err error) {
	var baseMenu []SysBaseMenu
	var SysAuthorityMenus []SysAuthorityMenu
	err = global.GLOAB_DB.Where("sys_authority_authority_id = ?", info.AuthorityId).Find(&SysAuthorityMenus).Error
	if err != nil {
		return
	}

	var MenuIds []string

	for i := range SysAuthorityMenus {
		MenuIds = append(MenuIds, SysAuthorityMenus[i].MenuId)
	}

	err = global.GLOAB_DB.Where("id in (?) ", MenuIds).Order("sort").Find(&baseMenu).Error

	for i := range baseMenu {
		menus = append(menus, SysMenu{
			SysBaseMenu: baseMenu[i],
			AuthorityId: info.AuthorityId,
			MenuId:      strconv.Itoa(int(baseMenu[i].ID)),
			Parameters:  baseMenu[i].Parameters,
		})
	}
	return menus, err
}
func (m *SysBaseMenu) GetInfoList() (list interface{}, total int64, err error) {
	var menuList []SysBaseMenu
	treeMap, err := m.getBaseMenuTreeMap()
	menuList = treeMap["0"]
	for i := 0; i < len(menuList); i++ {
		err = m.getBaseChildrenList(&menuList[i], treeMap)
	}
	return menuList, total, err
}
func (m *SysBaseMenu) getBaseChildrenList(menu *SysBaseMenu, treeMap map[string][]SysBaseMenu) (err error) {
	menu.Children = treeMap[strconv.Itoa(int(menu.ID))]
	for i := 0; i < len(menu.Children); i++ {
		err = m.getBaseChildrenList(&menu.Children[i], treeMap)
	}
	return err
}
func (m *SysBaseMenu) getBaseMenuTreeMap() (treeMap map[string][]SysBaseMenu, err error) {
	var allMenus []SysBaseMenu
	treeMap = make(map[string][]SysBaseMenu)
	err = global.GLOAB_DB.Order("sort").Preload("MenuBtn").Preload("Parameters").Find(&allMenus).Error
	for _, v := range allMenus {
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}
	return treeMap, err
}
