package repo

import (
	"database/sql"
	"time"

	"github.com/ansel1/merry"
	"github.com/jmoiron/sqlx"

	"github.com/mreider/koto/backend/common"
)

var (
	ErrMessageNotFound = common.ErrNotFound.WithMessage("message not found")

	maxTimestamp        = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.Local)
	defaultMessageCount = 10
)

type Message struct {
	ID                    string         `json:"id" db:"id"`
	ParentID              sql.NullString `json:"parent_id" db:"parent_id"`
	UserID                string         `json:"user_id" db:"user_id"`
	UserName              string         `json:"user_name" db:"user_name"`
	Text                  string         `json:"text" db:"text"`
	AttachmentID          string         `json:"attachment_id" db:"attachment_id"`
	AttachmentType        string         `json:"attachment_type" db:"attachment_type"`
	AttachmentThumbnailID string         `json:"attachment_thumbnail_id" db:"attachment_thumbnail_id"`
	CreatedAt             time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at" db:"updated_at"`
	Likes                 int            `json:"likes" db:"likes"`
	LikedByMe             bool           `json:"liked_by_me" db:"liked_by_me"`
}

type MessageLike struct {
	MessageID string    `json:"message_id" db:"message_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	UserName  string    `json:"user_name" db:"user_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type MessageRepo interface {
	Messages(currentUserID string, userIDs []string, from time.Time, count int) ([]Message, error)
	Message(currentUserID string, messageID string) (Message, error)
	AddMessage(parentID string, message Message) error
	EditMessageText(userID, messageID, text string, updatedAt time.Time) error
	EditMessageAttachment(userID, messageID, attachmentID, attachmentType, attachmentThumbnailID string, updatedAt time.Time) error
	DeleteMessage(userID, messageID string) error
	Comments(currentUserID string, messageIDs []string) (map[string][]Message, error)
	LikeMessage(userID, messageID string) (likes int, err error)
	MessagesLikes(messageIDs []string) (likes map[string][]MessageLike, err error)
	MessageLikes(messageID string) (likes []MessageLike, err error)
	SetMessageVisibility(userID, messageID string, visibility bool) error
}

type messageRepo struct {
	db *sqlx.DB
}

func NewMessages(db *sqlx.DB) MessageRepo {
	return &messageRepo{
		db: db,
	}
}

func (r *messageRepo) Messages(currentUserID string, userIDs []string, from time.Time, count int) ([]Message, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	if from.IsZero() {
		from = maxTimestamp
	}
	if count <= 0 {
		count = defaultMessageCount
	}

	var messages []Message
	query, args, err := sqlx.In(`
			select id, parent_id, user_id, user_name, text, attachment_id, attachment_type, attachment_thumbnail_id, created_at, updated_at,
				   (select count(*) from message_likes where message_id = m.id) likes,
				   case when exists(select * from message_likes where message_id = m.id and user_id = ?) then true else false end liked_by_me
			from messages m
			where user_id in (?) and parent_id is null
				and created_at < ?
				and not exists(select * from message_visibility mv where mv.user_id = ? and mv.message_id = m.id and mv.visibility = false)
			order by created_at desc, "id"
			limit ?`,
		currentUserID, userIDs, from, currentUserID, count)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&messages, query, args...)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	return messages, nil
}

func (r *messageRepo) Message(currentUserID string, messageID string) (Message, error) {
	var message Message
	err := r.db.Get(&message, `
		select id, parent_id, user_id, user_name, text, attachment_id, attachment_type, attachment_thumbnail_id, created_at, updated_at,
		       (select count(*) from message_likes where message_id = messages.id) likes,
		       case when exists(select * from message_likes where message_id = messages.id and user_id = $1) then true else false end liked_by_me
		from messages
		where id = $2`, currentUserID, messageID)
	if err != nil {
		if merry.Is(err, sql.ErrNoRows) {
			return message, ErrMessageNotFound.Here()
		}
		return message, merry.Wrap(err)
	}
	return message, nil
}

func (r *messageRepo) AddMessage(parentID string, message Message) error {
	_, err := r.db.Exec(`
		insert into messages(id, parent_id, user_id, user_name, text, attachment_id, attachment_type, attachment_thumbnail_id, created_at, updated_at)
		select $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		where not exists(select * from messages where id = $1)`,
		message.ID, sql.NullString{String: parentID, Valid: parentID != ""},
		message.UserID, message.UserName,
		message.Text, message.AttachmentID, message.AttachmentType, message.AttachmentThumbnailID,
		message.CreatedAt, message.UpdatedAt)
	if err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func (r *messageRepo) EditMessageText(userID, messageID, text string, updatedAt time.Time) error {
	res, err := r.db.Exec(`
		update messages
		set text = $1, updated_at = $2
		where id = $3 and user_id = $4`,
		text, updatedAt, messageID, userID)
	if err != nil {
		return merry.Wrap(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return merry.Wrap(err)
	}
	if rowsAffected < 1 {
		return ErrMessageNotFound.Here()
	}
	return nil
}

func (r *messageRepo) EditMessageAttachment(userID, messageID, attachmentID, attachmentType, attachmentThumbnailID string, updatedAt time.Time) error {
	return common.RunInTransaction(r.db, func(tx *sqlx.Tx) error {
		var message Message
		err := tx.Get(&message, "select attachment_id, attachment_thumbnail_id from messages where id = $1 and user_id = $2",
			messageID, userID)
		if err != nil && !merry.Is(err, sql.ErrNoRows) {
			return merry.Wrap(err)
		}
		if message.AttachmentID != "" && message.AttachmentID != attachmentID {
			_, err = tx.Exec(`
				insert into blob_pending_deletes(blob_id, deleted_at)
				values ($1, $2)`,
				message.AttachmentID, updatedAt)
			if err != nil {
				return merry.Wrap(err)
			}
		}
		if message.AttachmentThumbnailID != "" && message.AttachmentThumbnailID != message.AttachmentID && message.AttachmentThumbnailID != attachmentThumbnailID {
			_, err = tx.Exec(`
				insert into blob_pending_deletes(blob_id, deleted_at)
				values ($1, $2)`,
				message.AttachmentThumbnailID, updatedAt)
			if err != nil {
				return merry.Wrap(err)
			}
		}

		res, err := r.db.Exec(`
		update messages
		set attachment_id = $1, attachment_type = $2, attachment_thumbnail_id = $3, updated_at = $4
		where id = $5 and user_id = $6`,
			attachmentID, attachmentType, attachmentThumbnailID, updatedAt, messageID, userID)
		if err != nil {
			return merry.Wrap(err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return merry.Wrap(err)
		}
		if rowsAffected < 1 {
			return ErrMessageNotFound.Here()
		}
		return nil
	})
}

func (r *messageRepo) DeleteMessage(userID, messageID string) error {
	return common.RunInTransaction(r.db, func(tx *sqlx.Tx) error {
		var messages []Message
		err := tx.Select(&messages, "select attachment_id, attachment_thumbnail_id from messages where (id = $1 and user_id = $2) or parent_id = $1",
			messageID, userID)
		if err != nil {
			return merry.Wrap(err)
		}
		now := common.CurrentTimestamp()
		for _, msg := range messages {
			if msg.AttachmentID != "" {
				_, err = tx.Exec(`
				insert into blob_pending_deletes(blob_id, deleted_at)
				values ($1, $2)`,
					msg.AttachmentID, now)
				if err != nil {
					return merry.Wrap(err)
				}
			}
			if msg.AttachmentThumbnailID != "" && msg.AttachmentThumbnailID != msg.AttachmentID {
				_, err = tx.Exec(`
				insert into blob_pending_deletes(blob_id, deleted_at)
				values ($1, $2)`,
					msg.AttachmentThumbnailID, now)
				if err != nil {
					return merry.Wrap(err)
				}
			}
		}

		_, err = tx.Exec(`
		delete from message_likes
		where message_id in (
		    select id     
		    from messages
			where parent_id = $1
		  		and (select user_id from messages where messages.id = $1) = $2)`,
			messageID, userID)
		if err != nil {
			return merry.Wrap(err)
		}

		_, err = tx.Exec(`
		delete from messages
		where parent_id = $1
		  and (select user_id from messages where messages.id = $1) = $2`,
			messageID, userID)
		if err != nil {
			return merry.Wrap(err)
		}

		_, err = tx.Exec(`
		delete from message_likes
		where message_id in (
		    select id
		    from messages
			where id = $1 and user_id = $2)`,
			messageID, userID)
		if err != nil {
			return merry.Wrap(err)
		}

		res, err := tx.Exec(`
		delete from messages
		where id = $1 and user_id = $2`,
			messageID, userID)
		if err != nil {
			return merry.Wrap(err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return merry.Wrap(err)
		}
		if rowsAffected < 1 {
			return ErrMessageNotFound.Here()
		}

		return nil
	})
}

func (r *messageRepo) Comments(currentUserID string, messageIDs []string) (map[string][]Message, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}

	var comments []Message
	query, args, err := sqlx.In(`
			select id, parent_id, user_id, user_name, text, attachment_id, attachment_type, attachment_thumbnail_id, created_at, updated_at,
				   (select count(*) from message_likes where message_id = m.id) likes,
				   case when exists(select * from message_likes where message_id = m.id and user_id = ?) then true else false end liked_by_me
			from messages m
			where parent_id in (?)
				and not exists(select * from message_visibility mv where mv.user_id = ? and mv.message_id = m.id and mv.visibility = false)
			order by created_at, id`,
		currentUserID, messageIDs, currentUserID)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&comments, query, args...)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	result := make(map[string][]Message)
	for _, comment := range comments {
		result[comment.ParentID.String] = append(result[comment.ParentID.String], comment)
	}
	return result, nil
}

func (r *messageRepo) LikeMessage(userID, messageID string) (likes int, err error) {
	_, err = r.db.Exec(`
		insert into message_likes(message_id, user_id, created_at)
		select $1, $2, $3
		where not exists(select * from message_likes where message_id = $1 and user_id = $2)`,
		messageID, userID, common.CurrentTimestamp())
	if err != nil {
		return -1, merry.Wrap(err)
	}
	err = r.db.Get(&likes, "select count(*) from message_likes where message_id = $1", messageID)
	if err != nil {
		return -1, merry.Wrap(err)
	}
	return likes, nil
}

func (r *messageRepo) MessagesLikes(messageIDs []string) (likes map[string][]MessageLike, err error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}

	var plainLikes []MessageLike
	query, args, err := sqlx.In(`
		select ml.message_id, ml.user_id, u.name user_name, ml.created_at
		from message_likes ml
			inner join users u on u.id = ml.user_id
		where ml.message_id in (?)
		order by ml.message_id, ml.created_at`, messageIDs)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	query = r.db.Rebind(query)

	err = r.db.Select(&plainLikes, query, args...)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	if len(plainLikes) == 0 {
		return nil, nil
	}
	likes = make(map[string][]MessageLike)
	for _, like := range plainLikes {
		likes[like.MessageID] = append(likes[like.MessageID], like)
	}
	return likes, nil
}

func (r *messageRepo) MessageLikes(messageID string) (likes []MessageLike, err error) {
	allLikes, err := r.MessagesLikes([]string{messageID})
	if err != nil {
		return nil, merry.Wrap(err)
	}
	if len(allLikes) == 0 {
		return nil, nil
	}

	for _, messageLikes := range allLikes {
		return messageLikes, nil
	}

	return likes, nil
}

func (r *messageRepo) SetMessageVisibility(userID, messageID string, visibility bool) error {
	if visibility {
		_, err := r.db.Exec(`
			delete from message_visibility
			where user_id = $1 and message_id = $2;`,
			userID, messageID)
		if err != nil {
			return merry.Wrap(err)
		}
	} else {
		_, err := r.db.Exec(`
			insert into message_visibility(user_id, message_id, visibility, created_at)
			select $1, $2, false, $3
			where not exists(select * from message_visibility where user_id = $1 and message_id = $2)`,
			userID, messageID, common.CurrentTimestamp())
		if err != nil {
			return merry.Wrap(err)
		}
	}
	return nil
}
