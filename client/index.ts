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
  private greenDot: HTMLDivElement;
  private commandTimeout: number | null = null;

  private readonly defaultClasses = ['bg-red-600', 'hover:scale-105', 'hover:bg-red-500'];
  private readonly recordingClasses = ['bg-red-700', 'animate-pulse', 'ring-4', 'ring-red-900'];

  constructor() {
    this.btn = document.getElementById('record-button') as HTMLButtonElement;
    this.transcriptEl = document.getElementById('transcript') as HTMLParagraphElement;
    this.interimEl = document.getElementById('interim') as HTMLParagraphElement;
    this.outputContainer = document.getElementById('output-container') as HTMLDivElement;
    this.placeholder = document.getElementById('placeholder') as HTMLDivElement;
    this.statusText = document.getElementById('status-text') as HTMLDivElement;
    this.greenDot = document.getElementById('green-dot') as HTMLDivElement;

    this.transcriptEl.innerText = "";
    this.interimEl.innerText = "";
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
      this.placeholder.textContent = "Listening for commands...";
      this.placeholder.classList.remove('hidden');
    } else {
      this.btn.classList.remove(...this.recordingClasses);
      this.btn.classList.add(...this.defaultClasses);
      this.statusText.classList.add('opacity-0');
      this.placeholder.textContent = "Tap button to speak...";
      this.placeholder.classList.remove('hidden');
      this.transcriptEl.innerText = "";
    }
  }

  public showCommand(command: string) {
    this.placeholder.classList.add('hidden');
    this.transcriptEl.innerText = `[ ${command.toUpperCase()} ]`;
    
    if (this.commandTimeout) {
        window.clearTimeout(this.commandTimeout);
    }

    this.commandTimeout = window.setTimeout(() => {
        this.transcriptEl.innerText = "";
        this.placeholder.classList.remove('hidden');
    }, 2000);
  }

  public clearText() {
    this.transcriptEl.innerText = '';
    this.interimEl.innerText = '';
    this.placeholder.classList.remove('hidden');
  }
}

/**
 * SECTION 4: SNIPER CORE CLASS
 */
class SniperCore {
  private audio: AudioManager;
  private ui: UIManager;
  private recognition: SpeechRecognition | null = null;
  
  private lastActionCommand: string = '';
  
  // Deduplication & Throttling State
  private previousProcessedToken: string = '';
  private lastResultIndex: number = -1; 
  private lastExecutionTime: number = 0;
  private readonly THROTTLE_DELAY = 200; // 200ms constraint

  private state = {
    isRecording: false,
    isLogging: true,
    shouldContinue: false
  };

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
      // 1. New Phrase Detection
      if (event.resultIndex !== this.lastResultIndex) {
          this.lastResultIndex = event.resultIndex;
          this.previousProcessedToken = '';
      }

      for (let i = event.resultIndex; i < event.results.length; ++i) {
        const result = event.results[i];
        if (!result || !result.length) continue;
        const alternative = result[0];
        if (!alternative) continue;
        
        // 2. Strict Single Word Check
        const transcript = alternative.transcript.trim();
        
        if (transcript.includes(' ')) {
            continue;
        }

        // 3. Process the Single Word
        const processed = this.preprocessNumber(transcript);
        
        const baseCommands = [
            // Controls
            'north', 'south', 'east', 'west', 'option', 'alt', 'command', 'control', 
            'write', 'click', 'left', 'right', 'up', 'down', 'on', 'off', 'exit', 'again', 'shift',
            // Phonetic Alphabet
            'alpha', 'bravo', 'charlie', 'delta', 'echo', 'foxtrot', 'golf', 'hotel', 'india', 
            'juliet', 'kilo', 'lima', 'mike', 'november', 'oscar', 'papa', 'quebec', 'romeo', 
            'sierra', 'tango', 'uniform', 'victor', 'whiskey', 'xray', 'yankee', 'zulu'
        ];
        
        const isCommand = baseCommands.includes(processed);
        const isNumber = /^\d+$/.test(processed);
        // We keep isLetter in case something slips through, but phonetic words are preferred
        const isLetter = /^[a-z]$/.test(processed);

        if (isCommand || isNumber || isLetter) {
            
            // 4. Deduplication
            if (processed !== this.previousProcessedToken) {
                
                // 5. Global Throttle Check
                const now = Date.now();
                if (now - this.lastExecutionTime > this.THROTTLE_DELAY) {
                    
                    this.previousProcessedToken = processed;
                    this.lastExecutionTime = now;
                    console.log(`[Sniper] Executing: ${processed}`);
                    this.ui.showCommand(processed);
                    this.handleCommands(processed);
                } else {
                    console.log(`[Sniper] Throttled: ${processed} (Limit 200ms)`);
                }
            }
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

  private async sendToBackend(command: string) {
    if (command.includes(' ') || command.trim().length === 0) {
        console.warn(`[Sniper] Blocked multi-word/empty: "${command}"`);
        return;
    }

    this.audio.play('click');
    try {
      fetch('http://localhost:8000/api/data', { 
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: command })
      }).catch(e => console.error("Fetch error", e));
    } catch (err) {
      console.warn('[Sniper] Backend connection failed.');
    }
  }

  private handleCommands(text: string): void {
    const command = text.toLowerCase().trim();

    if (command === 'again') {
        if (this.lastActionCommand) {
            this.handleCommands(this.lastActionCommand);
            return;
        }
    }

    if (!/^\d+$/.test(command) && command !== 'again') {
        this.lastActionCommand = command;
    }
      
    // Pass phonetic words directly to backend, they will be handled there
    console.log(command)
    this.sendToBackend(command);
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