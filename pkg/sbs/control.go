package sbs

import "github.com/tcfw/vpc/pkg/sbs/control"

func (s *Server) watchController(once bool) {
	attachments := s.controller.AttachmentsCh(true)
	volumes := s.controller.VolumesCh(true)

	for {
		select {
		case <-s.shutdownCh:
			return
		case attachAction := <-attachments:
			s.handleAttachmentUpdate(attachAction)
			break
		case volAction := <-volumes:
			s.handleVolumeUpdate(volAction)
			break
		}

		if once {
			break
		}
	}
}

func (s *Server) handleVolumeUpdate(action control.VolumeAction) {
	switch action.Op {
	case control.OpAdd:
		if _, ok := s.volumes[action.Desc.Id]; ok {
			return
		}
		s.log.WithField("vol", action.Desc.Id).Info("got new volume")
		if err := s.AddVolume(action.Desc, action.Peers); err != nil {
			s.log.WithField("vol", action.Desc.Id).WithError(err).Error("failed to add new volume")
		}
	}
}

func (s *Server) handleAttachmentUpdate(action control.VolumeAction) {
	s.log.Infof("got new attachment %s", action.Desc.Id)

	if s.nbd == nil {
		s.nbd = NewNBDServer(s, NBDDefaultPort, s.controller)
	}

	peers := s.GetPeers(action.Peers)

	if vol, isLocal := s.volumes[action.Desc.Id]; isLocal {
		err := s.nbd.Attach(vol, peers)
		if err != nil {
			s.log.WithError(err).Error("failed to attach volume")
		}

		return
	}

	//add a half configured volume
	vol := &Volume{
		id:             action.Desc.Id,
		PlacementPeers: map[string]struct{}{},
		desc:           action.Desc,
	}

	err := s.nbd.Attach(vol, peers)
	if err != nil {
		s.log.WithError(err).Error("failed to attach volume")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.volumes[vol.id] = vol
}
