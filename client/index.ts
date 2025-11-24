/// <reference lib="dom" />

/**
 * SECTION 1: TYPE DEFINITIONS
 * (No changes here)
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
      'sniper-on': new Audio('/static/sniper-on.wav'),
      'sniper-off': new Audio('/static/sniper-off.wav'),
      'sniper-exit': new Audio('/static/sniper-exit.wav'),
      'sniper-clear': new Audio('/static/sniper-clear.wav'),
      'sniper-copy': new Audio('/static/sniper-copy.wav'),
      'sniper-search': new Audio('/static/sniper-search.wav'), 
      'sniper-visit': new Audio('/static/sniper-visit.wav'), 
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
 * (No changes here)
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
    if (isLogging && final) {
      this.transcriptEl.innerText += final + ' ';
    }
     
    if (isLogging) {
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

  // Array to track all windows opened by this session (search, visit, and open)
  private openedWindows: Window[] = [];

  // --- CUSTOM SHORTCUTS DICTIONARY ---
  private webShortcuts: Record<string, string> = {
    'ai': 'https://gemini.google.com',
    'chat gpt': 'https://chatgpt.com',
    'youtube': 'https://youtube.com',
    'github': 'https://github.com',
    'local': 'http://localhost:3000',
    'mail': 'https://gmail.com'
  };

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

      if (finalChunk) {
        const processed = this.handleCommands(finalChunk);
        if (!processed.capturedByCommand) {
           this.ui.updateText(finalChunk, interimChunk, this.state.isLogging);
           if(this.state.isLogging) this.audio.play('click');
        }
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
    const command = text.toLowerCase().trim().replace(/[?!]/g, ''); 
     
    // --- DYNAMIC COMMANDS ---
     
    // COMMAND: "open [name]" -> Uses the dictionary
    if (command.startsWith('open')) {
      const target = command.replace('open', '').trim();
      
      // Check if the target exists in our webShortcuts dictionary
      if (this.webShortcuts[target]) {
        this.audio.play('sniper-visit');
        const url = this.webShortcuts[target];
        
        const newWin = window.open(url, '_blank');
        
        // CRITICAL: Track this window so 'simplify' can close it later
        if (newWin) this.openedWindows.push(newWin);
        
        this.ui.clearText();
        return { capturedByCommand: true };
      }
    }

    // COMMAND: "visit [url]"
    if (command.startsWith('visit')) {
      this.audio.play('sniper-visit');
      const rawUrl = command.replace('visit', '').trim();
      this.openDirectUrl(rawUrl);
      this.ui.clearText();
      return { capturedByCommand: true };
    }

    // COMMAND: "search [query]"
    if (command.startsWith('search')) {
      this.audio.play('sniper-search');
      const query = command.replace('search', '').trim();
      this.openSearch(query);
      this.ui.clearText();
      return { capturedByCommand: true };
    }

    // --- STATIC COMMANDS ---
    switch (command.replace(/[.,]/g, '')) { 
      case 'exit':
        this.audio.play('sniper-exit');
        this.ui.clearText();
        this.state.shouldContinue = false;
        this.stop();
        return { capturedByCommand: true };
      
      // --- NEW COMMAND ADDED HERE ---
      case 'help mouse':
        this.audio.play('sniper-visit');
        const mouseWin = window.open('/mouse', '_blank');
        if (mouseWin) this.openedWindows.push(mouseWin);
        this.ui.clearText();
        return { capturedByCommand: true };

      // --- NEW: Help Characters ---
      case 'help signs':
        this.audio.play('sniper-visit');
        const charWin = window.open('/signs', '_blank');
      if (charWin) this.openedWindows.push(charWin);
      this.ui.clearText();
      return { capturedByCommand: true };

      case 'reveal':
        this.sendToBackend('reveal');
        this.ui.clearText();
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

      case 'clear':
        this.audio.play('sniper-clear');
        this.ui.clearText();
        return { capturedByCommand: true };

      case 'keep':
        this.audio.play('sniper-copy');
        const currentText = this.ui.getText();
        if (currentText) {
          navigator.clipboard.writeText(currentText);
        }
        this.ui.clearText();
        return { capturedByCommand: true };
      
      // Closes all windows opened by 'search', 'visit', OR 'open'
      case 'simplify':
        this.audio.play('sniper-clear'); 
        this.closeOpenedWindows();
        this.ui.clearText();
        return { capturedByCommand: true };

    default:
      // CHECK: Only send to backend if the system is currently "On" (logging)
      if (this.state.isLogging) {
        this.sendToBackend(command);
      }
      return { capturedByCommand: false };
    }
  }

  private closeOpenedWindows() {
    let closedCount = 0;
    this.openedWindows.forEach(win => {
      // Check if window object exists and isn't already closed manually
      if (win && !win.closed) {
        win.close();
        closedCount++;
      }
    });
    // Reset the array after closing
    this.openedWindows = [];
    console.log(`Sniper Simplified: Closed ${closedCount} tabs.`);
  }

  private openSearch(query: string) {
    if (!query) return;
    const normalized = query.replace(/ dot /g, '.').replace(/ period /g, '.');
    const searchUrl = `https://www.google.com/search?q=${encodeURIComponent(normalized)}`;
    
    const newWin = window.open(searchUrl, '_blank');
    if (newWin) this.openedWindows.push(newWin);
  }

  private openDirectUrl(transcript: string) {
    if (!transcript) return;

    let url = transcript.toLowerCase()
      .replace(/ dot /g, '.')
      .replace(/ period /g, '.')
      .replace(/ slash /g, '/')
      .replace(/\s+/g, ''); 

    if (!url.startsWith('http')) {
      url = 'https://' + url;
    }
    
    const newWin = window.open(url, '_blank');
    if (newWin) this.openedWindows.push(newWin);
  }

  private bindEvents() {
    const btn = this.ui.getRecordButton();/// <reference lib="dom" />

/**
 * SECTION 1: TYPE DEFINITIONS
 * (No changes here)
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
      'sniper-on': new Audio('/static/sniper-on.wav'),
      'sniper-off': new Audio('/static/sniper-off.wav'),
      'sniper-exit': new Audio('/static/sniper-exit.wav'),
      'sniper-clear': new Audio('/static/sniper-clear.wav'),
      'sniper-copy': new Audio('/static/sniper-copy.wav'),
      'sniper-search': new Audio('/static/sniper-search.wav'), 
      'sniper-visit': new Audio('/static/sniper-visit.wav'), 
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
 * (No changes here)
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
    if (isLogging && final) {
      this.transcriptEl.innerText += final + ' ';
    }
      
    if (isLogging) {
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

  // Array to track all windows opened by this session (search, visit, and open)
  private openedWindows: Window[] = [];

  // --- CUSTOM SHORTCUTS DICTIONARY ---
  private webShortcuts: Record<string, string> = {
    'ai': 'https://gemini.google.com',
    'chat gpt': 'https://chatgpt.com',
    'youtube': 'https://youtube.com',
    'github': 'https://github.com',
    'local': 'http://localhost:3000',
    'mail': 'https://gmail.com'
  };

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

      if (finalChunk) {
        const processed = this.handleCommands(finalChunk);
        if (!processed.capturedByCommand) {
           this.ui.updateText(finalChunk, interimChunk, this.state.isLogging);
           if(this.state.isLogging) this.audio.play('click');
        }
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
    const command = text.toLowerCase().trim().replace(/[?!]/g, ''); 
      
    // --- DYNAMIC COMMANDS ---
      
    // COMMAND: "open [name]" -> Uses the dictionary
    if (command.startsWith('open')) {
      const target = command.replace('open', '').trim();
       
      // Check if the target exists in our webShortcuts dictionary
      if (this.webShortcuts[target]) {
        this.audio.play('sniper-visit');
        const url = this.webShortcuts[target];
        
        const newWin = window.open(url, '_blank');
        
        // CRITICAL: Track this window so 'simplify' can close it later
        if (newWin) this.openedWindows.push(newWin);
        
        this.ui.clearText();
        return { capturedByCommand: true };
      }
    }

    // COMMAND: "visit [url]"
    if (command.startsWith('visit')) {
      this.audio.play('sniper-visit');
      const rawUrl = command.replace('visit', '').trim();
      this.openDirectUrl(rawUrl);
      this.ui.clearText();
      return { capturedByCommand: true };
    }

    // COMMAND: "search [query]"
    if (command.startsWith('search')) {
      this.audio.play('sniper-search');
      const query = command.replace('search', '').trim();
      this.openSearch(query);
      this.ui.clearText();
      return { capturedByCommand: true };
    }

    // --- STATIC COMMANDS ---
    switch (command.replace(/[.,]/g, '')) { 
      case 'exit':
        this.audio.play('sniper-exit');
        this.ui.clearText();
        this.state.shouldContinue = false;
        this.stop();
        return { capturedByCommand: true };
       
      // --- NEW COMMAND ADDED HERE ---
      case 'help mouse':
        this.audio.play('sniper-visit');
        const mouseWin = window.open('/mouse', '_blank');
        if (mouseWin) this.openedWindows.push(mouseWin);
        this.ui.clearText();
        return { capturedByCommand: true };

      // --- NEW: Help Characters ---
      case 'help characters':
        this.audio.play('sniper-visit');
        const charWin = window.open('/characters', '_blank');
      if (charWin) this.openedWindows.push(charWin);
      this.ui.clearText();
      return { capturedByCommand: true };

      case 'reveal':
        this.sendToBackend('reveal');
        this.ui.clearText();
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

      case 'clear':
        this.audio.play('sniper-clear');
        this.ui.clearText();
        return { capturedByCommand: true };

      case 'keep':
        this.audio.play('sniper-copy');
        const currentText = this.ui.getText();
        if (currentText) {
          navigator.clipboard.writeText(currentText);
        }
        this.ui.clearText();
        return { capturedByCommand: true };
       
      // Closes all windows opened by 'search', 'visit', OR 'open'
      case 'simplify':
        this.audio.play('sniper-clear'); 
        this.closeOpenedWindows();
        this.ui.clearText();
        return { capturedByCommand: true };

      default:
        // FIX: Only send data to backend if the system is actually logging (On)
        if (this.state.isLogging) {
            this.sendToBackend(command);
        }
        return { capturedByCommand: false };
    }
  }

  private closeOpenedWindows() {
    let closedCount = 0;
    this.openedWindows.forEach(win => {
      // Check if window object exists and isn't already closed manually
      if (win && !win.closed) {
        win.close();
        closedCount++;
      }
    });
    // Reset the array after closing
    this.openedWindows = [];
    console.log(`Sniper Simplified: Closed ${closedCount} tabs.`);
  }

  private openSearch(query: string) {
    if (!query) return;
    const normalized = query.replace(/ dot /g, '.').replace(/ period /g, '.');
    const searchUrl = `https://www.google.com/search?q=${encodeURIComponent(normalized)}`;
    
    const newWin = window.open(searchUrl, '_blank');
    if (newWin) this.openedWindows.push(newWin);
  }

  private openDirectUrl(transcript: string) {
    if (!transcript) return;

    let url = transcript.toLowerCase()
      .replace(/ dot /g, '.')
      .replace(/ period /g, '.')
      .replace(/ slash /g, '/')
      .replace(/\s+/g, ''); 

    if (!url.startsWith('http')) {
      url = 'https://' + url;
    }
    
    const newWin = window.open(url, '_blank');
    if (newWin) this.openedWindows.push(newWin);
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