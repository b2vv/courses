package domain

type AccountStatus string

const (
	AccountActive  AccountStatus = "active"
	AccountBlocked AccountStatus = "blocked"
	AccountDeleted AccountStatus = "deleted"
)

// UserSort is the whitelisted ordering of the admin user roster. The empty
// value and SortByEmail both mean the default (email, then id); SortByName
// orders by display name, SortByRecent by newest account first. Any other
// value is refused at the transport boundary rather than reaching SQL.
type UserSort string

const (
	SortByEmail  UserSort = "email"
	SortByName   UserSort = "name"
	SortByRecent UserSort = "recent"
)

type Actor struct {
	UserID string
}

func (user User) AccountStatus() AccountStatus {
	if user.DeletedAt != nil {
		return AccountDeleted
	}
	if user.DisabledAt != nil {
		return AccountBlocked
	}
	return AccountActive
}

func CanBlock(actor Actor, target User, otherActiveAdmins int) error {
	if actor.UserID == target.ID {
		return ErrCannotTargetSelf
	}
	if target.Role == Admin && target.Active() && otherActiveAdmins == 0 {
		return ErrLastAdmin
	}
	return nil
}

func CanDelete(actor Actor, target User) error {
	if actor.UserID == target.ID {
		return ErrCannotTargetSelf
	}
	if target.Role == Admin {
		return ErrAdminImmutable
	}
	return nil
}

func ValidPassword(password string) bool {
	return len(password) > 0 && len(password) <= 1024
}
