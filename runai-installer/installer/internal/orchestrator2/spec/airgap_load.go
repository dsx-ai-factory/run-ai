package spec

import "errors"

// AirgapLoad is the fully resolved input of the airgap load command: the
// credentials are read, so nothing downstream performs configuration I/O. The
// artifact lock is not here — it travels inside the bundle and is only readable
// once the bundle is extracted.
type AirgapLoad struct {
	BundlePath  string
	WorkingDir  string
	ValuesDir   string
	Destination string
	Credential  RegistryCredential
}

func (l *AirgapLoad) validate() error {
	if l == nil {
		return errors.New("airgap load settings are required")
	}
	if l.BundlePath == "" {
		return errors.New("bundle path is required")
	}
	if l.WorkingDir == "" {
		return errors.New("working directory is required")
	}
	if l.ValuesDir == "" {
		return errors.New("values directory is required")
	}
	if l.Destination == "" {
		return errors.New("destination registry is required")
	}
	return nil
}
