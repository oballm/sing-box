package snell

import (
	"testing"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/internal/useridtest"
)

func TestAssignUserIDsStable(t *testing.T) {
	h := &Inbound{}
	useridtest.Run(t, func(names ...string) []int {
		users := make([]option.SnellUser, len(names))
		for i, name := range names {
			users[i] = option.SnellUser{Name: name, UserKey: "key-" + name}
		}
		ids, _ := h.assignUserIDs(users)
		return ids
	}, func() (map[int]string, map[string]int) { return h.userNameMap, h.userIDByName })
}
