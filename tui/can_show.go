package tui

func (miniApp *MiniApp) HideFunc(hideFunc func() bool) {
	miniApp.hideFunc = hideFunc
}
