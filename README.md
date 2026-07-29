<div align="center">
    <img width="380" height="112" alt="Dynamite" src="https://github.com/user-attachments/assets/f7d94c24-9362-4f81-97e3-62f3bba8e40b" />
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
<!--
    ____                            ><  __
   / __ \__  ______  ____ _____ ___  \ / /____
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

<br/>

[dynamite.webm](https://github.com/user-attachments/assets/442d57de-ce0c-4ff0-9e6e-a3f07c994d9f)

<img width="3452" height="1769" alt="dynamite-screenshots-composition-1" src="https://github.com/user-attachments/assets/a07c1665-f7b5-4f49-8c89-04211870fab2" />


<br/>

## 🚧 WORK IN PROGRESS 🚧

> [!Warning]
> This is a work in progress!
> 
> Breaking changes may occur!

<br/>

## 🤖 AI NOTICE 🤖

> [!NOTE]
> At least up until the first release,
> 
> this project will be exclusively hand-crafted.

<br/>

## ❔ Why

I wanted a TUI for quickly finding and browsing Amazon Dynamo-DB items. It needed quality of life features such as the ability to toggle columns on or off, easily copying items or fields, and sorting by a given field.

I couldn't find one that felt exactly right to me, so I decided to build one myself.  

<br/>

## 📦 Installation

Install the package using go:

```bash
# install
go install github.com/wolfwfr/dynamite/cmd/dynamite@latest

# execute
dynamite
```

Or build it from source:

```bash
# obtain
git clone git@github.com/wolfwfr/dynamite.git
cd dynamite

# build
go build -o dynamite ./cmd/dynamite/

# execute
./dynamite
```

<br/>

## 🔨 Getting Started

### Help

For help, simply run:
```bash
dynamite --help
```

### Execute

Install `Dynamite`.
Then execute with the valid AWS credentials:

> [!NOTE]
>
> `Dynamite` can only perform read-operations.

<br/>

**With AWS Credentials in Environment**
```bash
# AWS_SESSION_TOKEN=*******
# AWS_PROFILE=******
dynamite
```

**With an AWS Profile Flag**
```bash
dynamite --aws_profile="my-profile"
```

### TUI

- **Navigate** with arrow-keys or vim-bindings
- **Quit** with `ctrl+c` at any point
- **Get Help** with `?`
- **Tab** between panes to shift focus for navigation or scrolling
- **Select** a table with `Enter`
- **Search** with `/`
- **Escape** search-mode and dialogs with `Esc`
- **Move Back** to the tables view with `Backspace`
- **Much more**: see the help menu (with `?`)

<br/>

## ✨ Features

Among others, Dynamite offers:

- **Easy Authentication**: AWS authentication through environment or profile
- **Region Selection**: select and switch AWS region within the TUI
- **Fuzzy Finding**: quickly search and find what you need
- **Syntax Highlighting**: JSON/YAML views of your items with highlighting
- **Visibility Toggle**: only display the columns you're interested in
- **Easy Sorting**: quickly sort your results by any field (S, B, N)
- **Flexible Formatting**: Display your items as JSON or YAML
- **Quick Copy**: Copy table name, item field or the item JSON/YAML immediately
- **Scan/Query**: Scan and Query your table, select index, order, and set keys
- **Filter**: Apply DynamoDB's filter capability to narrow down your search
- **Open in Browser**: Swiftly open the table or item in your browser and edit
- **ZOOM**: Don't need the second pane? Zoom in and only display what you need

<br/>

## 🛣 Roadmap

☝️ This is a work in progress and the following is required for a first release:

- **Code Polish**: the code and its style require some polishing
- **Testing**: Improve and extend unit testing
- **Compatibility**: Test in different terminals & at different resolution scales
- **Theme Configuration**: use the config file to configure the colours to your liking
- **README Polish**: expand the README with images and video among others

✌️ Other features I have in mind are:

- **CLI Extension**: use CLI flags to hop straight into a table of choice or launch a query.
- **Pane Configurability**: configure width distribution of the 2 panes
- **Transforms**: transform column values, e.g. unix timestamps to human readable

<br/>

## ✋ Non Goals

- **ADMIN Mode**: Although I'm considering it, I'm currently flagging write
operations as a non-goal
- **Full API Compatibility**: Full integration with all of the aws-sdk-go-v2 dynamo-db related functions

<br/>

## 🫴 Alternatives

- **[Sacha](https://github.com/Sachamama/sacha)** another 2-pane TUI that also integrates with S3, EC2, Lambda, and more!
- **[ddv](https://github.com/lusingander/ddv)** a blazing fast dynamo-DB viewer for the terminal, written in Rust

<br/>

## 🩼 Troubleshooting

**Scrolling down doesn't automatically retrieve the next page**

If not all pages have been retrieved, it is possible that pagination is disabled because of an enabled search (default key: `/`) or because page-retrieval had been explicitly canceled by pressing the `Esc` key during page-retrieval (in which case a ~PAGING~ box should appear in the bottom-left corner). Re-enable pagination with the `c` key, or view the help menu (default key: `?`) for the appropriate key-binding.

**My scan or query is not returning the expected results**

If filter parameters were left enabled, then they will affect your search results. Check for the `FILTER` box in the bottom-left corner, if it is depicted, then filters are being applied. Open the filter-parameters dialog (default key: `ctrl+f`), reset (default key: `ctrl+r`), and commit (default key: `alt+enter`). This should remove any applied filters to your current operation. Note that Dynamite maintains a separate set of filter-parameters for scan- & query-modes and stores a session per table that restores any scan-, query-, and filter-parameters when re-selecting that table within the same Dynamite session.
