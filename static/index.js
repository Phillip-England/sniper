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
  defaultClasses = ["bg-red-600", "hover:scale-105", "hover:bg-red-500"];
  recordingClasses = ["bg-red-700", "animate-pulse", "ring-4", "ring-red-900"];
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

// client/SniperCore.ts
class SniperCore {
  audio;
  ui;
  recognition = null;
  lastCommand = "";
  commandTriggers = [];
  silenceTimer = null;
  currentInterim = "";
  state = {
    isRecording: false,
    isLogging: true,
    shouldContinue: false
  };
  constructor(audio, ui) {
    this.audio = audio;
    this.ui = ui;
    this.initializeSpeechEngine();
    this.bindEvents();
    this.loadCommandTriggers();
  }
  async loadCommandTriggers() {
    try {
      const response = await fetch("/api/commands/min");
      if (!response.ok) {
        throw new Error(`Failed to fetch commands: ${response.statusText}`);
      }
      const commands = await response.json();
      this.commandTriggers = commands.reduce((acc, cmd) => {
        return acc.concat(cmd.called_by);
      }, []);
      console.log(`[Sniper] Loaded ${this.commandTriggers.length} command triggers.`);
    } catch (err) {
      console.warn("[Sniper] Could not load command registry. Command validation may be limited.", err);
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
    this.recognition.onresult = (event) => {
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
        this.ui.updateText("", interimChunk, this.state.isLogging);
        if (interimChunk.trim().length > 0) {
          this.silenceTimer = setTimeout(() => {
            console.log("[Sniper] Force finalizing stuck interim result...");
            this.executeFinalSequence(this.currentInterim, "");
            this.currentInterim = "";
          }, 1000);
        }
      }
    };
    this.recognition.onerror = (event) => {
      if (event.error === "not-allowed" || event.error === "service-not-allowed") {
        this.state.shouldContinue = false;
        this.stop();
      }
    };
  }
  executeFinalSequence(finalText, interimText) {
    this.ui.updateText(finalText, interimText, this.state.isLogging);
    const processed = this.handleCommands(finalText);
    if (!processed.capturedByCommand && this.state.isLogging) {
      this.audio.play("click");
    }
  }
  async sendToBackend(command) {
    try {
      console.log(`[Sniper] Sending to backend: ${command}`);
      await fetch("http://localhost:8000/api/data", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ command })
      });
    } catch (err) {
      console.warn("[Sniper] Backend connection failed. Is localhost:8000 running?");
    }
  }
  handleCommands(text) {
    const command = text.toLowerCase().trim().replace(/[?!]/g, "");
    if (command) {
      this.lastCommand = command;
    }
    switch (command.replace(/[.,]/g, "")) {
      case "exit":
        this.audio.play("sniper-exit");
        this.ui.clearText();
        this.state.shouldContinue = false;
        this.stop();
        return { capturedByCommand: true };
      case "off":
        this.audio.play("sniper-off");
        this.ui.clearText();
        this.state.isLogging = false;
        this.ui.updateGreenDot(this.state.isRecording, this.state.isLogging);
        return { capturedByCommand: true };
      case "on":
        this.audio.play("sniper-on");
        this.state.isLogging = true;
        this.ui.updateGreenDot(this.state.isRecording, this.state.isLogging);
        return { capturedByCommand: true };
      default:
        if (this.state.isLogging) {
          this.sendToBackend(command);
        }
        return { capturedByCommand: false };
    }
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
document.addEventListener("DOMContentLoaded", () => {
  const audioManager = new AudioManager;
  const uiManager = new UIManager;
  const app = new SniperCore(audioManager, uiManager);
});
