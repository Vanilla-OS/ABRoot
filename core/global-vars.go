package core

var Verbose = false

func SetVerbose(status bool) {
	Verbose = status
}

func GetVerbose() bool {
	return Verbose
}
