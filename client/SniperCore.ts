import { AudioManager } from "./AudioManager";
import { CommandCenter } from "./CommandCenter";
import { Throttler } from "./Thottler";
import { UIManager } from "./UIManager";
import { SniperService } from "./SniperService";
import type { IWindow, SpeechRecognition, SpeechRecognitionErrorEvent, SpeechRecognitionEvent } from "./SpeechTypes";
import { PhraseMode } from "./PhraseMode";
import { RapidMode } from "./RapidMode";



// The Interface for any future mode (Rapid, Phrase, etc.)
export interface IRecognitionMode {
  handleResult(event: SpeechRecognitionEvent): void;
}



/**
 * SECTION 3: SNIPER CORE CLASS
 */
export class SniperCore {
  // Made public so the Modes can access them
  public audio: AudioManager;
  public ui: UIManager;
  public api: SniperService;
  public commandCenter: CommandCenter;
  
  public state = {
    isRecording: false,
    isLogging: true,
    shouldContinue: false
  };

  private recognition: SpeechRecognition | null = null;
  
  // The Strategy
  private mode: IRecognitionMode;

  constructor(audio: AudioManager, ui: UIManager) {
    this.audio = audio;
    this.ui = ui;
    this.commandCenter = new CommandCenter();
    this.api = new SniperService();
    
    // Initialize default mode
    this.mode = new RapidMode(this);

    this.initializeSpeechEngine();
    this.bindEvents();
  }

  // Method to swap modes dynamically in the future
  public setMode(modeType: 'phrase' | 'rapid') {
    if (modeType === 'rapid') {
      console.log("[Sniper] Switching to Rapid Mode");
      this.mode = new RapidMode(this);
      this.audio.play('sniper-on'); // sound cue for mode switch
    } else {
      console.log("[Sniper] Switching to Phrase Mode");
      this.mode = new PhraseMode(this);
      this.audio.play('click'); // sound cue for mode switch
    }
  }

  private initializeSpeechEngine() {
    const SpeechRecognitionCtor = (window as unknown as IWindow).SpeechRecognition ||
      (window as unknown as IWindow).webkitSpeechRecognition;
    if (!SpeechRecognitionCtor) {
      alert("Browser not supported. Try Chrome/Safari.");
      return;
    }
    this.recognition = new SpeechRecognitionCtor();
    this.recognition.continuous = true;
    this.recognition.interimResults = true;
    this.recognition.lang = 'en-US';
    this.setupRecognitionHandlers();
  }

  private setupRecognitionHandlers() {
    if (!this.recognition) return;

    this.recognition.onstart = () => {
      if (!this.state.isRecording) {
        this.audio.play('sniper-on');
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
      if (event.error === 'not-allowed' || event.error === 'service-not-allowed') {
        this.state.shouldContinue = false;
        this.stop();
      }
    };
  }

  private bindEvents() {
    const btn = this.ui.getRecordButton();
    if (btn) {
      btn.addEventListener('click', () => {
        if (this.state.isRecording) {
          this.stop();
        } else {
          this.start();
        }
      });
    }
  }

  public start() {
    this.state.shouldContinue = true;
    this.state.isLogging = true;
    this.recognition?.start();
  }

  public stop() {
    this.state.shouldContinue = false;
    this.recognition?.stop();
  }


}