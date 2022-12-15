package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func Stringify(data interface{}) string {
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return string(bytes)
}

func PrintFatal(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
	os.Exit(1)
}
