package snell

import (
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

// UpdateUsers replaces the users of a multi-user inbound without a restart.
// An inbound created without users (single PSK) has no user table to update.
func (h *Inbound) UpdateUsers(users []option.SnellUser) error {
	if h.multiService == nil {
		return E.New("snell: inbound was created without users, user update needs a restart")
	}
	h.userMu.Lock()
	defer h.userMu.Unlock()
	oldNameMap, oldIDByName, oldNextUserID := h.userNameMap, h.userIDByName, h.nextUserID
	userList, keyList := h.assignUserIDs(users)
	err := h.multiService.UpdateUsers(userList, keyList)
	if err != nil {
		// The service kept the old users; keep the matching names.
		h.userNameMap, h.userIDByName, h.nextUserID = oldNameMap, oldIDByName, oldNextUserID
	}
	return err
}

// assignUserIDs gives each user.Name (the user's UUID in xboard-node) a
// stable int that persists across UpdateUsers calls, so a connection keeps
// its owner when other users are added or removed. Users were identified by
// list index before, which shifts on any removal and moves traffic between
// users. Must be serialized via h.userMu (or called before the inbound is
// shared).
func (h *Inbound) assignUserIDs(users []option.SnellUser) ([]int, [][]byte) {
	userList := make([]int, 0, len(users))
	keyList := make([][]byte, 0, len(users))
	newNameMap := make(map[int]string, len(users))
	newIDByName := make(map[string]int, len(users))
	for _, user := range users {
		id, ok := h.userIDByName[user.Name]
		if user.Name == "" || !ok {
			id = h.nextUserID
			h.nextUserID++
		}
		userList = append(userList, id)
		keyList = append(keyList, []byte(user.UserKey))
		if user.Name != "" {
			newNameMap[id] = user.Name
			newIDByName[user.Name] = id
		}
	}
	h.userNameMap = newNameMap
	h.userIDByName = newIDByName
	return userList, keyList
}
