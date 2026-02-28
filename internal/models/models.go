package models

// URL represents a shortened URL record.
type URL struct {
	ID          int64  `db:"id" json:"id"`
	ShortCode   string `db:"short_code" json:"short_code"`
	OriginalURL string `db:"original_url" json:"original_url"`
	CreatedAt   string `db:"created_at" json:"created_at"`
	Clicks      int64  `db:"clicks" json:"clicks"`
}

// Click represents a single click/visit on a shortened URL.
type Click struct {
	ID        int64  `db:"id" json:"id"`
	URLID     int64  `db:"url_id" json:"url_id"`
	ClickedAt string `db:"clicked_at" json:"clicked_at"`
	Referrer  string `db:"referrer" json:"referrer"`
	UserAgent string `db:"user_agent" json:"user_agent"`
	IPAddress string `db:"ip_address" json:"ip_address"`
}
