package user

import (
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/google/uuid"
)

func PaginateResult(users []*domainuser.User, limit int32) ([]*domainuser.User, domainuser.PageInfo) {
	hasMore := false
	var next *domainuser.UserCursorKey

	if int32(len(users)) > limit {
		hasMore = true
		last := users[limit-1]
		lastID, _ := uuid.Parse(last.ID)
		next = &domainuser.UserCursorKey{Email: last.Email, ID: lastID}
		users = users[:limit]
	} else if len(users) > 0 {
		last := users[len(users)-1]
		lastID, _ := uuid.Parse(last.ID)
		next = &domainuser.UserCursorKey{Email: last.Email, ID: lastID}
	}

	return users, domainuser.PageInfo{Next: next, HasMore: hasMore}
}
