export class SniperService {
    private readonly baseUrl = 'http://localhost:9090';

    /**
     * Sends the processed command string to the backend API.
     */
    public async sendCommand(command: string): Promise<void> {
        try {
            console.log(`[SniperService] Sending: ${command}`);
            await fetch(`${this.baseUrl}/api/data`, { 
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ command: command })
            });
        } catch (err) {
            console.warn('[SniperService] Connection failed. Is the backend running?', err);
        }
    }
}