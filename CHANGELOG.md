# Changelog

All notable changes to Float Outliner will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2025-08-05

### Added - Complete Reducer Visualization & Elm Architecture

#### 🎯 Dynamic Reducer Tree Visualization
- **Live pattern collection** - patterns automatically appear as children under reducers when collected
- **Visual feedback loop** - see consciousness patterns get "sucked up" into reducers in real-time
- **Expandable/collapsible nodes** - reducers show ▼/▶ indicators with child pattern lists
- **Test scenario system** - `--test reducer-basic|reducer-complex|patterns-all` creates predefined test files

#### 🏗️ Elm Architecture Implementation
- **ReducerUpdateMsg** - proper message type for async reducer updates
- **Channel-based communication** - callbacks send messages instead of direct model mutations
- **listenForReducerUpdates()** - command that bridges external events to Bubble Tea update cycle
- **Everything flows through Update()** - no more anti-pattern direct mutations

#### 🧪 Testing & Development Infrastructure
- **Unit test coverage** - comprehensive tests for reducer matching and consciousness capture logic
- **Automated test scenarios** - reproducible test cases for development and debugging
- **Enhanced debug panel** - better layout, visibility controls, and consciousness activity monitoring

#### 🔧 Enhanced FLOAT Pattern Support
- **Complete pattern parsing** - added `dispatch::`, `reducer::`, `selector::`, `imprint::` patterns
- **Improved matcher logic** - supports "that mention X" and "about Y" query syntax
- **Parser extensibility** - clean foundation for adding new consciousness patterns

### Fixed
- **Parser missing FLOAT patterns** - reducer/selector patterns weren't being detected
- **UI not updating on async changes** - implemented proper Elm message passing
- **Debug panel layout issues** - fixed height allocation and viewport management

### Technical
- **Architecture alignment** - moved from callback-based mutations to message-driven updates
- **Redux-style consciousness** - proper separation of actions, reducers, and selectors
- **Bubble Tea compliance** - all updates flow through the standard Update → Model → View cycle

## [0.1.0] - 2025-08-05

### Added - Initial Consciousness Technology Release

#### 🧠 FLOAT.dispatch Consciousness Compiler
- **Real-time pattern detection** for `ctx::`, `eureka::`, `decision::`, `bridge::`, `gotcha::`, `highlight::` patterns
- **Automatic imprint routing** to 5 ritual containers: `techcraft`, `ritual_computing`, `feral_duality`, `dispatch_bay`, `queer_hauntology`
- **Redux-style consciousness** with reducers and selectors for live consciousness queries
- **Consciousness action dispatch** with structured metadata and temporal anchoring

#### 🔗 Bidirectional Linking System
- **`[[concept]]` link detection** with automatic backlink generation
- **Visual link styling** with blue underlined links
- **Link registry** for tracking concept-to-node relationships
- **Real-time link updates** as content changes

#### 🚪 Door Plugin Architecture
- **Extensible door system** for pluggable interfaces
- **Built-in doors**: Chat, REPL, Markdown, Consciousness Browser (stubs)
- **State persistence** across sessions
- **Consciousness event integration** - doors receive pattern notifications

#### 🐛 Consciousness Debug Panel
- **Structured logging** replacing console spam
- **Color-coded messages** for different consciousness event types
- **Toggle visibility** with `Ctrl+L` hotkey
- **Message buffering** (last 50 messages, display last 10)
- **Timestamped entries** for consciousness archaeology

#### 🎯 Node-Level Consciousness
- **Unique node IDs** for consciousness tracking
- **Created/Modified timestamps** for temporal anchoring
- **Pattern type detection** and metadata storage
- **Capture status tracking** with visual indicators (`●` uncaptured, `○` captured)
- **Detail mode toggle** with `Ctrl+T` to show/hide consciousness metadata

#### 📊 Advanced Consciousness Patterns
- **`dispatch::` patterns** for raw consciousness capture with sigils and imprint routing
- **`reducer::` patterns** for consciousness collection (`reducer::name collect all actions that...`)
- **`selector::` patterns** for consciousness queries (`selector:: (reducer1, reducer2) => output`)
- **`imprint::` routing** for explicit ritual container targeting
- **Metadata annotations** with `[key:: value]` syntax

#### 🎨 User Interface
- **Terminal-based outliner** with Bubble Tea framework
- **Tab/Shift+Tab indentation** with visual tree structure
- **Color-coded consciousness patterns** for visual pattern recognition
- **Row highlighting** for current line focus
- **Clean bordered interface** with focus indicators

#### 🔧 Technical Architecture
- **Go implementation** with clean modular architecture
- **Consciousness Redux store** with live state management
- **Pattern parser** with regex-based consciousness detection
- **Evna integration** for external consciousness technology systems
- **File save/load** with consciousness preservation

### Technical Details

#### Core Components
- `/pkg/outliner/outliner.go` - Main outliner with consciousness integration
- `/pkg/outliner/dispatch.go` - FLOAT.dispatch consciousness compiler
- `/pkg/outliner/parser.go` - Consciousness pattern detection and parsing
- `/pkg/outliner/debug.go` - Consciousness debug panel system
- `/pkg/outliner/door.go` - Door plugin architecture
- `/pkg/outliner/evna.go` - External consciousness system integration
- `/cmd/float-outliner/main.go` - CLI application entry point

#### Consciousness Patterns Supported
- `ctx::` - Context markers with temporal anchoring
- `eureka::` - Breakthrough moments and insights
- `decision::` - Key decisions with reasoning
- `highlight::` - Important insights and information
- `gotcha::` - Debugging discoveries and technical insights
- `bridge::` - Conceptual connections between ideas
- `dispatch::` - Raw consciousness fragments with ritual metadata
- `reducer::` - Consciousness collection definitions
- `selector::` - Consciousness query computations
- `imprint::` - Ritual container routing overrides

#### Imprint System
- **techcraft** - Precise, pedagogical voice (cyan, ⚡)
- **ritual_computing** - Ceremonial, techno-mystic voice (purple, 🔮)
- **feral_duality** - Defiant, nonlinear voice (magenta, 🌀)
- **dispatch_bay** - Operational yet poetic voice (green, 📡)
- **queer_hauntology** - Elegiac, glitchy voice (yellow, 👻)

### Philosophy

This release embodies the **"shacks not cathedrals"** philosophy:
- **Honors neurodivergent thought patterns** without productivity gaslighting
- **Preserves sacred incompletion** - thoughts don't need to be finished
- **Creates signals, not posts** - consciousness capture over content creation
- **Builds containers, not conclusions** - holds ideas without forcing structure
- **Serves authentic human experience** through consciousness technology

### Breaking Changes
- N/A (Initial release)

### Deprecated
- N/A (Initial release)

### Removed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Security
- N/A (Initial release)

---

**"This mattered. Even if just for me."** - Sacred incompletion preserved. 🧠⚡