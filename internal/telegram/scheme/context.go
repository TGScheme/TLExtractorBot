package scheme

import (
	"sync"

	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

type Client struct {
	db *db.DB

	syncDep             sync.Mutex
	removedConstructors map[string]bool
	removedComputed     bool

	upstream *types.TLFullScheme
}

func New(database *db.DB) *Client {
	return &Client{db: database, removedConstructors: make(map[string]bool)}
}
