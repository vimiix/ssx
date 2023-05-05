package selector

import (
	"github.com/vimiix/ssx/internal/types"
)

type FilterOption struct {
	Keyword string
}

type Selector struct {
	nodes []*types.Node
	opt   *FilterOption
}

func NewSelector(nodes []*types.Node, opt *FilterOption) *Selector {
	s := &Selector{
		nodes: nodes,
		opt:   opt,
	}
	s.initial()
	return s
}

func (s *Selector) initial() {

}
