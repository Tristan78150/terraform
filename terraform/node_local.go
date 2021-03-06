package terraform

import (
	"log"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/lang"
)

// NodePlannableLocal represents a named local value in a configuration module,
// which has not yet been expanded.
type NodePlannableLocal struct {
	Addr   addrs.LocalValue
	Module addrs.Module
	Config *configs.Local
}

var (
	_ RemovableIfNotTargeted     = (*NodePlannableLocal)(nil)
	_ GraphNodeReferenceable     = (*NodePlannableLocal)(nil)
	_ GraphNodeReferencer        = (*NodePlannableLocal)(nil)
	_ GraphNodeDynamicExpandable = (*NodePlannableLocal)(nil)
)

func (n *NodePlannableLocal) Name() string {
	path := n.Module.String()
	addr := n.Addr.String()
	if path != "" {
		return path + "." + addr
	}
	return addr
}

// GraphNodeModulePath
func (n *NodePlannableLocal) ModulePath() addrs.Module {
	return n.Module
}

// RemovableIfNotTargeted
func (n *NodePlannableLocal) RemoveIfNotTargeted() bool {
	return true
}

// GraphNodeReferenceable
func (n *NodePlannableLocal) ReferenceableAddrs() []addrs.Referenceable {
	return []addrs.Referenceable{n.Addr}
}

// GraphNodeReferencer
func (n *NodePlannableLocal) References() []*addrs.Reference {
	refs, _ := lang.ReferencesInExpr(n.Config.Expr)
	return appendResourceDestroyReferences(refs)
}

func (n *NodePlannableLocal) DynamicExpand(ctx EvalContext) (*Graph, error) {
	var g Graph
	expander := ctx.InstanceExpander()
	for _, module := range expander.ExpandModule(n.Module) {
		o := &NodeLocal{
			Addr:   n.Addr.Absolute(module),
			Config: n.Config,
		}
		log.Printf("[TRACE] Expanding local: adding %s as %T", o.Addr.String(), o)
		g.Add(o)
	}
	return &g, nil
}

// NodeLocal represents a named local value in a particular module.
//
// Local value nodes only have one operation, common to all walk types:
// evaluate the result and place it in state.
type NodeLocal struct {
	Addr   addrs.AbsLocalValue
	Config *configs.Local
}

var (
	_ GraphNodeModuleInstance = (*NodeLocal)(nil)
	_ RemovableIfNotTargeted  = (*NodeLocal)(nil)
	_ GraphNodeReferenceable  = (*NodeLocal)(nil)
	_ GraphNodeReferencer     = (*NodeLocal)(nil)
	_ GraphNodeEvalable       = (*NodeLocal)(nil)
	_ dag.GraphNodeDotter     = (*NodeLocal)(nil)
)

func (n *NodeLocal) Name() string {
	return n.Addr.String()
}

// GraphNodeModuleInstance
func (n *NodeLocal) Path() addrs.ModuleInstance {
	return n.Addr.Module
}

// GraphNodeModulePath
func (n *NodeLocal) ModulePath() addrs.Module {
	return n.Addr.Module.Module()
}

// RemovableIfNotTargeted
func (n *NodeLocal) RemoveIfNotTargeted() bool {
	return true
}

// GraphNodeReferenceable
func (n *NodeLocal) ReferenceableAddrs() []addrs.Referenceable {
	return []addrs.Referenceable{n.Addr.LocalValue}
}

// GraphNodeReferencer
func (n *NodeLocal) References() []*addrs.Reference {
	refs, _ := lang.ReferencesInExpr(n.Config.Expr)
	return appendResourceDestroyReferences(refs)
}

// GraphNodeEvalable
func (n *NodeLocal) EvalTree() EvalNode {
	return &EvalLocal{
		Addr: n.Addr.LocalValue,
		Expr: n.Config.Expr,
	}
}

// dag.GraphNodeDotter impl.
func (n *NodeLocal) DotNode(name string, opts *dag.DotOpts) *dag.DotNode {
	return &dag.DotNode{
		Name: name,
		Attrs: map[string]string{
			"label": n.Name(),
			"shape": "note",
		},
	}
}
