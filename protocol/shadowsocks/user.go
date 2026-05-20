package shadowsocks

import (
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
)

func (h *MultiInbound) UpdateUsersByOptions(users []option.ShadowsocksUser) error {
	h.userMu.Lock()
	defer h.userMu.Unlock()
	h.users = users
	names := common.Map(users, func(user option.ShadowsocksUser) string {
		return user.Name
	})
	passwords := common.Map(users, func(user option.ShadowsocksUser) string {
		return user.Password
	})
	userList := h.assignUserIDs(names)
	return h.service.UpdateUsersWithPasswords(userList, passwords)
}

// assignUserIDs gives each name (UUID, in upstream usage) a stable int that
// persists across UpdateUsers calls. The previous code used the iteration
// index directly, so any reorder or deletion would shift indices and
// silently re-attribute traffic between users. Not safe for concurrent
// callers — must be serialized via h.userMu.
func (h *MultiInbound) assignUserIDs(names []string) []int {
	userList := make([]int, 0, len(names))
	newNameMap := make(map[int]string, len(names))
	newIDByName := make(map[string]int, len(names))
	for _, name := range names {
		var id int
		if name != "" {
			if existing, ok := h.userIDByName[name]; ok {
				id = existing
			} else {
				id = h.nextUserID
				h.nextUserID++
			}
		} else {
			id = h.nextUserID
			h.nextUserID++
		}
		userList = append(userList, id)
		if name != "" {
			newNameMap[id] = name
			newIDByName[name] = id
		}
	}
	h.userNameMap = newNameMap
	h.userIDByName = newIDByName
	return userList
}
