# prompter

A small Go library for interactive CLI prompts — secured passwords, text input, and selection menus.

## Install

```bash
go get github.com/olekukonko/prompter
```

## Usage

### Secret (hidden input)

```go
s := prompter.NewSecret("Password")
r, err := s.Run()
if err != nil {
    log.Fatal(err)
}
fmt.Println(r.String())
r.Zero() // wipe from memory when done
```

With confirmation and validation:

```go
r, err := prompter.NewSecret("Password",
    prompter.WithRequired(true),
    prompter.WithLength(8, 128),
    prompter.WithValidator(func(b []byte) error {
        if !bytes.Contains(b, []byte("!")) {
            return errors.New("must contain !")
        }
        return nil
    }),
).WithConfirmation("Confirm password").Run()
```

### Text input

```go
r, err := prompter.NewInput("Username",
    prompter.WithRequired(true),
    prompter.WithLength(3, 32),
).Run()
```

> **Note:** `Input` does not zero memory. Do not use it for secrets.

### Selection menu

```go
idx, err := prompter.Select("Choose env", []string{"dev", "staging", "prod"})

// or get the value directly
val, err := prompter.SelectValue("Choose env", []string{"dev", "staging", "prod"})
```

### Confirm

```go
ok, err := prompter.Confirm("Continue?")
```

### Context / timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

r, err := s.RunContext(ctx)
```

## Options

| Option | Description |
|---|---|
| `WithRequired(bool)` | Fail on empty input |
| `WithLength(min, max int)` | Min/max byte length (0 = no limit) |
| `WithMaxRetries(n int)` | Max attempts before error (0 = unlimited) |
| `WithValidator(fn)` | Custom validation function |
| `WithFormatter(fn)` | Custom prompt formatting |
| `WithErrorCallback(fn)` | Called on each validation failure |
| `WithInput(r io.Reader)` | Override stdin (useful for testing) |

## Security

- Secret input uses `term.ReadPassword` on terminals (no echo)
- Secret bytes are zeroed from memory after use — call `Result.Zero()` when done
- Input is capped at 4096 bytes to prevent memory exhaustion
- Prompt strings are sanitized to strip ANSI escape sequences
- Context cancellation zeros any buffered secret before discarding it
- `Input` does **not** zero memory; do not use it for passwords