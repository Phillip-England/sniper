export enum CommandStrategy {
  Rapid,
  Phrase,
}

// Structure for the data coming from /api/commands/min
interface CommandRegistryItem {
  name: string;
  called_by: string[];
}

export class CommandCenter {
  triggers: string[] = [];
  lastConsumed: string = "";
  strategy: CommandStrategy;

  constructor() {
    this.load();
    this.strategy = CommandStrategy.Phrase;
  }

  public async load() {
    try {
      const response = await fetch("/api/commands/min");
      if (!response.ok) {
        throw new Error(`Failed to fetch commands: ${response.statusText}`);
      }
      const commands = await response.json() as CommandRegistryItem[];
      this.triggers = commands.reduce((acc, cmd) => {
        return acc.concat(cmd.called_by);
      }, [] as string[]);
      console.log(
        `[CommandCenter] Loaded ${this.triggers.length} command triggers.`,
      );
    } catch (err) {
      console.warn("[CommandCenter] Could not load command registry.", err);
    }
  }

  public getTriggers(): string[] {
    return this.triggers;
  }

  /**
   * specific method to check if a phrase is a valid trigger
   */
  public isValidTrigger(phrase: string): boolean {
    // Ensure we check case-insensitively
    return this.triggers.includes(phrase.toLowerCase());
  }

  /**
   * [NEW] Consumes a command string.
   * If it matches a valid trigger, it remembers it and returns true.
   */
  public consume(input: string): boolean {
    const cleanInput = input.trim();

    if (this.isValidTrigger(cleanInput)) {
      this.lastConsumed = cleanInput;
      console.log(`[CommandCenter] consumed: "${this.lastConsumed}"`);
      return true;
    }

    return false;
  }

  /**
   * [NEW] Getter to retrieve the last remembered command
   */
  public getLastConsumed(): string {
    return this.lastConsumed;
  }
}
