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

chunk_size = 1280    # 80ms chunks at 16kHz mono (1280 samples)
threshold = 0.20     # Neural score threshold for openWakeWord
cooldown_sec = 2.0   # Cooldown period
last_trigger_time = 0.0

current_state = "WAKE_SEARCH"
recorded_chunks = []
silence_chunks = 0
has_speech_started = False

VAD_ENERGY_THRESHOLD = 200   # RMS Energy threshold for speech vs silence
SILENCE_CUTOFF_CHUNKS = 15   # 1.2s silence cutoff
MAX_RECORDING_CHUNKS  = 62   # ~5.0s max recording duration

def get_audio_int16_and_rms(raw_float):
    int16_chunk = (raw_float * 32767.0).astype(np.int16)
    rms = float(np.sqrt(np.mean(int16_chunk.astype(np.float32)**2)))
    return int16_chunk, rms

def audio_callback(indata, frames, time_info, status):
    global current_state, recorded_chunks, last_trigger_time, silence_chunks, has_speech_started
    
    if current_state == "PAUSED":
        return

    raw_float = indata[:, 0]
    audio_chunk, rms_energy = get_audio_int16_and_rms(raw_float)
    
    if current_state == "WAKE_SEARCH":
        now = time.time()
        if now - last_trigger_time < cooldown_sec:
            return

        # Feed real-time 80ms audio frame to openWakeWord neural engine
        oww_model.predict(audio_chunk)
        
        # Check predictions for loaded wake word models (hey_jarvis, alexa, hey_rhasspy, etc.)
        for model_name, scores in oww_model.prediction_buffer.items():
            score = scores[-1]
            if score > threshold:
                last_trigger_time = now
                
                # Format clean wake word trigger name
                clean_name = "Hey Jarvis"
                if "alexa" in model_name:
                    clean_name = "Alexa"
                elif "rhasspy" in model_name:
                    clean_name = "Rhasspy"
                
                # 1. Trigger Wake Event
                print(f"WAKE:{clean_name}:{score:.2f}", flush=True)
                
                # 2. Reset openWakeWord model internal buffers
                try:
                    oww_model.reset()
                except:
                    oww_model.prediction_buffer[model_name] = [0.0] * len(scores)
                
                # 3. Transition to RECORDING state from the SAME mic stream
                current_state = "RECORDING"
                recorded_chunks = []
                silence_chunks = 0
                has_speech_started = False
                break

    elif current_state == "RECORDING":
        recorded_chunks.append(audio_chunk)
        
        is_speech = rms_energy >= VAD_ENERGY_THRESHOLD
        if is_speech:
            has_speech_started = True
            silence_chunks = 0
        else:
            if has_speech_started:
                silence_chunks += 1
                
        should_stop = (has_speech_started and silence_chunks >= SILENCE_CUTOFF_CHUNKS) or (len(recorded_chunks) >= MAX_RECORDING_CHUNKS)
        
        if should_stop and len(recorded_chunks) >= 10:
            full_audio = np.concatenate(recorded_chunks)
            
            temp_wav = os.path.join(tempfile.gettempdir(), f"jarvis_cmd_{int(time.time()*1000)}.wav")
            with wave.open(temp_wav, 'wb') as wf:
                wf.setnchannels(1)
                wf.setsampwidth(2) # 16-bit PCM
                wf.setframerate(16000)
                wf.writeframes(full_audio.tobytes())
            
            # Emit WAV path to stdout for Go & OpenAI Whisper STT
            print(f"COMMAND_WAV:{temp_wav}", flush=True)
            
            try:
                oww_model.reset()
            except:
                pass
            last_trigger_time = time.time()
            current_state = "WAKE_SEARCH"
            recorded_chunks = []
            silence_chunks = 0
            has_speech_started = False

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
                try:
                    oww_model.reset()
                except:
                    pass
                last_trigger_time = time.time()
                current_state = "WAKE_SEARCH"
                recorded_chunks = []
                silence_chunks = 0
                has_speech_started = False
                print("RESUMED", flush=True)
except Exception as e:
    print(f"STREAM_ERROR:{e}", flush=True)
