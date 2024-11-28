package tighistory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"tig/internal/tigfile"
)

// NTree is an N-ary tree structure. First child is the main child (the root branch)
type NTree[T any] struct {
	Parent *NTree[T]  `json:"-"` // Can't store parent because of cyclic json marshalling
	Childs []NTree[T] `json:"childs"`
	Value  T          `json:"value"`
}

func New[T any]() NTree[T] {
	return NTree[T]{}
}

func (tree *NTree[T]) Add(value T) {
	tree.Childs = append(tree.Childs, NTree[T]{Parent: tree, Value: value})
}

func (tree *NTree[T]) GetMainChild(value T) *NTree[T] {
	if len(tree.Childs) > 0 {
		return &tree.Childs[0]
	} else {
		return nil
	}
}

func (tree *NTree[T]) Save(filepath string) error {
	b, err := json.Marshal(tree)
	if err != nil {
		return fmt.Errorf("tree:Save(): %w", err)
	}
	err = tigfile.WriteFileBytes(filepath, b)
	if err != nil {
		return fmt.Errorf("tree:Save(): %w", err)
	}
	return nil
}

func (tree *NTree[T]) Load(filepath string) error {
	b, err := tigfile.ReadFileBytes(filepath, -1)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("tree:Load(): %w", err)
	}
	err = json.Unmarshal(b, tree)
	if err != nil {
		return fmt.Errorf("tree:Load(): %w", err)
	}
	// Restore parenting
	ptr := tree
	for {
		for i := 0; i < len(ptr.Childs); i++ {
			ptr.Childs[i].Parent = ptr
		}
		break
	}
	return nil
}
