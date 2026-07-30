import os
import sys
import time
import tempfile
import wave
import numpy as np

try:
    import sounddevice as sd
    import openwakeword
    from openwakeword.model import Model
except ImportError as e:
    print(f"IMPORT_ERROR:{e}", flush=True)
    sys.exit(1)

# Initialize openWakeWord engine with ONNX models
try:
    oww_model = Model(inference_framework="onnx")
    print("OPENWAKEWORD_READY", flush=True)
except Exception as e:
    try:
        import openwakeword.utils
        openwakeword.utils.download_models()
        oww_model = Model(inference_framework="onnx")
        print("OPENWAKEWORD_READY", flush=True)
    except Exception as ex:
        print(f"MODEL_ERROR:{ex}", flush=True)
        sys.exit(1)

chunk_size = 1280 # 80ms chunks at 16kHz mono (1280 samples)
threshold = 0.45   # Neural score threshold (0.0 to 1.0)

# State Management
# STATE: "WAKE_SEARCH" | "RECORDING" | "PAUSED"
current_state = "WAKE_SEARCH"
recorded_chunks = []
recording_target_chunks = 50 # 50 * 80ms = 4.0 seconds of audio

def audio_callback(indata, frames, time_info, status):
    global current_state, recorded_chunks
    
    if current_state == "PAUSED":
        return

    # Convert float32 16kHz mono audio chunk to int16
    audio_chunk = (indata[:, 0] * 32767).astype(np.int16)
    
    if current_state == "WAKE_SEARCH":
        # Feed real-time 80ms audio frame to openWakeWord neural engine
        oww_model.predict(audio_chunk)
        
        # Check predictions for all loaded wake word models
        for model_name, scores in oww_model.prediction_buffer.items():
            score = scores[-1]
            if score > threshold:
                # 1. Trigger Wake Event
                print(f"WAKE:{model_name}:{score:.2f}", flush=True)
                
                # 2. Reset score buffer to prevent re-triggering
                oww_model.prediction_buffer[model_name] = [0.0] * len(scores)
                
                # 3. Transition directly to RECORDING state from the SAME mic stream!
                current_state = "RECORDING"
                recorded_chunks = []
                break

    elif current_state == "RECORDING":
        recorded_chunks.append(audio_chunk)
        
        # Check if we have accumulated target duration (4 seconds)
        if len(recorded_chunks) >= recording_target_chunks:
            # Concatenate recorded audio chunks
            full_audio = np.concatenate(recorded_chunks)
            
            # Save to temporary WAV file (16kHz 16-bit mono)
            temp_wav = os.path.join(tempfile.gettempdir(), f"jarvis_cmd_{int(time.time()*1000)}.wav")
            with wave.open(temp_wav, 'wb') as wf:
                wf.setnchannels(1)
                wf.setsampwidth(2) # 16-bit PCM
                wf.setframerate(16000)
                wf.writeframes(full_audio.tobytes())
            
            # Emit WAV path to stdout for Go & OpenAI Whisper STT
            print(f"COMMAND_WAV:{temp_wav}", flush=True)
            
            # Transition back to WAKE_SEARCH (or wait for Go signal)
            current_state = "WAKE_SEARCH"
            recorded_chunks = []

# Start single unified microphone stream (16kHz mono)
try:
    with sd.InputStream(channels=1, samplerate=16000, blocksize=chunk_size, callback=audio_callback):
        while True:
            line = sys.stdin.readline()
            if not line or "QUIT" in line:
                break
            elif "PAUSE" in line:
                current_state = "PAUSED"
                print("PAUSED", flush=True)
            elif "RESUME" in line:
                current_state = "WAKE_SEARCH"
                recorded_chunks = []
                print("RESUMED", flush=True)
except Exception as e:
    print(f"STREAM_ERROR:{e}", flush=True)
