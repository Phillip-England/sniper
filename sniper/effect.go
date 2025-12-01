package sniper

// EffectFunc is the signature for an effect (middleware).
// It takes the Engine and a 'next' function which represents the next link in the chain.
type EffectFunc func(e *Engine, next func() error) error

// EffectChain wraps a core action function with a slice of effects.
// It executes effects in order: effects[0] wraps effects[1], which wraps... the handler.
func EffectChain(e *Engine, handler func() error, effects ...EffectFunc) error {
	// If there are no effects, just run the core handler.
	if len(effects) == 0 {
		return handler()
	}

	// We wrap the handler with the effects.
	// We loop backwards so that effects[0] becomes the outermost wrapper.
	next := handler
	for i := len(effects) - 1; i >= 0; i-- {
		eff := effects[i]
		currentNext := next
		next = func() error {
			return eff(e, currentNext)
		}
	}

	// Execute the outermost function
	return next()
}
