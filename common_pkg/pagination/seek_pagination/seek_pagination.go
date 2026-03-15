package seek_pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)


/*
Timeline:
  Page 1 fetched at T=0:  [User1, User2, User3, User4, User5]  ← anchor = User5
  New user added at T=1:  [NewUser, User1, User2, User3, User4, User5, ...]
  Page 2 fetched at T=2:  ???

Without anchoring → User5 appears twice (or gets skipped). With anchoring, we split into two "rails":
	New rail: items newer than the anchor (injected at top)
	Historical rail: items older than the anchor (normal next-page seek)

T=0  Client fetches Page 1
     DB has: [U10, U9, U8, U7, U6, U5, U4, U3, U2, U1]
     Response: data=[U10..U1], anchor=U10, last=U1, new_rail_offset=0

T=1  U11, U12 added to DB

T=2  Client fetches Page 2 (cursor from page 1)
     New rail query : created_at > U10.created_at  → [U12, U11]  (offset=0, limit=2)
     Remaining slots: 10 - 2 = 8
     Old rail query : id < U1.id AND id <= U10.id  → [U9..U2]   (8 rows)
     Response: data=[U12, U11, U9, U8...U2], new_count=2, next cursor: last=U2, new_rail_offset=2

T=3  Client fetches Page 3
     New rail query : offset=2 → [] (no more new items)
     Old rail : id < U2.id → [U1]
     Response: data=[U1], has_next=false

*/

// Cursor is the opaque token passed between client and server.
// It encodes everything needed to resume pagination exactly where it left off,
// including a session anchor so newly inserted rows don't corrupt page results.
type Cursor struct {
	// Seek position: the last item the client received on the historical rail
	LastID        uint      `json:"lid"`
	LastCreatedAt time.Time `json:"lca"`

	// Anchor: set once on page 1, never changes for the life of the session.
	// Rows newer than this are on the "new rail"; rows older are on the "history rail".
	AnchorID        uint      `json:"aid"`
	AnchorCreatedAt time.Time `json:"aca"`

	// NewRailOffset tracks how many new-rail rows have already been delivered,
	// so we can paginate the new rail with OFFSET (acceptable: new rail is tiny).
	NewRailOffset int `json:"nro"`
}

// Encode encodes cursor to a base64 opaque token for the client
func (c *Cursor) EncodeCursor() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", errors.New("failed to encode cursor")
	}
	// return base64.URLEncoding.EncodeToString(b), nil // base64.URLEncoding which produces padding characters (=). When put in a URL query string these must be percent-encoded as %3D. If the client passes the raw token without encoding it, decoding silently fails
	return base64.RawURLEncoding.EncodeToString(b), nil // base64.RawURLEncoding omits padding, so the token is URL-safe without further encoding
}

// DecodeCursor decodes an opaque cursor token
func DecodeCursor(token string) (*Cursor, error) {
	if token == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(token) 
	if err != nil {
		return nil, errors.New("invalid cursor: bad encoding")
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, errors.New("invalid cursor: malformed payload")
	}
	return &c, nil
}


