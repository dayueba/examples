package snowflake

import (
	"testing"
	"log"
)


func TestGenerateId(t *testing.T) {
	log.Println(GenerateId())
}