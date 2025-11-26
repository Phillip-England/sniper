import pyaudio
import os
import json
from vosk import Model, KaldiRecognizer

# --- CONFIGURATION ---

# 1. Ensure this path points to the model folder you downloaded and unzipped.
MODEL_PATH = "vosk" 

# 2. Audio Settings (Do not change unless necessary)
RATE = 16000 
CHUNK = 8192
CHANNELS = 1

def start_transcription():
    """Initializes Vosk and PyAudio and runs the transcription loop."""
    if not os.path.exists(MODEL_PATH):
        print(f"Error: Model path not found at {MODEL_PATH}")
        print("Please download the small English model and unzip it into the script directory.")
        return

    # 1. Load the Vosk Model and Recognizer
    print(f"Loading Vosk model from: {MODEL_PATH}...")
    model = Model(MODEL_PATH)
    recognizer = KaldiRecognizer(model, RATE)

    # 2. Initialize PyAudio (Microphone Input)
    audio = pyaudio.PyAudio()
    stream = audio.open(
        format=pyaudio.paInt16,
        channels=CHANNELS,
        rate=RATE,
        input=True,
        frames_per_buffer=CHUNK,
    )

    print("\n\n🎤 Listening... Speak now to see real-time transcription.")
    print("   (Press Ctrl+C to stop.)")

    try:
        # 3. Main Transcription Loop
        while True:
            data = stream.read(CHUNK, exception_on_overflow=False)
            
            # Process audio data with Vosk
            if recognizer.AcceptWaveform(data):
                # Final result received (the user finished speaking)
                result_json = json.loads(recognizer.Result())
                final_text = result_json.get("text", "")
                if final_text:
                    print(f"\n✅ Final: {final_text}")
            else:
                # Get and print the partial (interim) result for a real-time feel
                partial_json = json.loads(recognizer.PartialResult())
                partial_text = partial_json.get("partial", "")
                
                # Overwrite the previous partial result on the same line
                print(f"👂 Partial: {partial_text}\r", end="")

    except KeyboardInterrupt:
        print("\nStopping transcription...")
    except Exception as e:
        print(f"\nAn error occurred: {e}")
    finally:
        # Cleanup
        stream.stop_stream()
        stream.close()
        audio.terminate()
        print("Microphone stream closed.")

if __name__ == "__main__":
    start_transcription()