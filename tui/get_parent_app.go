package tui

func GetParentApp(app *MiniApp) *MiniApp {
	for _, miniApp := range miniApps {
		if miniApp.id == app.parent {
			return miniApp
		}
	}
	return nil
}
