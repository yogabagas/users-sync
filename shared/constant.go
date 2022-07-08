package shared

type StatusType int

const (
	AuthToken = ""
)

const (
	StatusImported StatusType = iota
	StatusFinished
	StatusFailInMasterData
	StatusFailInAuth
	StatusFailInAuthz
)

func (s StatusType) String() string {
	status := [...]string{
		"data imported",
		"data finished",
		"data failed to process in masterdata service",
		"data failed to process in authentication service",
		"data failed to process in authorization service",
	}
	return status[s]
}
