package domain

type AccountStatus string

const (
	AccountActive  AccountStatus = "active"
	AccountBlocked AccountStatus = "blocked"
	AccountDeleted AccountStatus = "deleted"
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
