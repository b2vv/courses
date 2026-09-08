package domain

type OwnershipRenewal uint8

const (
	OwnershipRenewed OwnershipRenewal = iota + 1
	OwnershipMissing
	OwnershipTakenOver
)
