package eventpub

import (
	"context"
	"github.com/dapr-platform/common"
)

func PublishInternalMessage(ctx context.Context, msg *common.InternalMessage) error {

	return common.GetDaprClient().PublishEvent(ctx, common.PUBSUB_NAME, common.INTERNAL_MESSAGE_TOPIC_NAME, msg)
}

func PublishCommonMessage(ctx context.Context, msg *common.CommonMessage) error {
	return common.GetDaprClient().PublishEvent(ctx, common.PUBSUB_NAME, common.WEB_MESSAGE_TOPIC_NAME, msg)
}
