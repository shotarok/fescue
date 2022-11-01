# fescue
Calendar Heatmap Generation of the read articles on Feedly.

```bash
go run main.go -file token.json -date 2022-10-31 | ~/go/bin/calendarheatmap > chart.png
```

![](./chart.png)

[NOTE] Please take a look at [sfanous/go-feedly](https://github.com/sfanous/go-feedly) for the format of token.json
