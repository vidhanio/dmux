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
"/ping text:string"
```

'nuff said.
