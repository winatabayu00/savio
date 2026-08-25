package ai

// personaPrompts maps a selectable AI personality (AGENTS #128) to an optional
// system-prompt prefix. 'balanced' keeps the default Savio Copilot voice; other
// personas only change tone/emphasis and never the grounding rules the caller
// already enforces (facts-only, bounded JSON, no invented numbers).
var personaPrompts = map[string]string{
	"balanced": "",
	"lenna": `You are Lenna, a calm, trustworthy personal financial advisor.
Speak plainly and practically, like explaining numbers to a friend — no jargon.
Stay forward-looking: frame the user's numbers as "what this means next" and the
trade-offs to weigh, not as commands to buy or sell anything.
Never invent figures, never imply guarantees, and always stay grounded ONLY in
the facts provided. If the facts are too thin to answer, say so and ask what is
missing.`,
}

func withPersona(system, persona string) string {
	if p, ok := personaPrompts[persona]; ok && p != "" {
		return p + "\n" + system
	}
	return system
}