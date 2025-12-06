import type { IRecognitionMode, SniperCore } from "./SniperCore";
import type { SpeechRecognitionEvent } from "./SpeechTypes";
import { StaticCommandHandler } from "./StaticCommandHandler";

export class RapidMode implements IRecognitionMode {
  private core: SniperCore;
  private sysCmd: StaticCommandHandler;

  constructor(core: SniperCore) {
    this.core = core;
    this.sysCmd = new StaticCommandHandler(core);
  }

  public handleResult(event: SpeechRecognitionEvent): void {
    for (let i = event.resultIndex; i < event.results.length; ++i) {
      const result = event.results[i];
      if (!result || !result.length) continue;
      
      const alternative = result[0];
      if (!alternative) continue;
      
      const transcript = alternative.transcript;

      // 1. Delegate to Shared System Handler
      if (this.sysCmd.process(transcript)) {
        return; 
      }

      // 2. UI Updates
      if (result.isFinal) {
        this.core.ui.updateText(transcript, '', this.core.state.isLogging);
      } else {
        this.core.ui.updateText('', transcript, this.core.state.isLogging);
      }

      // 3. Rapid Logic (Fire Immediately on Final)
      if (result.isFinal && this.core.state.isLogging) {
        this.core.commandCenter.consume(transcript);
        this.core.api.sendCommand(transcript.trim());
        this.core.audio.play('click');
      }
    }
  }
}