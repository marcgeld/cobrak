# Configuration System

cobrak supports configuration through a `~/.cobrak/settings.toml` file. This allows you to set default values that apply to all commands.

## Configuration File

The configuration file is located at: `~/.cobrak/settings.toml`

### Default Values

When you first use cobrak, default settings are:
```toml
output = "text"
namespace = ""
context = ""
top = 20
```

## Configuration Options

### `output`
- **Type**: string
- **Allowed values**: `text`, `json`, `yaml`
- **Default**: `text`
- **Description**: Default output format for commands

### `namespace`
- **Type**: string
- **Default**: `""` (empty = all namespaces)
- **Description**: Default namespace to inspect

### `context`
- **Type**: string
- **Default**: `""` (empty = current context)
- **Description**: Default Kubernetes context to use

### `top`
- **Type**: integer
- **Default**: `20`
- **Description**: Default number of top offenders to show

## Managing Configuration

### View Current Configuration

```bash
./cobrak config show
```

Output:
```
Configuration file: /Users/marcus/.cobrak/settings.toml

output:    text (text, json, yaml)
namespace:  (empty = all namespaces)
context:   (empty = current context)
top:       20

Note: Command-line flags override these settings
```

### Set a Configuration Value

```bash
# Set default output to JSON
./cobrak config set output json

# Set default namespace
./cobrak config set namespace kube-system

# Set default context
./cobrak config set context my-cluster

# Set default top value
./cobrak config set top 50
```

### Reset to Defaults

```bash
./cobrak config reset
```

## Usage Examples

### Example 1: Change Default Output to JSON

```bash
# Set JSON as default
./cobrak config set output json

# Now all commands use JSON by default
./cobrak resources          # Uses JSON (from config)
./cobrak resources --output=text   # Override with text
```

### Example 2: Default to Production Namespace

```bash
# Set production as default namespace
./cobrak config set namespace production

# Monitor production cluster by default
./cobrak resources          # Shows production resources
./cobrak nodeinfo          # Shows production nodes
./cobrak resources --namespace=staging  # Override for staging
```

### Example 3: Configuration File Example

```toml
# ~/.cobrak/settings.toml
output = "json"
namespace = "production"
context = "production-cluster"
top = 50
```

With this configuration:
- All commands output JSON by default
- All commands inspect the production namespace
- Commands use the production-cluster context
- Show top 50 results

## Flag Override Precedence

Command-line flags always take precedence over configuration file settings:

```bash
# Configuration has output=json, namespace=kube-system

# Uses JSON (from config)
./cobrak resources

# Uses text (flag overrides config)
./cobrak resources --output=text

# Uses default namespace (from config)
./cobrak resources

# Uses production namespace (flag overrides config)
./cobrak resources --namespace=production
```

## Complete Example Workflow

```bash
# 1. View current configuration
./cobrak config show

# 2. Set preferred output format
./cobrak config set output yaml

# 3. Set default namespace for work
./cobrak config set namespace monitoring

# 4. Set default top value
./cobrak config set top 30

# 5. Verify configuration
./cobrak config show

# 6. Use cobrak with these defaults
./cobrak resources          # YAML output, monitoring namespace, top 30
./cobrak nodeinfo          # Uses same defaults
./cobrak capacity          # Uses same defaults

# 7. Override a setting when needed
./cobrak resources --output=json --namespace=kube-system

# 8. Reset to defaults if needed
./cobrak config reset
```

## Configuration File Format

The configuration file uses TOML format. You can edit it manually if needed:

```bash
# View raw config file
cat ~/.cobrak/settings.toml

# Edit manually
vim ~/.cobrak/settings.toml
```

TOML syntax for cobrak settings:
```toml
output = "json"        # String values use quotes
namespace = ""         # Empty string is allowed
context = ""          # Empty string is allowed
top = 50              # Numbers without quotes
```

## Environment Variables

Note: Environment variables are NOT directly supported for config values. Use the `cobrak config set` command or edit `~/.cobrak/settings.toml` directly.

However, `KUBECONFIG` and context are still handled via environment variables and command-line flags.

## Troubleshooting

### Config file not found

If the config file doesn't exist, cobrak will use default values automatically. No error is raised.

```bash
# Create initial config
./cobrak config show

# Set a value (creates file if needed)
./cobrak config set output json
```

### Invalid configuration value

```bash
# Error: invalid top value (must be number)
./cobrak config set top invalid

# Correct:
./cobrak config set top 50
```

### Permission denied

```bash
# Fix permissions on config directory
chmod 755 ~/.cobrak

# Fix permissions on config file
chmod 644 ~/.cobrak/settings.toml
```

## Command Reference

```bash
# Show configuration
cobrak config show

# Set a value
cobrak config set <key> <value>

# Reset to defaults
cobrak config reset

# Get help
cobrak config --help
cobrak config set --help
cobrak config show --help
cobrak config reset --help
```

## Implementation Details

- Configuration is stored in TOML format
- File location: `~/.cobrak/settings.toml`
- Directory created automatically if needed
- File permissions: 0644
- Default directory permissions: 0755
- Command-line flags always override config file settings
- Invalid config files will cause errors - they must be valid TOML

