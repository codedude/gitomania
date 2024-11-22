package head

import "tig/internal/tgcommit"

type TigHead struct {
	Head *tgcommit.TigCommit // Which version we're on
}
