package datastructs

import (
	"fmt"
)

// GraphScan is used to collect information while scanning a graph
type GraphScan[T comparable] struct {
	visited  map[T]bool
	order    []T
	tempMark map[T]bool // for cycle detection
}

// DependenciesGraph is a directed graph that can be used for ordering tasks that depend
// on one another. for example: task A depends on B, B depends on C, etc.
type DependenciesGraph[T comparable] struct {
	adjList map[T][]T
}

func NewDependenciesGraph[T comparable]() *DependenciesGraph[T] {
	return &DependenciesGraph[T]{
		adjList: make(map[T][]T),
	}
}

// AddDependency adds a directed edge from a node to its preRequisite
func (g *DependenciesGraph[T]) AddDependency(node, preRequisite T) {
	g.adjList[node] = append(g.adjList[node], preRequisite)
}

// FindOrder perform DFS scan of the graph. This returns=≠ the necessary order to make sure
// that when we process a node all its preRequisites already processed
func (g *DependenciesGraph[T]) FindOrder() ([]T, error) {
	scan := GraphScan[T]{
		visited:  make(map[T]bool),
		order:    make([]T, 0, len(g.adjList)),
		tempMark: make(map[T]bool),
	}

	for node := range g.adjList {
		if !scan.visited[node] {
			if err := g.visit(node, &scan); err != nil {
				return nil, err
			}
		}
	}
	return scan.order, nil
}

// visit function for DFS with cycle detection using temporary marks
func (g *DependenciesGraph[T]) visit(node T, scan *GraphScan[T]) error {
	if scan.tempMark[node] {
		return fmt.Errorf("cycle detected: %v", node)
	}
	if scan.visited[node] {
		return nil
	}

	// Mark this node temporarily
	scan.tempMark[node] = true

	for _, neighbor := range g.adjList[node] {
		if err := g.visit(neighbor, scan); err != nil {
			return err
		}
	}

	// Remove temporary mark and mark as fully visited
	scan.tempMark[node] = false
	scan.visited[node] = true

	// Add to order
	scan.order = append(scan.order, node)
	return nil
}
