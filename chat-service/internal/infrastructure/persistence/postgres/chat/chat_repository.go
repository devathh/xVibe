package chatpg

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/devathh/xvibe/chat/internal/domain/chat"
	"github.com/devathh/xvibe/chat/internal/domain/member"
	"github.com/devathh/xvibe/chat/pkg/consts"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChatRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (cr *ChatRepository) Save(
	ctx context.Context,
	chat *chat.ChatModel,
	memberIds []uuid.UUID,
) (*chat.ChatModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	model := toModelChatModel(chat)
	members := toModelMembers(memberIds, chat.ID())

	if err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(model).Error; err != nil {
			return err
		}

		var ids int
		if err := tx.Table("user_models").Where("id IN ?", memberIds).Select("count(*)").Find(&ids).Error; err != nil {
			return err
		}

		if ids != len(memberIds) {
			return consts.ErrUsersDontExist
		}

		if err := tx.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(members).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, consts.ErrUsersDontExist) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to save chat: %w", err)
	}

	return toDomainChatModel(model), nil
}

func (cr *ChatRepository) Delete(
	ctx context.Context,
	chatID uuid.UUID,
	userID uuid.UUID,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var model ChatModel
	if err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&model, chatID).Error; err != nil {
			return err
		}

		if model.OwnerID != userID && model.TypeID != int(chat.TYPE_SELF) {
			return consts.ErrUserIsntOwner
		}

		if err := tx.Delete(&model, "id = ?", chatID).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		if errors.Is(err, consts.ErrUserIsntOwner) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrChatNotFound
		}

		return fmt.Errorf("failed to delete chat: %w", err)
	}

	return nil
}

func (cr *ChatRepository) Update(
	ctx context.Context,
	chatID uuid.UUID,
	userID uuid.UUID,
	title string,
) (*chat.ChatModel, []*member.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	var (
		model     ChatModel
		memberIds []uuid.UUID
		members   []UserMember
	)

	if err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&model, chatID).Error; err != nil {
			return err
		}

		if model.OwnerID != userID || model.TypeID != int(chat.TYPE_GROUP) {
			return consts.ErrUserIsntOwner
		}

		if err := tx.Model(&model).Where("id = ?", chatID).Update("title", strings.TrimSpace(title)).Error; err != nil {
			return err
		}

		if err := tx.Model(&ChatMembers{}).Where("chat_id = ?", chatID).Select("user_id").Find(&memberIds).Error; err != nil {
			return err
		}

		if err := tx.Table("user_models").Where("id IN ?", memberIds).Find(&members).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, consts.ErrChatNotFound
		}

		if errors.Is(err, consts.ErrUserIsntOwner) {
			return nil, nil, err
		}

		return nil, nil, fmt.Errorf("failed to update chat: %w", err)
	}

	return toDomainChatModel(&model), toDomainUserMembers(members, model.OwnerID), nil
}

func (cr *ChatRepository) AddMembers(
	ctx context.Context,
	memberIds []uuid.UUID,
	chatID uuid.UUID,
) (*chat.ChatModel, []*member.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	var (
		model            ChatModel
		memberModelIds   []uuid.UUID
		memberUserModels []UserMember
	)

	members := toModelMembers(memberIds, chatID)
	if err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&model, chatID).Error; err != nil {
			return err
		}

		if err := tx.Model(&ChatMembers{}).Where("chat_id = ?", chatID).Select("user_id").Find(&memberModelIds).Error; err != nil {
			return err
		}

		if err := tx.Table("user_models").Where("id IN ?", memberModelIds).Find(&memberUserModels).Error; err != nil {
			return err
		}

		if len(memberUserModels) != len(memberModelIds) {
			return consts.ErrUsersDontExist
		}

		return tx.Clauses(clause.OnConflict{
			DoNothing: true,
		}).CreateInBatches(members, 100).Error
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, consts.ErrChatNotFound
		}

		if errors.Is(err, consts.ErrUsersDontExist) {
			return nil, nil, err
		}

		return nil, nil, fmt.Errorf("failed to add new members into chat: %w", err)
	}

	return toDomainChatModel(&model), toDomainUserMembers(memberUserModels, model.OwnerID), nil
}

func (cr *ChatRepository) DeleteMembers(
	ctx context.Context,
	chatID uuid.UUID,
	memberIds []uuid.UUID,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model ChatModel
		if err := tx.Select("owner_id").Where("id = ?", chatID).First(&model).Error; err != nil {
			return err
		}

		if slices.Contains(memberIds, model.OwnerID) {
			return consts.ErrDeleteOwner
		}

		return tx.Where("chat_id = ? AND user_id IN ?", chatID, memberIds).
			Delete(&ChatMembers{}).Error
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrChatNotFound
		}

		if errors.Is(err, consts.ErrDeleteOwner) {
			return err
		}

		return fmt.Errorf("failed to delete members: %w", err)
	}

	return nil
}

func (cr *ChatRepository) GetSelfChats(
	ctx context.Context,
	userID uuid.UUID,
) ([]*chat.ChatModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var models []ChatModel
	if err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []uuid.UUID
		if err := tx.Model(&ChatMembers{}).Limit(1000).Where("user_id = ?", userID).Select("chat_id").Find(&ids).Error; err != nil {
			return err
		}

		if err := tx.Where("id IN ?", ids).Order("created_at DESC").Find(&models).Error; err != nil {
			return err
		}

		return cr.fillSelfChats(tx, models, userID)
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to get self chats: %w", err)
	}

	chats := make([]*chat.ChatModel, len(models))
	for idx, chat := range models {
		chats[idx] = toDomainChatModel(&chat)
	}

	return chats, nil
}

func (cr *ChatRepository) fillSelfChats(tx *gorm.DB, models []ChatModel, userID uuid.UUID) error {
	chatIDs := make([]uuid.UUID, len(models))
	chatMap := make(map[uuid.UUID]*ChatModel, len(models))
	for i, c := range models {
		if c.TypeID != int(chat.TYPE_SELF) {
			continue
		}

		chatIDs[i] = c.ID
		chatMap[c.ID] = &models[i]
	}

	type ChatPartner struct {
		ChatID uuid.UUID `gorm:"column:chat_id"`
		UserID uuid.UUID `gorm:"column:user_id"`
	}

	var partners []ChatPartner
	err := tx.Table("chat_members").
		Select("chat_id, user_id").
		Where("chat_id IN ? AND user_id != ?", chatIDs, userID).
		Find(&partners).Error
	if err != nil {
		return err
	}

	partnerUserIDs := make([]uuid.UUID, len(partners))
	chatToPartner := make(map[uuid.UUID]uuid.UUID, len(partners))
	for i, p := range partners {
		partnerUserIDs[i] = p.UserID
		chatToPartner[p.ChatID] = p.UserID
	}

	type UserNames struct {
		ID        uuid.UUID `gorm:"column:id"`
		FirstName string    `gorm:"column:firstname"`
	}

	var names []UserNames
	err = tx.Table("user_models").
		Select("id, firstname").
		Where("id IN ?", partnerUserIDs).
		Find(&names).Error
	if err != nil {
		return err
	}

	userIDToName := make(map[uuid.UUID]string, len(names))
	for _, n := range names {
		userIDToName[n.ID] = n.FirstName
	}

	for chatID, partnerID := range chatToPartner {
		if chat, ok := chatMap[chatID]; ok {
			if name, exists := userIDToName[partnerID]; exists {
				chat.Title = name
			}
		}
	}

	return nil
}

func (cr *ChatRepository) GetChat(
	ctx context.Context,
	chatID uuid.UUID,
) (*chat.ChatModel, []*member.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	var (
		model     ChatModel
		memberIds []uuid.UUID
		members   []UserMember
	)

	if err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&model, chatID).Error; err != nil {
			return err
		}

		if err := tx.Model(&ChatMembers{}).Where("chat_id = ?", chatID).Pluck("user_id", &memberIds).Error; err != nil {
			return err
		}

		return tx.Table("user_models").Where("id IN ?", memberIds).Find(&members).Error
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, consts.ErrChatNotFound
		}

		return nil, nil, fmt.Errorf("failed to get chat with members: %w", err)
	}

	return toDomainChatModel(&model), toDomainUserMembers(members, model.OwnerID), nil
}
