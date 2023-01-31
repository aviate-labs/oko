# Commands

## `init`

Initializes the Oko package file.

### Options

|name|value|
|---|---|
|**compiler**|*compiler version*|

## `download`

Downloads all packages specified in the Oko package file.

## `install`

Allows you to install packages from GitHub or link local directories.

### Sub Commands

#### `github`

Allows you to install packages from GitHub.

Expects `{org}/{repo}`, i.e. if you want to install the package at https://github.com/internet-computer/testing.mo you will have to pass `internet-computer/testing.mo` to the first argument.

##### Arguments

1. url
2. version

##### Options

|name|value|
|---|---|
|**name**|*package name*|

#### `local`

Allows you to link local packages as dependencies.

##### Arguments

1. path

##### Options

|name|value|
|---|---|
|**name**|*package name*|

## `remove`

Allows you to remove packages by name.

### Arguments

1. name

## `migrate`

Allows you to migrate Vessel config files to Oko.

### Options

|name|value|
|---|---|
|**delete**||
|**keep**||

## `sources`

prints moc package sources

## `bin`

Motoko compiler stuff

### Sub Commands

#### `download`

downloads the Motoko compiler

#### `show`

prints out the path to the bin dir
