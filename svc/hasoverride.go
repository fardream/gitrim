package svc

type RequestWithPossibleOverride interface {
	GetId() string
	GetOverrideFromBranch() string
	GetOverrideToBranch() string
}

func HasOverrides[TReq RequestWithPossibleOverride](a TReq) bool {
	return a.GetOverrideFromBranch() != "" || a.GetOverrideToBranch() != ""
}
