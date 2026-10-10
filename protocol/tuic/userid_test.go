package tuic

import (
	"fmt"
	"testing"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/internal/useridtest"
)

func TestAssignUserIDsStable(t *testing.T) {
	h := &Inbound{}
	useridtest.Run(t, func(names ...string) []int {
		users := make([]option.TUICUser, len(names))
		for i, name := range names {
			userUUID := name
			if userUUID == "" { // TUIC requires a UUID even for unnamed users
				userUUID = fmt.Sprintf("00000000-0000-4000-8000-%012d", i)
			}
			users[i] = option.TUICUser{Name: name, UUID: userUUID, Password: "password"}
		}
		ids, _, _, err := h.assignUserIDs(users)
		if err != nil {
			t.Fatal(err)
		}
		return ids
	}, func() (map[int]string, map[string]int) { return h.userNameMap, h.userIDByName })
}
