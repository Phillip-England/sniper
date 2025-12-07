import type { IRecognitionMode, SniperCore } from "./SniperCore";
import type { SpeechRecognitionEvent } from "./SpeechTypes";
import { StaticCommandHandler } from "./StaticCommandHandler";
import { Throttler } from "./Thottler";

export class RapidMode implements IRecognitionMode {
  core: SniperCore;
  sysCmd: StaticCommandHandler;
  throttler: Throttler
  sameWordThrottler: Throttler
  prevWord: string

  constructor(core: SniperCore) {
    this.core = core;
    this.sysCmd = new StaticCommandHandler(core);
    this.throttler = new Throttler(400)
    this.sameWordThrottler = new Throttler(1100)
    this.prevWord = ''
  }

  name() {
    return "rapid"
  }

  async handleResult(event: SpeechRecognitionEvent): Promise<void> {
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

      let words = transcript.split(' ')
      let lastWord = words[words.length-1]
      if (!lastWord) {
        continue
      }

      if (this.throttler.isWaiting()) {
        continue
      }



      this.throttler.wait()
      this.sameWordThrottler.wait()

      if (this.prevWord == lastWord && this.sameWordThrottler.isWaiting()) {
        continue
      }

      let status = await this.core.api.sendCommand(lastWord, this.core.mode);
      if (status != 200) {
        this.core.ui.updateText("", "", this.core.state.isLogging);
        return
      }
      this.prevWord = lastWord
      this.core.audio.play("click");
      // 2. UI Updates
      if (result.isFinal) {
        this.core.ui.updateText(transcript, "", this.core.state.isLogging);
      } else {
        this.core.ui.updateText("", transcript, this.core.state.isLogging);
      }
    }
  }
}

const isNumericString = (str: string) => {
  if (typeof str !== 'string') {
    return false;
  }
  return Number.isFinite(+str) && str.trim() !== ''; 
};