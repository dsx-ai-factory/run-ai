package structs

type OrgUnitKey struct {
	Scope string
	Id    string
}

type OrgUnitMap map[OrgUnitKey]OrgUnitKey
