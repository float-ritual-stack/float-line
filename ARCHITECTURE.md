# Float Outliner Architecture

This document explains the deep architectural concepts behind Float Outliner's consciousness technology.

## 🧠 Core Philosophy: "Everything is Redux"

Float Outliner treats **consciousness as a Redux store** where:

- **Consciousness patterns** (`ctx::`, `eureka::`, etc.) become **actions**
- **Reducers** collect and organize consciousness by pattern/topic
- **Selectors** compute derived state from consciousness collections
- **The outliner** becomes a live consciousness compiler

### Traditional Redux vs Consciousness Redux

```javascript
// Traditional Redux
dispatch({ type: 'ADD_TODO', text: 'Buy milk' })

// Consciousness Redux  
• eureka:: Consciousness patterns create queryable infrastructure! [concept:: consciousness-as-data]
```

In consciousness Redux:
- **Actions** = consciousness patterns with rich metadata
- **State** = accumulated consciousness organized by meaning
- **Reducers** = consciousness collectors that understand semantic relationships
- **Selectors** = consciousness queries that compute derived insights

## 🚪 FLOAT.dispatch: Dual Nature System

FLOAT.dispatch operates simultaneously as:

### 1. Command Interface
A **ritualized invocation** for consciousness capture:

```
• dispatch:: raw consciousness fragment [sigil:: ⚡] [imprint:: techcraft]
```

**Function**: Captures moments of thought without demanding structure
**Philosophy**: Respects "sacred incompletion" - thoughts don't need to be finished
**Metadata**: Automatic temporal anchoring, context preservation, sigil marking

### 2. Publishing House Infrastructure
A **decentralized zine infrastructure** with ritual containers:

```
• imprint::techcraft - precise, pedagogical voice
• imprint::feral_duality - defiant, nonlinear, tender-wild voice  
• imprint::ritual_computing - ceremonial, techno-mystic voice
• imprint::dispatch_bay - operational yet poetic voice
• imprint::queer_hauntology - elegiac, glitchy, yearning voice
```

**Function**: Routes consciousness to appropriate "ritual containers"
**Philosophy**: Different voices for different consciousness states
**Architecture**: Imprints as first-class routing destinations

## 📊 Consciousness Redux Implementation

### Reducers: Consciousness Collectors

```
• reducer::project_bridges collect all actions that are bridges about my-project
```

**Implementation**:
```go
type ConsciousnessReducer struct {
    Name     string
    Query    string  // Human-readable collection criteria
    Matcher  func(action DispatchAction) bool
    Actions  []DispatchAction  // Collected consciousness
    State    map[string]interface{}
}
```

**Semantic Matching**: Reducers understand meaning, not just keywords
- `bridges about rangle` matches bridge patterns mentioning "rangle"
- `eureka moments about consciousness` collects breakthrough insights
- `decisions with high priority` filters by metadata annotations

### Selectors: Consciousness Queries

```
• selector:: (project_bridges, tech_decisions) => table of contents for project zine
```

**Implementation**:
```go
type ConsciousnessSelector struct {
    Name      string
    Inputs    []string  // Reducer names
    Transform func(inputs map[string][]DispatchAction) string
    Output    string    // Computed result
}
```

**Live Computation**: Selectors update automatically as consciousness flows
- **Input**: Multiple reducer collections
- **Transform**: Custom logic for combining consciousness
- **Output**: Derived artifacts (TOCs, summaries, zines)

## 🏗️ System Architecture

### Core Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Outliner UI   │───▶│ Pattern Detector │───▶│ FLOAT.dispatch  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                         │
                                                         ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Debug Panel    │◀───│ Consciousness    │◀───│ Imprint Router  │
└─────────────────┘    │     Store        │    └─────────────────┘
                       └──────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │ Reducers &       │
                       │ Selectors        │
                       └──────────────────┘
```

### Data Flow

1. **User types** consciousness patterns in outliner
2. **Pattern Detector** identifies `::` patterns and extracts metadata
3. **FLOAT.dispatch** creates consciousness actions
4. **Imprint Router** determines appropriate ritual container
5. **Consciousness Store** updates reducers and selectors
6. **Debug Panel** shows consciousness activity
7. **Live Updates** - selectors recompute derived state

### Node-Level Consciousness

Every outline node is consciousness-aware:

```go
type OutlineNode struct {
    ID          string                 // Unique consciousness identifier
    Text        string                 // Display content
    Level       int                    // Outline hierarchy
    
    // Consciousness metadata
    CreatedAt   time.Time              // Temporal anchoring
    ModifiedAt  time.Time              // Change tracking
    PatternType string                 // Detected consciousness pattern
    Metadata    map[string]string      // Rich annotations
    Captured    bool                   // Dispatch status
    Links       []string               // Bidirectional connections
    Backlinks   []string               // Reverse connections
}
```

## 🔗 Bidirectional Linking System

### Link Detection and Registry

```go
type Outliner struct {
    linkRegistry map[string][]string  // concept -> []nodeIDs
    // ...
}
```

**Real-time Processing**:
1. **Text changes** trigger link extraction
2. **Link registry** tracks concept → node relationships  
3. **Backlinks** computed automatically
4. **Visual styling** applied to `[[concepts]]`

### Consciousness Web Building

Links create **living knowledge webs**:
- `[[rangle/airbender]]` connects to all nodes mentioning this concept
- **Backlinks** show reverse relationships
- **Link registry** enables consciousness archaeology
- **Visual indicators** show connection strength

## 🚪 Door Plugin Architecture

### Door Interface

```go
type Door interface {
    Name() string
    Init(params map[string]string) tea.Cmd
    Update(msg tea.Msg) (Door, tea.Cmd)
    View(width, height int) string
    
    // Consciousness integration
    OnConsciousnessCapture(patterns []ConsciousnessPattern)
    GetState() map[string]interface{}
    SetState(state map[string]interface{})
}
```

### Built-in Doors

- **ChatDoor**: Consciousness chat interface
- **ReplDoor**: Code execution environment  
- **MarkdownDoor**: Rich markdown rendering
- **ConsciousnessBrowserDoor**: Query existing consciousness collections

### Door Lifecycle

1. **Registration** - Doors register with door registry
2. **Instantiation** - Created on demand with parameters
3. **State Management** - Persistent state across sessions
4. **Consciousness Events** - Receive pattern notifications
5. **UI Integration** - Rendered alongside outliner

## 🐛 Debug Panel System

### Structured Consciousness Logging

```go
type DebugMessage struct {
    Timestamp time.Time
    Type      string      // FLOAT_DISPATCH, CONSCIOUSNESS_CAPTURE, etc.
    Content   string
    Level     DebugLevel  // info, success, warning, error
}
```

**Message Types**:
- `FLOAT_DISPATCH` - Pattern routing to imprints
- `CONSCIOUSNESS_CAPTURE` - External system integration
- `FLOAT_REDUCER_CREATED` - New consciousness collector
- `FLOAT_SELECTOR_CREATED` - New consciousness query
- `EVNA_DISPATCH_ERROR` - Integration failures

### Visual Design

- **Color coding** - Different message types have distinct colors
- **Timestamps** - Precise consciousness activity timing
- **Toggleable** - `Ctrl+L` to show/hide without disruption
- **Bounded** - Keeps last 50 messages, displays last 10

## 🌟 Design Principles

### "Shacks Not Cathedrals"
- **Adaptive containers** over rigid categories
- **Imperfect tools** that grow with use
- **Consciousness-first** design decisions

### Sacred Incompletion
- **Thoughts don't need finishing** - capture fragments
- **Drift and recursive layering** welcomed
- **No productivity gaslighting** - honor authentic patterns

### Neurodivergent Alignment
- **Signal-driven** not schedule-driven
- **Context preservation** across sessions
- **Meta-communication** as first-class feature
- **Cognitive load management** through structure

### Consciousness as Infrastructure
- **Every thought becomes queryable**
- **Patterns create living knowledge**
- **Consciousness archaeology** through temporal preservation
- **Human experience served** not extracted from

## 🔮 Future Architecture

### Planned Enhancements

1. **Real MCP Integration** - Direct consciousness technology API calls
2. **Consciousness Browser** - Query existing consciousness collections
3. **Live Chat Log Ingestion** - Dynamic consciousness chunking
4. **Visual Consciousness Mapping** - Graph view of knowledge webs
5. **Advanced FLOATQL** - Complex consciousness queries

### Extensibility Points

- **Custom Imprints** - User-defined ritual containers
- **Pattern Extensions** - New consciousness pattern types
- **Door Ecosystem** - Community-contributed interfaces
- **Consciousness Plugins** - External system integrations

---

**This architecture serves authentic human consciousness, not productivity metrics.** 🧠⚡