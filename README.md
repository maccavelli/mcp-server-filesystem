> **Mirror notice:** This repository is a one-way published export of a
> privately hosted project. History is squashed into sync snapshots, and pull
> requests cannot be merged here directly — open an issue instead. Changes
> land in the private source and are re-exported.

<!-- markdownlint-disable MD013 MD060 MD033 -->

# MagicFilesystem MCP Sub-Server

A high-performance Model Context Protocol (MCP) sub-server providing secure, sandboxed filesystem operations.

## 🏗️ Core Pillars: Why, What, and How

### 1. Why the Server Exists
AI agents running in IDEs require direct access to the filesystem to read, edit, search, and manage project files. However, granting unrestricted shell access poses substantial security risks (e.g., accidental deletions, data extraction, or running malicious scripts). `mcp-server-filesystem` exists to provide a secure, strict, and sandboxed interface that isolates file modifications to explicit, user-approved directories while preventing directory traversal attacks.

### 2. What the Server Does
The filesystem server exposes standard filesystem primitives bounded by an allowlist sandbox:
- **Precision Reading/Writing**: Reads text files, streams binary/media assets, and writes/overwrites files.
- **Directory Traversal Safeguards**: Lists directory contents, filters files by glob pattern, and builds structured directory trees.
- **Surgical Edits**: Edits target ranges of files with validation and patch diff output.
- **Diagnostics**: Checks file metadata, computes secure checksums (SHA-256), and retrieves internal logs.

### 3. How it Does What it Does
- **Strict Allowlisting Sandbox**: Validates all paths against a set of root directories provided via CLI arguments or MCP roots.
- **Path Cleaning & Symlink Resolving**: Uses `filepath.Clean` and `filepath.EvalSymlinks` to normalize and evaluate every target path. Any path resolving outside the defined roots throws a strict traversal error.
- **Go Standard Library & Concurrency**: Leverages standard Go `os` routines and maps tool execution safely over standard I/O channels.

---

## Quick Start

### Step 1: Place the Binary

Download the `mcp-server-filesystem` binary for your platform and place it in a directory on your system `PATH`.

#### Linux

```bash
# Move the binary to your local bin directory
mv mcp-server-filesystem ~/.local/bin/mcp-server-filesystem
chmod +x ~/.local/bin/mcp-server-filesystem
```

#### macOS

```bash
# Move the binary to your local bin directory
mv mcp-server-filesystem /usr/local/bin/mcp-server-filesystem
chmod +x /usr/local/bin/mcp-server-filesystem
```

#### Windows (PowerShell)

```powershell
# Create a directory for the binary if it doesn't exist
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\Programs\filesystem"

# Move the binary
Move-Item mcp-server-filesystem.exe "$env:LOCALAPPDATA\Programs\filesystem\mcp-server-filesystem.exe"

# Add to your PATH (current user, persistent)
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
[Environment]::SetEnvironmentVariable("Path", "$currentPath;$env:LOCALAPPDATA\Programs\filesystem", "User")
```

---

### Step 2: Initialize Configuration

`mcp-server-filesystem` is a stateless sub-server. It does not require local configuration files or API tokens. However, **you must pass allowed directories as arguments** when running the server to define the sandbox.

---

### Step 3: Configure Your IDE

> **⚠️ IMPORTANT ORCHESTRATOR MESSAGING**
>
> While the standalone IDE configurations below are provided for testing and debugging, `mcp-server-filesystem` is designed to be run as a downstream node behind the **`magictools` orchestrator** in production environments.
>
> When running in production, you should **only** configure `magictools` in your IDE, which will automatically proxy requests to `filesystem` as needed.

If you are testing the server standalone, configure your IDE to launch the binary directly. **Make sure to include the directories you want to allow access to in the `args` array.**

#### Antigravity (Google DeepMind)

| OS | Configuration File Path |
|---|---|
| Linux | `~/.gemini/antigravity/mcp_config.json` |
| macOS | `~/.gemini/antigravity/mcp_config.json` |
| Windows | `%USERPROFILE%\.gemini\antigravity\mcp_config.json` |

**Linux / macOS:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/home/youruser/.local/bin/mcp-server-filesystem",
      "args": [
        "/home/youruser/projects",
        "/home/youruser/downloads"
      ],
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\filesystem\\mcp-server-filesystem.exe",
      "args": [
        "C:\\Users\\YourUser\\Projects",
        "C:\\Users\\YourUser\\Downloads"
      ],
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### Visual Studio Code (GitHub Copilot / Native MCP)

| OS | User-Level Configuration File Path |
|---|---|
| Linux | `~/.config/Code/User/mcp.json` |
| macOS | `~/Library/Application Support/Code/User/mcp.json` |
| Windows | `%APPDATA%\Code\User\mcp.json` |

**Linux:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/home/youruser/.local/bin/mcp-server-filesystem",
      "args": [
        "/home/youruser/projects"
      ],
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**macOS:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/usr/local/bin/mcp-server-filesystem",
      "args": [
        "/Users/youruser/projects"
      ],
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\filesystem\\mcp-server-filesystem.exe",
      "args": [
        "C:\\Users\\YourUser\\Projects"
      ],
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### VSCode — Cline Extension

| OS | Configuration File Path |
|---|---|
| Linux | `~/.cline/data/settings/cline_mcp_settings.json` |
| macOS | `~/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json` |
| Windows | `%APPDATA%\Code\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json` |

**Linux:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/home/youruser/.local/bin/mcp-server-filesystem",
      "args": [
        "/home/youruser/projects"
      ],
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**macOS:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/usr/local/bin/mcp-server-filesystem",
      "args": [
        "/Users/youruser/projects"
      ],
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\filesystem\\mcp-server-filesystem.exe",
      "args": [
        "C:\\Users\\YourUser\\Projects"
      ],
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### Claude Desktop

| OS | Configuration File Path |
|---|---|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |

**macOS:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/usr/local/bin/mcp-server-filesystem",
      "args": [
        "/Users/youruser/projects"
      ],
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\filesystem\\mcp-server-filesystem.exe",
      "args": [
        "C:\\Users\\YourUser\\Projects"
      ],
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### Claude Code (CLI)

Claude Code uses a CLI command to register MCP servers. Ensure you pass the allowed directories at the end.

**Linux:**

```bash
claude mcp add filesystem -s user -- /home/youruser/.local/bin/mcp-server-filesystem /home/youruser/projects
```

**macOS:**

```bash
claude mcp add filesystem -s user -- /usr/local/bin/mcp-server-filesystem /Users/youruser/projects
```

**Windows (PowerShell):**

```powershell
claude mcp add filesystem -s user -- "C:\Users\YourUser\AppData\Local\Programs\filesystem\mcp-server-filesystem.exe" "C:\Users\YourUser\Projects"
```

#### Cursor

| OS | Global Configuration File Path |
|---|---|
| Linux | `~/.cursor/mcp.json` |
| macOS | `~/.cursor/mcp.json` |
| Windows | `%USERPROFILE%\.cursor\mcp.json` |

**Linux:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/home/youruser/.local/bin/mcp-server-filesystem",
      "args": [
        "/home/youruser/projects"
      ],
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**macOS:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/usr/local/bin/mcp-server-filesystem",
      "args": [
        "/Users/youruser/projects"
      ],
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\filesystem\\mcp-server-filesystem.exe",
      "args": [
        "C:\\Users\\YourUser\\Projects"
      ],
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### JetBrains IDEs (IntelliJ, GoLand, WebStorm, PyCharm)

JetBrains IDEs configure MCP servers via the AI Assistant settings or a local configuration file.

| OS | Configuration File Path |
|---|---|
| Linux | `~/.config/JetBrains/AI/mcp.json` (or via UI: Settings > Tools > AI Assistant > MCP Servers) |
| macOS | `~/Library/Application Support/JetBrains/AI/mcp.json` (or via UI: Settings > Tools > AI Assistant > MCP Servers) |
| Windows | `%APPDATA%\JetBrains\AI\mcp.json` (or via UI: Settings > Tools > AI Assistant > MCP Servers) |

**Linux:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/home/youruser/.local/bin/mcp-server-filesystem",
      "args": [
        "/home/youruser/projects"
      ],
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**macOS:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/usr/local/bin/mcp-server-filesystem",
      "args": [
        "/Users/youruser/projects"
      ],
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\filesystem\\mcp-server-filesystem.exe",
      "args": [
        "C:\\Users\\YourUser\\Projects"
      ],
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

---

## ⚙️ Configuration File (`config.yaml`)

On startup, `mcp-server-filesystem` automatically initializes a configuration folder and an empty configuration template at `~/.config/mcp-server-filesystem/config.yaml`. 

This file is currently kept as a template for user extension and does not enforce required keys. The primary sandbox boundaries are defined by CLI arguments or IDE roots instead.

---

## 🛠️ MCP Tools & Resources Reference

Once the server is running, the following tools and resources are exposed:

### Tools

| Tool | Parameters | Description |
|---|---|---|
| `read_file` | `path` (string) | Reads the text content of a file within the sandbox. Capped at 50MB. |
| `read_media_file` | `path` (string) | Reads a binary file (images, PDFs, audio/video) and returns base64 data. |
| `read_multiple_files` | `paths` (array of strings) | Concurrently reads multiple files in a single batch. |
| `write_to_file` | `path` (string), `content` (string) | Creates a new file or overwrites an existing file with new content. |
| `edit_file` | `path` (string), `edits` (array of edit operations) | Applies targeted chunked line replacements to a file. |
| `create_directory` | `path` (string) | Creates a directory and any necessary parent directories. |
| `list_directory` | `path` (string) | Lists files and subdirectories immediately inside a target folder. |
| `list_directory_with_sizes`| `path` (string) | Lists items inside a folder along with their sizes in bytes. |
| `directory_tree` | `path` (string) | Recursively prints a tree view of a folder up to a maximum depth of 20. |
| `move_file` | `source` (string), `destination` (string) | Safely renames or moves a file/directory to a new sandboxed path. |
| `search_files` | `path` (string), `query` (string) | Performs glob searches for filenames or text patterns within files. |
| `get_file_info` | `path` (string) | Retrieves file statistics (size, permissions, mod time, dir status). |
| `list_allowed_directories`| None | Returns the list of directories currently allowlisted in the sandbox. |
| `copy_path` | `source` (string), `destination` (string) | Copies a file or folder to a destination path inside the sandbox. |
| `remove_path` | `path` (string) | Deletes a file or directory within the allowed sandbox. |
| `append_file` | `path` (string), `content` (string) | Appends content to the end of a text file. |
| `get_file_hash` | `path` (string) | Computes the SHA-256 checksum of a file. |
| `get_internal_logs` | None | **[SERVER: filesystem]** Returns the tail of the execution log ring buffer. |

### Resources

| Resource URI | Description |
|---|---|
| `filesystem://logs` | Exposes the tail of the internal server diagnostics logs. |

---

## 📋 Data Storage Locations

| Data | Linux | macOS | Windows |
|---|---|---|---|
| **Configuration** | `~/.config/mcp-server-filesystem/config.yaml` | `~/Library/Application Support/mcp-server-filesystem/config.yaml` | `%APPDATA%\mcp-server-filesystem\config.yaml` |
| **Server Logs** | `stderr` (captured by IDE) | `stderr` (captured by IDE) | `stderr` (captured by IDE) |

---

*Built with ❤️. Part of the MagicTools Intelligence Suite.*
