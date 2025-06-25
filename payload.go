package polymarketdataclient

// NOTE: This file is for the payloads that are sent by the server.
// More info can be found here: https://github.com/Polymarket/real-time-data-client

// Trade represents a user's trade on a conditional token market.
type Trade struct {
	Asset           string  `json:"asset"`           // ERC1155 token ID of conditional token being traded
	Bio             string  `json:"bio"`             // Bio of the user of the trade
	ConditionID     string  `json:"conditionId"`     // ID of market which is also the CTF condition ID
	EventSlug       string  `json:"eventSlug"`       // Slug of the event
	Icon            string  `json:"icon"`            // URL to the market icon image
	Name            string  `json:"name"`            // Name of the user of the trade
	Outcome         string  `json:"outcome"`         // Human readable outcome of the market
	OutcomeIndex    int     `json:"outcomeIndex"`    // Index of the outcome
	Price           float64 `json:"price"`           // Price of the trade
	ProfileImage    string  `json:"profileImage"`    // URL to the user profile image
	ProxyWallet     string  `json:"proxyWallet"`     // Address of the user proxy wallet
	Pseudonym       string  `json:"pseudonym"`       // Pseudonym of the user
	Side            string  `json:"side"`            // Side of the trade (BUY/SELL)
	Size            float32 `json:"size"`            // Size of the trade
	Slug            string  `json:"slug"`            // Slug of the market
	Timestamp       int64   `json:"timestamp"`       // Timestamp of the trade
	Title           string  `json:"title"`           // Title of the event
	TransactionHash string  `json:"transactionHash"` // Hash of the transaction
}

// Comment represents a comment on an event or series.
type Comment struct {
	ID               string  `json:"id"`               // Unique identifier of comment
	Body             string  `json:"body"`             // Content of the comment
	ParentEntityType string  `json:"parentEntityType"` // Type of the parent entity (Event or Series)
	ParentEntityID   float64 `json:"parentEntityID"`   // ID of the parent entity
	ParentCommentID  string  `json:"parentCommentID"`  // ID of the parent comment
	UserAddress      string  `json:"userAddress"`      // Address of the user
	ReplyAddress     string  `json:"replyAddress"`     // Address of the reply user
	CreatedAt        string  `json:"createdAt"`        // Creation timestamp
	UpdatedAt        string  `json:"updatedAt"`        // Last update timestamp
}

// Reaction represents a reaction to a comment.
type Reaction struct {
	ID           string  `json:"id"`           // Unique identifier of reaction
	CommentID    float64 `json:"commentID"`    // ID of the comment
	ReactionType string  `json:"reactionType"` // Type of the reaction
	Icon         string  `json:"icon"`         // Icon representing the reaction
	UserAddress  string  `json:"userAddress"`  // Address of the user
	CreatedAt    string  `json:"createdAt"`    // Creation timestamp
}

// Request represents a request to trade a conditional token.
type Request struct {
	RequestID    string  `json:"requestId"`    // Unique identifier for the request
	ProxyAddress string  `json:"proxyAddress"` // User proxy address
	Market       string  `json:"market"`       // ID of market which is also the CTF condition ID
	Token        string  `json:"token"`        // ERC1155 token ID of conditional token being traded
	Complement   string  `json:"complement"`   // Complement ERC1155 token ID of conditional token being traded
	State        string  `json:"state"`        // Current state of the request
	Side         string  `json:"side"`         // Indicates buy or sell side
	SizeIn       float64 `json:"sizeIn"`       // Input size of the request
	SizeOut      float64 `json:"sizeOut"`      // Output size of the request
	Price        float64 `json:"price"`        // Price from in/out sizes
	Expiry       int64   `json:"expiry"`       // Expiry timestamp (UNIX format)
}

// Quote represents a response to a trade request.
type Quote struct {
	QuoteID      string  `json:"quoteId"`      // Unique identifier for the quote
	RequestID    string  `json:"requestId"`    // Associated request identifier
	ProxyAddress string  `json:"proxyAddress"` // User proxy address
	Token        string  `json:"token"`        // ERC1155 token ID of conditional token being traded
	State        string  `json:"state"`        // Current state of the quote
	Side         string  `json:"side"`         // Indicates buy or sell side
	SizeIn       float64 `json:"sizeIn"`       // Input size of the quote
	SizeOut      float64 `json:"sizeOut"`      // Output size of the quote
	Condition    string  `json:"condition"`    // ID of market which is also the CTF condition ID
	Complement   string  `json:"complement"`   // Complement ERC1155 token ID of conditional token being traded
	Expiry       int64   `json:"expiry"`       // Expiry timestamp (UNIX format)
}
