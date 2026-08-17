package scheme

func (ctx *Client) InvalidateCaches() {
	ctx.syncDep.Lock()
	defer ctx.syncDep.Unlock()
	ctx.removedConstructors = make(map[string]bool)
	ctx.removedComputed = false
	ctx.upstream = nil
}
