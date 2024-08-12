package tui

import "log"

func (miniApp *MiniApp) NewAppPage() *MiniApp {
	if miniApp.parent != 0 {
		log.Fatal("NewAppPage() can only be called on the root MiniApp")
	}
	pseudoMiniApp := NewMiniApp(miniApp.appName)
	pseudoMiniApp.parent = miniApp.id
	return pseudoMiniApp
}
