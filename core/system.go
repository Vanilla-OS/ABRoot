package core

import "errors"

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

// ABSystem represents the system
type ABSystem struct {
	Checks   *Checks
	RootM    *ABRootManager
	Registry *Registry
	CurImage *ABImage
}

// NewABSystem creates a new system
func NewABSystem() (*ABSystem, error) {
	i, err := NewABImageFromRoot()
	if err != nil {
		return nil, err
	}

	c := NewChecks()
	r := NewRegistry()
	rm := NewABRootManager()

	return &ABSystem{
		Checks:   c,
		RootM:    rm,
		Registry: r,
		CurImage: i,
	}, nil
}

// CheckAll performs all checks from the Checks struct
func (s *ABSystem) CheckAll() error {
	err := s.Checks.PerformAllChecks()
	if err != nil {
		return err
	}
	return nil
}

// CheckUpdate checks if there is an update available
func (s *ABSystem) CheckUpdate() bool {
	return s.Registry.HasUpdate(s.CurImage.Digest)
}

// Upgrade upgrades the system
func (s *ABSystem) Upgrade() error {
	if !s.CheckUpdate() {
		return errors.New("no update available")
	}

	_, err := s.RootM.GetFuture()
	if err != nil {
		return err
	}

	return nil // TODO
}
