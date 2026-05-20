package tuic

import (
	"github.com/gofrs/uuid/v5"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/exceptions"
)

func (h *Inbound) UpdateUsers(users []option.TUICUser) error {
	h.userMu.Lock()
	defer h.userMu.Unlock()
	userList, userUUIDList, userPasswordList, err := h.assignUserIDs(users)
	if err != nil {
		return err
	}
	h.server.UpdateUsers(userList, userUUIDList, userPasswordList)
	return nil
}

// assignUserIDs gives each user.Name (UUID, in upstream usage) a stable int
// that persists across UpdateUsers calls. The previous code used the
// iteration index directly, so any reorder or deletion would shift indices
// and silently re-attribute traffic between users. Not safe for concurrent
// callers — must be serialized via h.userMu.
func (h *Inbound) assignUserIDs(users []option.TUICUser) ([]int, [][16]byte, []string, error) {
	userList := make([]int, 0, len(users))
	userUUIDList := make([][16]byte, 0, len(users))
	userPasswordList := make([]string, 0, len(users))
	newNameMap := make(map[int]string, len(users))
	newIDByName := make(map[string]int, len(users))
	for index, user := range users {
		if user.UUID == "" {
			return nil, nil, nil, exceptions.New("missing uuid for user ", index)
		}
		userUUID, err := uuid.FromString(user.UUID)
		if err != nil {
			return nil, nil, nil, exceptions.Cause(err, "invalid uuid for user ", index)
		}
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
		userUUIDList = append(userUUIDList, userUUID)
		userPasswordList = append(userPasswordList, user.Password)
		if user.Name != "" {
			newNameMap[id] = user.Name
			newIDByName[user.Name] = id
		}
	}
	h.userNameMap = newNameMap
	h.userIDByName = newIDByName
	return userList, userUUIDList, userPasswordList, nil
}
