package scheme

import "sync"

var Client *context

type context struct {
	removedConstructors []string
	syncDep             sync.Mutex
}
