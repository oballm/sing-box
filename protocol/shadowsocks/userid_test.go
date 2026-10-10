package shadowsocks

import (
	"testing"

	"github.com/sagernet/sing-box/protocol/internal/useridtest"
)

func TestAssignUserIDsStable(t *testing.T) {
	h := &MultiInbound{}
	useridtest.Run(t, func(names ...string) []int {
		return h.assignUserIDs(names)
	}, func() (map[int]string, map[string]int) { return h.userNameMap, h.userIDByName })
}
