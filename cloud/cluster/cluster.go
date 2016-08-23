package cluster

import (
	"sort"
)

// Cluster represents a cluster of Nodes.

type Cluster struct {
	State *State
	reqCh chan interface{}
	ch    chan interface{}
	// updateCh chan []NodeUpdate
	// StateCh  chan []Node
	subs []chan []NodeUpdate
}

func NewCluster(state []Node, ch chan interface{}) *Cluster {
	// updateCh chan []NodeUpdate, stateCh chan []Node) *Cluster {
	s := MakeState(state)
	c := &Cluster{
		State: s,
		reqCh: make(chan interface{}),
		ch:    ch,
		// updateCh: updateCh,
		// StateCh:  stateCh,
		subs: nil,
	}
	go c.loop()
	return c
}

func (c *Cluster) Members() []Node {
	ch := make(chan []Node)
	c.reqCh <- ch
	return <-ch
}

func (c *Cluster) Subscribe() Subscriber {
	ch := make(chan Subscriber)
	c.reqCh <- ch
	return <-ch
}

func (c *Cluster) Close() error {
	for _, s := range c.subs {
		c.handleReq(s)
	}
	c.subs = nil
	close(c.reqCh)
	return nil
}

func (c *Cluster) done() bool {
	// return c.updateCh == nil && c.StateCh == nil &&
	return c.ch == nil && c.reqCh == nil
}

func (c *Cluster) loop() {
	for !c.done() {
		select {
		case nodesOrUpdates, ok := <-c.ch:
			if !ok {
				c.ch = nil
				continue
			}
			outgoing := []NodeUpdate{}
			if updates, ok := nodesOrUpdates.([]NodeUpdate); ok {
				outgoing = c.State.FilterAndUpdate(updates)
			} else if nodes, ok := nodesOrUpdates.([]Node); ok {
				outgoing = c.State.SetAndDiff(nodes)
			}
			for _, sub := range c.subs {
				sub <- outgoing
			}
		// case updates, ok := <-c.updateCh:
		// 	if !ok {
		// 		c.updateCh = nil
		// 		continue
		// 	}
		// 	filtered := c.State.FilterAndUpdate(updates)
		// 	for _, sub := range c.subs {
		// 		sub <- filtered
		// 	}
		// case nodes, ok := <-c.StateCh:
		// 	if !ok {
		// 		c.StateCh = nil
		// 		continue
		// 	}
		// 	outgoing := c.State.SetAndDiff(nodes)
		// 	for _, sub := range c.subs {
		// 		sub <- outgoing
		// 	}
		case req, ok := <-c.reqCh:
			if !ok {
				c.reqCh = nil
				continue
			}
			c.handleReq(req)
		}
	}
	for _, sub := range c.subs {
		close(sub)
	}
}

func (c *Cluster) handleReq(req interface{}) {
	switch req := req.(type) {
	case chan []Node:
		// Members()
		req <- c.current()
	case chan Subscriber:
		// Subscribe()
		ch := make(chan []NodeUpdate)
		s := newSubscriber(c.current(), c, ch)
		c.subs = append(c.subs, ch)
		req <- s
	case chan []NodeUpdate:
		// close of a subscription
		for i, sub := range c.subs {
			if sub == req {
				c.subs = append(
					c.subs[0:i],
					c.subs[i+1:]...)
				close(req)
				break
			}
		}
	}
}

func (c *Cluster) closeSubscription(s *Subscriber) {
	c.reqCh <- s.inCh
}

func (c *Cluster) current() []Node {
	var r []Node
	for _, v := range c.State.Nodes {
		r = append(r, v)
	}
	sort.Sort(NodeSorter(r))
	return r
}
