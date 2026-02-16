package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	authpb "github.com/devathh/xvibe/auth-service/api/auth/v1"
	"github.com/devathh/xvibe/auth-service/internal/domain/filem"
	"github.com/devathh/xvibe/auth-service/internal/domain/session"
	"github.com/devathh/xvibe/auth-service/internal/domain/user"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/config"
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthService interface {
	Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.Token, error)
	Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.Token, error)
	Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.Token, error)
	Update(ctx context.Context, req *authpb.UpdateRequest) (*authpb.User, error)
	GetSelf(ctx context.Context) (*authpb.User, error)
	GetUserByID(ctx context.Context, req *authpb.GetByIDRequest) (*authpb.User, error)
	GetUsersByUsername(ctx context.Context, req *authpb.GetByUsernameRequest) (*authpb.Users, error)
	LogoutAll(ctx context.Context) error
	GetPublicKey(ctx context.Context) (*authpb.PublicKey, error)
}

type authService struct {
	cfg         *config.Config
	log         *slog.Logger
	userRepo    user.UserRepository
	userCache   user.UserCache
	sessionRepo session.SessionRepository
	jwtMngr     session.JwtManager
	filemRepo   filem.FilemRepository
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	userRepo user.UserRepository,
	userCache user.UserCache,
	sessionRepo session.SessionRepository,
	jwtMngr session.JwtManager,
	filemRepo filem.FilemRepository,
) AuthService {
	return &authService{
		cfg:         cfg,
		log:         log,
		userRepo:    userRepo,
		userCache:   userCache,
		sessionRepo: sessionRepo,
		jwtMngr:     jwtMngr,
		filemRepo:   filemRepo,
	}
}

// Creating a new user
// and generate a new pair of tokens (r+a)
func (as *authService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	as.log.Debug("start to register user", slog.String("email", req.Email))

	ip, err := as.getIP(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userAgent, err := as.getUserAgent(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	as.log.Debug("try to create new domain model")
	user, err := as.createNewUser(req)
	if err != nil {
		return nil, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	as.log.Debug("saving user into db", slog.String("user_id", user.ID().String()))
	savedUser, err := as.userRepo.Save(ctxTimeout, user)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		as.log.Error("failed to save user into db", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	as.log.Debug("create new user's session", slog.String("user_id", user.ID().String()))
	access, refresh, err := as.createSession(ctxTimeout, savedUser, ip, userAgent)
	if err != nil {
		return nil, err
	}

	return &authpb.Token{
		Access:           access,
		Refresh:          refresh,
		RefreshExpiresAt: timestamppb.New(time.Now().UTC().Add(as.cfg.Service.Session.RefreshTTL)),
		AccessExpiresAt:  timestamppb.New(time.Now().UTC().Add(as.cfg.Service.Session.AccessTTL)),
	}, nil
}

// Fetching a user by email
// and generate a new pair of tokens (r+a)
func (as *authService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	as.log.Debug("start to login user", slog.String("email", req.Email))

	ip, err := as.getIP(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userAgent, err := as.getUserAgent(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	as.log.Debug("try to get user from db", slog.String("email", req.Email))
	targetUser, err := as.userRepo.GetByEmail(ctxTimeout, user.Email(req.Email))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		as.log.Error("failed to get user by email", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	as.log.Debug("checking password")
	if !targetUser.PasswordHash().Compare(req.Password) {
		return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidCredentials.Error())
	}

	as.log.Debug("creating new user's session", slog.String("user_id", targetUser.ID().String()))
	access, refresh, err := as.createSession(ctxTimeout, targetUser, ip, userAgent)
	if err != nil {
		return nil, err
	}

	return &authpb.Token{
		Access:           access,
		Refresh:          refresh,
		RefreshExpiresAt: timestamppb.New(time.Now().UTC().Add(as.cfg.Service.Session.RefreshTTL)),
		AccessExpiresAt:  timestamppb.New(time.Now().UTC().Add(as.cfg.Service.Session.AccessTTL)),
	}, nil
}

// Get session by refresh token
// 1. generate a new pair of tokens (r+a)
// 2. delete old session
func (as *authService) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	as.log.Debug("start to refresh session")

	ip, err := as.getIP(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userAgent, err := as.getUserAgent(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	as.log.Debug("try to get old session")
	session, err := as.sessionRepo.Get(ctxTimeout, req.RefreshToken)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrSessionDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		as.log.Error("failed to get refresh session", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	as.log.Debug("checking fingerprint")
	if !session.Fingerprint().Compare(ip, userAgent) {
		return nil, status.Error(codes.PermissionDenied, "invalid finger print")
	}

	as.log.Debug("generate access")
	newAccess, err := as.jwtMngr.GenerateAccess(session.UserID(), session.Email())
	if err != nil {
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	as.log.Debug("generate refresh")
	newRefresh, err := as.jwtMngr.GenerateRefresh()
	if err != nil {
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	as.log.Debug("save new session")
	if err := as.sessionRepo.Save(ctxTimeout, newRefresh, session); err != nil {
		as.log.Error("failed to save new session", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	as.log.Debug("delete old session")
	if err := as.sessionRepo.Del(ctxTimeout, req.RefreshToken); err != nil {
		as.log.Error("failed to delete old session", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return &authpb.Token{
		Access:           newAccess,
		Refresh:          newRefresh,
		RefreshExpiresAt: timestamppb.New(time.Now().UTC().Add(as.cfg.Service.Session.RefreshTTL)),
		AccessExpiresAt:  timestamppb.New(time.Now().UTC().Add(as.cfg.Service.Session.AccessTTL)),
	}, nil
}

// Update fields of user
// by the update's mask
func (as *authService) Update(ctx context.Context, req *authpb.UpdateRequest) (*authpb.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userId, err := as.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidToken.Error())
	}

	as.log.Debug("start to update user", slog.String("user_id", userId.String()))

	updUser := user.ForUpdate(
		userId,
		req.Updates.Firstname,
		req.Updates.Lastname,
		user.Username(req.Updates.Username),
		user.Email(req.Updates.Email),
	)

	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	as.log.Debug("try to update fields in db", slog.String("user_id", updUser.ID().String()))
	updatedUser, err := as.userRepo.Update(
		ctxTimeout,
		updUser,
		req.UpdateMask.Paths,
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrUserAlreadyTaken) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		as.log.Error("failed to update user", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	go as.deleteUserCache(context.Background(), updatedUser.ID())
	return as.returnUser(updatedUser), nil
}

// Fetch user by id
// (user's id is taken from the token)
func (as *authService) GetSelf(ctx context.Context) (*authpb.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidToken.Error())
	}

	as.log.Debug("start to fetch user", slog.String("user_id", userID.String()))

	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	if cachedUser, err := as.userCache.Get(ctxTimeout, userID); err == nil {
		as.log.Debug("user is taked from cache")
		return as.returnUser(cachedUser), nil
	} else if !errors.Is(err, consts.ErrUserDoesntExist) {
		as.log.Error("failed to get user from cache", slog.String("error", err.Error()))
	}

	as.log.Debug("try to fetch user", slog.String("user_id", userID.String()))
	user, err := as.userRepo.GetByID(ctxTimeout, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		as.log.Error("failed to get user by id", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	go as.saveUserCache(context.Background(), user)
	return as.returnUser(user), nil
}

// Fetch user by id
// (user's id is taken from the request)
func (as *authService) GetUserByID(ctx context.Context, req *authpb.GetByIDRequest) (*authpb.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidUserID.Error())
	}

	as.log.Debug("start to fetch user by id", slog.String("user_id", userID.String()))

	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	if cachedUser, err := as.userCache.Get(ctxTimeout, userID); err == nil {
		as.log.Debug("user is taked from cache")
		return as.returnUser(cachedUser), nil
	} else if !errors.Is(err, consts.ErrUserDoesntExist) {
		as.log.Error("failed to get user from cache", slog.String("error", err.Error()))
	}

	as.log.Debug("try to get user by id", slog.String("user_id", userID.String()))
	user, err := as.userRepo.GetByID(ctxTimeout, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		as.log.Error("failed to get user by id", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	go as.saveUserCache(context.Background(), user)
	return as.returnUser(user), nil
}

// Find all users, where username like request
// (limit: 100)
func (as *authService) GetUsersByUsername(ctx context.Context, req *authpb.GetByUsernameRequest) (*authpb.Users, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	users, err := as.userRepo.GetByUsername(ctxTimeout, req.Username)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrInvalidUsername) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		as.log.Error("failed to get users by username", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	response := authpb.Users{
		Users: make([]*authpb.User, len(users)),
	}
	for idx, user := range users {
		response.Users[idx] = &authpb.User{
			Id:        user.ID().String(),
			Email:     user.Email().Value(),
			Firstname: user.Firstname(),
			Lastname:  user.Lastname(),
			Username:  user.Username().Value(),
			CreatedAt: timestamppb.New(user.CreatedAt()),
			UpdatedAt: timestamppb.New(user.UpdatedAt()),
		}
	}

	return &response, nil
}

// Clear all user's sessions
// by jwt-token of user
func (as *authService) LogoutAll(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return status.Error(codes.Canceled, err.Error())
	}

	userID, err := as.getUserID(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, consts.ErrInvalidToken.Error())
	}

	as.log.Debug("start to logout all sessions", slog.String("user_id", userID.String()))

	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	as.log.Debug("try to del all sessions")
	if err := as.sessionRepo.LogoutAll(ctxTimeout, userID); err != nil {
		as.log.Error("failed to logout all sessions", slog.String("error", err.Error()))
		return status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return nil
}

// Get public key pem
func (as *authService) GetPublicKey(ctx context.Context) (*authpb.PublicKey, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	publicKey, err := as.filemRepo.GetPublicKey(ctxTimeout)
	if err != nil {
		as.log.Error("failed to get public key", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return &authpb.PublicKey{
		Filename: publicKey.Name(),
		Content:  publicKey.Content(),
	}, nil
}

func (as *authService) saveUserCache(ctx context.Context, user *user.User) {
	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	if err := as.userCache.Save(ctxTimeout, user); err != nil {
		as.log.Error("failed to save user into cache", slog.String("user_id", user.ID().String()), slog.String("error", err.Error()))
	}
}

func (as *authService) createDomainSession(user *user.User, ip, userAgent string) (*session.Session, error) {
	fingerPrint := session.NewFingerPrint(ip, userAgent)

	return session.Create(
		user.ID(),
		user.Email().Value(),
		fingerPrint,
	)
}

func (as *authService) getIP(ctx context.Context) (string, error) {
	raw := ctx.Value(session.ClientIPKey)
	if raw == nil {
		return "", consts.ErrInvalidClientIP
	}

	if ip, ok := raw.(string); ok {
		return ip, nil
	}

	return "", consts.ErrInvalidClientIP
}

func (as *authService) getUserAgent(ctx context.Context) (string, error) {
	raw := ctx.Value(session.UserAgentKey)
	if raw == nil {
		return "", consts.ErrInvalidUserAgent
	}

	if userAgent, ok := raw.(string); ok {
		return userAgent, nil
	}

	return "", consts.ErrInvalidUserAgent
}

func (as *authService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(session.UserIDKey)
	if raw == nil {
		return uuid.Nil, consts.ErrInvalidUserID
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, consts.ErrInvalidUserID
}

func (as *authService) generateTokens(user *user.User) (string, string, error) {
	access, err := as.jwtMngr.GenerateAccess(user.ID(), user.Email().Value())
	if err != nil {
		return "", "", err
	}

	refresh, err := as.jwtMngr.GenerateRefresh()
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (as *authService) createNewUser(req *authpb.RegisterRequest) (*user.User, error) {
	passwordHash, err := user.NewPasswordHash(req.Password)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	username, err := user.NewUsername()
	if err != nil {
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	user, err := user.New(
		user.Email(req.Email),
		passwordHash,
		req.Firstname,
		req.Lastname,
		username,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return user, nil
}

func (as *authService) createSession(ctx context.Context, user *user.User, ip, userAgent string) (string, string, error) {
	access, refresh, err := as.generateTokens(user)
	if err != nil {
		as.log.Error("failed to generate tokens", slog.String("error", err.Error()))
		return "", "", status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	session, err := as.createDomainSession(user, ip, userAgent)
	if err != nil {
		as.log.Error("failed to create new session", slog.String("error", err.Error()))
		return "", "", status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	if err := as.sessionRepo.Save(ctx, refresh, session); err != nil {
		as.log.Error("failed to save session", slog.String("error", err.Error()))
		return "", "", status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return access, refresh, nil
}

func (as *authService) returnUser(user *user.User) *authpb.User {
	return &authpb.User{
		Id:        user.ID().String(),
		Email:     user.Email().Value(),
		Firstname: user.Firstname(),
		Lastname:  user.Lastname(),
		Username:  user.Username().Value(),
		CreatedAt: timestamppb.New(user.CreatedAt()),
		UpdatedAt: timestamppb.New(user.UpdatedAt()),
	}
}

func (as *authService) deleteUserCache(ctx context.Context, id uuid.UUID) {
	ctxTimeout, cancel := context.WithTimeout(ctx, as.cfg.Server.Timeout)
	defer cancel()

	if err := as.userCache.Del(ctxTimeout, id); err != nil {
		as.log.Error("failed to delete user from cache by id",
			slog.String("user_id", id.String()),
			slog.String("error", err.Error()),
		)
	}
}
