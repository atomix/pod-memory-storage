// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	counterv1 "github.com/atomix/runtime/primitives/pkg/counter/v1"
	countermapv1 "github.com/atomix/runtime/primitives/pkg/countermap/v1"
	electionv1 "github.com/atomix/runtime/primitives/pkg/election/v1"
	indexedmapv1 "github.com/atomix/runtime/primitives/pkg/indexedmap/v1"
	lockv1 "github.com/atomix/runtime/primitives/pkg/lock/v1"
	mapv1 "github.com/atomix/runtime/primitives/pkg/map/v1"
	multimapv1 "github.com/atomix/runtime/primitives/pkg/multimap/v1"
	setv1 "github.com/atomix/runtime/primitives/pkg/set/v1"
	valuev1 "github.com/atomix/runtime/primitives/pkg/value/v1"
	"github.com/atomix/runtime/sdk/pkg/network"
	"github.com/atomix/runtime/sdk/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/protocol/node"
	"github.com/atomix/runtime/sdk/pkg/protocol/statemachine"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"sync"
)

func newNode(network network.Network, opts ...node.Option) *node.Node {
	node := node.NewNode(network, newProtocol(), opts...)
	counterv1.RegisterServer(node)
	countermapv1.RegisterServer(node)
	electionv1.RegisterServer(node)
	indexedmapv1.RegisterServer(node)
	lockv1.RegisterServer(node)
	mapv1.RegisterServer(node)
	multimapv1.RegisterServer(node)
	setv1.RegisterServer(node)
	valuev1.RegisterServer(node)
	return node
}

func newProtocol() node.Protocol {
	return &podMemoryProtocol{
		partition: node.NewPartition(1, newExecutor()),
	}
}

type podMemoryProtocol struct {
	partition node.Partition
}

func (p *podMemoryProtocol) Partitions() []node.Partition {
	return []node.Partition{p.partition}
}

func (p *podMemoryProtocol) Partition(partitionID protocol.PartitionID) (node.Partition, bool) {
	if p.partition.ID() != partitionID {
		return nil, false
	}
	return p.partition, true
}

func newExecutor() node.Executor {
	registry := statemachine.NewPrimitiveTypeRegistry()
	counterv1.RegisterStateMachine(registry)
	countermapv1.RegisterStateMachine(registry)
	electionv1.RegisterStateMachine(registry)
	indexedmapv1.RegisterStateMachine(registry)
	lockv1.RegisterStateMachine(registry)
	mapv1.RegisterStateMachine(registry)
	multimapv1.RegisterStateMachine(registry)
	setv1.RegisterStateMachine(registry)
	valuev1.RegisterStateMachine(registry)
	return &podMemoryExecutor{
		sm: statemachine.NewStateMachine(registry),
	}
}

type podMemoryExecutor struct {
	sm statemachine.StateMachine
	mu sync.RWMutex
}

func (e *podMemoryExecutor) Propose(ctx context.Context, proposal *protocol.ProposalInput, stream streams.WriteStream[*protocol.ProposalOutput]) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sm.Propose(proposal, stream)
	return nil
}

func (e *podMemoryExecutor) Query(ctx context.Context, query *protocol.QueryInput, stream streams.WriteStream[*protocol.QueryOutput]) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.sm.Query(query, stream)
	return nil
}
