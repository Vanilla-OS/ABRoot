package core

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
	return nil // TODO
}
