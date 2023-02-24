# Commands

## `init`

Initializes the Oko package file.

```shell
oko init
```

### Options

|name|value|
|---|---|
|**compiler**|*compiler version*|

## `download`

Downloads all packages specified in the Oko package file.

Name aliases: `d`

```shell
oko download
```

## `install`

Allows you to install packages from GitHub or link local directories.

Name aliases: `i`

### Sub Commands

#### `github`

Allows you to install packages from GitHub.

Expects `{org}/{repo}`, i.e. if you want to install the package at https://github.com/internet-computer/testing.mo you will have to pass `internet-computer/testing.mo` to the first argument.

Instead of specifying a specific version, `latest` can be used.

Name aliases: `gh`

```shell
oko install github <url> <version>
```

##### Arguments

1. url
2. version

##### Options

|name|value|
|---|---|
|**name**|*package name*|

#### `local`

Allows you to link local packages as dependencies.

Name aliases: `l`

```shell
oko install local <path>
```

##### Arguments

1. path

##### Options

|name|value|
|---|---|
|**name**|*package name*|

## `remove`

Allows you to remove packages by name.

Name aliases: `r`

```shell
oko remove <name>
```

### Arguments

1. name

## `migrate`

Allows you to migrate Vessel config files to Oko.

```shell
oko migrate
```

### Options

|name|value|
|---|---|
|**delete**||
|**keep**||

## `sources`

prints moc package sources

```shell
oko sources
```

## `bin`

Motoko compiler stuff

Name aliases: `b`

### Sub Commands

#### `download`

downloads the Motoko compiler

Name aliases: `d`

```shell
oko bin download
```

##### Options

|name|value|
|---|---|
|**didc**||

#### `show`

prints out the path to the bin dir

Name aliases: `s`

```shell
oko bin show
```
