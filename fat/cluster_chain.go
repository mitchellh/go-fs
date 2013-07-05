package fat

import (
	"fmt"
	"github.com/mitchellh/go-fs"
	"math"
)

type ClusterChain struct {
	device       fs.BlockDevice
	fat          *FAT
	startCluster uint32
	writeOffset  uint32
}

// Write will write to the cluster chain, expanding it if necessary.
func (c *ClusterChain) Write(p []byte) (n int, err error) {
	bpc := c.fat.bs.BytesPerCluster()
	chain := c.fat.Chain(c.startCluster)
	chainLength := uint32(len(chain)) * bpc

	if chainLength < c.writeOffset+uint32(len(p)) {
		// We need to grow the chain
		bytesNeeded := (c.writeOffset + uint32(len(p))) - chainLength
		clustersNeeded := int(math.Ceil(float64(bytesNeeded) / float64(bpc)))
		chain, err = c.fat.ResizeChain(c.startCluster, len(chain)+clustersNeeded)
		if err != nil {
			return
		}

		// Write the FAT out
		if err = c.fat.WriteToDevice(c.device); err != nil {
			return
		}
	}

	dataOffset := uint32(0)
	for dataOffset < uint32(len(p)) {
		chainIdx := c.writeOffset / bpc
		clusterOffset := c.fat.bs.ClusterOffset(int(chain[chainIdx]))
		clusterOffset += c.writeOffset % bpc
		dataOffsetEnd := dataOffset + bpc
		dataOffsetEnd -= c.writeOffset % bpc
		dataOffsetEnd = uint32(math.Min(float64(dataOffsetEnd), float64(len(p))))

		var nw int
		fmt.Printf("CHAIN IDX: %d\n", chainIdx)
		fmt.Printf("WRITING CLUSTER: %d at %d\n", chain[chainIdx], clusterOffset)
		fmt.Printf("WRITING DATA: %d : %d\n", dataOffset, dataOffsetEnd)
		nw, err = c.device.WriteAt(p[dataOffset:dataOffsetEnd], int64(clusterOffset))
		if err != nil {
			return
		}

		c.writeOffset += uint32(nw)
		dataOffset += uint32(nw)
		n += nw
	}

	return
}
