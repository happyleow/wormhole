package publicrpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/certusone/wormhole/node/pkg/db"
	publicrpcv1 "github.com/certusone/wormhole/node/pkg/proto/publicrpc/v1"
	"github.com/certusone/wormhole/node/pkg/vaa"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PublicrpcServer implements the publicrpc gRPC service.
type PublicrpcServer struct {
	publicrpcv1.UnsafePublicRPCServiceServer
	logger *zap.Logger
	db     *db.Database
	gst    *common.GuardianSetState
}

func NewPublicrpcServer(
	logger *zap.Logger,
	db *db.Database,
	gst *common.GuardianSetState,
) *PublicrpcServer {
	return &PublicrpcServer{
		logger: logger.Named("publicrpcserver"),
		db:     db,
		gst:    gst,
	}
}

func (s *PublicrpcServer) GetLastHeartbeats(ctx context.Context, req *publicrpcv1.GetLastHeartbeatsRequest) (*publicrpcv1.GetLastHeartbeatsResponse, error) {
	gs := s.gst.Get()
	if gs == nil {
		return nil, status.Error(codes.Unavailable, "guardian set not fetched from chain yet")
	}

	resp := &publicrpcv1.GetLastHeartbeatsResponse{
		Entries: make([]*publicrpcv1.GetLastHeartbeatsResponse_Entry, 0),
	}

	// Fetch all heartbeats (including from nodes not in the guardian set - which
	// can happen either with --disableHeartbeatVerify or when the guardian set changes)
	for addr, v := range s.gst.GetAll() {
		for peerId, hb := range v {
			resp.Entries = append(resp.Entries, &publicrpcv1.GetLastHeartbeatsResponse_Entry{
				VerifiedGuardianAddr: addr.Hex(),
				P2PNodeAddr:          peerId.Pretty(),
				RawHeartbeat:         hb,
			})
		}
	}

	return resp, nil
}

func (s *PublicrpcServer) GetSignedVAA(ctx context.Context, req *publicrpcv1.GetSignedVAARequest) (*publicrpcv1.GetSignedVAAResponse, error) {
	address, err := hex.DecodeString(req.MessageId.EmitterAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to decode address: %v", err))
	}
	if len(address) != 32 {
		return nil, status.Error(codes.InvalidArgument, "address must be 32 bytes")
	}
	if req.MessageId == nil {
		return nil, status.Error(codes.InvalidArgument, "no message ID specified")
	}

	addr := vaa.Address{}
	copy(addr[:], address)

	b, err := s.db.GetSignedVAABytes(db.VAAID{
		EmitterChain:   vaa.ChainID(req.MessageId.EmitterChain.Number()),
		EmitterAddress: addr,
		Sequence:       uint64(req.MessageId.Sequence),
	})

	if err != nil {
		if err == db.ErrVAANotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		s.logger.Error("failed to fetch VAA", zap.Error(err), zap.Any("request", req))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &publicrpcv1.GetSignedVAAResponse{
		VaaBytes: b,
	}, nil
}

func (s *PublicrpcServer) GetCurrentGuardianSet(ctx context.Context, req *publicrpcv1.GetCurrentGuardianSetRequest) (*publicrpcv1.GetCurrentGuardianSetResponse, error) {
	gs := s.gst.Get()
	if gs == nil {
		return nil, status.Error(codes.Unavailable, "guardian set not fetched from chain yet")
	}

	resp := &publicrpcv1.GetCurrentGuardianSetResponse{
		GuardianSet: &publicrpcv1.GuardianSet{
			Index:     gs.Index,
			Addresses: make([]string, len(gs.Keys)),
		},
	}

	for i, v := range gs.Keys {
		resp.GuardianSet.Addresses[i] = v.Hex()
	}

	return resp, nil
}
