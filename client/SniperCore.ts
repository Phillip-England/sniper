import { AudioManager } from "./AudioManager";
import { CommandCenter } from "./CommandCenter";
import { Throttler } from "./Thottler";
import { UIManager } from "./UIManager";
import { SniperService } from "./SniperService";
import type {
  IWindow,
  SpeechRecognition,
  SpeechRecognitionErrorEvent,
  SpeechRecognitionEvent,
} from "./SpeechTypes";
import { PhraseMode } from "./PhraseMode";
import { RapidMode } from "./RapidMode";

// The Interface for any future mode (Rapid, Phrase, etc.)
export interface IRecognitionMode {
  handleResult(event: SpeechRecognitionEvent): void;
}

/**
 * SECTION 3: SNIPER CORE CLASS (Now an exported object factory)
 */
export class SniperCore {
  // Made public so the Modes can access them
  audio: AudioManager;
  ui: UIManager;
  api: SniperService;
  commandCenter: CommandCenter;

  state = {
    isRecording: false,
    isLogging: true,
    shouldContinue: false,
  };

  recognition: SpeechRecognition | null = null;

  // The Strategy
  mode: IRecognitionMode;

  /**
   * Constructor (No modifier, making it accessible within the module scope for instantiation)
   * Only handles basic property instantiation.
   */
  constructor() {
    this.audio = new AudioManager();
    this.ui = new UIManager();
    this.commandCenter = new CommandCenter();
    this.api = new SniperService();

    // Initialize default mode
    this.mode = new PhraseMode(this);
  }

  /**
   * Static Factory Method. (No modifier, making it implicitly public)
   * Creates the instance, initializes the engine, binds events,
   * and returns the fully prepared object.
   */
  static async new(): Promise<SniperCore> { // Removed 'async' and 'Promise<...>' as it was synchronous in the previous step
    const core = new SniperCore();
    core.initializeSpeechEngine();
    core.bindEvents();
    return core;
  }

  // Method to swap modes dynamically in the future (No modifier, making it implicitly public)
  setMode(modeType: "phrase" | "rapid") {
    if (modeType === "rapid") {
      console.log("[Sniper] Switching to Rapid Mode");
      this.mode = new RapidMode(this);
      this.audio.play("sniper-on"); // sound cue for mode switch
    } else {
      console.log("[Sniper] Switching to Phrase Mode");
      this.mode = new PhraseMode(this);
      this.audio.play("click"); // sound cue for mode switch
    }
  }

  initializeSpeechEngine() {
    const SpeechRecognitionCtor =
      (window as unknown as IWindow).SpeechRecognition ||
      (window as unknown as IWindow).webkitSpeechRecognition;
    if (!SpeechRecognitionCtor) {
      alert("Browser not supported. Try Chrome/Safari.");
      return;
    }
    this.recognition = new SpeechRecognitionCtor();
    this.recognition.continuous = true;
    this.recognition.interimResults = true;
    this.recognition.lang = "en-US";
    this.setupRecognitionHandlers();
  }

  setupRecognitionHandlers() {
    if (!this.recognition) return;

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

    // DELEGATION HAPPENS HERE
    this.recognition.onresult = (event: SpeechRecognitionEvent) => {
      this.mode.handleResult(event);
    };

    this.recognition.onerror = (event: SpeechRecognitionErrorEvent) => {
      if (
        event.error === "not-allowed" ||
        event.error === "service-not-allowed"
      ) {
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