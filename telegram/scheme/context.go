package scheme

import "sync"

var Client *context

type context struct {
	removedConstructors []string
	removedComputed     bool
	syncDep             sync.Mutex
}
