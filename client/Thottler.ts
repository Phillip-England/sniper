export class Throttler {
  private duration: number;
  private timeoutId: any = null; // Using 'any' to support both Node.js and Browser environments
  private _waiting: boolean = false;

  constructor(durationMs: number) {
    this.duration = durationMs;
  }

  /**
   * Resets the timer. The state will remain "waiting" for the
   * duration specified in the constructor.
   */
  public wait(): void {
    // If there is an active timer, clear it to reset the clock
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
    }

    this._waiting = true;

    // Start the new timer
    this.timeoutId = setTimeout(() => {
      this._waiting = false;
      this.timeoutId = null;
    }, this.duration);
  }

  /**
   * Returns true if the timer is currently active.
   */
  public isWaiting(): boolean {
    return this._waiting;
  }

  /**
   * Optional: Call this if you need to force stop the throttler immediately.
   */
  public cancel(): void {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
    this._waiting = false;
  }
}
