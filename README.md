<div align="center">
    ____                              _ __     
   / __ \__  ______  ____ _____ ___  (_) /____ 
  / / / / / / / __ \/ __ `/ __ `__ \/ / __/ _ \
 / /_/ / /_/ / / / / /_/ / / / / / / / /_/  __/
/_____/\__, /_/ /_/\__,_/_/ /_/ /_/_/\__/\___/ 
      /____/                                   
  <h1>🧨🧨🧨</h1>
</div>

<!--
________                                 __________
___  __ \____  ______________ _______ ______(_)_  /_____
__  / / /_  / / /_  __ \  __ `/_  __ `__ \_  /_  __/  _ \
_  /_/ /_  /_/ /_  / / / /_/ /_  / / / / /  / / /_ /  __/
/_____/ _\__, / /_/ /_/\__,_/ /_/ /_/ /_//_/  \__/ \___/
        /____/
-->
<!--
    ____                              _ __
   / __ \__  ______  ____ _____ ___  (_) /____
  / / / / / / / __ \/ __ `/ __ `__ \/ / __/ _ \
 / /_/ / /_/ / / / / /_/ / / / / / / / /_/  __/
/_____/\__, /_/ /_/\__,_/_/ /_/ /_/_/\__/\___/
      /____/
-->

<div align="center">
  <p>
    Amazon Dynamo-DB Query Engine for the Terminal. <br/>
    A fast 2-pane TUI full of QOL features.
  </p>
</div>

<div align="center">
🚧#################################🚧
#                                   #
#              NOTICE               #
#                                   #
#    This is a work in progress!    #
#    Breaking changes may occur!    #
#                                   #
🚧#################################🚧
</div>

<div align="center">
🤖#################################🤖
#                                   #
#            AI NOTICE              #
#                                   #
#   At least up until the first     #
#   release, this project will be   #
#      exclusively hand-coded       #
#                                   #
🤖#################################🤖
</div>

## ❔ Why

I wanted a TUI for quickly finding and browsing Amazon Dynamo-DB items. It
needed quality of life features such as the ability to toggle columns on or off,
easily copying items or fields, and sorting by a given field.

I couldn't find one that felt exactly right to me, so I decided to build one
myself.

## 📦 Installation

Install the package using go:

```bash
go install github.com/wolfwfr/dynamite/cmd@latest
```

Or build it from source:

```bash
# obtain
git clone github.com/wolfwfr/dynamite.git
cd dynamite

# build
go build -o dynamite ./cmd/

# execute
./dynamite
```

## ✨ Features

Among others, Dynamite offers:

- **Easy Authentication**: AWS authentication through environment or profile
- **Region Selection**: select and switch AWS region within the TUI
- **Fuzzy Finding**: quicly search and find what you need
- **Visibility Toggle**: only display the columns you're interested in
- **easy sorting**: quickly sort your results by any field (S, B, N)
- **Flexible Formatting**: Display your items as JSON or YAML
- **Quick Copy**: Copy table name, item field or the item JSON/YAML immediately
- **Scan/Query**: Scan and Query your table, select index, order, and set keys
- **ZOOM**: Don't need the second pane? Zoom in and only display what you need

## 🛣 Roadmap

This is a work in progress and the following is required for a first release:

- **Code polish**: the code and its style require some polishing
- **Testing**: Improve and extend unit testing
- **Compatibility**: Test in different terminals & at different resolution scales
- **Theme Configuration**: use the config file to configure the colours to your
  liking
- **README Polish**: expand the README with images and video among others

Other features I have in mind are:

- **DynamoDB Filter**: implement integration with scan/query filter options
- **CLI Extension**: use CLI flags to hop straight into a table of choice or
launch a query.
- **Pane Configurability**: configure width distribution of the 2 panes
- **Transforms**: transform column values, e.g. unix timestamps to human
readable

## ✋ Non Goals

- **ADMIN mode**: Although I'm considering it, I'm currently flagging write
operations as a non-goal
- **Full API compatibility**: Full integration with all of the aws-sdk-go-v2

## 🫴 Alternatives

- **[Sacha](https://github.com/Sachamama/sacha)** another 2-pane TUI that also integrates with S3, EC2, Lambda, and more!
- **[ddv](https://github.com/lusingander/ddv)** a blazingly fast dynamo-DB
viewer for the terminal, written in Rust
