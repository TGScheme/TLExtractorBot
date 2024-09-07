package debug_menu

import (
	"TLExtractor/debug_menu/types"
	"TLExtractor/github"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/charmbracelet/huh"
)

func newTDeskSelect(typeName string) *types.TDeskSelect {
	selector := &types.TDeskSelect{}
	selector.Select = huh.NewSelect[string]().
		Title(fmt.Sprintf("Select %s TDesk Commit:", typeName)).
		Value(&selector.Value).
		OptionsFunc(func() []huh.Option[string] {
			commits, err := github.Client.GetCommits("telegramdesktop", "tdesktop", "Telegram/SourceFiles/mtproto/scheme/api.tl")
			if err != nil {
				gologging.Fatal(err)
			}
			selector.CommitList = commits
			var commitMessages []string
			for _, commit := range commits {
				commitMessages = append(commitMessages, selector.NameFormat(commit))
			}
			return huh.NewOptions(commitMessages...)
		}, "commit")
	return selector
}
