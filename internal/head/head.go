package head

import "tig/internal/commit"

type TigHead struct {
	Head *commit.TigCommit // Which version we're on
}
