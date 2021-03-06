package services

import (
	"bytes"
	"context"
	"image/jpeg"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/ansel1/merry"
	"github.com/gofrs/uuid"
	"github.com/h2non/filetype"
	"github.com/twitchtv/twirp"

	"github.com/mreider/koto/backend/common"
	"github.com/mreider/koto/backend/messagehub/repo"
	"github.com/mreider/koto/backend/messagehub/rpc"
	"github.com/mreider/koto/backend/messagehub/services/message"
	"github.com/mreider/koto/backend/token"
)

const (
	fileTypeBufSize = 8192
)

type messageService struct {
	*BaseService
}

func NewMessage(base *BaseService) rpc.MessageService {
	return &messageService{
		BaseService: base,
	}
}

func (s *messageService) Post(ctx context.Context, r *rpc.MessagePostRequest) (*rpc.MessagePostResponse, error) {
	user := s.getUser(ctx)

	_, claims, err := s.tokenParser.Parse(r.Token, "post-message")
	if err != nil {
		if merry.Is(err, token.ErrInvalidToken) {
			return nil, twirp.NewError(twirp.InvalidArgument, "invalid token")
		}
		return nil, err
	}

	if user.ID != claims["id"].(string) ||
		strings.TrimSuffix(s.externalAddress, "/") != strings.TrimSuffix(claims["hub"].(string), "/") {
		return nil, twirp.NewError(twirp.InvalidArgument, "invalid token")
	}

	rawFriendIDs := claims["friends"].([]interface{})
	friends := make([]string, len(rawFriendIDs))
	for i, rawID := range rawFriendIDs {
		friends[i] = rawID.(string)
	}

	messageID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	attachmentThumbnailID, attachmentType, err := s.processAttachment(ctx, r.AttachmentId)
	if err != nil {
		return nil, err
	}

	now := common.CurrentTimestamp()
	msg := repo.Message{
		ID:                    messageID.String(),
		UserID:                claims["id"].(string),
		UserName:              claims["name"].(string),
		Text:                  r.Text,
		AttachmentID:          r.AttachmentId,
		AttachmentType:        attachmentType,
		AttachmentThumbnailID: attachmentThumbnailID,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	err = s.repos.Message.AddMessage("", msg)
	if err != nil {
		return nil, err
	}

	s.notificationSender.SendNotification(friends, msg.UserName+" posted a new message", "message/post", map[string]interface{}{
		"user_id":    msg.UserID,
		"message_id": msg.ID,
	})

	userTags := message.FindUserTags(msg.Text)
	users, err := s.repos.User.FindUsersByName(userTags)
	if err != nil {
		return nil, err
	}
	notifyUsers := make([]string, 0, len(users))
	for _, u := range users {
		if u.ID != msg.UserID {
			notifyUsers = append(notifyUsers, u.ID)
		}
	}
	s.notificationSender.SendNotification(notifyUsers, msg.UserName+" tagged you in a message", "message/tag", map[string]interface{}{
		"user_id":    msg.UserID,
		"message_id": msg.ID,
	})

	attachmentLink, err := s.createBlobLink(ctx, msg.AttachmentID)
	if err != nil {
		return nil, err
	}

	attachmentThumbnailLink, err := s.createBlobLink(ctx, msg.AttachmentThumbnailID)
	if err != nil {
		return nil, err
	}

	return &rpc.MessagePostResponse{
		Message: &rpc.Message{
			Id:                  msg.ID,
			UserId:              msg.UserID,
			UserName:            msg.UserName,
			Text:                msg.Text,
			Attachment:          attachmentLink,
			AttachmentType:      attachmentType,
			AttachmentThumbnail: attachmentThumbnailLink,
			CreatedAt:           common.TimeToRPCString(msg.CreatedAt),
			UpdatedAt:           common.TimeToRPCString(msg.UpdatedAt),
			Likes:               int32(msg.Likes),
			LikedByMe:           msg.LikedByMe,
		},
	}, nil
}

func (s *messageService) Messages(ctx context.Context, r *rpc.MessageMessagesRequest) (*rpc.MessageMessagesResponse, error) {
	user := s.getUser(ctx)

	_, claims, err := s.tokenParser.Parse(r.Token, "get-messages")
	if err != nil {
		if merry.Is(err, token.ErrInvalidToken) {
			return nil, twirp.NewError(twirp.InvalidArgument, "invalid token")
		}
		return nil, err
	}

	if user.ID != claims["id"].(string) ||
		strings.TrimSuffix(s.externalAddress, "/") != strings.TrimSuffix(claims["hub"].(string), "/") {
		return nil, twirp.NewError(twirp.InvalidArgument, "invalid token")
	}

	rawUserIDs := claims["users"].([]interface{})
	userIDs := make([]string, len(rawUserIDs))
	for i, rawUserID := range rawUserIDs {
		userIDs[i] = rawUserID.(string)
	}

	var from time.Time
	if r.From != "" {
		from, err = common.RPCStringToTime(r.From)
		if err != nil {
			return nil, twirp.InvalidArgumentError("from", err.Error())
		}
	}

	messages, err := s.repos.Message.Messages(user.ID, userIDs, from, int(r.Count))
	if err != nil {
		return nil, err
	}

	messageIDs := make([]string, len(messages))
	rpcMessages := make([]*rpc.Message, len(messages))
	rpcMessageMap := make(map[string]*rpc.Message, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.ID
		attachmentLink, err := s.createBlobLink(ctx, msg.AttachmentID)
		if err != nil {
			return nil, err
		}
		attachmentThumbnailLink, err := s.createBlobLink(ctx, msg.AttachmentThumbnailID)
		if err != nil {
			return nil, err
		}

		rpcMessages[i] = &rpc.Message{
			Id:                  msg.ID,
			UserId:              msg.UserID,
			UserName:            msg.UserName,
			Text:                msg.Text,
			Attachment:          attachmentLink,
			AttachmentType:      msg.AttachmentType,
			AttachmentThumbnail: attachmentThumbnailLink,
			CreatedAt:           common.TimeToRPCString(msg.CreatedAt),
			UpdatedAt:           common.TimeToRPCString(msg.UpdatedAt),
			Likes:               int32(msg.Likes),
			LikedByMe:           msg.LikedByMe,
		}
		rpcMessageMap[msg.ID] = rpcMessages[i]
	}

	allLikes, err := s.repos.Message.MessagesLikes(messageIDs)
	if err != nil {
		return nil, err
	}
	for msgID, likes := range allLikes {
		rpcLikes := make([]*rpc.MessageLike, len(likes))
		for i, like := range likes {
			rpcLikes[i] = &rpc.MessageLike{
				UserId:   like.UserID,
				UserName: like.UserName,
				LikedAt:  common.TimeToRPCString(like.CreatedAt),
			}
		}
		rpcMessageMap[msgID].LikedBy = rpcLikes
	}

	comments, err := s.repos.Message.Comments(user.ID, messageIDs)
	if err != nil {
		return nil, err
	}
	for messageID, messageComments := range comments {
		rpcComments := make([]*rpc.Message, len(messageComments))
		for i, comment := range messageComments {
			attachmentLink, err := s.createBlobLink(ctx, comment.AttachmentID)
			if err != nil {
				return nil, err
			}
			attachmentThumbnailLink, err := s.createBlobLink(ctx, comment.AttachmentThumbnailID)
			if err != nil {
				return nil, err
			}

			rpcComments[i] = &rpc.Message{
				Id:                  comment.ID,
				UserId:              comment.UserID,
				UserName:            comment.UserName,
				Text:                comment.Text,
				Attachment:          attachmentLink,
				AttachmentType:      comment.AttachmentType,
				AttachmentThumbnail: attachmentThumbnailLink,
				CreatedAt:           common.TimeToRPCString(comment.CreatedAt),
				UpdatedAt:           common.TimeToRPCString(comment.UpdatedAt),
				Likes:               int32(comment.Likes),
				LikedByMe:           comment.LikedByMe,
			}
		}
		rpcMessageMap[messageID].Comments = rpcComments
	}

	return &rpc.MessageMessagesResponse{
		Messages: rpcMessages,
	}, nil
}

func (s *messageService) Message(ctx context.Context, r *rpc.MessageMessageRequest) (*rpc.MessageMessageResponse, error) {
	user := s.getUser(ctx)

	_, claims, err := s.tokenParser.Parse(r.Token, "get-messages")
	if err != nil {
		if merry.Is(err, token.ErrInvalidToken) {
			return nil, twirp.NewError(twirp.InvalidArgument, "invalid token")
		}
		return nil, err
	}

	if user.ID != claims["id"].(string) ||
		strings.TrimSuffix(s.externalAddress, "/") != strings.TrimSuffix(claims["hub"].(string), "/") {
		return nil, twirp.NewError(twirp.InvalidArgument, "invalid token")
	}

	rawUserIDs := claims["users"].([]interface{})
	userIDs := make(map[string]bool, len(rawUserIDs))
	for _, rawUserID := range rawUserIDs {
		userIDs[rawUserID.(string)] = true
	}

	msg, err := s.repos.Message.Message(user.ID, r.MessageId)
	if err != nil {
		if merry.Is(err, repo.ErrMessageNotFound) {
			return nil, twirp.NotFoundError("message not found")
		}
		return nil, err
	}

	if !userIDs[msg.UserID] {
		return nil, twirp.NotFoundError("message not found")
	}

	attachmentLink, err := s.createBlobLink(ctx, msg.AttachmentID)
	if err != nil {
		return nil, err
	}
	attachmentThumbnailLink, err := s.createBlobLink(ctx, msg.AttachmentThumbnailID)
	if err != nil {
		return nil, err
	}

	rpcMessage := &rpc.Message{
		Id:                  msg.ID,
		UserId:              msg.UserID,
		UserName:            msg.UserName,
		Text:                msg.Text,
		Attachment:          attachmentLink,
		AttachmentType:      msg.AttachmentType,
		AttachmentThumbnail: attachmentThumbnailLink,
		CreatedAt:           common.TimeToRPCString(msg.CreatedAt),
		UpdatedAt:           common.TimeToRPCString(msg.UpdatedAt),
		Likes:               int32(msg.Likes),
		LikedByMe:           msg.LikedByMe,
	}

	allLikes, err := s.repos.Message.MessagesLikes([]string{msg.ID})
	if err != nil {
		return nil, err
	}
	for _, likes := range allLikes {
		rpcLikes := make([]*rpc.MessageLike, len(likes))
		for i, like := range likes {
			rpcLikes[i] = &rpc.MessageLike{
				UserId:   like.UserID,
				UserName: like.UserName,
				LikedAt:  common.TimeToRPCString(like.CreatedAt),
			}
		}
		rpcMessage.LikedBy = rpcLikes
	}

	comments, err := s.repos.Message.Comments(user.ID, []string{msg.ID})
	if err != nil {
		return nil, err
	}
	for _, messageComments := range comments {
		rpcComments := make([]*rpc.Message, len(messageComments))
		for i, comment := range messageComments {
			attachmentLink, err := s.createBlobLink(ctx, comment.AttachmentID)
			if err != nil {
				return nil, err
			}
			attachmentThumbnailLink, err := s.createBlobLink(ctx, comment.AttachmentThumbnailID)
			if err != nil {
				return nil, err
			}

			rpcComments[i] = &rpc.Message{
				Id:                  comment.ID,
				UserId:              comment.UserID,
				UserName:            comment.UserName,
				Text:                comment.Text,
				Attachment:          attachmentLink,
				AttachmentType:      comment.AttachmentType,
				AttachmentThumbnail: attachmentThumbnailLink,
				CreatedAt:           common.TimeToRPCString(comment.CreatedAt),
				UpdatedAt:           common.TimeToRPCString(comment.UpdatedAt),
				Likes:               int32(comment.Likes),
				LikedByMe:           comment.LikedByMe,
			}
		}
		rpcMessage.Comments = rpcComments
	}

	return &rpc.MessageMessageResponse{
		Message: rpcMessage,
	}, nil
}

func (s *messageService) Edit(ctx context.Context, r *rpc.MessageEditRequest) (*rpc.MessageEditResponse, error) {
	user := s.getUser(ctx)
	now := common.CurrentTimestamp()
	if r.TextChanged {
		err := s.repos.Message.EditMessageText(user.ID, r.MessageId, r.Text, now)
		if err != nil {
			if merry.Is(err, repo.ErrMessageNotFound) {
				return nil, twirp.NotFoundError(err.Error())
			}
			return nil, err
		}
	}
	if r.AttachmentChanged {
		attachmentThumbnailID, attachmentType, err := s.processAttachment(ctx, r.AttachmentId)
		if err != nil {
			return nil, err
		}

		err = s.repos.Message.EditMessageAttachment(user.ID, r.MessageId, r.AttachmentId, attachmentType, attachmentThumbnailID, now)
		if err != nil {
			if merry.Is(err, repo.ErrMessageNotFound) {
				return nil, twirp.NotFoundError(err.Error())
			}
			return nil, err
		}
	}

	msg, err := s.repos.Message.Message(user.ID, r.MessageId)
	if err != nil {
		if merry.Is(err, repo.ErrMessageNotFound) {
			return nil, twirp.NotFoundError(err.Error())
		}
		return nil, err
	}

	attachmentLink, err := s.createBlobLink(ctx, msg.AttachmentID)
	if err != nil {
		return nil, err
	}
	attachmentThumbnailLink, err := s.createBlobLink(ctx, msg.AttachmentThumbnailID)
	if err != nil {
		return nil, err
	}

	return &rpc.MessageEditResponse{
		Message: &rpc.Message{
			Id:                  msg.ID,
			UserId:              msg.UserID,
			UserName:            msg.UserName,
			Text:                msg.Text,
			Attachment:          attachmentLink,
			AttachmentType:      msg.AttachmentType,
			AttachmentThumbnail: attachmentThumbnailLink,
			CreatedAt:           common.TimeToRPCString(msg.CreatedAt),
			UpdatedAt:           common.TimeToRPCString(msg.UpdatedAt),
			Likes:               int32(msg.Likes),
			LikedByMe:           msg.LikedByMe,
		},
	}, nil
}

func (s *messageService) Delete(ctx context.Context, r *rpc.MessageDeleteRequest) (_ *rpc.Empty, err error) {
	user := s.getUser(ctx)

	err = s.repos.Message.DeleteMessage(user.ID, r.MessageId)
	if err != nil {
		if merry.Is(err, repo.ErrMessageNotFound) {
			return nil, twirp.NotFoundError(err.Error())
		}
		return nil, err
	}
	return &rpc.Empty{}, nil
}

func (s *messageService) PostComment(ctx context.Context, r *rpc.MessagePostCommentRequest) (*rpc.MessagePostCommentResponse, error) {
	user := s.getUser(ctx)

	_, claims, err := s.tokenParser.Parse(r.Token, "get-messages")
	if err != nil {
		if merry.Is(err, token.ErrInvalidToken) {
			return nil, twirp.NewError(twirp.InvalidArgument, "invalid token")
		}
		return nil, err
	}

	if user.ID != claims["id"].(string) ||
		strings.TrimSuffix(s.externalAddress, "/") != strings.TrimSuffix(claims["hub"].(string), "/") {
		return nil, twirp.NewError(twirp.InvalidArgument, "invalid token")
	}

	msg, err := s.repos.Message.Message(user.ID, r.MessageId)
	if err != nil {
		if merry.Is(err, repo.ErrMessageNotFound) {
			return nil, twirp.NotFoundError(err.Error())
		}
		return nil, err
	}

	rawUserIDs := claims["users"].([]interface{})
	found := false
	for _, rawUserID := range rawUserIDs {
		userID := rawUserID.(string)
		if userID == msg.UserID {
			found = true
			break
		}
	}

	if !found {
		return nil, twirp.NotFoundError(repo.ErrMessageNotFound.Error())
	}

	commentID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	attachmentThumbnailID, attachmentType, err := s.processAttachment(ctx, r.AttachmentId)
	if err != nil {
		return nil, err
	}

	now := common.CurrentTimestamp()
	comment := repo.Message{
		ID:                    commentID.String(),
		UserID:                claims["id"].(string),
		UserName:              claims["name"].(string),
		Text:                  r.Text,
		AttachmentID:          r.AttachmentId,
		AttachmentType:        attachmentType,
		AttachmentThumbnailID: attachmentThumbnailID,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	err = s.repos.Message.AddMessage(r.MessageId, comment)
	if err != nil {
		return nil, err
	}

	if user.ID != msg.UserID {
		s.notificationSender.SendNotification([]string{msg.UserID}, user.Name+" posted a new comment", "comment/post", map[string]interface{}{
			"user_id":    user.ID,
			"message_id": msg.ID,
			"comment_id": comment.ID,
		})
	}

	userTags := message.FindUserTags(comment.Text)
	users, err := s.repos.User.FindUsersByName(userTags)
	if err != nil {
		return nil, err
	}

	notifyUsers := make([]string, 0, len(users))
	for _, u := range users {
		if u.ID != comment.UserID {
			notifyUsers = append(notifyUsers, u.ID)
		}
	}
	s.notificationSender.SendNotification(notifyUsers, comment.UserName+" tagged you in a comment", "comment/tag", map[string]interface{}{
		"user_id":    comment.UserID,
		"message_id": msg.ID,
		"comment_id": comment.ID,
	})

	attachmentLink, err := s.createBlobLink(ctx, comment.AttachmentID)
	if err != nil {
		return nil, err
	}

	attachmentThumbnailLink, err := s.createBlobLink(ctx, comment.AttachmentThumbnailID)
	if err != nil {
		return nil, err
	}

	return &rpc.MessagePostCommentResponse{
		Comment: &rpc.Message{
			Id:                  comment.ID,
			UserId:              comment.UserID,
			UserName:            comment.UserName,
			Text:                comment.Text,
			Attachment:          attachmentLink,
			AttachmentType:      attachmentType,
			AttachmentThumbnail: attachmentThumbnailLink,
			CreatedAt:           common.TimeToRPCString(comment.CreatedAt),
			UpdatedAt:           common.TimeToRPCString(comment.UpdatedAt),
			Likes:               int32(comment.Likes),
			LikedByMe:           comment.LikedByMe,
		},
	}, nil
}

func (s *messageService) EditComment(ctx context.Context, r *rpc.MessageEditCommentRequest) (*rpc.MessageEditCommentResponse, error) {
	user := s.getUser(ctx)
	now := common.CurrentTimestamp()
	if r.TextChanged {
		err := s.repos.Message.EditMessageText(user.ID, r.CommentId, r.Text, now)
		if err != nil {
			if merry.Is(err, repo.ErrMessageNotFound) {
				return nil, twirp.NotFoundError("comment not found")
			}
			return nil, err
		}
	}
	if r.AttachmentChanged {
		attachmentThumbnailID, attachmentType, err := s.processAttachment(ctx, r.AttachmentId)
		if err != nil {
			return nil, err
		}

		err = s.repos.Message.EditMessageAttachment(user.ID, r.CommentId, r.AttachmentId, attachmentType, attachmentThumbnailID, now)
		if err != nil {
			if merry.Is(err, repo.ErrMessageNotFound) {
				return nil, twirp.NotFoundError(err.Error())
			}
			return nil, err
		}
	}

	comment, err := s.repos.Message.Message(user.ID, r.CommentId)
	if err != nil {
		if merry.Is(err, repo.ErrMessageNotFound) {
			return nil, twirp.NotFoundError("comment not found")
		}
		return nil, err
	}

	attachmentLink, err := s.createBlobLink(ctx, comment.AttachmentID)
	if err != nil {
		return nil, err
	}

	attachmentThumbnailLink, err := s.createBlobLink(ctx, comment.AttachmentThumbnailID)
	if err != nil {
		return nil, err
	}

	return &rpc.MessageEditCommentResponse{
		Comment: &rpc.Message{
			Id:                  comment.ID,
			UserId:              comment.UserID,
			UserName:            comment.UserName,
			Text:                comment.Text,
			Attachment:          attachmentLink,
			AttachmentType:      comment.AttachmentType,
			AttachmentThumbnail: attachmentThumbnailLink,
			CreatedAt:           common.TimeToRPCString(comment.CreatedAt),
			UpdatedAt:           common.TimeToRPCString(comment.UpdatedAt),
			Likes:               int32(comment.Likes),
			LikedByMe:           comment.LikedByMe,
		},
	}, nil
}

func (s *messageService) DeleteComment(ctx context.Context, r *rpc.MessageDeleteCommentRequest) (_ *rpc.Empty, err error) {
	user := s.getUser(ctx)
	err = s.repos.Message.DeleteMessage(user.ID, r.CommentId)
	if err != nil {
		if merry.Is(err, repo.ErrMessageNotFound) {
			return nil, twirp.NotFoundError("comment not found")
		}
		return nil, err
	}
	return &rpc.Empty{}, nil
}

func (s *messageService) getAttachmentType(ctx context.Context, attachmentID string) (string, error) {
	if attachmentID == "" {
		return "", nil
	}

	buf, err := s.s3Storage.ReadN(ctx, attachmentID, fileTypeBufSize)
	if err != nil {
		return "", merry.Wrap(err)
	}
	t, err := filetype.Match(buf)
	if err != nil {
		return "", merry.Wrap(err)
	}
	return t.MIME.Value, nil
}

func (s *messageService) getAttachmentThumbnailID(ctx context.Context, attachmentID, attachmentType string) (string, error) {
	if strings.HasPrefix(attachmentType, "image/") {
		return attachmentID, nil
	}

	if !strings.HasPrefix(attachmentType, "video/") {
		return "", nil
	}

	link, err := s.createBlobLink(ctx, attachmentID)
	if err != nil {
		return "", merry.Wrap(err)
	}
	thumbnail, err := common.VideoThumbnail(link)
	if err != nil {
		return "", merry.Wrap(err)
	}
	if len(thumbnail) == 0 {
		return "", nil
	}

	ext := filepath.Ext(attachmentID)
	attachmentThumbnailID := strings.TrimSuffix(attachmentID, ext) + "-thumbnail.jpg"
	err = s.s3Storage.PutObject(ctx, attachmentThumbnailID, thumbnail, "image/jpeg")
	if err != nil {
		return "", err
	}
	return attachmentThumbnailID, nil
}

func (s *messageService) LikeMessage(ctx context.Context, r *rpc.MessageLikeMessageRequest) (*rpc.MessageLikeMessageResponse, error) {
	user := s.getUser(ctx)

	msg, err := s.repos.Message.Message(user.ID, r.MessageId)
	if err != nil {
		if !merry.Is(err, repo.ErrMessageNotFound) {
			return nil, err
		}
		return &rpc.MessageLikeMessageResponse{
			Likes: -1,
		}, nil
	}

	if msg.ParentID.Valid {
		return nil, twirp.InvalidArgumentError("message_id", "is not a message")
	}

	newLikeCount, err := s.repos.Message.LikeMessage(user.ID, msg.ID)
	if err != nil {
		return nil, err
	}
	s.notificationSender.SendNotification([]string{msg.UserID}, user.Name+" liked your post", "message/like", map[string]interface{}{
		"user_id":    user.ID,
		"message_id": msg.ID,
	})
	return &rpc.MessageLikeMessageResponse{
		Likes: int32(newLikeCount),
	}, nil
}

func (s *messageService) LikeComment(ctx context.Context, r *rpc.MessageLikeCommentRequest) (*rpc.MessageLikeCommentResponse, error) {
	user := s.getUser(ctx)

	comment, err := s.repos.Message.Message(user.ID, r.CommentId)
	if err != nil {
		if !merry.Is(err, repo.ErrMessageNotFound) {
			return nil, err
		}
		return &rpc.MessageLikeCommentResponse{
			Likes: -1,
		}, nil
	}

	if !comment.ParentID.Valid {
		return nil, twirp.InvalidArgumentError("comment_id", "is not a comment")
	}

	newLikeCount, err := s.repos.Message.LikeMessage(user.ID, comment.ID)
	if err != nil {
		return nil, err
	}
	s.notificationSender.SendNotification([]string{comment.UserID}, user.Name+" liked your comment", "comment/like", map[string]interface{}{
		"user_id":    user.ID,
		"message_id": comment.ParentID.String,
		"comment_id": comment.ID,
	})
	return &rpc.MessageLikeCommentResponse{
		Likes: int32(newLikeCount),
	}, nil
}

func (s *messageService) MessageLikes(_ context.Context, r *rpc.MessageMessageLikesRequest) (*rpc.MessageMessageLikesResponse, error) {
	likes, err := s.repos.Message.MessageLikes(r.MessageId)
	if err != nil {
		return nil, err
	}
	rpcLikes := make([]*rpc.MessageLike, len(likes))
	for i, like := range likes {
		rpcLikes[i] = &rpc.MessageLike{
			UserId:   like.UserID,
			UserName: like.UserName,
			LikedAt:  common.TimeToRPCString(like.CreatedAt),
		}
	}
	return &rpc.MessageMessageLikesResponse{
		Likes: rpcLikes,
	}, nil
}

func (s *messageService) CommentLikes(_ context.Context, r *rpc.MessageCommentLikesRequest) (*rpc.MessageCommentLikesResponse, error) {
	likes, err := s.repos.Message.MessageLikes(r.CommentId)
	if err != nil {
		return nil, err
	}
	rpcLikes := make([]*rpc.MessageLike, len(likes))
	for i, like := range likes {
		rpcLikes[i] = &rpc.MessageLike{
			UserId:   like.UserID,
			UserName: like.UserName,
			LikedAt:  common.TimeToRPCString(like.CreatedAt),
		}
	}
	return &rpc.MessageCommentLikesResponse{
		Likes: rpcLikes,
	}, nil
}

func (s *messageService) processAttachment(ctx context.Context, attachmentID string) (attachmentThumbnailID, attachmentType string, err error) {
	attachmentType, err = s.getAttachmentType(ctx, attachmentID)
	if err != nil {
		return "", "", err
	}

	attachmentThumbnailID, err = s.getAttachmentThumbnailID(ctx, attachmentID, attachmentType)
	if err != nil {
		return "", "", err
	}

	if attachmentType != "image/jpeg" {
		return attachmentThumbnailID, attachmentType, nil
	}

	var buf bytes.Buffer
	err = s.s3Storage.Read(ctx, attachmentID, &buf)
	if err != nil {
		log.Println("can't read attachment:", err)
		return attachmentThumbnailID, attachmentType, nil
	}

	orientation := common.GetImageOrientation(bytes.NewReader(buf.Bytes()))
	if orientation == "1" {
		return attachmentThumbnailID, attachmentType, nil
	}
	if img, err := common.DecodeImageAndFixOrientation(bytes.NewReader(buf.Bytes()), orientation); err == nil {
		buf.Reset()
		if err := jpeg.Encode(&buf, img, nil); err == nil {
			_ = s.s3Storage.PutObject(ctx, attachmentID, buf.Bytes(), attachmentType)
		}
	}
	return attachmentThumbnailID, attachmentType, nil
}

func (s *messageService) SetMessageVisibility(ctx context.Context, r *rpc.MessageSetMessageVisibilityRequest) (*rpc.Empty, error) {
	user := s.getUser(ctx)
	err := s.repos.Message.SetMessageVisibility(user.ID, r.MessageId, r.Visibility)
	if err != nil {
		return nil, err
	}
	return &rpc.Empty{}, nil
}

func (s *messageService) SetCommentVisibility(ctx context.Context, r *rpc.MessageSetCommentVisibilityRequest) (*rpc.Empty, error) {
	user := s.getUser(ctx)
	err := s.repos.Message.SetMessageVisibility(user.ID, r.CommentId, r.Visibility)
	if err != nil {
		return nil, err
	}
	return &rpc.Empty{}, nil
}
