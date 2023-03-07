package build

var Commit string

const version = "2.0.0"

func GetVersion() string {
	return "v" + version + "-" + Commit
}
