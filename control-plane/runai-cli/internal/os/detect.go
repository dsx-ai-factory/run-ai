package os

import "runtime"

func DetectOsType() Type {
	switch runtime.GOOS {
	case Windows.String():
		return Windows
	case Linux.String():
		return Linux
	case Darwin.String():
		return Darwin
	default:
		return Unknown
	}
}

func DetectArch() Arch {
	switch runtime.GOARCH {
	case AMD64.String():
		return AMD64
	case X86.String():
		return X86
	case ARM64.String():
		return ARM64
	default:
		return UnknownArch
	}
}
