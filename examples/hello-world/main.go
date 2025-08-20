package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/totmicro/mops-sdk"
)

func main() {
	plugin := sdk.NewPluginBuilder("hello-world-standalone", "1.0.0", "A standalone hello world plugin using the SDK").
		SetAuthor("MOPS Team").
		SetLicense("MIT").
		SetHomepage("https://github.com/totmicro/mops-sdk").
		AddTag("example").
		AddTag("standalone").
		AddTag("hello").
		SetMopsVersions("1.0.0", "2.0.0").
		SetDefaultConfig(map[string]any{
			"greeting":      "Hello",
			"enable_colors": true,
			"timeout":       30,
		}).
		
		// Add a simple action executor
		WithSimpleExecutor("hello", func(entry sdk.MenuEntry, input string) sdk.ActionResult {
			name := strings.TrimSpace(input)
			if name == "" {
				name = "World"
			}
			
			greeting := "Hello"
			if greetingParam, ok := entry.Params["greeting"].(string); ok {
				greeting = greetingParam
			}
			
			return sdk.ActionResult{
				Success:    true,
				Output:     fmt.Sprintf("%s, %s! 🌍", greeting, name),
				ShowOutput: true,
				Title:      "Greeting",
			}
		}).
		
		// Add a dynamic provider for different greetings
		WithSimpleProvider("greetings", "Provides different greeting options", func(param string) ([]sdk.MenuEntry, error) {
			greetings := []struct {
				key      string
				greeting string
				emoji    string
			}{
				{"1", "Hello", "👋"},
				{"2", "Hi", "🙋"},
				{"3", "Hola", "🇪🇸"},
				{"4", "Bonjour", "🇫🇷"},
				{"5", "Guten Tag", "🇩🇪"},
				{"6", "Ciao", "🇮🇹"},
			}
			
			var entries []sdk.MenuEntry
			for _, g := range greetings {
				entries = append(entries, sdk.MenuEntry{
					Key:    g.key,
					Label:  fmt.Sprintf("%s %s %s", g.emoji, g.greeting, g.emoji),
					Action: "hello",
					Params: map[string]interface{}{
						"greeting": g.greeting,
					},
				})
			}
			
			return entries, nil
		}).
		
		// Add an interactive function for real-time chat
		WithInteractiveFunction("chat", func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
			outputChan <- "🤖 Welcome to the Hello World chat! Type 'quit' to exit."
			outputChan <- "💬 What's your name?"
			
			var userName string
			
			for {
				select {
				case input := <-inputChan:
					input = strings.TrimSpace(input)
					
					if input == "quit" || input == "exit" {
						outputChan <- "👋 Goodbye! Thanks for chatting!"
						return nil
					}
					
					if userName == "" {
						userName = input
						outputChan <- fmt.Sprintf("🎉 Nice to meet you, %s!", userName)
						outputChan <- "💬 Try saying something! (or type 'quit' to exit)"
						continue
					}
					
					// Echo back with different responses
					responses := []string{
						fmt.Sprintf("😊 That's interesting, %s!", userName),
						fmt.Sprintf("🤔 Tell me more about that, %s!", userName),
						fmt.Sprintf("👍 I see, %s. That's cool!", userName),
						fmt.Sprintf("🚀 Awesome, %s!", userName),
					}
					
					response := responses[len(input)%len(responses)]
					outputChan <- response
					
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}).
		
		// Add CLI commands
		WithCLICommand("hello", "Print a hello message", func(args []string) error {
			name := "World"
			if len(args) > 0 {
				name = strings.Join(args, " ")
			}
			fmt.Printf("👋 Hello, %s! 🌍\n", name)
			return nil
		}).
		
		WithCLICommand("greet", "Print a greeting in different languages", func(args []string) error {
			greetings := map[string]string{
				"en": "Hello",
				"es": "Hola", 
				"fr": "Bonjour",
				"de": "Guten Tag",
				"it": "Ciao",
			}
			
			lang := "en"
			name := "World"
			
			if len(args) > 0 {
				if _, exists := greetings[args[0]]; exists {
					lang = args[0]
					if len(args) > 1 {
						name = strings.Join(args[1:], " ")
					}
				} else {
					name = strings.Join(args, " ")
				}
			}
			
			fmt.Printf("%s, %s! 🌍\n", greetings[lang], name)
			return nil
		}).
		
		WithCLICommand("status", "Show plugin status", func(args []string) error {
			fmt.Printf("📊 Hello World Plugin Status\n")
			fmt.Printf("✅ Plugin: hello-world-standalone v1.0.0\n")
			fmt.Printf("🕐 Current time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			fmt.Printf("🔧 Status: Ready\n")
			
			if len(args) > 0 && args[0] == "--verbose" {
				fmt.Printf("\n📋 Detailed Information:\n")
				fmt.Printf("   - Author: MOPS Team\n")
				fmt.Printf("   - License: MIT\n")
				fmt.Printf("   - Available actions: hello\n")
				fmt.Printf("   - Available providers: greetings\n")
				fmt.Printf("   - Available interactive functions: chat\n")
			}
			
			return nil
		}).
		
		Build()

	sdk.Main(plugin)
}
