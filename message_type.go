package polymarketdataclient

type MessageType string

const (
	// MessageTypeAll is the type for all messages.
	// This means that the client will receive all messages of the topic, for example "requests" or "quotes"
	MessageTypeAll MessageType = "*"

	// MessageTypeTrades is the type for trades
	MessageTypeTrades MessageType = "trades"

	// MessageTypeOrdersMatched is the type for orders matched
	MessageTypeOrdersMatched MessageType = "orders_matched"

	// MessageTypeCommentCreated is the type for comment created
	MessageTypeCommentCreated MessageType = "comment_created"

	// MessageTypeCommentRemoved is the type for comment removed
	MessageTypeCommentRemoved MessageType = "comment_removed"

	// MessageTypeReactionCreated is the type for reaction created
	MessageTypeReactionCreated MessageType = "reaction_created"

	// MessageTypeReactionRemoved is the type for reaction removed
	MessageTypeReactionRemoved MessageType = "reaction_removed"

	// MessageTypeRequestCreated is the type for request created
	MessageTypeRequestCreated MessageType = "request_created"

	// MessageTypeRequestEdited is the type for request edited
	MessageTypeRequestEdited MessageType = "request_edited"

	// MessageTypeRequestCanceled is the type for request canceled
	MessageTypeRequestCanceled MessageType = "request_canceled"

	// MessageTypeRequestExpired is the type for request expired
	MessageTypeRequestExpired MessageType = "request_expired"

	// MessageTypeQuoteCreated is the type for quote created
	MessageTypeQuoteCreated MessageType = "quote_created"

	// MessageTypeQuoteEdited is the type for quote edited
	MessageTypeQuoteEdited MessageType = "quote_edited"

	// MessageTypeQuoteCanceled is the type for quote canceled
	MessageTypeQuoteCanceled MessageType = "quote_canceled"

	// MessageTypeQuoteExpired is the type for quote expired
	MessageTypeQuoteExpired MessageType = "quote_expired"
)
