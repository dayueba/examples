package data

import (
	"time"
	"strconv"

	"comment-service/internal/biz"

	"gorm.io/gorm"
	"github.com/go-kratos/kratos/v2/log"
)

type CommentContent struct {
	CommentID   int64     `gorm:"column:comment_id;primary_key"` // 对应 comment_index.id
	AtMemberIds string    `gorm:"column:at_member_ids;NOT NULL"` // 对象id
	Message     string    `gorm:"column:message;NOT NULL"`       // 评论内容
	Meta        string    `gorm:"column:meta;NOT NULL"`          // 评论元数据，背景，字体
	IP          string    `gorm:"column:ip;NOT NULL"`
	Platform    string    `gorm:"column:platform;NOT NULL"` // 平台
	Device      string    `gorm:"column:device;NOT NULL"`   // 设备
	CreateTime  time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdateTime  time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (m *CommentContent) TableName() string {
	return "comment_content_"
}

func TableOfCommentContent(table *CommentContent, commentId int64) func(db *gorm.DB) *gorm.DB {
  return func(db *gorm.DB) *gorm.DB {
			tableName := table.TableName() + strconv.Itoa(int(commentId % 10))
			return db.Table(tableName)
  }
}

type CommentIndex struct {
	ID         int64     `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	ObjID      int64     `gorm:"column:obj_id;NOT NULL"`               // 对象id
	ObjType    int       `gorm:"column:obj_type;NOT NULL"`             // 对象类型
	MemberID   int64     `gorm:"column:member_id"`                     // 发表者用户id
	Root       int64     `gorm:"column:root;default:0;NOT NULL"`       // 根评论Id，不为0是回复评论
	Parent     int64     `gorm:"column:parent;default:0;NOT NULL"`     // 父评论id，为0是root评论
	Floor      int       `gorm:"column:floor"`                         // 评论楼层
	Count      int       `gorm:"column:count"`                         // 评论总数
	RootCount  int       `gorm:"column:root_count;default:0;NOT NULL"` // 根评论总数
	Like       int       `gorm:"column:like;default:0;NOT NULL"`       // 点赞数
	Hate       int       `gorm:"column:hate"`                          // 点踩数
	State      int       `gorm:"column:state;default:0"`               // 状态：0正常，1隐藏
	Attrs      int       `gorm:"column:attrs"`                         // 属性
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (m *CommentIndex) TableName() string {
	return "comment_index_"
}

func TableOfCommentIndex(table *CommentIndex, commentId int64) func(db *gorm.DB) *gorm.DB {
  return func(db *gorm.DB) *gorm.DB {
			tableName := table.TableName() + strconv.Itoa(int(commentId % 10))
			return db.Table(tableName)
  }
}

type CommentSubject struct {
	ID         int64     `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	ObjID      int       `gorm:"column:obj_id"`     // 对象id
	ObjType    int       `gorm:"column:obj_type"`   // 对象类型
	MemberID   int       `gorm:"column:member_id"`  // 作者用户id
	Count      int       `gorm:"column:count"`      // 评论总数
	RootCount  int       `gorm:"column:root_count"` // 根评论总数
	AllCount   int       `gorm:"column:all_count"`  // 评论加回复总数
	State      int       `gorm:"column:state"`      // 状态，0正常，1隐藏
	Attrs      int       `gorm:"column:attrs"`      // 属性 0 运营置顶，1 up置顶，2 大数据过滤
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (m *CommentSubject) TableName() string {
	return "comment_subject"
}

func TableOfCommentSubject(table *CommentIndex, commentId int64) func(db *gorm.DB) *gorm.DB {
  return func(db *gorm.DB) *gorm.DB {
			tableName := table.TableName() + strconv.Itoa(int(commentId % 3))
			return db.Table(tableName)
  }
}

var _ biz.CommentRepo = (*commentRepo)(nil)

type commentRepo struct {
	data *Data
	log  *log.Helper
}

func NewCommentRepo(data *Data, logger log.Logger) biz.CommentRepo {
	return &commentRepo{data: data, log: log.NewHelper(logger)}
}
