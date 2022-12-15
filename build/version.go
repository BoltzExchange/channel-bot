package build

var Commit string

const version = "1.2.1"

func GetVersion() string {
	return "v" + version + "-" + Commit
}
