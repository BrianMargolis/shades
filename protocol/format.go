package protocol

/**
The protocol that the daemon and the various clients use to talk to each other
is very simple and works over plain text. Each message is delimited by a \n
character, and is of the structure VERB:NOUN

Where the valid verbs are:
	- subscribe:[client name] (valid noun is any string, the name of the client)
  - unsubscribe:
	- propose:[light|dark]
	- set:[light|dark]
	- palette:[json object of color name to hex string]
  - get

An empty noun is valid (e.g. 'subscribe:'). A noun may contain colons, so only
the first one delimits the verb.
**/

func Subscribe(name string) []byte {
	return []byte("subscribe:" + name + "\n")
}

func Unsubscribe() []byte {
	return []byte("unsubscribe:" + "\n")
}

func Propose(theme string) []byte {
	return []byte("propose:" + theme + "\n")
}

func Set(theme string) []byte {
	return []byte("set:" + theme + "\n")
}

// Palette carries the resolved colors of the theme a set is about to name, so
// a client that cannot read shades.yaml itself does not have to keep its own
// copy of the palette. It is always sent before the set it belongs to.
func Palette(payload string) []byte {
	return []byte("palette:" + payload + "\n")
}

func Get() []byte {
	return []byte("get:" + "\n")
}
