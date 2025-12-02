package sniper

import "time"

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

// WaitBefore returns an EffectFunc that sleeps for the specified milliseconds
// BEFORE executing the next function in the chain.
func WaitBefore(ms int) EffectFunc {
	return func(e *Engine, next func() error) error {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return next()
	}
}

// WaitAfter returns an EffectFunc that executes the next function in the chain,
// and then sleeps for the specified milliseconds AFTER it completes.
func WaitAfter(ms int) EffectFunc {
	return func(e *Engine, next func() error) error {
		// Execute the action first
		err := next()
		if err != nil {
			return err
		}

		// If successful, wait the specified duration
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return nil
	}
}

// KillAfter returns an EffectFunc that sets the Engine.IsOperating flag to false
// AFTER the command executes successfully.
func KillAfter() EffectFunc {
	return func(e *Engine, next func() error) error {
		// Execute the action first
		err := next()
		if err != nil {
			return err
		}

		// Change IsOperating to false to stop the engine
		e.IsOperating = false
		return nil
	}
}

// ClickBefore returns an EffectFunc that performs a mouse click
// BEFORE executing the next function in the chain.
func ClickBefore() EffectFunc {
	return func(e *Engine, next func() error) error {
		// Click to focus or position cursor
		e.Mouse.DoubleClick()
		time.Sleep(time.Millisecond * 50)
		return next()
	}
}

// ClickAfter returns an EffectFunc that performs a mouse click
// AFTER executing the next function in the chain.
func ClickAfter() EffectFunc {
	return func(e *Engine, next func() error) error {
		// Execute the action first
		err := next()
		if err != nil {
			return err
		}

		// Click mouse after the action completes
		e.Mouse.DoubleClick()
		time.Sleep(time.Millisecond * 50)
		return nil
	}
}
