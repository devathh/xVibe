package messagepg

import (
	"context"
	"errors"
	"fmt"

	"github.com/devathh/xvibe/message-service/internal/domain/message"
	"github.com/devathh/xvibe/message-service/pkg/consts"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (mr *MessageRepository) GetWrappedDEK(ctx context.Context, chatID uuid.UUID) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var key struct {
		DEK []byte `gorm:"column:wrapped_dek"`
	}

	if err := mr.db.WithContext(ctx).Table("chat_models").Where("id = ?", chatID).Take(&key).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrChatNotFound
		}

		return nil, fmt.Errorf("failed to get dek key of chat: %w", err)
	}

	return key.DEK, nil
}

func (mr *MessageRepository) Save(ctx context.Context, msg *message.Message) (*message.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	model := toModel(msg)
	if err := mr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exists bool
		if err := tx.Table("chat_members").Where("chat_id = ? AND user_id = ?", msg.ChatID(), msg.AuthorID()).
			Select("1").Take(&exists).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return consts.ErrNoEnoughRights
			}
			return err
		}

		if !exists {
			return consts.ErrNoEnoughRights
		}

		return tx.Create(&model).Error
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, consts.ErrNoEnoughRights) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to save msg: %w", err)
	}

	return toDomain(model), nil
}

func (mr *MessageRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := mr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model MessageModel
		if err := tx.Take(&model, id).Error; err != nil {
			return err
		}

		if model.AuthorID != userID {
			return consts.ErrNoEnoughRights
		}

		return tx.Where("id = ?", id).Delete(&model).Error
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		if errors.Is(err, consts.ErrNoEnoughRights) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrMsgNotFound
		}

		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (mr *MessageRepository) GetHistory(ctx context.Context, chatID, userID, beforeID uuid.UUID, limit uint32) ([]*message.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if limit > 100 {
		return nil, consts.ErrTooBigLimit
	}

	var models []MessageModel
	if err := mr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exists bool
		if err := mr.db.WithContext(ctx).Table("chat_members").
			Where("chat_id = ? AND user_id = ?", chatID, userID).
			Select("1").Take(&exists).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return consts.ErrNoEnoughRights
			}
			return err
		}

		if !exists {
			return consts.ErrNoEnoughRights
		}

		query := tx.Limit(int(limit)).Where("chat_id = ?", chatID)
		if beforeID != uuid.Nil {
			query = query.Where("id < ?", beforeID)
		}

		if err := query.Order("id DESC").Find(&models).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, consts.ErrNoEnoughRights) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to get history of chat: %w", err)
	}

	msgs := make([]*message.Message, len(models))
	for idx, model := range models {
		msgs[idx] = toDomain(&model)
	}

	return msgs, nil
}

func (mr *MessageRepository) IsUserMember(ctx context.Context, chatID, userID uuid.UUID) bool {
	if err := ctx.Err(); err != nil {
		return false
	}

	var exists bool
	if err := mr.db.WithContext(ctx).Table("chat_members").
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Select("1").Take(&exists).Error; err != nil {
		return false
	}

	return exists
}
