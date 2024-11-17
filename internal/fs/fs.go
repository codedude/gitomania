package fs

type TigFileSnapshot struct {
	Hash     string           // Content hash of the file at snapshot
	Path     string           // Path of the snapshot in Tig (based on hash)
	File     *TigFile         // Never nil
	Previous *TigFileSnapshot // Never nil
}

type TigFile struct {
	Path   string           // Path of the file in the client project
	Latest *TigFileSnapshot // Never nil
	Oldest *TigFileSnapshot // Never nil
}
