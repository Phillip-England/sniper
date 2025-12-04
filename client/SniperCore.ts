import { AudioManager } from "./AudioManager";
import { UIManager } from "./UIManager";

/**
 * SECTION 1: TYPE DEFINITIONS
 */

export interface SpeechRecognitionEvent extends Event {
  resultIndex: number;
  results: SpeechRecognitionResultList;
}

export interface SpeechRecognitionResultList {
  length: number;
  item(index: number): SpeechRecognitionResult;
  [index: number]: SpeechRecognitionResult;
}

export interface SpeechRecognitionResult {
  isFinal: boolean;
  length: number;
  item(index: number): SpeechRecognitionAlternative;
  [index: number]: SpeechRecognitionAlternative;
}

export interface SpeechRecognitionAlternative {
  transcript: string;
  confidence: number;
}

export interface SpeechRecognitionErrorEvent extends Event {
  error: string;
  message: string;
}

export interface SpeechRecognition extends EventTarget {
  continuous: boolean;
  interimResults: boolean;
  lang: string;
  start(): void;
  stop(): void;
  abort(): void;
  onstart: ((this: SpeechRecognition, ev: Event) => any) | null;
  onend: ((this: SpeechRecognition, ev: Event) => any) | null;
  onresult: ((this: SpeechRecognition, ev: SpeechRecognitionEvent) => any) | null;
  onerror: ((this: SpeechRecognition, ev: SpeechRecognitionErrorEvent) => any) | null;
}

export interface SpeechRecognitionConstructor {
  new (): SpeechRecognition;
}

export interface IWindow extends Window {
  SpeechRecognition?: SpeechRecognitionConstructor;
  webkitSpeechRecognition?: SpeechRecognitionConstructor;
}


// Structure for the data coming from /api/commands/min
interface CommandRegistryItem {
  name: string;
  called_by: string[];
}



/**
 * SECTION 4: SNIPER CORE CLASS
 */
export class SniperCore {
  private audio: AudioManager;
  private ui: UIManager;
  private recognition: SpeechRecognition | null = null;
  
  // Track the very last command executed for the 'repeat' functionality
  private lastCommand: string = '';
  // Stores all valid command triggers fetched from the backend
  private commandTriggers: string[] = [];

  // --- NEW PROPERTIES FOR FREEZE FIX ---
  private silenceTimer: any = null; 
  private currentInterim: string = '';

  private state = {
    isRecording: false,
    isLogging: true,
    shouldContinue: false
  }; 

  constructor(audio: AudioManager, ui: UIManager) {
    this.audio = audio;
    this.ui = ui;
    this.initializeSpeechEngine();
    this.bindEvents();
    
    // Initialize the command array from the backend
    this.loadCommandTriggers();
  }

  // Fetches the registry and populates this.commandTriggers
  private async loadCommandTriggers() {
    try {
      const response = await fetch('/api/commands/min');
      if (!response.ok) {
        throw new Error(`Failed to fetch commands: ${response.statusText}`);
      }
      
      const commands = await response.json() as CommandRegistryItem[];
      
      // Flatten all 'called_by' arrays into a single array of strings
      this.commandTriggers = commands.reduce((acc, cmd) => {
        return acc.concat(cmd.called_by);
      }, [] as string[]);

      console.log(`[Sniper] Loaded ${this.commandTriggers.length} command triggers.`);
      // console.log(this.commandTriggers); // Uncomment to see the full list
      
    } catch (err) {
      console.warn("[Sniper] Could not load command registry. Command validation may be limited.", err);
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

    this.recognition.onresult = (event: SpeechRecognitionEvent) => {
      // 1. Clear existing timer immediately on any activity
      if (this.silenceTimer) {
        clearTimeout(this.silenceTimer);
        this.silenceTimer = null;
      }

      let finalChunk = '';
      let interimChunk = '';

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

      // Save interim for the timer logic
      this.currentInterim = interimChunk;

      if (finalChunk) {
        // CASE A: Natural Final Result received
        this.executeFinalSequence(finalChunk, interimChunk);
      } else {
        // CASE B: Only interim results. Update UI and start debounce timer.
        this.ui.updateText('', interimChunk, this.state.isLogging);

        // If we have text pending, start the 1-second timer
        if (interimChunk.trim().length > 0) {
            this.silenceTimer = setTimeout(() => {
                console.log("[Sniper] Force finalizing stuck interim result...");
                // Force execute with the stuck interim text
                this.executeFinalSequence(this.currentInterim, ''); 
                this.currentInterim = ''; 
            }, 1000); 
        }
      }
    };

    this.recognition.onerror = (event: SpeechRecognitionErrorEvent) => {
      if (event.error === 'not-allowed' || event.error === 'service-not-allowed') {
        this.state.shouldContinue = false;
        this.stop();
      }
    };
  }

  /**
   * Helper to handle the logic when a command is considered "Done"
   * (Either by natural isFinal or by the debounce timer)
   */
  private executeFinalSequence(finalText: string, interimText: string) {
    // 1. Immediately update UI with the new command (replaces old command)
    this.ui.updateText(finalText, interimText, this.state.isLogging);
    
    // 2. Process the command
    const processed = this.handleCommands(finalText);
    
    // 3. Audio feedback if not a specific command
    if (!processed.capturedByCommand && this.state.isLogging) {
       this.audio.play('click');
    }
    
  }

  private async sendToBackend(command: string) {
    try {
      console.log(`[Sniper] Sending to backend: ${command}`);
      await fetch('http://localhost:9090/api/data', { 
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ command: command })
      });
    } catch (err) {
      console.warn('[Sniper] Backend connection failed. Is localhost:9090 running?');
    }
  }

  private handleCommands(text: string): { capturedByCommand: boolean } {
    const command = text.toLowerCase().trim().replace(/[?!]/g, ''); 
      
    if (command) {
        this.lastCommand = command;
    }
      
    // --- STATIC COMMANDS ---
    switch (command.replace(/[.,]/g, '')) { 
      // ... existing cases ...
      case 'exit':
        this.audio.play('sniper-exit');
        this.ui.clearText(); 
        this.state.shouldContinue = false;
        this.stop();
        return { capturedByCommand: true };
       
      case 'off':
        this.audio.play('sniper-off');
        this.ui.clearText(); 
        this.state.isLogging = false;
        this.ui.updateGreenDot(this.state.isRecording, this.state.isLogging);
        return { capturedByCommand: true };

      case 'on':
        this.audio.play('sniper-on');
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
