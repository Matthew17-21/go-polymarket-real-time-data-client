package polymarketdataclient

type Topic string

const (
	// TopicActivity is the topic for activity
	TopicActivity Topic = "activity"

	// TopicComments is the topic for comments
	TopicComments Topic = "comments"

	// TopicRfq is the topic for RFQ
	TopicRfq Topic = "rfq"
)
