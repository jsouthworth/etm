package agent

type XfrmFn func(interface {} ...interface{})interface{}

type Agent struct {
	atom.Atom
	ch chan XfrmFn
}

func New(s interface{} ) *Agent {
	a := &Agent{state: atom.New(s), ch: make(chan XfrmFn, 100)}
	go a.process()
	return a
}

func (a *Agent) Dispatch(f XfrmFn) {
	a.ch <-f
}
func (a *Agent) process() {
	for {
		f := <- a.ch
		a.Swap(f)
	}
}