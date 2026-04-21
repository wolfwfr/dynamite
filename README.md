# TODO

## Description

The dynamite TUI will support the following (example) workflow:

1. Open dynamite
2. It attempts to load a config file from `~/.config/dynamite` or at location of
   `--config` or `-c` flag.
3. It selects a region based on `--region` or `-r` flag, or the config file, or
   defaults to `us-east-1`, in that order.
4. It uses an AWS profile from the `--profile` or `-p` flag, or the config file,
   or defaults to looking for AWS credentials in ENV.
5. If no credentials were found, it opens a dialog that lists supported
   credential provisioning methods and asks for a profile, or for the user to
   quit the application and configure environment variables.
6. Upon resolving AWS credentials, it lists tables in the `tables` view.
7. The user can change the AWS region and browse the listed tables, observing
   table details in the preview pane.
8. The user can select a table, which opens the `table` view.
9. In the table view, the user can `scan`, `query`, or `administrate` the table.
10. The user can select query parameters, search results, paginate through
    results, deselect attributes for display, and transform unix timestamps, and
    dump results to JSON.

### Views

**TABLES**

The `tables` view depicts three boxes on screen, two main ones that split the
screen vertically and one `row` type box at the bottom.

The left box depicts table names and allows table selection. Navigation is
operated through arrow keys or hjkl. Selection is operated through up/down
motions. Hitting 'Enter' on a selected table instructs the TUI to navigate to
the **TABLE** view.

The right box depicts the selected table details. Scrolling is operated through
ctrl+u and ctrl+d.

The bottom box depicts environment & help information. It depicts
`[r]egion: <AWS-region>`, and `[h]elp`.

_Supported view operations_:

| name        | hotkey | description                              |
| ----------- | ------ | ---------------------------------------- |
| region      | 'r'    | select a different AWS-region\*          |
| search      | '/'    | use fuzzy-finding to find a table        |
| dump        | 'd'    | dump entire table contents to local JSON |
| dump config | 'D'    | open table dump config dialog\*\*        |
| quit        | 'q'    | quit the application                     |

\*the config allows for defining 'starred' regions, which can be selected more
easily  
\*\*table dump config dialog allows closing of config with or without dumping

---

**TABLE**

The `table` view depicts four boxes on screen, two main ones that split the
screen vertically, one `row` type box at the bottom, and finally one small `row`
style box in the left-lower corner that depicts the `mode`.

The left box depicts the items in the current page as single-line rows. At the
top, it depicts the page-number and the total number of pages (Based on table
item-count). Navigating pages is operated through `tab` and `shift+tab`.
Navigation of the keys is operated through the arrow keys or hjkl. Selection is
operated through up/down motions. Selecting 'Enter' on an item opens the item
JSON in the default editor.

The right box depicts a JSON representation of the item. Scrolling is operated
through ctrl+u and ctrl+d.

The bottom box depicts environment & help information. It depicts
`[i]ndex/table: <index/table>`, `[h]elp`, and symbols for active `[/]filter`,
`[t].

The bottom `mode` box depicts the `mode`, which is one of three:

- scan (default)
- query
- admin

`Scan mode` is opened by default and simply depicts the first page of scan
results. Scan and query page-sizes can be configured in the config but are small
by default (nr of lines on screen). Scan mode allows a filter.

When `query mode` is opened without a previously configured query, it
automatically opens the **Table Query** dialog.

`Admin mode` allows limited table admin operations such as table deletion or
provisioning. This is a low priority feature and is not planned for
implementation anytime soon.

_Supported view operations_:

| name          | mode  | hotkey      | description                             |
| ------------- | ----- | ----------- | --------------------------------------- |
| scan          | S/Q/A | 'S'         | switch to 'scan' mode                   |
| query         | S/Q/A | 'Q'         | switch to 'query' mode                  |
| admin         | S/Q/A | 'A'         | switch to 'admin' mode                  |
| quit          | S/Q/A | 'q'         | quit the application                    |
| filter        | S/Q   | 'f'         | open filter dialog                      |
| filter fields | S/Q   | 'F'         | (de)select what fields to depict        |
| transform     | S/Q   | 't'         | open transform attribute dailog         |
| zoom left     | S/Q   | 'z'         | hide preview and only depict keys       |
| change query  | Q     | 'c'         | open query dialog                       |
| search        | S/Q   | '/'         | use fuzzy-finding to find items in view |
| dump          | S/Q   | 'd'         | dump the selected item to JSON\*        |
| dump config   | S/Q   | 'D'         | open item dump config dialog\*\*        |
| next-page     | S/Q   | 'tab'       | navigate to the next page               |
| previous-page | S/Q   | 'shift-tab' | navigate to the previous page           |

\*dump with lower-case 'd' immediately dumps  
\*\*item dump config dialog allows closing of config with or without dumping

---

**TABLE KEYS**

This view is the `zoomed left` view for the **TABLE** view.

---

### Dialogs

**Region Selection**

This dialog allows the user to select an AWS-region. It depicts
`starred regions` (if available) at the top, in a dedicated section.

The supported regions are built-in and extra regions can be configured in the
config file (just in case).

_dialog operations_:

| name   | hotkey  | description                   |
| ------ | ------- | ----------------------------- |
| select | 'Enter' | select region                 |
| close  | 'c'     | close without changing region |

---

**Missing Credentials**

This dialog explains to the user that `dynamite` requires AWS credentials and
can obtain credentials either from a locally configured AWS profile or from
appropriate environment variables.

It comes with the following message:

```text
DYNAMITE requires AWS credentials to operate. These can be provided explicitly,
by relating a locally configured AWS profile here, directly in the config (see
[readme](https://github.com/wolfwfr/dynamite), or as `--profile` or `-p` flag
when booting DYNAMITE.

Alternatively, you can close DYNAMITE and boot it in an environment with AWS
credentials available as environment variables (e.g. by using [aws-vault](https://github.com/99designs/aws-vault)).

Setting a profile here will save it to the config at <config-location>. If you
ever want to change it, you can change it there.
```

_dialog fields_:

| name    | input type | description            |
| ------- | ---------- | ---------------------- |
| profile | string     | local AWS profile name |

_dialog operations_:

| name  | hotkey | description     |
| ----- | ------ | --------------- |
| apply | 'a'    | apply and close |
| quit  | 'q'    | quit DYNAMITE   |

---

**Table Query**

This dialog allows the user to configure a query for the table.

_dialog fields_:

| name                | input type | description         |
| ------------------- | ---------- | ------------------- |
| partition key value | string     | partition key value |
| sort key value      | string     | sort key value      |

_dialog operations_:

| name  | hotkey  | description     |
| ----- | ------- | --------------- |
| apply | 'Enter' | apply and close |

---

**Table Index**

This dialog allows the user to select the base table or one of the available
indexes.

_dialog fields_:

| name           | input type                                  | description    |
| -------------- | ------------------------------------------- | -------------- |
| table or index | drowdown (base table and available indexes) | table or index |

_dialog operations_:

| name              | hotkey | description                 |
| ----------------- | ------ | --------------------------- |
| apply             | 'a'    | apply and close             |
| apply and execute | 'A'    | apply and execute           |
| clear             | 'C'    | remove custom configuration |

---

**Item Dump Configuration**

This dialog allows the user to change the item dump configuration.

_dialog fields_:

| name          | input type                 | description                              |
| ------------- | -------------------------- | ---------------------------------------- |
| file-name     | string                     | file-name, with fmt directives for attrs |
| file-location | string                     | absolute path to file-location           |
| file-type     | drop-down (JSON/YAML/DYNA) | file-type                                |

_dialog operations_:

| name              | hotkey | description                 |
| ----------------- | ------ | --------------------------- |
| apply             | 'a'    | apply and close             |
| apply and execute | 'A'    | apply and execute           |
| clear             | 'C'    | remove custom configuration |

---

**Table Dump Configuration**

This dialog allows the user to change the table dump configuration.

_dialog fields_:

| name          | input type                 | description                    |
| ------------- | -------------------------- | ------------------------------ |
| file-name     | string                     | file-name                      |
| file-location | string                     | absolute path to file-location |
| file-type     | drop-down (JSON/YAML/DYNA) | file-type                      |

_dialog operations_:

| name              | hotkey | description                 |
| ----------------- | ------ | --------------------------- |
| apply             | 'a'    | apply and close             |
| apply and execute | 'A'    | apply and execute           |
| clear             | 'C'    | remove custom configuration |

---

**Scan/Query Filter**

This dialog allows the user to select a filter for a `scan` or `query`
operation.

_dialog fields_:

| name           | input type                                | description          |
| -------------- | ----------------------------------------- | -------------------- |
| attribute-name | drop-down (available attributes)          | attribute name       |
| condition      | drop-down (available conditions on dyndb) | condition            |
| value-type     | drop-down (string/num/bin/bool/null)      | attribute value type |
| value          | string                                    | condition value      |

_dialog operations_:

| name  | hotkey | description       |
| ----- | ------ | ----------------- |
| apply | 'a'    | apply and close   |
| clear | 'C'    | remove any filter |

---

**Transform Attribute**

This dialog allows the user to select a field-name from the available items and
define an available transformation. The only transformation currently available
will be a unix to RFC3339 (or similar) transformation.

_available transformations_:

- unix to RFC3339

_dialog fields_ (repeated):

| name           | input type                            | description                 |
| -------------- | ------------------------------------- | --------------------------- |
| atribute-name  | drop-down (available field names)     | the item attribute-name     |
| transformation | drop-down (available transformations) | the transformation to apply |

_dialog operations_:

| name  | hotkey | description          |
| ----- | ------ | -------------------- |
| apply | 'a'    | apply and close      |
| clear | 'C'    | remove any transform |

---

### Advanced Features (future maybe)

**Automatic Unix Timestamp Conversion**

The TUI automatically detects unix timestamps in the item and converts them to a
human readable format.

**Attribute Projection**

Perform `scan` or `query` with attribute projection.

**Attribute Sorting**

Sort ASC or DESC by a column of choice.
