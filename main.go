package main

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	// No need to create a device; robotgo controls the actual system mouse.

	fmt.Println("✅ RobotGo initialized.")
	fmt.Println("⏳ Waiting 2 seconds (switch to your desktop so you can see it)...")
	time.Sleep(2 * time.Second)

	fmt.Println("🚀 Moving mouse...")

	// Move the mouse in a square pattern
	for i := 0; i < 5; i++ {
		// Move 50 pixels Right (X positive, Y 0)
		robotgo.MoveRelative(50, 0)
		time.Sleep(200 * time.Millisecond)

		// Move 50 pixels Down (X 0, Y positive)
		robotgo.MoveRelative(0, 50)
		time.Sleep(200 * time.Millisecond)

		// Move 50 pixels Left (X negative, Y 0)
		robotgo.MoveRelative(-50, 0)
		time.Sleep(200 * time.Millisecond)

		// Move 50 pixels Up (X 0, Y negative)
		robotgo.MoveRelative(0, -50)
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Done.")
}
