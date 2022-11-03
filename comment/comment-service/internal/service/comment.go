package service

import (
	"context"

	pb "comment-service/api/comment/service/v1"
	"comment-service/internal/biz"
)

type CommentService struct {
	pb.UnimplementedCommentServer
	uc *biz.CommentUsecase
}

func NewCommentService(uc *biz.CommentUsecase) *CommentService {
	return &CommentService{uc: uc}
}

func (s *CommentService) CreateComment(ctx context.Context, req *pb.CreateCommentReq) (*pb.CreateCommentResp, error) {
	return &pb.CreateCommentResp{}, nil
}
func (s *CommentService) ListComments(ctx context.Context, req *pb.ListCommentsReq) (*pb.ListCommentsResp, error) {
	return &pb.ListCommentsResp{}, nil
}
func (s *CommentService) DeleteComments(ctx context.Context, req *pb.DeleteCommentsReq) (*pb.DeleteCommentsResp, error) {
	return &pb.DeleteCommentsResp{}, nil
}
func (s *CommentService) UpdateComment(ctx context.Context, req *pb.UpdateCommentReq) (*pb.UpdateCommentResp, error) {
	return &pb.UpdateCommentResp{}, nil
}
