package types

type ReferenceList map[string]CodeRef

func NewReferenceList() ReferenceList {
	return make(ReferenceList)
}
