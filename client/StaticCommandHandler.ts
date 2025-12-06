import { SniperCore } from "./SniperCore";

export class StaticCommandHandler {
  private core: SniperCore;

  constructor(core: SniperCore) {
    this.core = core;
  }

  /**
   * Checks if the text is a system command (exit, off, on).
   * Returns true if a command was executed.
   */
  public process(text: string): boolean {
    // Standardize cleaning regex for both modes
    const command = text.toLowerCase().trim().replace(/[?!.,]/g, "");

    switch (command) {

      case "rapid":
        this.core.setMode('rapid')
        return true

      case "phrase":
        this.core.setMode('phrase')
        return true

      case "exit":
        this.core.audio.play("sniper-exit");
        this.core.ui.clearText();
        this.core.state.shouldContinue = false;
        this.core.stop();
        return true;

      case "off":
        this.core.audio.play("sniper-off");
        this.core.ui.clearText();
        this.core.state.isLogging = false;
        this.core.ui.updateGreenDot(
          this.core.state.isRecording,
          this.core.state.isLogging,
        );
        return true;

      case "on":
        this.core.audio.play("sniper-on");
        this.core.state.isLogging = true;
        this.core.ui.updateGreenDot(
          this.core.state.isRecording,
          this.core.state.isLogging,
        );
        return true;
    }

    return false;
  }
}
