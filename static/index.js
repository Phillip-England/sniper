// client/index.ts
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

class UIManager {
  btn;
  transcriptEl;
  interimEl;
  outputContainer;
  placeholder;
  statusText;
  greenDot;
  commandTimeout = null;
  defaultClasses = ["bg-red-600", "hover:scale-105", "hover:bg-red-500"];
  recordingClasses = ["bg-red-700", "animate-pulse", "ring-4", "ring-red-900"];
  constructor() {
    this.btn = document.getElementById("record-button");
    this.transcriptEl = document.getElementById("transcript");
    this.interimEl = document.getElementById("interim");
    this.outputContainer = document.getElementById("output-container");
    this.placeholder = document.getElementById("placeholder");
    this.statusText = document.getElementById("status-text");
    this.greenDot = document.getElementById("green-dot");
    this.transcriptEl.innerText = "";
    this.interimEl.innerText = "";
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
      this.placeholder.textContent = "Listening for commands...";
      this.placeholder.classList.remove("hidden");
    } else {
      this.btn.classList.remove(...this.recordingClasses);
      this.btn.classList.add(...this.defaultClasses);
      this.statusText.classList.add("opacity-0");
      this.placeholder.textContent = "Tap button to speak...";
      this.placeholder.classList.remove("hidden");
      this.transcriptEl.innerText = "";
    }
  }
  showCommand(command) {
    this.placeholder.classList.add("hidden");
    this.transcriptEl.innerText = `[ ${command.toUpperCase()} ]`;
    if (this.commandTimeout) {
      window.clearTimeout(this.commandTimeout);
    }
    this.commandTimeout = window.setTimeout(() => {
      this.transcriptEl.innerText = "";
      this.placeholder.classList.remove("hidden");
    }, 2000);
  }
  clearText() {
    this.transcriptEl.innerText = "";
    this.interimEl.innerText = "";
    this.placeholder.classList.remove("hidden");
  }
}

class SniperCore {
  audio;
  ui;
  recognition = null;
  lastActionCommand = "";
  previousProcessedToken = "";
  lastResultIndex = -1;
  lastExecutionTime = 0;
  THROTTLE_DELAY = 400;
  state = {
    isRecording: false,
    isLogging: true,
    shouldContinue: false
  };
  numberMap = {
    zero: "0",
    one: "1",
    won: "1",
    two: "2",
    to: "2",
    too: "2",
    three: "3",
    tree: "3",
    four: "4",
    for: "4",
    five: "5",
    six: "6",
    seven: "7",
    eight: "8",
    ate: "8",
    nine: "9",
    ten: "10",
    eleven: "11",
    twelve: "12",
    thirteen: "13",
    fourteen: "14",
    fifteen: "15",
    sixteen: "16",
    seventeen: "17",
    eighteen: "18",
    nineteen: "19",
    twenty: "20",
    thirty: "30",
    forty: "40",
    fifty: "50",
    sixty: "60",
    seventy: "70",
    eighty: "80",
    ninety: "90",
    hundred: "100"
  };
  constructor(audio, ui) {
    this.audio = audio;
    this.ui = ui;
    this.initializeSpeechEngine();
    this.bindEvents();
  }
  preprocessNumber(input) {
    const cleanWord = input.toLowerCase().trim().replace(/[?!.,]/g, "");
    return this.numberMap[cleanWord] || cleanWord;
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
      if (event.resultIndex !== this.lastResultIndex) {
        this.lastResultIndex = event.resultIndex;
        this.previousProcessedToken = "";
      }
      for (let i = event.resultIndex;i < event.results.length; ++i) {
        const result = event.results[i];
        if (!result || !result.length)
          continue;
        const alternative = result[0];
        if (!alternative)
          continue;
        const transcript = alternative.transcript.trim();
        if (transcript.includes(" ")) {
          continue;
        }
        const processed = this.preprocessNumber(transcript);
        const baseCommands = ["history", "shift", "north", "south", "east", "west", "option", "alt", "command", "control", "write", "click", "left", "right", "up", "down", "on", "off", "exit", "again"];
        const isCommand = baseCommands.includes(processed);
        const isNumber = /^\d+$/.test(processed);
        const isLetter = /^[a-z]$/.test(processed);
        if (isCommand || isNumber || isLetter) {
          if (processed !== this.previousProcessedToken) {
            const now = Date.now();
            if (now - this.lastExecutionTime > this.THROTTLE_DELAY) {
              this.previousProcessedToken = processed;
              this.lastExecutionTime = now;
              console.log(`[Sniper] Executing: ${processed}`);
              this.ui.showCommand(processed);
              this.handleCommands(processed);
            } else {
              console.log(`[Sniper] Throttled: ${processed} (Limit 400ms)`);
            }
          }
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
  async sendToBackend(command) {
    if (command.includes(" ") || command.trim().length === 0) {
      console.warn(`[Sniper] Blocked multi-word/empty: "${command}"`);
      return;
    }
    this.audio.play("click");
    try {
      fetch("http://localhost:8000/api/data", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ command })
      }).catch((e) => console.error("Fetch error", e));
    } catch (err) {
      console.warn("[Sniper] Backend connection failed.");
    }
  }
  handleCommands(text) {
    const command = text.toLowerCase().trim();
    if (command === "again") {
      if (this.lastActionCommand) {
        this.handleCommands(this.lastActionCommand);
        return;
      }
    }
    if (!/^\d+$/.test(command) && command !== "again") {
      this.lastActionCommand = command;
    }
    switch (command) {
      case "left":
      case "write":
      case "right":
      case "up":
      case "down":
      case "click":
      case "control":
      case "option":
      case "alt":
      case "command":
      case "north":
      case "south":
      case "east":
      case "west":
      case "shift":
        this.sendToBackend(command);
        break;
      case "exit":
        this.audio.play("sniper-exit");
        this.ui.clearText();
        this.state.shouldContinue = false;
        this.stop();
        break;
      case "off":
        this.audio.play("sniper-off");
        this.ui.clearText();
        this.state.isLogging = false;
        this.ui.updateGreenDot(this.state.isRecording, this.state.isLogging);
        break;
      case "on":
        this.audio.play("sniper-on");
        this.state.isLogging = true;
        this.ui.updateGreenDot(this.state.isRecording, this.state.isLogging);
        break;
      default:
        if (/^\d+$/.test(command) || /^[a-z]$/.test(command)) {
          this.sendToBackend(command);
        }
        break;
    }
  }
  bindEvents() {
    const btn = this.ui.getRecordButton();
    if (btn) {
      btn.addEventListener("click", () => {
        if (this.state.isRecording)
          this.stop();
        else
          this.start();
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
document.addEventListener("DOMContentLoaded", () => {
  const audioManager = new AudioManager;
  const uiManager = new UIManager;
  const app = new SniperCore(audioManager, uiManager);
});
