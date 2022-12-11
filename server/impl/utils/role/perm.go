package role

import (
	"errors"
	"fmt"
	"sync"

	mcomRoles "gitlab.kenda.com.tw/kenda/mcom/utils/roles"

	"gitlab.kenda.com.tw/kenda/mui/server/protobuf/kenda"
	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
)

// funcRoleList define function roles relationship.
type funcRoleList map[kenda.FunctionOperationID]map[mcomRoles.Role]struct{}

// permissionList which stores function permission roles data
var permissionList funcRoleList

// HasPermission will check if those user's roles have the operation id's permission.
// it will return true if user has permission, otherwise false.
func HasPermission(id kenda.FunctionOperationID, roles []models.Role) bool {
	if rpm, ok := permissionList[id]; ok {
		for _, userRole := range roles {
			if _, ok := rpm[mcomRoles.Role(userRole)]; ok {
				return true
			}
		}
	}
	return false
}

var mu sync.Mutex

// InitPermission initialized permission list.
func InitPermission(perm map[string][]string) error {
	mu.Lock()
	defer mu.Unlock()
	if len(permissionList) > 0 {
		return errors.New("role permissions have been initialized")
	}
	permissionList = make(funcRoleList, len(perm))
	for name, roles := range perm {
		funcID, ok := kenda.FunctionOperationID_value[name]
		if !ok {
			return fmt.Errorf("function %s was not in the list", name)
		}
		roleMap, err := rolesToMap(roles)
		if err != nil {
			return err
		}
		permissionList[kenda.FunctionOperationID(funcID)] = roleMap
	}
	return nil
}

// ClearPermission clears permission list.
func ClearPermission() {
	mu.Lock()
	defer mu.Unlock()
	permissionList = nil
}

// rolesToMap convert role list into mapping list;
// return error if the role is not existed inside the library list.
func rolesToMap(roles []string) (map[mcomRoles.Role]struct{}, error) {
	funcRoles := make(map[mcomRoles.Role]struct{}, len(roles))
	for _, role := range roles {
		r, ok := mcomRoles.Role_value[role]
		if !ok {
			return nil, fmt.Errorf("not existed role: %s", role)
		}
		funcRoles[mcomRoles.Role(r)] = struct{}{}
	}
	return funcRoles, nil
}
