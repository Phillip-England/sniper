export class SniperService {
  private readonly baseUrl = "http://localhost:9090";

  /**
   * Sends the processed command string to the backend API.
   * @returns A promise that resolves to the HTTP status code of the response.
   */
  public async sendCommand(command: string): Promise<number> {
    try {
      console.log(`[SniperService] Sending: ${command}`);
      const response = await fetch(`${this.baseUrl}/api/data`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ command: command }),
      });

      // Check if the request was successful
      if (!response.ok) {
        // Log a warning if the response status is not in the 200-299 range
        console.warn(
          `[SniperService] Request failed with status: ${response.status} ${response.statusText}`,
        );
      }

      // Return the status code
      return response.status;
    } catch (err) {
      console.warn(
        "[SniperService] Connection failed. Is the backend running?",
        err,
      );
      // Return a status code of 0 or -1 to signify a network/connection error
      // that prevented a proper HTTP response (e.g., DNS failure, server not running).
      // 0 is often used for network errors in some contexts, but -1 is also common.
      return 0;
    }
  }
}