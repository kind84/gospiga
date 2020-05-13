package gospiga

import (
	"fmt"
	"io"
	"runtime"
)

// Populated during build, don't touch!
var (
	Version   = "undefined"
	GitRev    = "undefined"
	GitBranch = "undefined"
	BuildDate = "undefined"
)

func PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "Version:      %s\n", Version)
	fmt.Fprintf(w, "Git revision: %s\n", GitRev)
	fmt.Fprintf(w, "Git branch:   %s\n", GitBranch)
	fmt.Fprintf(w, "Go version:   %s\n", runtime.Version())
	fmt.Fprintf(w, "Built:        %s\n", BuildDate)
	fmt.Fprintf(w, "OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
