/// <reference lib="dom" />

/**
 * SECTION 1: TYPE DEFINITIONS
 */
interface SpeechRecognitionEvent extends Event {
  resultIndex: number;
  results: SpeechRecognitionResultList;
}

interface SpeechRecognitionResultList {
  length: number;
  item(index: number): SpeechRecognitionResult;
  [index: number]: SpeechRecognitionResult;
}

interface SpeechRecognitionResult {
  isFinal: boolean;
  length: number;
  item(index: number): SpeechRecognitionAlternative;
  [index: number]: SpeechRecognitionAlternative;
}

interface SpeechRecognitionAlternative {
  transcript: string;
  confidence: number;
}

interface SpeechRecognitionErrorEvent extends Event {
  error: string;
  message: string;
}

interface SpeechRecognition extends EventTarget {
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

interface SpeechRecognitionConstructor {
  new (): SpeechRecognition;
}

interface IWindow extends Window {
  SpeechRecognition?: SpeechRecognitionConstructor;
  webkitSpeechRecognition?: SpeechRecognitionConstructor;
}

/**
 * SECTION 2: AUDIO MANAGER CLASS
 */
type SoundType = 'sniper-on' | 'sniper-off' | 'sniper-exit' | 'click' | 'sniper-clear' | 'sniper-copy' | 'sniper-search' | 'sniper-visit';

class AudioManager {
  private sounds: Record<SoundType, HTMLAudioElement>;

  constructor() {
    this.sounds = {
      'sniper-on': new Audio('/static/on.wav'),
      'sniper-off': new Audio('/static/off.wav'),
      'sniper-exit': new Audio('/static/exit.wav'),
      'sniper-clear': new Audio('/static/clear.wav'),
      'sniper-copy': new Audio('/static/copy.wav'),
      'sniper-search': new Audio('/static/search.wav'), 
      'sniper-visit': new Audio('/static/visit.wav'), 
      'click': new Audio('/static/click.wav'),
    };
    this.preloadSounds();
  }

  private preloadSounds() {
    Object.values(this.sounds).forEach((sound) => {
      sound.preload = 'auto';
      sound.load();
    });
  }

  public play(type: SoundType) {
    const audio = this.sounds[type];
    audio.currentTime = 0;
    audio.play().catch((e) => {
      console.warn(`Audio playback failed for ${type}`, e);
    });
  }
}

/**
 * SECTION 3: UI MANAGER CLASS
 */
class UIManager {
  private btn: HTMLButtonElement;
  private transcriptEl: HTMLParagraphElement;
  private interimEl: HTMLParagraphElement;
  private outputContainer: HTMLDivElement;
  private placeholder: HTMLDivElement;
  private statusText: HTMLDivElement;
  private copyBtn: HTMLButtonElement;
  private greenDot: HTMLDivElement;

  private readonly defaultClasses = ['bg-red-600', 'hover:scale-105', 'hover:bg-red-500'];
  private readonly recordingClasses = ['bg-red-700', 'animate-pulse', 'ring-4', 'ring-red-900'];

  constructor() {
    this.btn = document.getElementById('record-button') as HTMLButtonElement;
    this.transcriptEl = document.getElementById('transcript') as HTMLParagraphElement;
    this.interimEl = document.getElementById('interim') as HTMLParagraphElement;
    this.outputContainer = document.getElementById('output-container') as HTMLDivElement;
    this.placeholder = document.getElementById('placeholder') as HTMLDivElement;
    this.statusText = document.getElementById('status-text') as HTMLDivElement;
    this.copyBtn = this.outputContainer.querySelector('button') as HTMLButtonElement;
    this.greenDot = document.getElementById('green-dot') as HTMLDivElement;

    this.setupCopyButton();
  }

  public getRecordButton(): HTMLButtonElement {
    return this.btn;
  }

  public updateGreenDot(isRecording: boolean, isLogging: boolean) {
    if (isRecording && isLogging) {
      this.greenDot.classList.remove('opacity-0');
    } else {
      this.greenDot.classList.add('opacity-0');
    }
  }

  public setRecordingState(isRecording: boolean) {
    if (isRecording) {
      this.btn.classList.remove(...this.defaultClasses);
      this.btn.classList.add(...this.recordingClasses);
      this.statusText.classList.remove('opacity-0');
      this.outputContainer.classList.remove('opacity-0', 'translate-y-10');
      this.placeholder.textContent = "Listening...";
    } else {
      this.btn.classList.remove(...this.recordingClasses);
      this.btn.classList.add(...this.defaultClasses);
      this.statusText.classList.add('opacity-0');
      this.placeholder.textContent = "Tap button to speak...";
      this.togglePlaceholder();
    }
  }

  public updateText(final: string, interim: string, isLogging: boolean) {
    if (isLogging) {
        if (final) {
            this.transcriptEl.innerText = final; 
        }
        this.interimEl.innerText = interim;
    } else {
      this.interimEl.innerText = '';
    }
    this.togglePlaceholder();
  }

  public clearText() {
    this.transcriptEl.innerText = '';
    this.interimEl.innerText = '';
    this.togglePlaceholder();
  }

  public getText(): string {
    return this.transcriptEl.innerText;
  }

  private togglePlaceholder() {
    if (this.transcriptEl.innerText || this.interimEl.innerText) {
      this.placeholder.classList.add('hidden');
    } else {
      this.placeholder.classList.remove('hidden');
    }
  }

  private setupCopyButton() {
    if (!this.copyBtn) return;
    this.copyBtn.onclick = null;
    this.copyBtn.addEventListener('click', () => {
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

/**
 * SECTION 4: SNIPER CORE CLASS
 */
class SniperCore {
  private audio: AudioManager;
  private ui: UIManager;
  private recognition: SpeechRecognition | null = null;
  
  // Track the very last command executed for the 'repeat' functionality
  private lastActionCommand: string = '';
  
  // Track the immediate last processed token to prevent duplicates (Debouncing)
  private previousProcessedToken: string = '';

  private state = {
    isRecording: false,
    isLogging: true,
    shouldContinue: false
  };

  // Maps spelled-out numbers to their digit string equivalents
  private numberMap: Record<string, string> = {
    "zero": "0", "one": "1", "won": "1", 
    "two": "2", "to": "2", "too": "2",
    "three": "3", "tree": "3",
    "four": "4", "for": "4", 
    "five": "5", 
    "six": "6", 
    "seven": "7", 
    "eight": "8", "ate": "8",
    "nine": "9", 
    "ten": "10",
    "eleven": "11", "twelve": "12", "thirteen": "13", 
    "fourteen": "14", "fifteen": "15", "sixteen": "16", 
    "seventeen": "17", "eighteen": "18", "nineteen": "19",
    "twenty": "20", "thirty": "30", "forty": "40", 
    "fifty": "50", "sixty": "60", "seventy": "70", 
    "eighty": "80", "ninety": "90", "hundred": "100"
  };

  constructor(audio: AudioManager, ui: UIManager) {
    this.audio = audio;
    this.ui = ui;
    this.initializeSpeechEngine();
    this.bindEvents();
  }

  /**
   * Cleans input word and converts spelled numbers to digits.
   * e.g., "Two" -> "2", "For" -> "4", "Left." -> "left"
   */
  private preprocessNumber(input: string): string {
    const cleanWord = input.toLowerCase().trim().replace(/[?!.,]/g, '');
    return this.numberMap[cleanWord] || cleanWord;
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
      let finalChunk = '';
      let interimChunk = '';
      let commandHandled = false;

      for (let i = event.resultIndex; i < event.results.length; ++i) {
        const result = event.results[i];
        if (!result || !result.length) continue;
        const alternative = result[0];
        if (!alternative) continue;
        const transcript = alternative.transcript;
        
        if (result.isFinal) {
          // Full sentence handling can go here later
          finalChunk += transcript;
        } else {
          const words = transcript.trim().split(/\s+/);
          const lastWord = words[words.length - 1];
          
          if (lastWord) {
            // 1. PREPROCESS: Convert to number or clean string
            const processed = this.preprocessNumber(lastWord);

            // 2. DEFINE VALID VOCABULARY
            const baseCommands = ['write', 'click', 'left', 'right', 'up', 'down', 'on', 'off', 'exit', 'again'];
            
            // Check if it is a known command OR a number (regex checks for digits)
            const isCommand = baseCommands.includes(processed);
            const isNumber = /^\d+$/.test(processed);

            if (isCommand || isNumber) {
                // 3. DEDUPLICATION CHECK
                // If the current processed word matches the immediate previous one, ignore it.
                // Exception: 'again' is allowed to be repeated if desired, 
                // but usually we prevent the *same word capture*. 
                // We'll enforce strict deduplication on input tokens here.
                if (processed !== this.previousProcessedToken) {
                    
                    // Update the tracker so we don't fire this again immediately
                    this.previousProcessedToken = processed;

                    // Execute logic
                    const outcome = this.handleCommands(processed);
                    if (outcome.capturedByCommand) {
                        commandHandled = true;
                    }
                }
            }
          }
          interimChunk += transcript;
        }
      }

      if (finalChunk) {
        this.ui.updateText(finalChunk, interimChunk, this.state.isLogging);
      } else if (commandHandled) {
        // If a command was handled, clear interim so it doesn't stick around
        this.ui.updateText('', '', this.state.isLogging);
      } else {
        this.ui.updateText('', interimChunk, this.state.isLogging);
      }
    };

    this.recognition.onerror = (event: SpeechRecognitionErrorEvent) => {
      if (event.error === 'not-allowed' || event.error === 'service-not-allowed') {
        this.state.shouldContinue = false;
        this.stop();
      }
    };
  }

  private async sendToBackend(command: string) {
    try {
      console.log(`[Sniper] Sending to backend: ${command}`);
      await fetch('http://localhost:8000/api/data', { 
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ command: command })
      });
    } catch (err) {
      console.warn('[Sniper] Backend connection failed. Is localhost:8000 running?');
    }
  }

  private handleCommands(text: string): { capturedByCommand: boolean } {
    const command = text.toLowerCase(); 
      
    // --- SPECIAL REPEAT LOGIC ---
    if (command === 'again') {
        // "again" repeats the LAST ACTION (not the last number)
        if (this.lastActionCommand) {
            console.log(`Repeating command: ${this.lastActionCommand}`);
            // We recursively call handleCommands with the stored action
            return this.handleCommands(this.lastActionCommand);
        } else {
            return { capturedByCommand: true };
        }
    }

    // Only update "lastActionCommand" if it is NOT a number
    // This ensures "Left -> 2 -> Again" repeats "Left", not "2"
    if (!/^\d+$/.test(command)) {
        this.lastActionCommand = command;
    }
      
    // --- STATIC COMMANDS ---
    switch (command) { 
      case 'left':
      case 'write':
      case 'right':
      case 'up':
      case 'down':
      case 'click':
        this.sendToBackend(command);
        return { capturedByCommand: true };

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
        // If it's a number (digits), send to backend
        if (/^\d+$/.test(command)) {
             this.sendToBackend(command);
             return { capturedByCommand: true };
        }
        return { capturedByCommand: false };
    }
  }

  private bindEvents() {
    const btn = this.ui.getRecordButton();
    if (btn) {
      btn.addEventListener('click', () => {
        if (this.state.isRecording) this.stop();
        else this.start();
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

/**
 * SECTION 5: INITIALIZATION
 */
document.addEventListener('DOMContentLoaded', () => {
  const audioManager = new AudioManager();
  const uiManager = new UIManager();
  const app = new SniperCore(audioManager, uiManager);
});