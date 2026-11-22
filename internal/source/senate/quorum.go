package senate

import "strings"

// Results a voting can have. The Senate publishes the quorum and the totals,
// but not whether the voting was approved, so the result is calculated.
const (
	resultApproved = "Aprobado"
	resultRejected = "Rechazado"
	// resultTie is a simple majority voting with as many votes in favor as
	// against. The Senate repeats those votings (23 of the 34 ties of the current
	// legislature are followed by a repeat in the same session), so it is
	// neither approved nor rejected.
	resultTie = "Empate"
)

// outcome calculates the result of a voting from its quorum.
//
//   - "Mayoría simple": more votes in favor than against.
//   - "Q.C." (qualified quorum): absolute majority of the senators in office.
//   - "Cuatro séptimos", "Dos tercios", "Tres quintos": that fraction of the
//     senators in office.
//
// Those quorums count the senators in office (seats), not the votes cast. It
// returns "" for a quorum it does not know, instead of guessing.
func outcome(quorum string, yes, no, seats int) string {
	q := strings.Join(words(quorum), " ") // lowercase, no accents, no punctuation

	switch {
	case strings.Contains(q, "mayoria simple"):
		switch {
		case yes > no:
			return resultApproved
		case yes < no:
			return resultRejected
		}
		return resultTie
	case strings.Contains(q, "cuatro septimos"):
		return reached(yes*7 >= 4*seats)
	case strings.Contains(q, "dos tercios"):
		return reached(yes*3 >= 2*seats)
	case strings.Contains(q, "tres quintos"):
		return reached(yes*5 >= 3*seats)
	case q == "q c", strings.Contains(q, "mayoria absoluta"):
		return reached(yes*2 > seats)
	}
	return ""
}

func reached(ok bool) string {
	if ok {
		return resultApproved
	}
	return resultRejected
}
