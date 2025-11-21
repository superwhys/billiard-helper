package types

type EmailCodeScene uint

const (
	EmailCodeSceneRegister EmailCodeScene = iota + 1
	EmailCodeSceneLogin
)

type LoginType uint

const (
	LoginTypePassword LoginType = iota + 1
	LoginTypeCode
)
