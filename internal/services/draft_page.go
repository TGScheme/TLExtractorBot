package services

import (
	"fmt"
	"sync"

	"github.com/Laky-64/gologging"
)

const telegraphSite = "https://telegra.ph/"

type draftPage struct {
	mutex  sync.Mutex
	key    string
	opened bool
	closed bool
	path   string
	url    string
}

func draftKey(pub publication) string {
	return fmt.Sprintf("%s:%s:%d:%d:%t", pub.update.Source, pub.update.VersionName, pub.update.BuildNumber, pub.fullScheme.Layer, pub.isPatch)
}

func (s *Service) restoreDraft(pub publication) *draftPage {
	draft := &draftPage{key: draftKey(pub)}
	settings, err := s.db.SettingsStore.GetSettings()
	if err != nil {
		gologging.Error("telegraph: unable to read the saved draft page:", err)
		return draft
	}
	if settings != nil && settings.DraftPagePath != "" && settings.DraftPageKey == draft.key {
		draft.opened = true
		draft.path, draft.url = settings.DraftPagePath, telegraphSite+settings.DraftPagePath
	}
	return draft
}

func (s *Service) forgetDraft() {
	if err := s.db.SettingsStore.SetDraftPage("", ""); err != nil {
		gologging.Error("telegraph: unable to clear the saved draft page:", err)
	}
}

func (d *draftPage) close() string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.closed = true
	return d.path
}

func (s *Service) openDraft(draft *draftPage, pub publication) (string, bool) {
	draft.mutex.Lock()
	defer draft.mutex.Unlock()
	if draft.closed {
		return "", false
	}
	if draft.opened {
		return draft.url, true
	}
	draft.opened = true
	pub.draft = true
	page, err := s.telegraph.CreatePage(fmt.Sprintf("Layer %d", pub.fullScheme.Layer), pub.page())
	if err != nil {
		gologging.Error("telegraph: unable to open the draft page:", err)
		return "", true
	}
	draft.path, draft.url = page.Path, page.URL
	if err = s.db.SettingsStore.SetDraftPage(draft.key, draft.path); err != nil {
		gologging.Error("telegraph: unable to save the draft page:", err)
	}
	return draft.url, true
}
