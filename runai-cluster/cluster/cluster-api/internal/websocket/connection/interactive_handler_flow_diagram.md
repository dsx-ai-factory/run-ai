# InteractiveHandler Flow Diagram

## Overview
The InteractiveHandler manages WebSocket connections for interactive terminal sessions, primarily handling terminal size updates and binary data forwarding.

## Main Flow Diagram

```mermaid
flowchart TD
    A[WebSocket Message Received] --> B{Message Type?}
    
    B -->|Binary| C[Copy to Buffer]
    B -->|Text| D[Parse JSON Message]
    B -->|Other| E[Return Error]
    
    C --> F[Return Message Length]
    E --> F
    
    D --> G{Message Type?}
    G -->|terminal_size| H[Handle Terminal Size]
    G -->|Other| I[Return Unsupported Error]
    
    H --> J{Width and Height > 0?}
    J -->|Yes| K[Update Terminal Size Queue]
    J -->|No| L[Return Validation Error]
    
    K --> M[Send Success Response]
    L --> N[Send Error Response]
    I --> N
    
    M --> F
    N --> F
    
    %% Next Flow
    O[Next Called] --> P[Get Next Terminal Size]
    P --> Q[Send Resize Update]
    Q --> R[Return Terminal Size]
    
    %% Close Flow
    S[Close Called] --> T[Stop Terminal Size Queue]
    T --> U[Return Success]
    
    %% Error Handling
    V[handleError] --> W[Create Error Response]
    W --> X[Marshal JSON]
    X --> Y[Send via WebSocket]
    
    %% Styling
    classDef success fill:#d4edda,stroke:#155724,stroke-width:2px
    classDef error fill:#f8d7da,stroke:#721c24,stroke-width:2px
    classDef process fill:#d1ecf1,stroke:#0c5460,stroke-width:2px
    classDef decision fill:#fff3cd,stroke:#856404,stroke-width:2px
    
    class C,F,M,R,U success
    class E,I,L,N error
    class D,H,K,P,Q,T,W,X,Y process
    class B,G,J decision
```

## Detailed Process Flow

```mermaid
sequenceDiagram
    participant WS as WebSocket
    participant IH as InteractiveHandler
    participant TSQ as TerminalSizeQueue
    participant JSON as JSON Parser
    
    Note over WS,JSON: Process() Flow
    WS->>IH: Process(msgType, msgRead, msgBuf)
    
    alt Binary Message
        IH->>IH: Copy msgRead to msgBuf
        IH-->>WS: Return message length
    else Text Message
        IH->>JSON: Unmarshal(msgRead)
        JSON-->>IH: MessageRequest
        
        alt Message Type == "terminal_size"
            IH->>IH: Extract width & height
            alt Valid dimensions
                IH->>TSQ: Resize(width, height)
                IH->>IH: Create success response
                IH->>WS: WriteText(response)
            else Invalid dimensions
                IH->>IH: Create error response
                IH->>WS: WriteText(error)
            end
        else Unsupported type
            IH->>IH: Create error response
            IH->>WS: WriteText(error)
        end
        
        IH-->>WS: Return message length
    else Unsupported Message Type
        IH-->>WS: Return error
    end
    
    Note over WS,JSON: Next() Flow
    IH->>TSQ: Next()
    TSQ-->>IH: TerminalSize
    IH->>IH: sendResizeUpdate()
    IH->>WS: WriteText(update)
    IH-->>IH: Return TerminalSize
    
    Note over WS,JSON: Close() Flow
    IH->>TSQ: Stop()
    IH-->>IH: Return success
```

## Data Structures

```mermaid
classDiagram
    class InteractiveHandler {
        +conn: WebSocket
        +terminalSizeQueue: *TerminalSizeQueue
        +id: atomic.Int32
        +NewInteractiveHandler(conn, queue)
        +Process(msgType, msgRead, msgBuf) (int, error)
        +handleMessage(msgRead) *MessageError
        +handleTerminalSize(request) *MessageError
        +Close() error
        +Next() *TerminalSize
        +sendResizeUpdate(next) error
        +handleError(errMsg) void
    }
    
    class MessageRequest {
        +Version: string
        +Id: int
        +Message: string
        +Type: string
        +Params: map[string]any
        +GetFloat64(key, def) float64
        +GetInt(key, def) int
    }
    
    class MessageResponse {
        +Version: string
        +Id: int
        +Status: int
        +Message: string
        +Type: string
    }
    
    class MessageError {
        +error: error
        +Id: int
    }
    
    class WebSocket {
        <<interface>>
        +net.Conn
        +WriteText(p []byte) error
        +SetMessageHandler(handler) void
    }
    
    InteractiveHandler --> WebSocket
    InteractiveHandler --> MessageRequest
    InteractiveHandler --> MessageResponse
    InteractiveHandler --> MessageError
```

## Key Features

1. **Message Processing**: Handles both binary and text WebSocket messages
2. **Terminal Size Management**: Processes terminal size updates and forwards them to the terminal size queue
3. **Error Handling**: Comprehensive error handling with structured error responses
4. **Response Generation**: Creates standardized JSON responses with version and status information
5. **Thread Safety**: Uses atomic operations for ID generation
6. **Resource Management**: Proper cleanup of terminal size queue on close

## Message Types

- **Binary Messages**: Forwarded directly to buffer
- **Text Messages**: JSON-encoded with specific message types
  - `terminal_size`: Updates terminal dimensions

## Error Scenarios

1. **Unsupported Message Type**: Non-binary, non-text messages
2. **Invalid JSON**: Malformed message structure
3. **Unsupported Message Type**: Non-terminal_size messages
4. **Invalid Dimensions**: Width or height ≤ 0
5. **Serialization Errors**: JSON marshaling failures
6. **WebSocket Write Errors**: Connection issues 