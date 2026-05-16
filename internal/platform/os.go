package platform

import "runtime"

func DetectOS() string {
	return runtime.GOOS
}

func DetectArch() string {
	return runtime.GOARCH
}
