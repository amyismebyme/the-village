package reddit

type ListingResponse struct {
	Data ListingData `json:"data"`
}

type ListingData struct {
	Children []ListingChild `json:"children"`
	After    *string        `json:"after"`
	Before   *string        `json:"before"`
}

type ListingChild struct {
	Kind string     `json:"kind"`
	Data RedditPost `json:"data"`
}

type RedditPost struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Title      string  `json:"title"`
	SelfText   string  `json:"selftext"`
	URL        string  `json:"url"`
	Permalink  string  `json:"permalink"`
	Subreddit  string  `json:"subreddit"`
	Author     string  `json:"author"`
	CreatedUTC float64 `json:"created_utc"`
	IsSelf     bool    `json:"is_self"`
	RemovedBy  *string `json:"removed_by"`
}
