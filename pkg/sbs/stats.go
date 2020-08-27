package sbs

import (
	"syscall"

	"github.com/tcfw/vpc/pkg/sbs/config"
)

//BlockStoreStats provides OS status on the block store/volume
type BlockStoreStats struct {
	Available uint64
	Used      uint64
	Allocated uint64
}

//BlockStoreStats provides status for the block store overall
func (s *Server) BlockStoreStats() (*BlockStoreStats, error) {
	stats := &BlockStoreStats{}

	//Allocated space from volume sizes
	for _, vol := range s.volumes {
		volSize, err := vol.Blocks.SizeOnDisk()
		if err != nil {
			return nil, err
		}

		stats.Allocated += uint64(vol.Size() * VolDescSizeMultiplier)
		stats.Used += volSize
	}

	available, err := s.blockStoreAvailability()
	if err == nil {
		stats.Available = available
	}

	return stats, nil
}

//blockStoreAvailability gets the OS' interpretation of available space
//for the block store dir
func (s *Server) blockStoreAvailability() (uint64, error) {
	var fStats syscall.Statfs_t
	if err := syscall.Statfs(config.BlockStoreDir(), &fStats); err != nil {
		return 0, err
	}

	return uint64(fStats.Bsize) * fStats.Bavail, nil
}

//BlockVolumeStats provides status for specific volumes
func (s *Server) BlockVolumeStats(vol *Volume) (*BlockStoreStats, error) {
	stats := &BlockStoreStats{}

	stats.Allocated = uint64(vol.Size() * VolDescSizeMultiplier)
	available, err := s.blockStoreAvailability()
	if err == nil {
		stats.Available = available
	}
	used, err := vol.Blocks.SizeOnDisk()
	if err != nil {
		return nil, err
	}
	stats.Used = used

	return stats, nil
}
