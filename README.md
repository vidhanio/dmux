# dmux

Turn

```go
&discordgo.ApplicationCommand{
    Name:        "echo",
    Description: "echo",
    Type:        discordgo.ChatApplicationCommand,
    Options: []*discordgo.ApplicationCommandOption{
        {
            Name:        "text",
            Description: "text",
            Type:        discordgo.ApplicationCommandOptionString,
        },
    },
}
```

into

```go
"/echo text:string"
```

'nuff said.
