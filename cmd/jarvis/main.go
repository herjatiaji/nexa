package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/heraji/jarvis/config"
	"github.com/heraji/jarvis/core"
	"github.com/heraji/jarvis/llm"
	"github.com/heraji/jarvis/tools"
	"github.com/heraji/jarvis/tools/apps"
	"github.com/heraji/jarvis/tools/filesystem"
	"github.com/heraji/jarvis/tools/terminal"
	"github.com/heraji/jarvis/voice"
)

const version = "0.1.0"

var (
	cyan   = color.New(color.FgCyan, color.Bold)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
	dim    = color.New(color.FgHiBlack)

	enableTTSFlag bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "jarvis",
		Short: "JARVIS — Personal AI Desktop Assistant",
		Long: `JARVIS is a personal AI assistant that can understand commands,
plan actions, execute tools, and automate workflows.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// When double-clicked in Explorer, default to interactive chat mode
			enableTTSFlag = true
			return chatCmd().RunE(cmd, args)
		},
	}

	rootCmd.AddCommand(chatCmd())
	rootCmd.AddCommand(askCmd())
	rootCmd.AddCommand(listenCmd())
	rootCmd.AddCommand(voiceCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// initAgent sets up the config, LLM, tools, and agent.
func initAgent() (*core.Agent, *config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("config error: %w", err)
	}

	// Override TTS config if flag is set
	if enableTTSFlag {
		cfg.EnableTTS = true
	}

	dim.Printf("🔌 Connecting to %s...\n", cfg.LLMProvider)
	llmProvider, err := llm.New(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("LLM error: %w", err)
	}

	registry := tools.NewRegistry()

	termTool := terminal.New()
	termTool.ConfirmFunc = func(command string) bool {
		yellow.Printf("⚠️  Dangerous command detected: %s\n", command)
		fmt.Print("Execute anyway? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		return answer == "y" || answer == "yes"
	}
	registry.Register(termTool)
	registry.Register(filesystem.New())
	registry.Register(apps.New())

	agent := core.NewAgent(llmProvider, registry, cfg.SystemPrompt, cfg.MaxIterations)

	agent.OnToolCall = func(toolName string, args string) {
		yellow.Printf("🔧 Using: %s\n", toolName)
		dim.Printf("   Args: %s\n", truncate(args, 100))
	}
	agent.OnToolResult = func(toolName string, result string) {
		dim.Printf("   Result: %s\n", truncate(result, 150))
	}

	ttsInfo := ""
	if cfg.EnableTTS {
		ttsInfo = " | 🔊 TTS Active"
	}

	dim.Printf("✅ Ready! (provider: %s%s)\n\n", cfg.LLMProvider, ttsInfo)
	return agent, cfg, nil
}

func chatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat session with JARVIS",
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, cfg, err := initAgent()
			if err != nil {
				return err
			}

			ttsEngine := voice.NewTTS(cfg.TTSVoice, cfg.TTSRate, cfg.EnableTTS)

			cyan.Println("╔══════════════════════════════════════╗")
			cyan.Println("║        JARVIS v" + version + "                 ║")
			cyan.Println("║   Personal AI Desktop Assistant      ║")
			cyan.Println("╚══════════════════════════════════════╝")
			fmt.Println()
			dim.Println("Type your message, or 'exit' to quit.")
			dim.Println("Use \\ at end of line for multi-line input.")
			fmt.Println()

			if cfg.EnableTTS {
				ttsEngine.SpeakAsync("At your service, sir. How may I help you?")
			}

			scanner := bufio.NewScanner(os.Stdin)
			for {
				cyan.Print("You > ")
				var inputLines []string

				for {
					if !scanner.Scan() {
						fmt.Println()
						return nil
					}
					line := scanner.Text()

					if strings.HasSuffix(line, "\\") {
						inputLines = append(inputLines, strings.TrimSuffix(line, "\\"))
						dim.Print("...  ")
						continue
					}
					inputLines = append(inputLines, line)
					break
				}

				input := strings.TrimSpace(strings.Join(inputLines, "\n"))

				if input == "" {
					continue
				}

				lower := strings.ToLower(input)
				if lower == "exit" || lower == "quit" || lower == "bye" {
					if cfg.EnableTTS {
						ttsEngine.Speak("Goodbye, sir.")
					}
					green.Println("\n👋 Goodbye! See you later.")
					return nil
				}

				if lower == "reset" || lower == "clear" {
					agent.Reset()
					dim.Println("🔄 Conversation cleared.\n")
					continue
				}

				dim.Println("⏳ Thinking...")
				response, err := agent.Run(input)
				if err != nil {
					red.Printf("❌ Error: %v\n\n", err)
					continue
				}

				fmt.Println()
				green.Print("JARVIS > ")
				fmt.Println(response)
				fmt.Println()

				if cfg.EnableTTS {
					ttsEngine.SpeakAsync(response)
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&enableTTSFlag, "tts", "t", false, "Enable Text-to-Speech voice output")
	return cmd
}

func askCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ask [question]",
		Short: "Ask JARVIS a single question",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, cfg, err := initAgent()
			if err != nil {
				return err
			}

			ttsEngine := voice.NewTTS(cfg.TTSVoice, cfg.TTSRate, cfg.EnableTTS)

			question := strings.Join(args, " ")
			dim.Println("⏳ Thinking...")

			response, err := agent.Run(question)
			if err != nil {
				return fmt.Errorf("error: %w", err)
			}

			fmt.Println()
			green.Print("JARVIS > ")
			fmt.Println(response)

			if cfg.EnableTTS {
				ttsEngine.Speak(response)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&enableTTSFlag, "tts", "t", false, "Enable Text-to-Speech voice output")
	return cmd
}

func listenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "listen",
		Short: "Start continuous hands-free voice trigger mode (say 'Friday' or 'Hey Friday' to activate)",
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, cfg, err := initAgent()
			if err != nil {
				return err
			}

			ttsEngine := voice.NewTTS(cfg.TTSVoice, cfg.TTSRate, true)

			cyan.Println("╔══════════════════════════════════════════════════════════════╗")
			cyan.Println("║         JARVIS Hands-Free Voice Activation Mode              ║")
			cyan.Println("║   Say 'Hey Friday' or 'Friday' to activate speech listening   ║")
			cyan.Println("╚══════════════════════════════════════════════════════════════╝")
			fmt.Println()

			// Check microphone health at startup
			micOK, micStatus := voice.CheckMicrophone()
			if micOK {
				green.Printf("🎤 %s\n", micStatus)
			} else {
				yellow.Printf("⚠️ %s\n", micStatus)
				dim.Println("💡 Tip: Set your active microphone as 'Default Recording Device' in Windows Sound Settings (Win+R -> mmsys.cpl).")
				dim.Println("   Open: Win+R -> mmsys.cpl -> Recording tab -> Right-click your USB mic -> Set as Default Device")
			}

			dim.Println("Press Ctrl+C to exit voice activation mode.")
			fmt.Println()

			// Start persistent speech engine (single process, stays alive)
			dim.Println("🔄 Starting persistent speech engine...")
			resultChan, cleanup, err := voice.StartPersistentListener()
			if err != nil {
				red.Printf("❌ Failed to start speech engine: %v\n", err)
				dim.Println("💡 Falling back to text-only mode. Type your commands below.")
				return runTextOnlyListenLoop(agent, cfg, ttsEngine)
			}
			defer cleanup()

			green.Println("✅ Persistent speech engine is running!")
			ttsEngine.SpeakAsync("Friday voice activation online. At your service, sir.")
			fmt.Println()

			for {
				result, ok := <-resultChan
				if !ok {
					dim.Println("Speech engine stopped.")
					break
				}

				if strings.HasPrefix(result, "WAKE:") {
					parts := strings.Split(result, ":")
					wakeWord := "Friday"
					if len(parts) >= 2 {
						wakeWord = parts[1]
					}

					green.Printf("⚡ Wake word detected: \"%s\" — FRIDAY activated!\n", wakeWord)

					// Prompt user with Piper Neural TTS while Python records next 4s of audio
					ttsEngine.SpeakAsync("Yes, sir?")
					cyan.Println("🎙️  Listening for command (OpenAI Whisper Large v3 STT)...")

				} else if strings.HasPrefix(result, "COMMAND_WAV:") {
					wavFile := strings.TrimPrefix(result, "COMMAND_WAV:")

					// Transcribe clean WAV file using OpenAI Whisper Large v3
					whisperEngine := voice.NewWhisperSTT(cfg.GroqAPIKey)
					userCommand, err := whisperEngine.Transcribe(wavFile)
					_ = os.Remove(wavFile)

					userCommand = strings.TrimSpace(userCommand)
					if err != nil || userCommand == "" {
						dim.Println("Didn't catch that. Say 'Hey Friday' or 'Jarvis' to try again.")
						continue
					}

					green.Printf("🎙️  Heard (Whisper STT): \"%s\"\n", userCommand)

					cmdLower := strings.ToLower(userCommand)
					if cmdLower == "exit" || cmdLower == "quit" || cmdLower == "goodbye" || cmdLower == "stop" {
						ttsEngine.Speak("Goodbye, sir.")
						green.Println("\n👋 Goodbye! See you later.")
						return nil
					}

					dim.Println("⏳ Thinking...")

					response, err := agent.Run(userCommand)
					if err != nil {
						red.Printf("❌ Error: %v\n\n", err)
						continue
					}

					fmt.Println()
					green.Print("JARVIS > ")
					fmt.Println(response)
					fmt.Println()

					ttsEngine.Speak(response)
				}
			}

			return nil
		},
	}
}

// runTextOnlyListenLoop provides a keyboard-only fallback when the speech engine fails.
func runTextOnlyListenLoop(agent *core.Agent, cfg *config.Config, ttsEngine *voice.TTS) error {
	ttsEngine.SpeakAsync("Voice engine unavailable. Text mode active, sir.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		cyan.Print("You > ")
		if !scanner.Scan() {
			return nil
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		lower := strings.ToLower(input)
		if lower == "exit" || lower == "quit" || lower == "bye" {
			ttsEngine.Speak("Goodbye, sir.")
			green.Println("\n👋 Goodbye!")
			return nil
		}

		dim.Println("⏳ Thinking...")
		response, err := agent.Run(input)
		if err != nil {
			red.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Println()
		green.Print("JARVIS > ")
		fmt.Println(response)
		fmt.Println()

		if cfg.EnableTTS {
			ttsEngine.Speak(response)
		}
	}
}

func voiceCmd() *cobra.Command {
	vCmd := &cobra.Command{
		Use:   "voice",
		Short: "Manage and sample JARVIS voice options",
	}

	sampleCmd := &cobra.Command{
		Use:   "sample",
		Short: "Play an authentic JARVIS voice sample greeting",
		RunE: func(cmd *cobra.Command, args []string) error {
			cyan.Println("🔊 Playing authentic JARVIS voice sample...")
			return voice.PlayJarvisSample()
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all installed system TTS voices",
		RunE: func(cmd *cobra.Command, args []string) error {
			cyan.Println("🎙️ Installed System Voices:")
			voices, err := voice.ListVoices()
			if err != nil {
				return err
			}
			for i, v := range voices {
				fmt.Printf("%d. %s\n", i+1, v)
			}
			return nil
		},
	}

	micCmd := &cobra.Command{
		Use:   "mic",
		Short: "Check microphone input status and diagnostic info",
		Run: func(cmd *cobra.Command, args []string) {
			cyan.Println("🎤 Checking Microphone Input Status...")
			ok, msg := voice.CheckMicrophone()
			if ok {
				green.Printf("✅ %s\n", msg)
			} else {
				red.Printf("❌ %s\n\n", msg)
				yellow.Println("💡 Troubleshooting Tips:")
				fmt.Println("1. Ensure your microphone/headset is plugged in and set as Default Recording Device in Windows Sound Settings.")
				fmt.Println("2. Open Windows Settings -> Privacy & Security -> Microphone.")
				fmt.Println("3. Ensure 'Microphone access' is turned ON and 'Let desktop apps access your microphone' is turned ON.")
			}
		},
	}

	vCmd.AddCommand(sampleCmd)
	vCmd.AddCommand(listCmd)
	vCmd.AddCommand(micCmd)
	return vCmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show JARVIS version",
		Run: func(cmd *cobra.Command, args []string) {
			cyan.Printf("JARVIS v%s\n", version)
		},
	}
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// containsWakeWord checks if the recognized speech text contains a wake word.
// Uses broad fuzzy matching to account for Windows Speech Recognition misrecognitions.
func containsWakeWord(text string) bool {
	wakePatterns := []string{
		"friday", "frida", "fry day", "fry de",
		"free day", "freed", "for a day",
		"hey fry", "hi fry", "hay fry",
	}

	for _, pattern := range wakePatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}
