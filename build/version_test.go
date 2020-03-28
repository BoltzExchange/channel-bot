package build

import "testing"

func TestGetVersion(t *testing.T) {
	result := "v" + version + "-" + Commit

	if GetVersion() != result {
		t.Error("Version is not formatted correctly: " + result)
	}
}
