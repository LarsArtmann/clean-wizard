# Shell Aliases

Convenience aliases for faster and easier clean-wizard usage.

## 📚 Table of Contents

- [Bash/Zsh](#bashzsh)
- [Fish](#fish)
- [PowerShell](#powershell)
- [Custom Aliases](#custom-aliases)

---

## 🐧 Bash/Zsh

Add to `~/.bashrc` (Bash) or `~/.zshrc` (Zsh):

```bash
# Clean-wizard aliases
alias cw='clean-wizard clean --mode standard'
alias cw-dry='clean-wizard clean --mode standard --dry-run'
alias cw-quick='clean-wizard clean --mode quick'
alias cw-aggressive='clean-wizard clean --mode aggressive'
alias cw-interactive='clean-wizard clean'

# Cleaner-specific aliases
alias cw-go='clean-wizard clean --cleaners go --mode standard'
alias cw-node='clean-wizard clean --cleaners node --mode standard'
alias cw-docker='clean-wizard clean --cleaners docker --mode aggressive'
alias cw-nix='clean-wizard clean --cleaners nix --mode aggressive'
alias cw-cargo='clean-wizard clean --cleaners cargo --mode standard'
alias cw-build='clean-wizard clean --cleaners buildcache --mode standard'
alias cw-temp='clean-wizard clean --cleaners temp --mode aggressive'
alias cw-system='clean-wizard clean --cleaners systemcache --mode standard'

# Developer aliases
alias cw-dev='clean-wizard clean --cleaners go,node,cargo,buildcache,temp --mode standard'
alias cw-full='clean-wizard clean --mode aggressive'
```

**Apply changes:**
```bash
source ~/.bashrc  # or: source ~/.zshrc
```

**Usage:**
```bash
cw              # Quick cleanup (standard mode)
cw-dry          # Preview cleanup
cw-quick         # Quick mode (minimal cleanup)
cw-aggressive     # Aggressive mode (maximum cleanup)
cw-go            # Clean Go packages only
cw-docker        # Clean Docker aggressively
cw-dev           # Developer cleanup (Go, Node, Cargo, Build, Temp)
```

---

## 🐟 Fish

Add to `~/.config/fish/config.fish`:

```fish
# Clean-wizard aliases
alias cw="clean-wizard clean --mode standard"
alias cw-dry="clean-wizard clean --mode standard --dry-run"
alias cw-quick="clean-wizard clean --mode quick"
alias cw-aggressive="clean-wizard clean --mode aggressive"
alias cw-interactive="clean-wizard clean"

# Cleaner-specific aliases
alias cw-go="clean-wizard clean --cleaners go --mode standard"
alias cw-node="clean-wizard clean --cleaners node --mode standard"
alias cw-docker="clean-wizard clean --cleaners docker --mode aggressive"
alias cw-nix="clean-wizard clean --cleaners nix --mode aggressive"
alias cw-cargo="clean-wizard clean --cleaners cargo --mode standard"
alias cw-build="clean-wizard clean --cleaners buildcache --mode standard"
alias cw-temp="clean-wizard clean --cleaners temp --mode aggressive"
alias cw-system="clean-wizard clean --cleaners systemcache --mode standard"

# Developer aliases
alias cw-dev="clean-wizard clean --cleaners go,node,cargo,buildcache,temp --mode standard"
alias cw-full="clean-wizard clean --mode aggressive"
```

**Apply changes:**
```fish
source ~/.config/fish/config.fish
```

**Usage:**
```fish
cw              # Quick cleanup (standard mode)
cw-dry          # Preview cleanup
cw-quick         # Quick mode (minimal cleanup)
cw-aggressive     # Aggressive mode (maximum cleanup)
cw-go            # Clean Go packages only
cw-docker        # Clean Docker aggressively
cw-dev           # Developer cleanup (Go, Node, Cargo, Build, Temp)
```

---

## 💻 PowerShell

Add to PowerShell profile (`$PROFILE`):

```powershell
# Clean-wizard aliases
function cw { clean-wizard clean --mode standard }
function cw-dry { clean-wizard clean --mode standard --dry-run }
function cw-quick { clean-wizard clean --mode quick }
function cw-aggressive { clean-wizard clean --mode aggressive }
function cw-interactive { clean-wizard clean }

# Cleaner-specific aliases
function cw-go { clean-wizard clean --cleaners go --mode standard }
function cw-node { clean-wizard clean --cleaners node --mode standard }
function cw-docker { clean-wizard clean --cleaners docker --mode aggressive }
function cw-nix { clean-wizard clean --cleaners nix --mode aggressive }
function cw-cargo { clean-wizard clean --cleaners cargo --mode standard }
function cw-build { clean-wizard clean --cleaners buildcache --mode standard }
function cw-temp { clean-wizard clean --cleaners temp --mode aggressive }
function cw-system { clean-wizard clean --cleaners systemcache --mode standard }

# Developer aliases
function cw-dev { clean-wizard clean --cleaners go,node,cargo,buildcache,temp --mode standard }
function cw-full { clean-wizard clean --mode aggressive }
```

**Find PowerShell profile:**
```powershell
$PROFILE
```

**Edit profile:**
```powershell
notepad $PROFILE
```

**Apply changes:**
```powershell
. $PROFILE
```

**Usage:**
```powershell
cw              # Quick cleanup (standard mode)
cw-dry          # Preview cleanup
cw-quick         # Quick mode (minimal cleanup)
cw-aggressive     # Aggressive mode (maximum cleanup)
cw-go            # Clean Go packages only
cw-docker        # Clean Docker aggressively
cw-dev           # Developer cleanup (Go, Node, Cargo, Build, Temp)
```

---

## ✏️ Custom Aliases

Create your own custom aliases for your specific workflow:

### Example 1: Daily cleanup before work
```bash
alias cw-workday='clean-wizard clean --cleaners go,node,docker --mode standard'
```

### Example 2: Weekly aggressive cleanup
```bash
alias cw-weekly='clean-wizard clean --mode aggressive'
```

### Example 3: Docker-focused cleanup
```bash
alias cw-docker-only='clean-wizard clean --cleaners docker --mode aggressive'
```

### Example 4: Preview then clean
```bash
alias cw-scan='clean-wizard clean --dry-run --mode standard && clean-wizard clean --mode standard'
```

---

## 🎯 Recommended Aliases

**Power users (use daily):**
```bash
# Quick daily cleanup
alias cw='clean-wizard clean --mode standard'

# Preview before cleanup
alias cw-dry='clean-wizard clean --dry-run --mode standard'
```

**Developers (use frequently):**
```bash
# Developer cleanup
alias cw-dev='clean-wizard clean --cleaners go,node,cargo,buildcache --mode standard'

# Docker cleanup
alias cw-docker='clean-wizard clean --cleaners docker --mode aggressive'
```

**Cleanup on schedule (use cron/launchd):**
```bash
# Full aggressive cleanup
alias cw-full='clean-wizard clean --mode aggressive'

# Quick cleanup
alias cw-quick='clean-wizard clean --mode quick'
```

---

## 🔍 Alias Reference

| Alias | Command | Purpose |
|--------|----------|---------|
| `cw` | `clean-wizard clean --mode standard` | Standard cleanup (recommended) |
| `cw-dry` | `clean-wizard clean --dry-run --mode standard` | Preview cleanup |
| `cw-quick` | `clean-wizard clean --mode quick` | Quick mode (minimal) |
| `cw-aggressive` | `clean-wizard clean --mode aggressive` | Aggressive mode (maximum) |
| `cw-go` | `clean-wizard clean --cleaners go --mode standard` | Go packages only |
| `cw-node` | `clean-wizard clean --cleaners node --mode standard` | Node.js packages only |
| `cw-docker` | `clean-wizard clean --cleaners docker --mode aggressive` | Docker only |
| `cw-nix` | `clean-wizard clean --cleaners nix --mode aggressive` | Nix only |
| `cw-cargo` | `clean-wizard clean --cleaners cargo --mode standard` | Cargo packages only |
| `cw-build` | `clean-wizard clean --cleaners buildcache --mode standard` | Build cache only |
| `cw-temp` | `clean-wizard clean --cleaners temp --mode aggressive` | Temp files only |
| `cw-system` | `clean-wizard clean --cleaners systemcache --mode standard` | System cache only |
| `cw-dev` | `clean-wizard clean --cleaners go,node,cargo,buildcache,temp --mode standard` | Developer cleanup |
| `cw-full` | `clean-wizard clean --mode aggressive` | Full aggressive cleanup |

---

## 📚 Additional Resources

- [Usage Guide](../HOW_TO_USE.md)
- [Command Reference](../USAGE.md)
- [Implementation Status](../IMPLEMENTATION_STATUS.md)
- [Troubleshooting](../TROUBLESHOOTING.md) (if available)

---

**Happy cleaning! 🧹✨**
