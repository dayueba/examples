package biz

import 	"github.com/go-kratos/kratos/v2/log"

type CommentRepo interface{

}

type CommentUsecase struct {
	log *log.Helper
	cr CommentRepo
}

func NewCommentUsecase(cr CommentRepo, logger log.Logger) *CommentUsecase {
	return &CommentUsecase{cr: cr, log: log.NewHelper(logger)}
}