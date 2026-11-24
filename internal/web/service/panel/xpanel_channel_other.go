//go:build !linux

package panel

import "fmt"

func brandedUpdateInfo() (*PanelUpdateInfo, error) {
	return nil, fmt.Errorf("Xpanel managed browser updates require Linux")
}
func startBrandedUpdate(dev bool) (int64, error) {
	return 0, fmt.Errorf("Xpanel managed browser updates require Linux")
}
func (s *PanelService) StartXpanelUpdate(expected string) (int64, error) {
	return 0, fmt.Errorf("Xpanel managed browser updates require Linux")
}
