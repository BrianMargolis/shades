# keith 

`keith` is a framework for changing the theme of everything in your terminal (and
to some extent, beyond) in a synchronized way. This is useful for people like
me, who use a lot of tools that are each themed separately, and who like to
switch between at least a light and dark theme, at the very least. 

The underlying architecture consists of 1 daemon and 1 or more clients,
communicating over UNIX socket. There are five messages in the protocol:
1. `subscribe:name` - begin receiving `set` messages
2. `unsubscribe:` - stop receiving `set` messages
3. `set:{theme}` - can only be sent by the daemon, a client should re-theme
   based on the value of `theme`
4. `propose:{theme}` - this is a request from a client to change the theme,
   which is useful for giving things like neovim interactive control over the
   theme. the server will broadcast `set`s in response to this
5. `get:{theme}` - firing this will result in the server firing a `set` back,
   useful in the startup context

Each message is delimited by a `\n` character.

Propose and get are optional functionality; the simplest client just
`subscribe`s, waits for and acts open any `set`, and fires an `unsubscribe` on
shutdown.

## Installation

TODO (clone and go build, I guess)

## Usage

The main way you'll interact with `keith` day-to-day is by invoking it to
change the theme.
```
keith everforest-dark
```

However, for `keith` to work at all, it must be running in server mode in the
background constantly. To do this, daemonize `keith -s` in whatever way is
appropriate for your machine. I use this on a mac, so I use launchd. A script
to perform this launchd setup lives in `scripts/setup-launchd.fish`.


## Implementing clients

Because UNIX sockets are a widely supported transport technology, and the
protocol we build on top of it is trivial to implement in any language, you can
integrate just about anything you can control programmatically with this
framework. See `client/` for examples of doing this in Go. 

However, batteries are included for many popular tools, including:
- neovim
- tmux
- alacritty
- fzf
- eza
- bat 
- btop
- mac dark light theme

All but one of the above clients can be run with a single `keith` invocation:
```
keith -c tmux alacritty fzf eza bat btop mac
```
which the launchd script will also help you daemonize on a mac.

The lone exception that is installed differently is the Neovim client, which is
a Lua plugin that embeds directly into Neovim.

## Implementing themes

TODO how this works on our end, but I think the idea is that themes are just
collections of light/dark variants, which are MOSTLY just their names - but
have the basic colormap as a fallback/crutch. each client is responsible for
ensuring reasonable fallback behavior and should prefer to use e.g. the
everforest theme for neovim than looking at our colormap This software is not
designed for the person who likes to try a new theme


## Why is this named keith?

Everybody knows that software is just little guys in your computer. There was
also a real guy outside of my computer named Keith Haring who made some art
that I like.

## TODOs
- [ ] turn the neovim stuff into a proper plugin
- [ ] basic configuration - stuff like paths
- [ ] advanced behavioral configuraton
- [ ] broader concept of theme, not just dark/light
- [ ] alias manager for themes
- [ ] integration with fzf?
- [ ] launchd setup script - want it to be nice and interactive, and ideally
  not fish specific
- [ ] write tool to more intelligently hunt down the alacritty theme - maybe we
  just have a build step that pre-pends the theme import.. ugly and now we're
  really dependent on `keith` to run alacritty at all.. maybe this is a
  separate tool?

