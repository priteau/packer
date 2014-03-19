package nimbus

import (
)

type artifact struct {
	image string
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (*artifact) Files() []string {
	// We have no files
	return nil
}

func (a *artifact) Id() string {
	return a.image
}

func (a *artifact) String() string {
	return a.image
}

func (a *artifact) Destroy() error {
	// TODO Implement destroy

	return nil
}
