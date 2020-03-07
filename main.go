package main

import (
	"github.com/demosdemon/termshare/cmd"
	_ "github.com/demosdemon/termshare/cmd/auth"
	_ "github.com/demosdemon/termshare/cmd/cat"
	_ "github.com/demosdemon/termshare/cmd/play"
	_ "github.com/demosdemon/termshare/cmd/rec"
	_ "github.com/demosdemon/termshare/cmd/upload"
)

func main() {
	_ = cmd.Execute(nil)
}
