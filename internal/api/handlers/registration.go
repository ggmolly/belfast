package handlers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	registrationPinPrefix = "B-"
)

var (
	registrationPinTTL      = 5 * time.Minute
	registrationRateLimit   = 5
	registrationRateWindow  = time.Minute
	registrationPinAttempts = 5
)

type RegistrationHandler struct {
	Config   config.AuthConfig
	Limiter  *auth.RateLimiter
	Validate *validator.Validate
}

func NewRegistrationHandler(cfg *config.Config) *RegistrationHandler {
	authCfg := auth.NormalizeConfig(config.AuthConfig{})
	if cfg != nil {
		authCfg = auth.NormalizeConfig(cfg.Auth)
	}
	return &RegistrationHandler{
		Config:   authCfg,
		Limiter:  auth.NewRateLimiter(),
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func RegisterRegistrationRoutes(party iris.Party, handler *RegistrationHandler) {
	party.Post("/challenges", handler.CreateChallenge)
	party.Get("/challenges/{id}", handler.GetChallengeStatus)
	party.Post("/challenges/{id}/verify", handler.VerifyChallenge)
}

// UserRegistrationChallenge godoc
// @Summary     Create registration challenge
// @Tags        Registration
// @Accept      json
// @Produce     json
// @Param       body  body  types.UserRegistrationChallengeRequest  true  "Challenge payload"
// @Success     200  {object}  UserRegistrationChallengeResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     429  {object}  APIErrorResponseDoc
// @Router      /api/v1/registration/challenges [post]
func (handler *RegistrationHandler) CreateChallenge(ctx iris.Context) {
	var req types.UserRegistrationChallengeRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "validation failed", validationErrors(err)))
		return
	}
	ip := auth.NormalizeIP(ctx.RemoteAddr())
	key := strings.Join([]string{"registration", ip, strconv.FormatUint(uint64(req.CommanderID), 10)}, ":")
	if !handler.Limiter.Allow(key, registrationRateLimit, registrationRateWindow) {
		ctx.StatusCode(iris.StatusTooManyRequests)
		_ = ctx.JSON(response.Error("auth.rate_limited", "too many attempts", nil))
		return
	}

	var commander orm.Commander
	if err := orm.GormDB.First(&commander, "commander_id = ?", req.CommanderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "player not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load player", nil))
		return
	}

	passwordHash, algo, err := auth.HashPassword(req.Password, handler.Config)
	if err != nil {
		if errors.Is(err, auth.ErrPasswordTooShort) {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("auth.password_too_short", "password too short", nil))
			return
		}
		if errors.Is(err, auth.ErrPasswordTooLong) {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("auth.password_too_long", "password too long", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to hash password", nil))
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(registrationPinTTL)
	var challenge *orm.UserRegistrationChallenge
	for i := 0; i < registrationPinAttempts; i++ {
		pin, err := generateRegistrationPin()
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to generate pin", nil))
			return
		}
		entry, err := orm.CreateUserRegistrationChallenge(req.CommanderID, pin, passwordHash, algo, expiresAt, now)
		if err != nil {
			switch {
			case errors.Is(err, orm.ErrRegistrationPinExists):
				continue
			case errors.Is(err, orm.ErrUserAccountExists):
				ctx.StatusCode(iris.StatusConflict)
				_ = ctx.JSON(response.Error("auth.account_exists", "account already exists", nil))
				return
			default:
				ctx.StatusCode(iris.StatusInternalServerError)
				_ = ctx.JSON(response.Error("internal_error", "failed to create challenge", nil))
				return
			}
		}
		challenge = entry
		break
	}
	if challenge == nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to generate unique pin", nil))
		return
	}

	pinValue := fmt.Sprintf("%s%s", registrationPinPrefix, challenge.Pin)
	mail := orm.Mail{
		ReceiverID: commander.CommanderID,
		Title:      "Registration PIN",
		Body:       fmt.Sprintf("Your registration PIN is %s. It expires at %s.", pinValue, challenge.ExpiresAt.UTC().Format(time.RFC3339)),
	}
	if err := commander.SendMail(&mail); err != nil {
		_ = orm.GormDB.Model(&orm.UserRegistrationChallenge{}).Where("id = ?", challenge.ID).
			Update("status", orm.UserRegistrationStatusExpired).Error
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to send registration mail", nil))
		return
	}
	notifyMailboxUpdate(commander.CommanderID)

	payload := types.UserRegistrationChallengeResponse{
		ChallengeID: challenge.ID,
		ExpiresAt:   challenge.ExpiresAt.UTC().Format(time.RFC3339),
	}
	_ = ctx.JSON(response.Success(payload))
}

// UserRegistrationStatus godoc
// @Summary     Get registration challenge status
// @Tags        Registration
// @Produce     json
// @Param       id  path  string  true  "Challenge ID"
// @Success     200  {object}  UserRegistrationStatusResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/registration/challenges/{id} [get]
func (handler *RegistrationHandler) GetChallengeStatus(ctx iris.Context) {
	id := ctx.Params().Get("id")
	challenge, err := orm.GetUserRegistrationChallenge(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "challenge not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load challenge", nil))
		return
	}
	if challenge.Status == orm.UserRegistrationStatusPending && !challenge.ExpiresAt.After(time.Now().UTC()) {
		_ = orm.GormDB.Model(&orm.UserRegistrationChallenge{}).Where("id = ?", challenge.ID).Update("status", orm.UserRegistrationStatusExpired).Error
		challenge.Status = orm.UserRegistrationStatusExpired
	}
	payload := types.UserRegistrationStatusResponse{Status: challenge.Status}
	_ = ctx.JSON(response.Success(payload))
}

// UserRegistrationVerify godoc
// @Summary     Verify registration challenge
// @Tags        Registration
// @Accept      json
// @Produce     json
// @Param       id  path  string  true  "Challenge ID"
// @Param       body  body  types.UserRegistrationVerifyRequest  true  "Verify payload"
// @Success     200  {object}  UserRegistrationStatusResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/registration/challenges/{id}/verify [post]
func (handler *RegistrationHandler) VerifyChallenge(ctx iris.Context) {
	id := ctx.Params().Get("id")
	if id == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "challenge id required", nil))
		return
	}
	var req types.UserRegistrationVerifyRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "validation failed", validationErrors(err)))
		return
	}
	pin, ok := normalizeRegistrationPin(req.Pin)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("auth.challenge_invalid", "invalid pin", nil))
		return
	}
	account, err := orm.ConsumeUserRegistrationChallengeByID(id, pin, time.Now().UTC())
	if err != nil {
		switch {
		case errors.Is(err, orm.ErrRegistrationChallengeNotFound):
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "challenge not found", nil))
		case errors.Is(err, orm.ErrRegistrationChallengeExpired):
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("auth.challenge_expired", "challenge expired", nil))
		case errors.Is(err, orm.ErrRegistrationChallengePinMismatch):
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("auth.challenge_invalid", "invalid pin", nil))
		case errors.Is(err, orm.ErrRegistrationChallengeConsumed):
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("auth.challenge_consumed", "challenge already used", nil))
		case errors.Is(err, orm.ErrUserAccountExists):
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("auth.account_exists", "account already exists", nil))
		default:
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to consume challenge", nil))
		}
		return
	}
	auth.LogUserAudit("registration.consume", &account.ID, account.CommanderID, nil)
	payload := types.UserRegistrationStatusResponse{Status: orm.UserRegistrationStatusConsumed}
	_ = ctx.JSON(response.Success(payload))
}

func generateRegistrationPin() (string, error) {
	max := big.NewInt(1000000)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func normalizeRegistrationPin(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, registrationPinPrefix) {
		trimmed = strings.TrimPrefix(trimmed, registrationPinPrefix)
	}
	if len(trimmed) != 6 {
		return "", false
	}
	for _, ch := range trimmed {
		if ch < '0' || ch > '9' {
			return "", false
		}
	}
	return trimmed, true
}

func notifyMailboxUpdate(commanderID uint32) {
	server := connection.BelfastInstance
	if server == nil {
		return
	}
	client, ok := server.FindClientByCommander(commanderID)
	if !ok {
		return
	}
	var totalCount int64
	if err := orm.GormDB.Model(&orm.Mail{}).Where("receiver_id = ? AND is_archived = ?", commanderID, false).Count(&totalCount).Error; err != nil {
		return
	}
	var unreadCount int64
	if err := orm.GormDB.Model(&orm.Mail{}).Where("receiver_id = ? AND is_archived = ? AND read = ?", commanderID, false, false).Count(&unreadCount).Error; err != nil {
		return
	}
	payload := protobuf.SC_30001{
		UnreadNumber: proto.Uint32(uint32(unreadCount)),
		TotalNumber:  proto.Uint32(uint32(totalCount)),
	}
	_, _, _ = client.SendMessage(30001, &payload)
}
