# Changelog

All notable changes to the MOPS SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.2] - 2025-09-26

### Changed
- **ToolInstaller Output**: Removed automatic icon prefixes from stderr output for cleaner command output
  - No longer prefixes stderr with ⚠️ warning icons
  - Raw command output is now passed through without modification
  - Maintains compatibility while providing cleaner user experience

### Fixed  
- **Output Consistency**: ToolInstaller output now matches direct shell command execution
- **Icon Pollution**: Eliminated unwanted emoji prefixes in tool installation output

### Technical Details
- Modified `ToolInstaller.streamOutput()` to remove stderr prefix `⚠️  `
- Updated error message formatting to remove emoji prefixes
- Maintained all functionality while improving output clarity

## [1.0.1] - Previous Release
- System-level safety net improvements
- Enhanced plugin resilience and error handling
- Cross-plugin communication enhancements