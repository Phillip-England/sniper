// client/AudioManager.ts
class AudioManager {
  sounds;
  constructor() {
    this.sounds = {
      "sniper-on": new Audio("/static/on.wav"),
      "sniper-off": new Audio("/static/off.wav"),
      "sniper-exit": new Audio("/static/exit.wav"),
      "sniper-clear": new Audio("/static/clear.wav"),
      "sniper-copy": new Audio("/static/copy.wav"),
      "sniper-search": new Audio("/static/search.wav"),
      "sniper-visit": new Audio("/static/visit.wav"),
      click: new Audio("/static/click.wav")
    };
    this.preloadSounds();
  }
  preloadSounds() {
    Object.values(this.sounds).forEach((sound) => {
      sound.preload = "auto";
      sound.load();
    });
  }
  play(type) {
    const audio = this.sounds[type];
    audio.currentTime = 0;
    audio.play().catch((e) => {
      console.warn(`Audio playback failed for ${type}`, e);
    });
  }
}

// client/CommandCenter.ts
class CommandCenter {
  triggers = [];
  lastConsumed = "";
  strategy;
  constructor() {
    this.load();
    this.strategy = 1 /* Phrase */;
  }
  async load() {
    try {
      const response = await fetch("/api/commands/min");
      if (!response.ok) {
        throw new Error(`Failed to fetch commands: ${response.statusText}`);
      }
      const commands = await response.json();
      this.triggers = commands.reduce((acc, cmd) => {
        return acc.concat(cmd.called_by);
      }, []);
      console.log(`[CommandCenter] Loaded ${this.triggers.length} command triggers.`);
    } catch (err) {
      console.warn("[CommandCenter] Could not load command registry.", err);
    }
  }
  getTriggers() {
    return this.triggers;
  }
  isValidTrigger(phrase) {
    return this.triggers.includes(phrase.toLowerCase());
  }
  consume(input) {
    const cleanInput = input.trim();
    if (this.isValidTrigger(cleanInput)) {
      this.lastConsumed = cleanInput;
      console.log(`[CommandCenter] consumed: "${this.lastConsumed}"`);
      return true;
    }
    return false;
  }
  getLastConsumed() {
    return this.lastConsumed;
  }
}

// client/UIManager.ts
class UIManager {
  btn;
  transcriptEl;
  interimEl;
  outputContainer;
  placeholder;
  statusText;
  copyBtn;
  greenDot;
  defaultClasses = [
    "bg-red-600",
    "hover:scale-105",
    "hover:bg-red-500"
  ];
  recordingClasses = [
    "bg-red-700",
    "animate-pulse",
    "ring-4",
    "ring-red-900"
  ];
  constructor() {
    this.btn = document.getElementById("record-button");
    this.transcriptEl = document.getElementById("transcript");
    this.interimEl = document.getElementById("interim");
    this.outputContainer = document.getElementById("output-container");
    this.placeholder = document.getElementById("placeholder");
    this.statusText = document.getElementById("status-text");
    this.copyBtn = this.outputContainer.querySelector("button");
    this.greenDot = document.getElementById("green-dot");
    this.setupCopyButton();
  }
  getRecordButton() {
    return this.btn;
  }
  updateGreenDot(isRecording, isLogging) {
    if (isRecording && isLogging) {
      this.greenDot.classList.remove("opacity-0");
    } else {
      this.greenDot.classList.add("opacity-0");
    }
  }
  setRecordingState(isRecording) {
    if (isRecording) {
      this.btn.classList.remove(...this.defaultClasses);
      this.btn.classList.add(...this.recordingClasses);
      this.statusText.classList.remove("opacity-0");
      this.outputContainer.classList.remove("opacity-0", "translate-y-10");
      this.placeholder.textContent = "Listening...";
    } else {
      this.btn.classList.remove(...this.recordingClasses);
      this.btn.classList.add(...this.defaultClasses);
      this.statusText.classList.add("opacity-0");
      this.placeholder.textContent = "Tap button to speak...";
      this.togglePlaceholder();
    }
  }
  updateText(final, interim, isLogging) {
    if (isLogging) {
      if (final) {
        this.transcriptEl.innerText = final;
      }
      this.interimEl.innerText = interim;
    } else {
      this.interimEl.innerText = "";
    }
    this.togglePlaceholder();
  }
  clearText() {
    this.transcriptEl.innerText = "";
    this.interimEl.innerText = "";
    this.togglePlaceholder();
  }
  getText() {
    return this.transcriptEl.innerText;
  }
  togglePlaceholder() {
    if (this.transcriptEl.innerText || this.interimEl.innerText) {
      this.placeholder.classList.add("hidden");
    } else {
      this.placeholder.classList.remove("hidden");
    }
  }
  setupCopyButton() {
    if (!this.copyBtn)
      return;
    this.copyBtn.onclick = null;
    this.copyBtn.addEventListener("click", () => {
      const text = this.transcriptEl.innerText;
      if (text) {
        navigator.clipboard.writeText(text);
        const originalText = this.copyBtn.innerText;
        this.copyBtn.innerText = "[ COPIED! ]";
        setTimeout(() => this.copyBtn.innerText = originalText, 2000);
      }
    });
  }
}

// client/SniperService.ts
class SniperService {
  baseUrl = "http://localhost:9090";
  async sendCommand(command, mode) {
    try {
      console.log(`[SniperService] Sending: ${command}`);
      let reqBody = JSON.stringify({
        command,
        mode: mode.name()
      });
      const response = await fetch(`${this.baseUrl}/api/data`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: reqBody
      });
      if (!response.ok) {
        console.warn(`[SniperService] Request failed with status: ${response.status} ${response.statusText}`);
      }
      return response.status;
    } catch (err) {
      console.warn("[SniperService] Connection failed. Is the backend running?", err);
      return 0;
    }
  }
}

// client/StaticCommandHandler.ts
class StaticCommandHandler {
  core;
  constructor(core) {
    this.core = core;
  }
  process(text) {
    const command = text.toLowerCase().trim().replace(/[?!.,]/g, "");
    switch (command) {
      case "rapid":
        this.core.setMode("rapid");
        return true;
      case "rabbit":
        this.core.setMode("rapid");
        return true;
      case "phrase":
        this.core.setMode("phrase");
        return true;
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
        this.core.ui.updateGreenDot(this.core.state.isRecording, this.core.state.isLogging);
        return true;
      case "on":
        this.core.audio.play("sniper-on");
        this.core.state.isLogging = true;
        this.core.ui.updateGreenDot(this.core.state.isRecording, this.core.state.isLogging);
        return true;
    }
    return false;
  }
}

// client/PhraseMode.ts
class PhraseMode {
  core;
  sysCmd;
  silenceTimer = null;
  currentInterim = "";
  constructor(core) {
    this.core = core;
    this.sysCmd = new StaticCommandHandler(core);
  }
  name() {
    return "phrase";
  }
  async handleResult(event) {
    if (this.silenceTimer) {
      clearTimeout(this.silenceTimer);
      this.silenceTimer = null;
    }
    let finalChunk = "";
    let interimChunk = "";
    for (let i = event.resultIndex;i < event.results.length; ++i) {
      const result = event.results[i];
      if (!result || !result.length)
        continue;
      const alternative = result[0];
      if (!alternative)
        continue;
      if (result.isFinal) {
        finalChunk += alternative.transcript;
      } else {
        interimChunk += alternative.transcript;
      }
    }
    this.currentInterim = interimChunk;
    if (finalChunk) {
      this.executeFinalSequence(finalChunk, interimChunk);
    } else {
      this.core.ui.updateText("", interimChunk, this.core.state.isLogging);
      if (interimChunk.trim().length > 0) {
        this.silenceTimer = setTimeout(() => {
          console.log("[Sniper] Force finalizing stuck interim result...");
          this.executeFinalSequence(this.currentInterim, "");
          this.currentInterim = "";
        }, 1000);
      }
    }
  }
  executeFinalSequence(finalText, interimText) {
    this.core.ui.updateText(finalText, interimText, this.core.state.isLogging);
    const wasSystemCommand = this.sysCmd.process(finalText);
    if (!wasSystemCommand) {
      if (this.core.state.isLogging) {
        this.core.api.sendCommand(finalText.trim(), this.core.mode);
        this.core.audio.play("click");
      }
    }
  }
}

// client/Thottler.ts
class Throttler {
  duration;
  timeoutId = null;
  _waiting = false;
  constructor(durationMs) {
    this.duration = durationMs;
  }
  wait() {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
    }
    this._waiting = true;
    this.timeoutId = setTimeout(() => {
      this._waiting = false;
      this.timeoutId = null;
    }, this.duration);
  }
  isWaiting() {
    return this._waiting;
  }
  cancel() {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
    this._waiting = false;
  }
}

// client/RapidMode.ts
class RapidMode {
  core;
  sysCmd;
  throttler;
  sameWordThrottler;
  prevWord;
  constructor(core) {
    this.core = core;
    this.sysCmd = new StaticCommandHandler(core);
    this.throttler = new Throttler(400);
    this.sameWordThrottler = new Throttler(1100);
    this.prevWord = "";
  }
  name() {
    return "rapid";
  }
  async handleResult(event) {
    for (let i = event.resultIndex;i < event.results.length; ++i) {
      const result = event.results[i];
      if (!result || !result.length)
        continue;
      const alternative = result[0];
      if (!alternative)
        continue;
      const transcript = alternative.transcript;
      if (this.sysCmd.process(transcript)) {
        return;
      }
      let words = transcript.split(" ");
      let lastWord = words[words.length - 1];
      if (!lastWord) {
        continue;
      }
      if (this.throttler.isWaiting()) {
        continue;
      }
      this.throttler.wait();
      this.sameWordThrottler.wait();
      if (this.prevWord == lastWord && this.sameWordThrottler.isWaiting()) {
        continue;
      }
      let status = await this.core.api.sendCommand(lastWord, this.core.mode);
      if (status != 200) {
        this.core.ui.updateText("", "", this.core.state.isLogging);
        return;
      }
      this.prevWord = lastWord;
      this.core.audio.play("click");
      if (result.isFinal) {
        this.core.ui.updateText(transcript, "", this.core.state.isLogging);
      } else {
        this.core.ui.updateText("", transcript, this.core.state.isLogging);
      }
    }
  }
}

// client/SniperCore.ts
class SniperCore {
  audio;
  ui;
  api;
  commandCenter;
  state = {
    isRecording: false,
    isLogging: true,
    shouldContinue: false
  };
  recognition = null;
  mode;
  constructor() {
    this.audio = new AudioManager;
    this.ui = new UIManager;
    this.commandCenter = new CommandCenter;
    this.api = new SniperService;
    this.mode = new RapidMode(this);
  }
  static async new() {
    const core = new SniperCore;
    core.initializeSpeechEngine();
    core.bindEvents();
    return core;
  }
  setMode(modeType) {
    if (modeType === "rapid") {
      console.log("[Sniper] Switching to Rapid Mode");
      this.mode = new RapidMode(this);
      this.audio.play("sniper-on");
    } else {
      console.log("[Sniper] Switching to Phrase Mode");
      this.mode = new PhraseMode(this);
      this.audio.play("click");
    }
  }
  initializeSpeechEngine() {
    const SpeechRecognitionCtor = window.SpeechRecognition || window.webkitSpeechRecognition;
    if (!SpeechRecognitionCtor) {
      alert("Browser not supported. Try Chrome/Safari.");
      return;
    }
    this.recognition = new SpeechRecognitionCtor;
    this.recognition.continuous = true;
    this.recognition.interimResults = true;
    this.recognition.lang = "en-US";
    this.setupRecognitionHandlers();
  }
  setupRecognitionHandlers() {
    if (!this.recognition)
      return;
    this.recognition.onstart = () => {
      if (!this.state.isRecording) {
        this.audio.play("sniper-on");
      }
      this.state.isRecording = true;
      this.ui.setRecordingState(true);
      this.ui.updateGreenDot(this.state.isRecording, this.state.isLogging);
    };
    this.recognition.onend = () => {
      if (this.state.shouldContinue) {
        this.recognition?.start();
        return;
      }
      this.state.isRecording = false;
      this.ui.setRecordingState(false);
      this.ui.updateGreenDot(this.state.isRecording, this.state.isLogging);
    };
    this.recognition.onresult = async (event) => {
      await this.mode.handleResult(event);
    };
    this.recognition.onerror = (event) => {
      if (event.error === "not-allowed" || event.error === "service-not-allowed") {
        this.state.shouldContinue = false;
        this.stop();
      }
    };
  }
  bindEvents() {
    const btn = this.ui.getRecordButton();
    if (btn) {
      btn.addEventListener("click", () => {
        if (this.state.isRecording) {
          this.stop();
        } else {
          this.start();
        }
      });
    }
  }
  start() {
    this.state.shouldContinue = true;
    this.state.isLogging = true;
    this.recognition?.start();
  }
  stop() {
    this.state.shouldContinue = false;
    this.recognition?.stop();
  }
}

// client/index.ts
document.addEventListener("DOMContentLoaded", async () => {
  const app = await SniperCore.new();
});
