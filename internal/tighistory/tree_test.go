package tighistory

import (
	"path"
	"testing"
	"time"
)

func TestTree(t *testing.T) {
	var err error
	tmpDirPath := t.TempDir()
	treePath := path.Join(tmpDirPath, "tree_test")

	treeOld := New[*TigCommit]()
	commit := &TigCommit{Author: "val", Msg: "Hello world", Date: time.Now().Unix()}
	treeOld.Add(commit)
	err = treeOld.Save(treePath)
	if err != nil {
		t.Fatalf("tree.Save(): %s", err)
	}

	treeNew := New[*TigCommit]()
	err = treeNew.Load(treePath)
	if err != nil {
		t.Fatalf("tree.Load(): %s", err)
	}

	if len(treeOld.Childs) != len(treeNew.Childs) {
		t.Fatalf("trees differ in size: %d", len(treeNew.Childs))
	}

	if treeNew.Childs[0].Parent != &treeNew {
		t.Fatalf("tree has wrong parent after umarshalling: %d", len(treeNew.Childs))
	}

}
