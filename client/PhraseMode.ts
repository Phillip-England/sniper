import type { IRecognitionMode, SniperCore } from "./SniperCore";
import type { SpeechRecognitionEvent } from "./SpeechTypes";
import { StaticCommandHandler } from "./StaticCommandHandler";

export class PhraseMode implements IRecognitionMode {
  private core: SniperCore;
  private sysCmd: StaticCommandHandler;
  private silenceTimer: any = null;
  private currentInterim: string = "";

  constructor(core: SniperCore) {
    this.core = core;
    this.sysCmd = new StaticCommandHandler(core);
  }

  public handleResult(event: SpeechRecognitionEvent): void {
    // 1. Clear existing silence timer
    if (this.silenceTimer) {
      clearTimeout(this.silenceTimer);
      this.silenceTimer = null;
    }

    let finalChunk = "";
    let interimChunk = "";

    // 2. Accumulate Results
    for (let i = event.resultIndex; i < event.results.length; ++i) {
      const result = event.results[i];
      if (!result || !result.length) continue;
      const alternative = result[0];
      if (!alternative) continue;

      if (result.isFinal) {
        finalChunk += alternative.transcript;
      } else {
        interimChunk += alternative.transcript;
      }
    }

    this.currentInterim = interimChunk;

    // 3. Determine Action
    if (finalChunk) {
      this.executeFinalSequence(finalChunk, interimChunk);
    } else {
      this.core.ui.updateText("", interimChunk, this.core.state.isLogging);

      // Force finalize if stuck in interim state for too long
      if (interimChunk.trim().length > 0) {
        this.silenceTimer = setTimeout(() => {
          console.log("[Sniper] Force finalizing stuck interim result...");
          this.executeFinalSequence(this.currentInterim, "");
          this.currentInterim = "";
        }, 1000);
      }
    }
  }

  private executeFinalSequence(finalText: string, interimText: string) {
    this.core.ui.updateText(finalText, interimText, this.core.state.isLogging);

    // 1. Delegate to Shared System Handler
    const wasSystemCommand = this.sysCmd.process(finalText);

    // 2. If not a system command, treat as a Phrase
    if (!wasSystemCommand) {
      if (this.core.state.isLogging) {
        this.core.api.sendCommand(finalText.trim());
        this.core.audio.play("click");
      }
    }
  }
}
