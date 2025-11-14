package permission

import (
	"encoding/json"

	"github.com/go-arcade/arcade/internal/engine/model/permission"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
)


type IRouterPermissionRepository interface {
	GetAllRoutePermissions() ([]permission.RoutePermissionDTO, error)
	GetRoutePermissionByPath(path, method string) (*permission.RoutePermissionDTO, error)
	CreateOrUpdateRoute(dto permission.RoutePermissionDTO) error
	InitBuiltinRoutes() error
	GetMenuRoutes() (map[string][]permission.RoutePermissionDTO, error)
	DeleteRoute(routeId string) error
	DisableRoute(routeId string) error
	EnableRoute(routeId string) error
}

type RouterPermissionRepo struct {
	db database.DB
}

func NewRouterPermissionRepo(db database.DB) IRouterPermissionRepository {
	return &RouterPermissionRepo{
		db: db,
	}
}

// GetAllRoutePermissions 获取所有路由权限配置
func (r *RouterPermissionRepo) GetAllRoutePermissions() ([]permission.RoutePermissionDTO, error) {
	var routes []permission.RouterPermission
	err := r.db.DB().Where("is_enabled = ?", 1).Order("`order` ASC").Find(&routes).Error
	if err != nil {
		return nil, err
	}

	result := make([]permission.RoutePermissionDTO, 0, len(routes))
	for _, route := range routes {
		var permissions []string
		if len(route.RequiredPermissions) > 0 {
			if err := json.Unmarshal(route.RequiredPermissions, &permissions); err != nil {
				log.Warnf("failed to unmarshal permissions for route %s: %v", route.RouteId, err)
				permissions = []string{}
			}
		}

		result = append(result, permission.RoutePermissionDTO{
			RouteId:             route.RouteId,
			Path:                route.Path,
			Method:              route.Method,
			Name:                route.Name,
			Group:               route.Group,
			Category:            route.Category,
			RequiredPermissions: permissions,
			Icon:                route.Icon,
			Order:               route.Order,
			IsMenu:              route.IsMenu == 1,
			Description:         route.Description,
		})
	}

	return result, nil
}

// GetRoutePermissionByPath 根据路径和方法获取路由权限
func (r *RouterPermissionRepo) GetRoutePermissionByPath(path, method string) (*permission.RoutePermissionDTO, error) {
	var route permission.RouterPermission
	err := r.db.DB().Where("path = ? AND method = ? AND is_enabled = ?", path, method, 1).First(&route).Error
	if err != nil {
		return nil, err
	}

	var permissions []string
	if len(route.RequiredPermissions) > 0 {
		if err := json.Unmarshal(route.RequiredPermissions, &permissions); err != nil {
			permissions = []string{}
		}
	}

	return &permission.RoutePermissionDTO{
		RouteId:             route.RouteId,
		Path:                route.Path,
		Method:              route.Method,
		Name:                route.Name,
		Group:               route.Group,
		Category:            route.Category,
		RequiredPermissions: permissions,
		Icon:                route.Icon,
		Order:               route.Order,
		IsMenu:              route.IsMenu == 1,
		Description:         route.Description,
	}, nil
}

// CreateOrUpdateRoute 创建或更新路由权限配置
func (r *RouterPermissionRepo) CreateOrUpdateRoute(dto permission.RoutePermissionDTO) error {
	permissionsJSON, err := json.Marshal(dto.RequiredPermissions)
	if err != nil {
		return err
	}

	isMenu := 0
	if dto.IsMenu {
		isMenu = 1
	}

	route := permission.RouterPermission{
		RouteId:             dto.RouteId,
		Path:                dto.Path,
		Method:              dto.Method,
		Name:                dto.Name,
		Group:               dto.Group,
		Category:            dto.Category,
		RequiredPermissions: permissionsJSON,
		Icon:                dto.Icon,
		Order:               dto.Order,
		IsMenu:              isMenu,
		IsEnabled:           1,
		Description:         dto.Description,
	}

	// 使用 Upsert 逻辑
	var existing permission.RouterPermission
	err = r.db.DB().Where("route_id = ?", dto.RouteId).First(&existing).Error
	if err != nil {
		// 不存在，创建
		return r.db.DB().Create(&route).Error
	}

	// 已存在，更新
	return r.db.DB().Model(&existing).Updates(route).Error
}

// InitBuiltinRoutes 初始化内置路由配置
func (r *RouterPermissionRepo) InitBuiltinRoutes() error {
	log.Info("initializing builtin routes...")

	for _, routeDTO := range permission.BuiltinRoutes {
		if err := r.CreateOrUpdateRoute(routeDTO); err != nil {
			log.Warnf("failed to init route %s: %v", routeDTO.RouteId, err)
			continue
		}
		log.Infof("initialized route: %s (%s %s)", routeDTO.Name, routeDTO.Method, routeDTO.Path)
	}

	log.Infof("builtin routes initialized: %d routes", len(permission.BuiltinRoutes))
	return nil
}

// GetMenuRoutes 获取菜单路由（按分组分类）
func (r *RouterPermissionRepo) GetMenuRoutes() (map[string][]permission.RoutePermissionDTO, error) {
	var routes []permission.RouterPermission
	err := r.db.DB().Where("is_enabled = ? AND is_menu = ?", 1, 1).Order("category ASC, `order` ASC").Find(&routes).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]permission.RoutePermissionDTO)

	for _, route := range routes {
		var permissions []string
		if len(route.RequiredPermissions) > 0 {
			if err := json.Unmarshal(route.RequiredPermissions, &permissions); err != nil {
				permissions = []string{}
			}
		}

		dto := permission.RoutePermissionDTO{
			RouteId:             route.RouteId,
			Path:                route.Path,
			Method:              route.Method,
			Name:                route.Name,
			Group:               route.Group,
			Category:            route.Category,
			RequiredPermissions: permissions,
			Icon:                route.Icon,
			Order:               route.Order,
			IsMenu:              true,
			Description:         route.Description,
		}

		result[route.Category] = append(result[route.Category], dto)
	}

	return result, nil
}

// DeleteRoute 删除路由配置
func (r *RouterPermissionRepo) DeleteRoute(routeId string) error {
	return r.db.DB().Where("route_id = ?", routeId).Delete(&permission.RouterPermission{}).Error
}

// DisableRoute disables a route (set is_enabled to 0)
func (r *RouterPermissionRepo) DisableRoute(routeId string) error {
	return r.db.DB().Model(&permission.RouterPermission{}).Where("route_id = ?", routeId).Update("is_enabled", 0).Error
}

// EnableRoute enables a route (set is_enabled to 1)
func (r *RouterPermissionRepo) EnableRoute(routeId string) error {
	return r.db.DB().Model(&permission.RouterPermission{}).Where("route_id = ?", routeId).Update("is_enabled", 1).Error
}
