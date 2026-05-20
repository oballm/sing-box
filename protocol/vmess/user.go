package vmess

import (
	"github.com/sagernet/sing-box/option"
)

func (h *Inbound) UpdateUsers(users []option.VMessUser) error {
	h.userMu.Lock()
	defer h.userMu.Unlock()
	h.users = users
	userList, uuidList, alterIDList := h.assignUserIDs(users)
	return h.service.UpdateUsers(userList, uuidList, alterIDList)
}

func (h *Inbound) DelUsers(uuids []string) error {
	h.userMu.Lock()
	defer h.userMu.Unlock()
	toDelete := make(map[string]struct{}, len(uuids))
	for _, uuid := range uuids {
		toDelete[uuid] = struct{}{}
	}
	remaining := make([]option.VMessUser, 0, len(h.users))
	for _, user := range h.users {
		if _, found := toDelete[user.UUID]; !found {
			remaining = append(remaining, user)
		}
	}
	h.users = remaining
	userList, uuidList, alterIDList := h.assignUserIDs(remaining)
	return h.service.UpdateUsers(userList, uuidList, alterIDList)
}

// assignUserIDs gives each user.Name (UUID, in upstream usage) a stable int
// that persists across UpdateUsers calls. The previous code used the
// iteration index directly, so any reorder or deletion would shift indices
// and silently re-attribute traffic between users. Not safe for concurrent
// callers — must be serialized via h.userMu.
func (h *Inbound) assignUserIDs(users []option.VMessUser) ([]int, []string, []int) {
	userList := make([]int, 0, len(users))
	uuidList := make([]string, 0, len(users))
	alterIDList := make([]int, 0, len(users))
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
		uuidList = append(uuidList, user.UUID)
		alterIDList = append(alterIDList, user.AlterId)
		if user.Name != "" {
			newNameMap[id] = user.Name
			newIDByName[user.Name] = id
		}
	}
	h.userNameMap = newNameMap
	h.userIDByName = newIDByName
	return userList, uuidList, alterIDList
}
