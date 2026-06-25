package svid

import (
	"context"
	"fmt"
	"strings"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire/pkg/server/datastore"
	"github.com/spiffe/spire/proto/spire/common"
)

const (
	k8sPSATSelectorType                = "k8s_psat"
	k8sPSATAgentNodeNameSelectorPrefix = "agent_node_name:"
)

func (s *Service) callerAgentNodeName(ctx context.Context, agentID spiffeid.ID) (string, error) {
	if s.ds == nil {
		return "", nil
	}
	selectors, err := s.ds.GetNodeSelectors(ctx, agentID.String(), datastore.RequireCurrent)
	if err != nil {
		return "", fmt.Errorf("get node selectors for %q: %w", agentID.String(), err)
	}
	return agentNodeNameFromSelectors(selectors)
}

func agentNodeNameFromSelectors(selectors []*common.Selector) (string, error) {
	var nodeName string
	for _, selector := range selectors {
		if selector == nil || selector.Type != k8sPSATSelectorType {
			continue
		}
		value, ok := strings.CutPrefix(selector.Value, k8sPSATAgentNodeNameSelectorPrefix)
		if !ok {
			continue
		}
		if value == "" {
			return "", fmt.Errorf("empty %s:%s selector", k8sPSATSelectorType, strings.TrimSuffix(k8sPSATAgentNodeNameSelectorPrefix, ":"))
		}
		if nodeName != "" && nodeName != value {
			return "", fmt.Errorf("conflicting agent node name selectors %q and %q", nodeName, value)
		}
		nodeName = value
	}
	return nodeName, nil
}
