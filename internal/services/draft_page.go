package services

import (
	"fmt"
	"sync"

	"github.com/Laky-64/gologging"
)

type draftPage struct {
	mutex  sync.Mutex
	opened bool
	closed bool
	path   string
	url    string
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
	return draft.url, true
}
