// client/index.ts
class AudioManager {
  sounds;
  constructor() {
    this.sounds = {
      "sniper-on": new Audio("/static/sniper-on.wav"),
      "sniper-off": new Audio("/static/sniper-off.wav"),
      "sniper-exit": new Audio("/static/sniper-exit.wav"),
      "sniper-clear": new Audio("/static/sniper-clear.wav"),
      "sniper-copy": new Audio("/static/sniper-copy.wav"),
      "sniper-search": new Audio("/static/sniper-search.wav"),
      "sniper-visit": new Audio("/static/sniper-visit.wav"),
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
    if (isLogging && final) {
      this.transcriptEl.innerText += final + " ";
    }
    if (isLogging) {
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

class SniperCore {
  audio;
  ui;
  recognition = null;
  openedWindows = [];
  webShortcuts = {
    ai: "https://gemini.google.com",
    "chat gpt": "https://chatgpt.com",
    youtube: "https://youtube.com",
    github: "https://github.com",
    local: "http://localhost:3000",
    mail: "https://gmail.com"
  };
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
      if (finalChunk) {
        const processed = this.handleCommands(finalChunk);
        if (!processed.capturedByCommand) {
          this.ui.updateText(finalChunk, interimChunk, this.state.isLogging);
          if (this.state.isLogging)
            this.audio.play("click");
        }
      } else {
        this.ui.updateText("", interimChunk, this.state.isLogging);
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
    if (command.startsWith("open")) {
      const target = command.replace("open", "").trim();
      if (this.webShortcuts[target]) {
        this.audio.play("sniper-visit");
        const url = this.webShortcuts[target];
        const newWin = window.open(url, "_blank");
        if (newWin)
          this.openedWindows.push(newWin);
        this.ui.clearText();
        return { capturedByCommand: true };
      }
    }
    if (command.startsWith("visit")) {
      this.audio.play("sniper-visit");
      const rawUrl = command.replace("visit", "").trim();
      this.openDirectUrl(rawUrl);
      this.ui.clearText();
      return { capturedByCommand: true };
    }
    if (command.startsWith("search")) {
      this.audio.play("sniper-search");
      const query = command.replace("search", "").trim();
      this.openSearch(query);
      this.ui.clearText();
      return { capturedByCommand: true };
    }
    switch (command.replace(/[.,]/g, "")) {
      case "exit":
        this.audio.play("sniper-exit");
        this.ui.clearText();
        this.state.shouldContinue = false;
        this.stop();
        return { capturedByCommand: true };
      case "help mouse":
        this.audio.play("sniper-visit");
        const mouseWin = window.open("/mouse", "_blank");
        if (mouseWin)
          this.openedWindows.push(mouseWin);
        this.ui.clearText();
        return { capturedByCommand: true };
      case "help characters":
        this.audio.play("sniper-visit");
        const charWin = window.open("/characters", "_blank");
        if (charWin)
          this.openedWindows.push(charWin);
        this.ui.clearText();
        return { capturedByCommand: true };
      case "reveal":
        this.sendToBackend("reveal");
        this.ui.clearText();
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
      case "clear":
        this.audio.play("sniper-clear");
        this.ui.clearText();
        return { capturedByCommand: true };
      case "keep":
        this.audio.play("sniper-copy");
        const currentText = this.ui.getText();
        if (currentText) {
          navigator.clipboard.writeText(currentText);
        }
        this.ui.clearText();
        return { capturedByCommand: true };
      case "simplify":
        this.audio.play("sniper-clear");
        this.closeOpenedWindows();
        this.ui.clearText();
        return { capturedByCommand: true };
      default:
        if (this.state.isLogging) {
          this.sendToBackend(command);
        }
        return { capturedByCommand: false };
    }
  }
  closeOpenedWindows() {
    let closedCount = 0;
    this.openedWindows.forEach((win) => {
      if (win && !win.closed) {
        win.close();
        closedCount++;
      }
    });
    this.openedWindows = [];
    console.log(`Sniper Simplified: Closed ${closedCount} tabs.`);
  }
  openSearch(query) {
    if (!query)
      return;
    const normalized = query.replace(/ dot /g, ".").replace(/ period /g, ".");
    const searchUrl = `https://www.google.com/search?q=${encodeURIComponent(normalized)}`;
    const newWin = window.open(searchUrl, "_blank");
    if (newWin)
      this.openedWindows.push(newWin);
  }
  openDirectUrl(transcript) {
    if (!transcript)
      return;
    let url = transcript.toLowerCase().replace(/ dot /g, ".").replace(/ period /g, ".").replace(/ slash /g, "/").replace(/\s+/g, "");
    if (!url.startsWith("http")) {
      url = "https://" + url;
    }
    const newWin = window.open(url, "_blank");
    if (newWin)
      this.openedWindows.push(newWin);
  }
  bindEvents() {
    const btn = this.ui.getRecordButton();

    class AudioManager2 {
      sounds;
      constructor() {
        this.sounds = {
          "sniper-on": new Audio("/static/sniper-on.wav"),
          "sniper-off": new Audio("/static/sniper-off.wav"),
          "sniper-exit": new Audio("/static/sniper-exit.wav"),
          "sniper-clear": new Audio("/static/sniper-clear.wav"),
          "sniper-copy": new Audio("/static/sniper-copy.wav"),
          "sniper-search": new Audio("/static/sniper-search.wav"),
          "sniper-visit": new Audio("/static/sniper-visit.wav"),
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

    class UIManager2 {
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
        if (isLogging && final) {
          this.transcriptEl.innerText += final + " ";
        }
        if (isLogging) {
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

    class SniperCore2 {
      audio;
      ui;
      recognition = null;
      openedWindows = [];
      webShortcuts = {
        ai: "https://gemini.google.com",
        "chat gpt": "https://chatgpt.com",
        youtube: "https://youtube.com",
        github: "https://github.com",
        local: "http://localhost:3000",
        mail: "https://gmail.com"
      };
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
          if (finalChunk) {
            const processed = this.handleCommands(finalChunk);
            if (!processed.capturedByCommand) {
              this.ui.updateText(finalChunk, interimChunk, this.state.isLogging);
              if (this.state.isLogging)
                this.audio.play("click");
            }
          } else {
            this.ui.updateText("", interimChunk, this.state.isLogging);
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
        if (command.startsWith("open")) {
          const target = command.replace("open", "").trim();
          if (this.webShortcuts[target]) {
            this.audio.play("sniper-visit");
            const url = this.webShortcuts[target];
            const newWin = window.open(url, "_blank");
            if (newWin)
              this.openedWindows.push(newWin);
            this.ui.clearText();
            return { capturedByCommand: true };
          }
        }
        if (command.startsWith("visit")) {
          this.audio.play("sniper-visit");
          const rawUrl = command.replace("visit", "").trim();
          this.openDirectUrl(rawUrl);
          this.ui.clearText();
          return { capturedByCommand: true };
        }
        if (command.startsWith("search")) {
          this.audio.play("sniper-search");
          const query = command.replace("search", "").trim();
          this.openSearch(query);
          this.ui.clearText();
          return { capturedByCommand: true };
        }
        switch (command.replace(/[.,]/g, "")) {
          case "exit":
            this.audio.play("sniper-exit");
            this.ui.clearText();
            this.state.shouldContinue = false;
            this.stop();
            return { capturedByCommand: true };
          case "help mouse":
            this.audio.play("sniper-visit");
            const mouseWin = window.open("/mouse", "_blank");
            if (mouseWin)
              this.openedWindows.push(mouseWin);
            this.ui.clearText();
            return { capturedByCommand: true };
          case "help characters":
            this.audio.play("sniper-visit");
            const charWin = window.open("/characters", "_blank");
            if (charWin)
              this.openedWindows.push(charWin);
            this.ui.clearText();
            return { capturedByCommand: true };
          case "reveal":
            this.sendToBackend("reveal");
            this.ui.clearText();
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
          case "clear":
            this.audio.play("sniper-clear");
            this.ui.clearText();
            return { capturedByCommand: true };
          case "keep":
            this.audio.play("sniper-copy");
            const currentText = this.ui.getText();
            if (currentText) {
              navigator.clipboard.writeText(currentText);
            }
            this.ui.clearText();
            return { capturedByCommand: true };
          case "simplify":
            this.audio.play("sniper-clear");
            this.closeOpenedWindows();
            this.ui.clearText();
            return { capturedByCommand: true };
          default:
            if (this.state.isLogging) {
              this.sendToBackend(command);
            }
            return { capturedByCommand: false };
        }
      }
      closeOpenedWindows() {
        let closedCount = 0;
        this.openedWindows.forEach((win) => {
          if (win && !win.closed) {
            win.close();
            closedCount++;
          }
        });
        this.openedWindows = [];
        console.log(`Sniper Simplified: Closed ${closedCount} tabs.`);
      }
      openSearch(query) {
        if (!query)
          return;
        const normalized = query.replace(/ dot /g, ".").replace(/ period /g, ".");
        const searchUrl = `https://www.google.com/search?q=${encodeURIComponent(normalized)}`;
        const newWin = window.open(searchUrl, "_blank");
        if (newWin)
          this.openedWindows.push(newWin);
      }
      openDirectUrl(transcript) {
        if (!transcript)
          return;
        let url = transcript.toLowerCase().replace(/ dot /g, ".").replace(/ period /g, ".").replace(/ slash /g, "/").replace(/\s+/g, "");
        if (!url.startsWith("http")) {
          url = "https://" + url;
        }
        const newWin = window.open(url, "_blank");
        if (newWin)
          this.openedWindows.push(newWin);
      }
      bindEvents() {
        const btn2 = this.ui.getRecordButton();
        if (btn2) {
          btn2.addEventListener("click", () => {
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
      const audioManager = new AudioManager2;
      const uiManager = new UIManager2;
      const app = new SniperCore2(audioManager, uiManager);
    });
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
