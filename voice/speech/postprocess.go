package speech

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// wavHeader represents the standard 44-byte RIFF WAV header.
type wavHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

// readWAV reads a WAV file and returns header + raw PCM samples as int16.
func readWAV(path string) (*wavHeader, []int16, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var hdr wavHeader
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return nil, nil, fmt.Errorf("invalid WAV header: %w", err)
	}

	if string(hdr.ChunkID[:]) != "RIFF" || string(hdr.Format[:]) != "WAVE" {
		return nil, nil, fmt.Errorf("not a valid WAV file")
	}

	if hdr.BitsPerSample != 16 {
		return nil, nil, fmt.Errorf("only 16-bit PCM WAV is supported, got %d-bit", hdr.BitsPerSample)
	}

	numSamples := int(hdr.Subchunk2Size) / 2
	samples := make([]int16, numSamples)
	if err := binary.Read(f, binary.LittleEndian, &samples); err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, nil, fmt.Errorf("error reading PCM data: %w", err)
	}

	return &hdr, samples, nil
}

// writeWAV writes PCM int16 samples back to a WAV file with the given header params.
func writeWAV(path string, hdr *wavHeader, samples []int16) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Update sizes
	hdr.Subchunk2Size = uint32(len(samples) * 2)
	hdr.ChunkSize = 36 + hdr.Subchunk2Size

	if err := binary.Write(f, binary.LittleEndian, hdr); err != nil {
		return err
	}
	return binary.Write(f, binary.LittleEndian, samples)
}

// NormalizeVolume peak-normalizes a WAV file to targetDB (e.g., -3.0 dBFS).
// This ensures consistent volume across all speech chunks without clipping.
func NormalizeVolume(wavPath string, targetDB float64) error {
	hdr, samples, err := readWAV(wavPath)
	if err != nil {
		return err
	}

	if len(samples) == 0 {
		return nil
	}

	// Find peak amplitude
	var peak int16
	for _, s := range samples {
		abs := s
		if abs < 0 {
			abs = -abs
		}
		if abs > peak {
			peak = abs
		}
	}

	if peak == 0 {
		return nil // Silent file, nothing to normalize
	}

	// Calculate gain factor
	// Target amplitude = 32767 * 10^(targetDB/20)
	targetAmp := 32767.0 * math.Pow(10, targetDB/20.0)
	gain := targetAmp / float64(peak)

	// Apply gain with soft clipping
	for i, s := range samples {
		amplified := float64(s) * gain
		// Soft clip to prevent harsh distortion
		if amplified > 32767 {
			amplified = 32767
		} else if amplified < -32768 {
			amplified = -32768
		}
		samples[i] = int16(amplified)
	}

	return writeWAV(wavPath, hdr, samples)
}

// TrimSilence removes leading and trailing silence from a WAV file.
// thresholdAmp is the amplitude below which samples are considered silence (e.g., 200).
func TrimSilence(wavPath string, thresholdAmp int16) error {
	hdr, samples, err := readWAV(wavPath)
	if err != nil {
		return err
	}

	if len(samples) == 0 {
		return nil
	}

	// Find first non-silent sample
	start := 0
	for start < len(samples) {
		amp := samples[start]
		if amp < 0 {
			amp = -amp
		}
		if amp > thresholdAmp {
			break
		}
		start++
	}

	// Find last non-silent sample
	end := len(samples) - 1
	for end > start {
		amp := samples[end]
		if amp < 0 {
			amp = -amp
		}
		if amp > thresholdAmp {
			break
		}
		end--
	}

	if start >= end {
		return nil // Entire file is silent
	}

	// Keep a small buffer (50ms) of silence at edges for natural onset/offset
	bufferSamples := int(hdr.SampleRate) * 50 / 1000 // 50ms buffer
	start -= bufferSamples
	if start < 0 {
		start = 0
	}
	end += bufferSamples
	if end >= len(samples) {
		end = len(samples) - 1
	}

	trimmed := samples[start : end+1]
	return writeWAV(wavPath, hdr, trimmed)
}

// ConcatWAVs joins multiple WAV files into a single output file.
// Inserts a brief crossfade (fadeMs) between chunks for smooth transitions.
func ConcatWAVs(paths []string, output string, fadeMs int) error {
	if len(paths) == 0 {
		return fmt.Errorf("no WAV files to concatenate")
	}

	if len(paths) == 1 {
		// Single file — just copy
		data, err := os.ReadFile(paths[0])
		if err != nil {
			return err
		}
		return os.WriteFile(output, data, 0644)
	}

	var allSamples []int16
	var baseHdr *wavHeader

	for i, p := range paths {
		hdr, samples, err := readWAV(p)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", p, err)
		}

		if i == 0 {
			baseHdr = hdr
			allSamples = append(allSamples, samples...)
		} else {
			// Apply crossfade between chunks for natural transition
			fadeSamples := int(hdr.SampleRate) * fadeMs / 1000
			if fadeSamples > len(allSamples) {
				fadeSamples = len(allSamples)
			}
			if fadeSamples > len(samples) {
				fadeSamples = len(samples)
			}

			if fadeSamples > 0 {
				// Crossfade: fade out tail of previous, fade in head of current
				for j := 0; j < fadeSamples; j++ {
					ratio := float64(j) / float64(fadeSamples)
					prevIdx := len(allSamples) - fadeSamples + j
					prevSample := float64(allSamples[prevIdx]) * (1.0 - ratio)
					currSample := float64(samples[j]) * ratio
					allSamples[prevIdx] = int16(prevSample + currSample)
				}
				// Append remaining samples after crossfade zone
				allSamples = append(allSamples, samples[fadeSamples:]...)
			} else {
				allSamples = append(allSamples, samples...)
			}
		}
	}

	return writeWAV(output, baseHdr, allSamples)
}
