package typesense

type ProjectDoc struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Chain           string `json:"chain"`
	ContractAddress string `json:"contractAddress"`
	Narrative       string `json:"narrative"`
	LaunchDate      string `json:"launchDate"`
	LaunchDateUnix  int64  `json:"launchDateUnix"`
	Twitter         string `json:"twitter"`
	Telegram        string `json:"telegram"`
	Discord         string `json:"discord"`
	Website         string `json:"website"`
	ImageUrl        string `json:"imageUrl"`
	LandingVideoUrl string `json:"landingVideoUrl"`
	DevXAccount     string `json:"devXAccount"`
	DevWallet       string `json:"devWallet"`
	IsVerified      bool   `json:"isVerified"`
	CreatedAt       string `json:"createdAt"`
	CreatedAtUnix   int64  `json:"createdAtUnix"`
	UpdatedAt       string `json:"updatedAt"`
	UpdatedAtUnix   int64  `json:"updatedAtUnix"`
	PositiveVotes   uint32 `json:"positiveVotes"`
	NegativeVotes   uint32 `json:"negativeVotes"`
	MongoID         string `json:"mongo_id"`
	Source          string `json:"source"`
	// Raw             any    `json:"raw"`
}
