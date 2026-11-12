package mieru

import (
	"github.com/sagernet/sing-box/option"

	mierupb "github.com/enfein/mieru/v3/pkg/appctl/appctlpb"
	"google.golang.org/protobuf/proto"
)

// UpdateUsers replaces the users of the running mieru server without a
// restart. Connections of removed users, or of users whose password
// changed, are closed by the server; other users are not affected.
func (h *Inbound) UpdateUsers(users []option.MieruUser) error {
	if err := validateMieruUsers(users); err != nil {
		return err
	}
	pbUsers := make([]*mierupb.User, 0, len(users))
	userNames := make([]string, 0, len(users))
	for _, user := range users {
		pbUsers = append(pbUsers, &mierupb.User{
			Name:     proto.String(user.Name),
			Password: proto.String(user.Password),
		})
		userNames = append(userNames, user.Name)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.server.UpdateUsers(pbUsers); err != nil {
		return err
	}
	h.userNames = userNames
	return nil
}
