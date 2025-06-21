# VideoCraft - Comprehensive Developer Documentation

VideoCraft is a security-first video generation platform with AI-powered progressive subtitles, Clean Architecture design, and comprehensive package documentation for developers and AI assistants.

## 📚 Complete Documentation Structure

### 🏗️ Internal Architecture Documentation
**Comprehensive CLAUDE.md files for every package and subpackage:**

#### API Layer
- [**internal/api/**](internal/api/CLAUDE.md) - API layer overview and architecture
- [**internal/api/http/**](internal/api/http/CLAUDE.md) - HTTP interface with Gin framework
- [**internal/api/http/middleware/**](internal/api/http/middleware/CLAUDE.md) - Security-first middleware stack
- [**internal/api/http/handlers/**](internal/api/http/handlers/CLAUDE.md) - Request handlers and routing
- [**internal/api/models/**](internal/api/models/CLAUDE.md) - Data structures and validation

#### Application Layer
- [**internal/app/**](internal/app/CLAUDE.md) - Configuration management and DI

#### Core Business Logic
- [**internal/core/**](internal/core/CLAUDE.md) - Core services overview
- [**internal/core/media/**](internal/core/media/CLAUDE.md) - Media processing layer
- [**internal/core/media/audio/**](internal/core/media/audio/CLAUDE.md) - URL-first audio analysis
- [**internal/core/media/video/**](internal/core/media/video/CLAUDE.md) - Video validation and metadata
- [**internal/core/media/image/**](internal/core/media/image/CLAUDE.md) - Secure image processing
- [**internal/core/media/subtitle/**](internal/core/media/subtitle/CLAUDE.md) - **Progressive Subtitles Innovation**
- [**internal/core/services/**](internal/core/services/CLAUDE.md) - Core services architecture
- [**internal/core/services/job/queue/**](internal/core/services/job/queue/CLAUDE.md) - Asynchronous job processing
- [**internal/core/services/transcription/**](internal/core/services/transcription/CLAUDE.md) - **Python-Go Whisper Integration**
- [**internal/core/video/**](internal/core/video/CLAUDE.md) - Video engine layer
- [**internal/core/video/composition/**](internal/core/video/composition/CLAUDE.md) - Service composition patterns
- [**internal/core/video/engine/**](internal/core/video/engine/CLAUDE.md) - FFmpeg integration and security

#### Infrastructure Layer
- [**internal/pkg/**](internal/pkg/CLAUDE.md) - Shared packages overview
- [**internal/pkg/logger/**](internal/pkg/logger/CLAUDE.md) - Structured logging with slog
- [**internal/pkg/errors/**](internal/pkg/errors/CLAUDE.md) - **Security-first error handling**
- [**internal/storage/**](internal/storage/CLAUDE.md) - Storage layer overview
- [**internal/storage/filesystem/**](internal/storage/filesystem/CLAUDE.md) - Secure filesystem operations

#### Root Architecture
- [**internal/**](internal/CLAUDE.md) - **Clean Architecture overview and system design**

### 📖 External Documentation (Legacy)
**Topic-based external documentation (for reference):**
- [docs/README.md](docs/README.md) - External documentation index
- [Architecture guides](docs/architecture/) - High-level system design
- [Security documentation](docs/security/) - Security architecture
- [API references](docs/api/) - External API documentation
- [Deployment guides](docs/deployment/) - Operations documentation

## 🏗️ Key Innovations & Architecture

### 🎯 Progressive Subtitles Innovation
VideoCraft's breakthrough feature revolutionizes subtitle timing:
- [**Technical Deep Dive**](internal/core/media/subtitle/CLAUDE.md) - Complete implementation details
- [**Audio Duration Analysis**](internal/core/media/audio/CLAUDE.md) - URL-first approach without downloads
- [**Real-Time Word Timing**](internal/core/services/transcription/CLAUDE.md) - Whisper AI integration
- **JSON Settings v0.0.1+** - Per-request subtitle customization with global fallback

### 🛡️ Security-First Architecture
Comprehensive multi-layer security design:
- [**Security Overview**](internal/api/http/middleware/CLAUDE.md) - Zero wildcards CORS policy
- [**Error Sanitization**](internal/pkg/errors/CLAUDE.md) - 40+ security patterns detection
- [**Path Protection**](internal/storage/filesystem/CLAUDE.md) - Advanced traversal prevention
- [**Command Injection Prevention**](internal/core/video/engine/CLAUDE.md) - Secure FFmpeg integration

### 🔗 Python-Go Integration Excellence
Seamless AI integration architecture:
- [**Daemon Architecture**](internal/core/services/transcription/CLAUDE.md) - Long-running Whisper process
- [**Communication Protocol**](internal/core/services/transcription/CLAUDE.md) - stdin/stdout JSON protocol
- [**Lifecycle Management**](internal/core/services/transcription/CLAUDE.md) - Automatic startup/shutdown
- [**Performance Optimization**](internal/core/services/transcription/CLAUDE.md) - 5-minute idle timeout

### 🎬 URL-First Media Analysis
Innovative approach to media processing:
- [**No-Download Analysis**](internal/core/media/audio/CLAUDE.md) - FFprobe URL analysis
- [**Google Drive Integration**](internal/core/media/audio/CLAUDE.md) - Direct URL resolution
- [**Streaming Support**](internal/core/media/video/CLAUDE.md) - Large file handling
- [**Security Validation**](internal/core/media/image/CLAUDE.md) - Format and content validation

## 📊 Architecture Patterns & Design

### 🏛️ Clean Architecture Implementation
- [**System Overview**](internal/CLAUDE.md) - Complete Clean Architecture implementation
- [**Layer Separation**](internal/api/CLAUDE.md) - Interface, Application, Core, Infrastructure
- [**Dependency Injection**](internal/core/video/composition/CLAUDE.md) - Service composition patterns
- [**Interface Design**](internal/pkg/CLAUDE.md) - Shared interfaces and utilities

### 🔄 Data Flow & Processing
- [**Request Pipeline**](internal/api/http/CLAUDE.md) - HTTP request lifecycle
- [**Job Processing**](internal/core/services/job/queue/CLAUDE.md) - Asynchronous video generation
- [**Media Pipeline**](internal/core/media/CLAUDE.md) - Audio, video, subtitle processing
- [**Storage Management**](internal/storage/CLAUDE.md) - Secure file operations

### 🛠️ Development Patterns
- [**Error Handling**](internal/pkg/errors/CLAUDE.md) - Domain-specific error types
- [**Logging Strategy**](internal/pkg/logger/CLAUDE.md) - Structured logging with slog
- [**Configuration Management**](internal/app/CLAUDE.md) - Environment-driven config
- [**Testing Approaches**] - Comprehensive testing strategies in each package

### 🔧 Integration Patterns
- [**HTTP Middleware**](internal/api/http/middleware/CLAUDE.md) - Security, CORS, rate limiting
- [**External Services**](internal/core/media/audio/CLAUDE.md) - FFprobe, Google Drive integration
- [**Background Processing**](internal/core/services/job/queue/CLAUDE.md) - Worker pools and job management
- [**AI Integration**](internal/core/services/transcription/CLAUDE.md) - Python-Go communication

## 🤖 AI & LLM Integration

### 📖 LLM-Optimized Documentation
VideoCraft provides comprehensive documentation optimized for AI assistants:

- **[/llms.txt](llms.txt)** - Essential documentation index for LLM consumption
- **[/llms-full.txt](llms-full.txt)** - Complete technical documentation
- **Internal CLAUDE.md files** - Package-specific detailed documentation
- **Architecture diagrams** - Visual Mermaid representations in every package

### 🔍 Documentation Features for AI
- **Structured Format** - Consistent CLAUDE.md format across all packages
- **Code Examples** - Complete implementation samples in every package
- **Security Patterns** - Detailed security implementations and validations
- **Testing Strategies** - Comprehensive test approaches documented
- **Performance Notes** - Optimization patterns and characteristics
- **Mermaid Diagrams** - Visual architecture in every component

### 🧠 AI-Friendly Patterns
- **Interface Documentation** - Complete Go interface definitions
- **Error Handling** - Standardized error types and security filtering
- **Configuration Examples** - YAML and environment variable samples
- **Usage Patterns** - Real-world implementation examples
- **Integration Examples** - Service composition and communication

## 🔗 Additional Resources

- [**GitHub Repository**](https://github.com/activadee/videocraft) - Source code and issues
- [**README.md**](README.md) - User documentation and quick start guide
- [**Security Documentation**](internal/pkg/errors/CLAUDE.md) - Comprehensive security implementation
- [**Architecture Overview**](internal/CLAUDE.md) - System design and Clean Architecture patterns

---

## 📈 Documentation Coverage

✅ **Complete Package Documentation**: All 18 packages/subpackages documented  
✅ **Security-First Design**: Comprehensive security patterns documented  
✅ **Clean Architecture**: Complete layer separation and interface documentation  
✅ **AI-Optimized**: LLM-friendly documentation with consistent structure  
✅ **Visual Architecture**: Mermaid diagrams in every component  
✅ **Code Examples**: Complete implementation samples throughout  
✅ **Testing Strategies**: Comprehensive test approaches documented  
✅ **Performance Patterns**: Optimization strategies and characteristics  

**Release Version**: v0.0.1 (Initial Release with Comprehensive Documentation)  
**Architecture**: Clean Architecture with Security-First Design  
**Coverage**: 100% - All packages and subpackages documented  

*For implementation details, start with [internal/CLAUDE.md](internal/CLAUDE.md) and navigate to specific packages.*