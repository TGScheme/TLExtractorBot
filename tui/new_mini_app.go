package tui

import (
	"TLExtractor/assets"
	"TLExtractor/tui/types"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
)

func NewMiniApp(appName string) *MiniApp {
	idRaw, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	miniApp := &MiniApp{
		id:      idRaw.Int64(),
		appName: appName,
		logo: strings.ReplaceAll(
			string(assets.Resources[fmt.Sprintf("%s.ascii", appName)]),
			"\r",
			"",
		),
		loadingMessage: "Checking...",
		check: func(checkType types.CheckType) error {
			return nil
		},
	}
	miniApps = append(miniApps, miniApp)
	return miniApp
}
