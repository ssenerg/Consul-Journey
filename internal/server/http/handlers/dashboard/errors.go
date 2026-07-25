package dashboard

import "consul-journey/internal/errors"

// -----------
// Node errors
// -----------

const (
	ErrPeerIDRequiredCode = "PEER_ID_REQUIRED"
	ErrPeerNotFoundCode   = "PEER_NOT_FOUND"
)

var (
	ErrPeerIDRequired = errors.New(400, ErrPeerIDRequiredCode, "Peer ID is required")
	ErrPeerNotFound   = errors.New(404, ErrPeerNotFoundCode, "Peer not found")
)
