package message

import "math/rand"

var mutterWitze = []string{
	"Deine Mutter ist so fett, ihre Blutgruppe ist Nutella.",
	"Aldi hat angerufen, deine Mutter hängt im Drehkreuz fest.",
	"Deine Mutter ist so fett, ihre Blutgruppe ist Nutella.",
	"Deine Mutter ist die Stärkste im Knast.",
	"Deine Mutter heißt Zonk und wohnt in Tor 3.",
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
	"Deine Mutter kostet nichts, außer Überwindung",
	"Deine Mutter ist so fett, ich wollte mich umdrehn aber da war sie auch.",
	"Deine Mutter arbeitet bei Nordsee als Geruch",
	"Deine Mutter zieht Katapulte nach Gondor.",
	"Deine Mutter ist so hässlich, bei ihr wird eingebrochen um die Vorhänge zuzuziehen.",
	"Wenn man deine Mutter überfahren will, muss man auf halber Strecke nachtanken.",
	"Deine Mutter ist so fett, sie guckt im Restaurant auf die Speisekarte und sagt: „Ok.“",
	"Deine Mutter ist so fett, bei ihr ist jeder Geburtstag ein runder Geburtstag",
	"Deine Mutter bellt wenn es klingelt.",
	"Deine Mutter hat so wenig Klasse, sie könnte eine marxistische Utopie sein.",
	"Deine Mutter ist so fett, sie macht Passfotos bei google earth.",
	"Amerikanische Forscher haben aus deiner Mutter herausgefunden.",
	"Deine Mutter ist so fett, dass sie aus dem Bett fällt, und zwar auf beiden Seiten zugleich.",
	"Deine Mutter sitzt bei Aldi unter der Kasse und macht: „Piep“",
	"Deine Mutter steht vor Aldi und ruft: „Ich bin billiger!“",
}

func MutterWitz() string {
	if len(mutterWitze) != 0 {
		return mutterWitze[rand.Intn(len(mutterWitze))]
	}
	return "Deine Mutter ist so fett, sie hat das Array mit den Witzen aufgegessen."
}
