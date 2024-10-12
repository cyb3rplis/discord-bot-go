package message

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/sound"
)

var mutterWitze = []string{
	"Deine Mutter ist so fett, ihre Blutgruppe ist Nutella.",
	"Aldi hat angerufen, deine Mutter hängt im Drehkreuz fest.",
	"Deine Mutter ist so fett, ihre Blutgruppe ist Nutella.",
	"Deine Mutter ist der Stärkste im Knast.",
	"Deine Mutter heißt Zonk und wohnt in Tor 3.",
	"Deine Mutter war schon als kleiner Junge hässlich.",
	"Deine Mutter ist der Fehler bei „Matrix“.",
	"Google Earth hat angerufen, deine Mutter steht im Bild.",
	"Deine Mutter sitzt besoffen im Schrank und sagt: „Willkommen in Narnia“.",
	"Deine Mutter stolpert übers W-LAN-Kabel.",
	"Der Dönerladen hat angerufen: Deine Mutter dreht sich nicht mehr.",
	"Deine Mutter kratzt an Bäumen nach Hartz IV.",
	"Deine Mutter schreckt mit ihrem Gesicht die Eier ab.",
	"Deine Mutter setzt sich in eine Badewanne voll mit Fanta, damit sie auch mal aus 'ner Limo winken kann.",
	"Deine Mutter dreht die Quadrate bei Tetris.",
	"Deine Mutter trug zur Hochzeit eine Burger-King-Krone.",
	"Deine Mutter ist so blöd, dass selbst Bob der Baumeister sagt: Ne, das schaffen wir nicht.",
	"Deine Mutter sammelt Laub für den Blätterteig.",
	"Deine Mutter ist so doof, die sitzt auf dem Fernseher und guckt Sofa.",
	"Deine Mutter krümelt beim Trinken.",
}

// AudioMessageHandler is created on any channel that the authenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := config.GetValueString("general", "prefix", ".")

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Content) == 0 { // Ignore empty messages
		fmt.Println("Empty content..")
		return
	}
	// Extract the command and arguments
	args := strings.Split(m.Content, " ")
	command := args[0]

	switch {
	//if the command starts with the prefix and is not a list or stop command
	case command == fmt.Sprintf("%shelp", prefix):
		// Default case: show help message
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🧐 Usage: \n > » List Categories: <%slist> \n", prefix))
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	// LIST CATEGORIES
	case command == fmt.Sprintf("%slist", prefix):
		// Get all sound folders to use for later
		soundFolders, err := sound.WalkSoundFolder()
		if err != nil {
			fmt.Println("Error getting sound subfolders")
			return
		}
		if len(soundFolders) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sound categories found.")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}
		content := []discordgo.MessageComponent{}
		row := discordgo.ActionsRow{}
		for i, folder := range soundFolders {
			// only 5 buttons per row - discord does not allow more
			if i > 0 && i%5 == 0 {
				content = append(content, row)
				row = discordgo.ActionsRow{}
			}
			row.Components = append(row.Components, discordgo.Button{
				Label:    folder,
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("list_sounds_%s", folder),
			})
		}
		// Append the last row if it has any components
		if len(row.Components) > 0 {
			content = append(content, row)
		}
		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content:    "➡ Click on a category [blue button]",
			Components: content,
		})
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	case strings.Contains(strings.ToLower(m.Content), "albani"):
		_, err := s.ChannelMessageSend(m.ChannelID, "Erfundene Sprache")
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	case strings.Contains(strings.ToLower(m.Content), "mutter"):
		_, err := s.ChannelMessageSend(m.ChannelID, mutterWitze[rand.Intn(len(mutterWitze))])
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	default:
		return
	}
}

func SendStopButton(s *discordgo.Session, m *discordgo.MessageCreate, soundName string) {
	content := []discordgo.MessageComponent{}
	row := discordgo.ActionsRow{}
	row.Components = append(row.Components, discordgo.Button{
		Label:    "Stop Sound",
		Style:    discordgo.PrimaryButton,
		CustomID: "stop_sound",
	})
	// Append the last row if it has any components
	if len(row.Components) > 0 {
		content = append(content, row)
	}
	_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content:    "➡ Currently Playing: " + soundName,
		Components: content,
	})
	if err != nil {
		fmt.Println("error sending message:", err)
		return
	}
}
