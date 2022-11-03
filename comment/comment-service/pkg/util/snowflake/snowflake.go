package snowflake


import (
	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func init() {
	var err error
	// todo: Node id
	node, err = snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
}

func GenerateId() int64 {
	return node.Generate().Int64()
}