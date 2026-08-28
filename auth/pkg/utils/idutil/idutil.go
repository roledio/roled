package idutil

import (
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	NanoIDDefaultSize = 21
)

// NewID generates a globally unique, opaque identifier for Roled entities.
//
// Roled intentionally uses a string-based ID generated from UUID v7 encoded
// with shortuuid, instead of auto-increment integers or fully custom ID formats.
//
// Rationale:
//
// 1. UUID v7 as entropy source
//   - UUID v7 is time-ordered at the bit level (unlike UUID v4).
//   - This provides better index locality for inserts compared to fully random IDs,
//     reducing the risk of severe B-Tree fragmentation in common relational databases.
//   - At the same time, UUID v7 avoids strictly monotonic behavior (e.g. ULID),
//     which can cause write hotspots in distributed or high-concurrency systems.
//
// 2. shortuuid encoding
//   - Encodes the 128-bit UUID into a shorter, URL-safe string (~22 chars).
//   - Uses a non-ambiguous, case-sensitive character set:
//   - No visually confusing characters (0/O, 1/I/l, etc).
//   - Case sensitivity is preserved, so IDs are never normalized or lowercased.
//   - Stored as plain text (CHAR/VARCHAR), making it easy to inspect, debug,
//     copy-paste, and query manually using SQL tools.
//
// 3. Opaque ID by design
//   - IDs are not intended to convey business meaning or ordering guarantees.
//   - Ordering, filtering, and range queries must rely on explicit columns
//     such as created_at / updated_at.
//   - This avoids accidental coupling between ID format and application logic.
//
// 4. Trade-offs vs alternatives
//   - vs auto-increment integer:
//     Pros: no data size leakage, safer for public APIs, globally unique.
//     Cons: larger storage footprint.
//   - vs fully random IDs (UUID v4, NanoID):
//     Pros: better insert locality and index behavior at scale.
//     Cons: slightly longer and marginally less random.
//   - vs strictly sortable IDs (ULID, Snowflake):
//     Pros: avoids hot partitions and monotonic write contention.
//     Cons: cannot be relied on for total ordering.
//
// Summary:
// This approach balances uniqueness, operational safety, index behavior,
// human readability, and long-term flexibility without over-optimizing for
// premature scaling concerns.
func NewID() string {
	uuid, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return shortuuid.DefaultEncoder.Encode(uuid)
}

// NanoID generates a random NanoID string with the given length.
// If length is not provided, it defaults to 21.
// The generated ID uses a custom set of characters that is used in shortuuid,
// which includes numbers and letters but excludes ambiguous characters such as 0, O, 1, I, l.
func NanoID(length ...int) string {
	size := NanoIDDefaultSize
	if len(length) > 0 && length[0] > 0 {
		size = length[0]
	}
	return gonanoid.MustGenerate(shortuuid.DefaultAlphabet, size)
}
