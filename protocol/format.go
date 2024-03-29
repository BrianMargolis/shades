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
  - get

An empty noun is valid (e.g. 'subscribe:').
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
	return []byte("set:" + theme + "")
}

func Get() []byte {
	return []byte("get:" + "\n")
}
