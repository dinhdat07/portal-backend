package portalgrpc

import (
	"context"
	"errors"
	"net"
	"portal-system/internal/auth"
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
const principalContextKey contextKey = "principal"

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
	principal, exists := GetPrincipal(ctx)
	if principal == nil || !exists {
		return nil, errors.New("missing principal in context")
	}

	return &domain.AuditUser{
		ID:       principal.UserID,
		Username: principal.Username,
		Email:    principal.Email,
		RoleCode: principal.RoleCode,
	}, nil
}

func getSessionIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	principal, exists := GetPrincipal(ctx)
	if principal == nil || !exists {
		return uuid.Nil, errors.New("missing principal in context")
	}

	if principal.SessionID == uuid.Nil {
		return uuid.Nil, gstatus.Error(codes.Unauthenticated, "missing session id")
	}

	return principal.SessionID, nil
}

func SetPrincipal(ctx context.Context, principal *auth.Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func GetPrincipal(ctx context.Context) (*auth.Principal, bool) {
	v := ctx.Value(principalContextKey)
	if v == nil {
		return nil, false
	}

	principal, ok := v.(*auth.Principal)
	return principal, ok
}
