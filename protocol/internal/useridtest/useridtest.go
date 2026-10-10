// Package useridtest holds the shared scenario for the per-protocol
// assignUserIDs tests.
//
// Inbounds hand the protocol service an opaque int per user and later map a
// connection's int back to a user name (the panel's UUID) for traffic
// accounting. The int therefore has to survive reorders, deletions and
// re-adds, and must never be handed to a different user: an established
// connection keeps the int it authenticated with, so reusing it would bill
// that connection's traffic to someone else.
package useridtest

import "testing"

// Assign runs one assignUserIDs round for the given user names (empty string
// = unnamed user) and returns the assigned IDs in input order.
type Assign func(names ...string) []int

// Maps returns the inbound's ID->name and name->ID maps after the last round.
type Maps func() (nameByID map[int]string, idByName map[string]int)

// Valid UUIDs, so protocols that parse the name/UUID (TUIC) accept them.
const (
	A = "00000000-0000-4000-8000-00000000000a"
	B = "00000000-0000-4000-8000-00000000000b"
	C = "00000000-0000-4000-8000-00000000000c"
	D = "00000000-0000-4000-8000-00000000000d"
)

func Run(t *testing.T, assign Assign, maps Maps) {
	t.Helper()
	issued := map[int]string{} // every ID ever handed out -> the user it went to

	round := func(names ...string) []int {
		t.Helper()
		ids := assign(names...)
		if len(ids) != len(names) {
			t.Fatalf("assign(%q) returned %d ids, want %d", names, len(ids), len(names))
		}
		for i, id := range ids {
			if prev, ok := issued[id]; ok && prev != names[i] {
				t.Fatalf("id %d reused: first %q, now %q", id, prev, names[i])
			}
			issued[id] = names[i]
		}
		nameByID, idByName := maps()
		named := 0
		for i, name := range names {
			if name == "" {
				if _, ok := nameByID[ids[i]]; ok {
					t.Fatalf("unnamed user's id %d resolves to a name", ids[i])
				}
				continue
			}
			named++
			if got := nameByID[ids[i]]; got != name {
				t.Fatalf("nameByID[%d] = %q, want %q", ids[i], got, name)
			}
			if got, ok := idByName[name]; !ok || got != ids[i] {
				t.Fatalf("idByName[%q] = %d (present %v), want %d", name, got, ok, ids[i])
			}
		}
		if len(nameByID) != named || len(idByName) != named {
			t.Fatalf("maps hold %d/%d entries after %q, want only the %d current users",
				len(nameByID), len(idByName), names, named)
		}
		return ids
	}

	first := round(A, B, C)
	idA, idB, idC := first[0], first[1], first[2]
	if idA == idB || idB == idC || idA == idC {
		t.Fatalf("ids not distinct: %v", first)
	}

	// Reorder and drop B: survivors keep their ids, B no longer resolves.
	if got := round(C, A); got[0] != idC || got[1] != idA {
		t.Fatalf("after reorder+delete got %v, want [%d %d]", got, idC, idA)
	}
	nameByID, _ := maps()
	if name := nameByID[idB]; name != "" {
		t.Fatalf("removed user's id %d still resolves to %q", idB, name)
	}

	// A new user must not inherit B's id (round() fails on any reuse).
	idD := round(A, C, D)[2]

	// Re-adding B may get a fresh id or its old one, never someone else's.
	if got := round(B, A, C, D); got[1] != idA || got[2] != idC || got[3] != idD {
		t.Fatalf("existing users' ids moved when B was re-added: %v", got)
	}

	// Unnamed users get distinct ids that collide with no named user.
	got := round("", "", A)
	if got[0] == got[1] {
		t.Fatalf("two unnamed users share id %d", got[0])
	}
	if got[2] != idA {
		t.Fatalf("A's id changed alongside unnamed users: %d, want %d", got[2], idA)
	}
}
