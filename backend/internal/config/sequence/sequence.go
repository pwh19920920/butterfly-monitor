package sequence

import "github.com/sirupsen/logrus"
import "github.com/pwh19920920/snowflake"

var flakeNode *snowflake.Node

func GetSequence() *snowflake.Node {
	if flakeNode == nil {
		node, err := snowflake.NewNode(1)
		if err != nil {
			logrus.Panic(err)
		}
		flakeNode = node
	}
	return flakeNode
}
