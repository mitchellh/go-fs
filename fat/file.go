package fat

import ()

type File struct {
	chain *ClusterChain
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.chain.Read(p)
}
