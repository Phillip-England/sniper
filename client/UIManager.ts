/**
 * SECTION 3: UI MANAGER CLASS
 */
export class UIManager {
  private btn: HTMLButtonElement;
  private transcriptEl: HTMLParagraphElement;
  private interimEl: HTMLParagraphElement;
  private outputContainer: HTMLDivElement;
  private placeholder: HTMLDivElement;
  private statusText: HTMLDivElement;
  private copyBtn: HTMLButtonElement;
  private greenDot: HTMLDivElement;

  private readonly defaultClasses = [
    "bg-red-600",
    "hover:scale-105",
    "hover:bg-red-500",
  ];
  private readonly recordingClasses = [
    "bg-red-700",
    "animate-pulse",
    "ring-4",
    "ring-red-900",
  ];

  constructor() {
    this.btn = document.getElementById("record-button") as HTMLButtonElement;
    this.transcriptEl = document.getElementById(
      "transcript",
    ) as HTMLParagraphElement;
    this.interimEl = document.getElementById("interim") as HTMLParagraphElement;
    this.outputContainer = document.getElementById(
      "output-container",
    ) as HTMLDivElement;
    this.placeholder = document.getElementById("placeholder") as HTMLDivElement;
    this.statusText = document.getElementById("status-text") as HTMLDivElement;
    this.copyBtn = this.outputContainer.querySelector(
      "button",
    ) as HTMLButtonElement;
    this.greenDot = document.getElementById("green-dot") as HTMLDivElement;

    this.setupCopyButton();
  }

  public getRecordButton(): HTMLButtonElement {
    return this.btn;
  }

  public updateGreenDot(isRecording: boolean, isLogging: boolean) {
    if (isRecording && isLogging) {
      this.greenDot.classList.remove("opacity-0");
    } else {
      this.greenDot.classList.add("opacity-0");
    }
  }

  public setRecordingState(isRecording: boolean) {
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

  public updateText(final: string, interim: string, isLogging: boolean) {
    if (isLogging) {
      // CHANGED: We now overwrite the text instead of appending (+Refactor)
      // This ensures only the most recent command is shown.
      if (final) {
        this.transcriptEl.innerText = final;
      }
      this.interimEl.innerText = interim;
    } else {
      this.interimEl.innerText = "";
    }
    this.togglePlaceholder();
  }

  public clearText() {
    this.transcriptEl.innerText = "";
    this.interimEl.innerText = "";
    this.togglePlaceholder();
  }

  public getText(): string {
    return this.transcriptEl.innerText;
  }

  private togglePlaceholder() {
    if (this.transcriptEl.innerText || this.interimEl.innerText) {
      this.placeholder.classList.add("hidden");
    } else {
      this.placeholder.classList.remove("hidden");
    }
  }

  private setupCopyButton() {
    if (!this.copyBtn) return;
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
