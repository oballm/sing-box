package hysteria2

import (
	"github.com/sagernet/sing-box/option"
)

func (h *Inbound) UpdateUsers(users []option.Hysteria2User) error {
	h.userMu.Lock()
	userList, userPasswordList := h.assignUserIDs(users)
	h.service.UpdateUsers(userList, userPasswordList)
	h.userMu.Unlock()
	return nil
}

// assignUserIDs gives each user.Name (UUID, in upstream usage) a stable int
// that persists across UpdateUsers calls. The previous code used the
// iteration index directly, so any reorder or deletion would shift indices
// and silently re-attribute traffic between users. Not safe for concurrent
// callers — must be serialized via h.userMu.
func (h *Inbound) assignUserIDs(users []option.Hysteria2User) ([]int, []string) {
	userList := make([]int, 0, len(users))
	userPasswordList := make([]string, 0, len(users))
	newNameMap := make(map[int]string, len(users))
	newIDByName := make(map[string]int, len(users))
	for _, user := range users {
		var id int
		if user.Name != "" {
			if existing, ok := h.userIDByName[user.Name]; ok {
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
		userPasswordList = append(userPasswordList, user.Password)
		if user.Name != "" {
			newNameMap[id] = user.Name
			newIDByName[user.Name] = id
		}
	}
	h.userNameMap = newNameMap
	h.userIDByName = newIDByName
	return userList, userPasswordList
}
