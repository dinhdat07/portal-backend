package portalgrpc

import (
	"context"
	"net"
	"portal-system/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	gstatus "google.golang.org/grpc/status"
)

type contextKey string

const AuditUserContextKey contextKey = "audit_user"
const SessionIDContextKey contextKey = "session_id"

func getAuditFromCtx(ctx context.Context) *domain.AuditMeta {
	meta := &domain.AuditMeta{}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err == nil {
			meta.IPAddress = host
		} else {
			meta.IPAddress = p.Addr.String()
		}
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ua := md.Get("user-agent")
		if len(ua) > 0 {
			meta.UserAgent = ua[0]
		}
	}

	return meta
}

func getActorFromCtx(ctx context.Context) (*domain.AuditUser, error) {
	v := ctx.Value(AuditUserContextKey)
	if v == nil {
		return nil, gstatus.Error(codes.Unauthenticated, "missing authenticated user")
	}

	actor, ok := v.(*domain.AuditUser)
	if !ok || actor == nil {
		return nil, gstatus.Error(codes.Unauthenticated, "invalid authenticated user")
	}

	return actor, nil
}

func getSessionIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(SessionIDContextKey)
	if v == nil {
		return uuid.Nil, gstatus.Error(codes.Unauthenticated, "missing session id")
	}

	switch sessionID := v.(type) {
	case uuid.UUID:
		if sessionID == uuid.Nil {
			return uuid.Nil, gstatus.Error(codes.Unauthenticated, "invalid session id")
		}
		return sessionID, nil
	case *uuid.UUID:
		if sessionID == nil || *sessionID == uuid.Nil {
			return uuid.Nil, gstatus.Error(codes.Unauthenticated, "invalid session id")
		}
		return *sessionID, nil
	case string:
		id, err := uuid.Parse(sessionID)
		if err != nil {
			return uuid.Nil, gstatus.Error(codes.Unauthenticated, "invalid session id")
		}
		return id, nil
	default:
		return uuid.Nil, gstatus.Error(codes.Unauthenticated, "invalid session id")
	}
}
