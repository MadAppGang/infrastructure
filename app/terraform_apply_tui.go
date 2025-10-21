package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Apply-specific types
type applyState struct {
	// Progress tracking
	startTime      time.Time
	totalResources int
	completed      []completedResource
	pending        []pendingResource
	currentOp      *currentOperation
	logs           []logEntry

	// Terraform process
	cmd            *exec.Cmd
	mu             sync.Mutex

	// State
	isApplying     bool
	applyComplete  bool
	hasErrors      bool
	errorCount     int
	warningCount   int

	// Diagnostic tracking (for associating with errors)
	diagnostics map[string]*DiagnosticInfo // resource address -> diagnostic

	// View state
	showFullLogs    bool
	showDetails     bool
	showErrorDetails bool
	selectedSection int  // 0=completed, 1=pending, 2=logs
	selectedError   int  // Index of selected error in completed list
	animationFrame  int  // For progress bar animation

	// Layout heights (calculated once, fixed during apply)
	headerHeight       int
	progressHeight     int
	currentOpHeight    int
	errorSummaryHeight int
	columnsHeight      int
	logsHeight         int
	footerHeight       int
}

type completedResource struct {
	Address         string
	Action          string
	Duration        time.Duration
	Timestamp       time.Time
	Success         bool
	Error           string // Short error message
	ErrorSummary    string // Diagnostic summary (if available)
	ErrorDetail     string // Full diagnostic detail (if available)
}

type pendingResource struct {
	Address string
	Action  string
	Type    string
}

type currentOperation struct {
	Address     string
	Action      string
	Progress    float64
	Status      string
	StartTime   time.Time
	ElapsedTime string // Elapsed time from Terraform messages like "10s"
}

type logEntry struct {
	Timestamp time.Time
	Level     string // info, warning, error
	Message   string
	Resource  string // optional resource address
}

// Messages for apply state updates
type applyStartMsg struct{}
type applyCompleteMsg struct{ success bool }
type applyErrorMsg struct{ err error }
type applyTickMsg struct{} // For animating progress bars

type resourceStartMsg struct {
	Address string
	Action  string
}

type resourceProgressMsg struct {
	Address  string
	Progress float64
	Status   string
}

type resourceCompleteMsg struct {
	Address      string
	Success      bool
	Error        string // Short error message
	ErrorSummary string // Diagnostic summary
	ErrorDetail  string // Full diagnostic detail
	Duration     time.Duration
}

type logMsg struct {
	Level    string
	Message  string
	Resource string
}

// TerraformJSONMessage represents the JSON output from terraform apply -json
type TerraformJSONMessage struct {
	Type       string          `json:"@type"`
	Level      string          `json:"@level"`
	Message    string          `json:"@message"`
	Module     string          `json:"@module"`
	Timestamp  string          `json:"@timestamp"`
	Diagnostic *DiagnosticInfo `json:"diagnostic,omitempty"`
	Hook       *HookInfo       `json:"hook,omitempty"`
	Change     *ChangeInfo     `json:"change,omitempty"`
}

type DiagnosticInfo struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail"`
	Address  string `json:"address,omitempty"`
}

type HookInfo struct {
	Resource       *ResourceInfo `json:"resource,omitempty"`
	Action         string        `json:"action,omitempty"`
	IDKey          string        `json:"id_key,omitempty"`
	IDValue        string        `json:"id_value,omitempty"`
	ElapsedSeconds float64       `json:"elapsed_seconds,omitempty"`
}

type ResourceInfo struct {
	Addr            string `json:"addr"`
	Module          string `json:"module"`
	Resource        string `json:"resource"`
	ResourceType    string `json:"resource_type"`
	ResourceName    string `json:"resource_name"`
	ResourceKey     interface{} `json:"resource_key,omitempty"`
	ImpliedProvider string `json:"implied_provider,omitempty"`
}

type ChangeInfo struct {
	Resource *ResourceInfo `json:"resource"`
	Action   string        `json:"action"`
}

// Calculate fixed layout heights for apply view
// NOTE: Heights stored here are CONTENT heights (what goes inside boxes).
// Boxes with borders will render as contentHeight + 2 lines on screen.
func (m *modernPlanModel) calculateApplyLayout(terminalHeight int) {
	if m.applyState == nil {
		return
	}

	// Fixed overhead
	headerHeight := 1
	footerHeight := 1

	// Calculate available height for all boxes
	// Add safety margin of 1 line to prevent overflow from spacing/newlines
	safetyMargin := 1
	availableHeight := terminalHeight - headerHeight - footerHeight - safetyMargin
	if availableHeight < 10 {
		availableHeight = 10 // absolute minimum
	}

	// Assign fixed values
	m.applyState.headerHeight = headerHeight
	m.applyState.footerHeight = footerHeight

	// Define size tiers - heights here are SCREEN heights (including borders)
	// Each box with a border takes contentHeight + 2 on screen
	// Thresholds are more conservative to ensure fit on real screens
	if availableHeight < 18 {
		// Tiny screen (< 21 total lines) - ultra-compact mode
		m.applyState.progressHeight = 0       // Hidden
		m.applyState.currentOpHeight = 4      // 4 lines total on screen -> content height 2
		m.applyState.errorSummaryHeight = 0   // Hidden
		m.applyState.columnsHeight = 5        // 5 lines total on screen -> content height 3
		// Logs get the rest: availableHeight - (4 + 5) = availableHeight - 9
		logsScreenHeight := availableHeight - 9
		if logsScreenHeight < 4 {
			logsScreenHeight = 4
		}
		m.applyState.logsHeight = logsScreenHeight
	} else if availableHeight < 30 {
		// Small screen (21-32 total lines) - compact mode
		// Target layout: currentOp(5), columns(7), logs(rest)
		m.applyState.progressHeight = 0       // Hidden to save space
		m.applyState.currentOpHeight = 5      // 5 lines total -> content 3
		m.applyState.errorSummaryHeight = 0   // Hidden to save space
		m.applyState.columnsHeight = 7        // 7 lines total -> content 5
		// Logs: availableHeight - (5 + 7) = availableHeight - 12
		logsScreenHeight := availableHeight - 12
		if logsScreenHeight < 5 {
			logsScreenHeight = 5
		}
		m.applyState.logsHeight = logsScreenHeight
	} else if availableHeight < 45 {
		// Medium screen (32-47 total lines) - balanced mode
		// This covers the 42-row terminal case
		m.applyState.progressHeight = 3       // Show progress bar (important feedback!)
		m.applyState.currentOpHeight = 6      // 6 lines total -> content 4
		m.applyState.errorSummaryHeight = 0   // Hidden to save space
		m.applyState.columnsHeight = 8        // 8 lines total -> content 6 (reduced from 10)
		// Logs: availableHeight - (3 + 6 + 8) = availableHeight - 17
		logsScreenHeight := availableHeight - 17
		if logsScreenHeight < 8 {
			logsScreenHeight = 8
		}
		m.applyState.logsHeight = logsScreenHeight
	} else {
		// Large screen (47+ total lines) - ideal mode
		m.applyState.progressHeight = 3       // 3 lines total -> content 1
		m.applyState.currentOpHeight = 8      // 8 lines total -> content 6
		m.applyState.errorSummaryHeight = 4   // 4 lines total -> content 2
		m.applyState.columnsHeight = 12       // 12 lines total -> content 10
		// Logs: availableHeight - (3 + 8 + 4 + 12) = availableHeight - 27
		logsScreenHeight := availableHeight - 27
		if logsScreenHeight < 10 {
			logsScreenHeight = 10
		}
		m.applyState.logsHeight = logsScreenHeight
	}
}

// Initialize apply state from the plan
func (m *modernPlanModel) initApplyState() {
	m.applyState = &applyState{
		startTime:      time.Now(),
		logs:           []logEntry{},
		pending:        []pendingResource{},
		completed:      []completedResource{},
		diagnostics:    make(map[string]*DiagnosticInfo),
		totalResources: m.stats.totalChanges,
	}

	// Calculate initial layout
	m.calculateApplyLayout(m.height)

	// Convert planned resources to pending
	for _, provider := range m.providers {
		for _, service := range provider.services {
			for _, resource := range service.resources {
				m.applyState.pending = append(m.applyState.pending, pendingResource{
					Address: resource.Address,
					Action:  resource.Change.Actions[0],
					Type:    resource.Type,
				})
			}
		}
	}

	// Add initial log
	m.applyState.logs = append(m.applyState.logs, logEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   fmt.Sprintf("Starting terraform apply with %d resources", m.applyState.totalResources),
	})
}

// Start terraform apply process
func (m *modernPlanModel) startTerraformApply() tea.Cmd {
	return func() tea.Msg {
		// Initialize apply state if needed
		if m.applyState == nil {
			m.initApplyState()
		}
		
		// Build command arguments
		args := []string{"apply", "-json", "-auto-approve"}
		
		// Add replace flags for marked resources
		for resource := range m.markedForReplace {
			args = append(args, fmt.Sprintf("-replace=%s", resource))
		}
		
		// Only use plan file if no replacements are marked
		// When using -replace, we need to let terraform create a new plan
		if len(m.markedForReplace) == 0 {
			// Add plan file
			args = append(args, "tfplan")
		}
		
		// Log the command being executed if there are replacements
		if len(m.markedForReplace) > 0 {
			m.sendLogMessage("info", fmt.Sprintf("🔄 Running terraform apply with %d resource replacements", len(m.markedForReplace)), "")
			for resource := range m.markedForReplace {
				m.sendLogMessage("info", fmt.Sprintf("  • Replacing: %s", resource), "")
			}
		}
		
		// Start terraform apply with JSON output
		cmd := exec.Command("terraform", args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return applyErrorMsg{err: err}
		}
		
		m.applyState.cmd = cmd
		
		if err := cmd.Start(); err != nil {
			return applyErrorMsg{err: err}
		}
		
		// Start parsing JSON output in goroutine
		go m.parseTerraformOutput(stdout)
		
		return applyStartMsg{}
	}
}

// Parse terraform JSON output
func (m *modernPlanModel) parseTerraformOutput(stdout interface{}) {
	scanner := bufio.NewScanner(stdout.(interface{ Read([]byte) (int, error) }))
	
	for scanner.Scan() {
		line := scanner.Text()
		var msg TerraformJSONMessage
		
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Not JSON, treat as plain log
			m.sendLogMessage("info", line, "")
			continue
		}
		
		
		// Process based on message type
		switch msg.Type {
		case "apply_start":
			m.handleApplyStart(&msg)
		case "apply_progress":
			m.handleApplyProgress(&msg)
		case "apply_complete":
			m.handleApplyComplete(&msg)
		case "apply_errored":
			m.handleApplyError(&msg)
		case "resource_drift":
			// Handle resource drift messages
			if msg.Hook != nil && msg.Hook.Resource != nil {
				m.sendLogMessage("info", fmt.Sprintf("Resource drift detected: %s", msg.Hook.Resource.Addr), msg.Hook.Resource.Addr)
			}
		case "diagnostic":
			m.handleDiagnostic(&msg)
		case "provision_start", "provision_progress", "provision_complete", "provision_errored":
			m.handleProvisionMessage(&msg)
		case "refresh_start":
			m.sendLogMessage("info", "🔄 Starting refresh...", "")
		case "refresh_complete":
			m.sendLogMessage("info", "✅ Refresh completed", "")
		default:
			
			// Check if this is an error message by content
			if msg.Message != "" {
				// Use the level from the message
				logLevel := "info"
				if msg.Level != "" {
					logLevel = msg.Level
				}
				
				// Check for completion patterns in the message
				if strings.Contains(msg.Message, ": Creation complete after") ||
				   strings.Contains(msg.Message, ": Modifications complete after") ||
				   strings.Contains(msg.Message, ": Destroy complete after") ||
				   strings.Contains(msg.Message, ": Destruction complete after") {
					// Parse successful completion
					parts := strings.SplitN(msg.Message, ":", 2)
					if len(parts) >= 1 {
						addr := strings.TrimSpace(parts[0])
						// Send a resource complete message with success
						m.sendMsg(resourceCompleteMsg{
							Address:  addr,
							Success:  true,
							Error:    "",
							Duration: 0,
						})
					}
				} else if msg.Level == "error" ||
				   (msg.Message != "" && (strings.Contains(msg.Message, ": Creation errored after") ||
				                         strings.Contains(msg.Message, ": Modification errored after") ||
				                         strings.Contains(msg.Message, ": Destruction errored after"))) {
					// Parse the resource address from error message
					if strings.Contains(msg.Message, "errored after") {
						// Override log level to error for these messages
						logLevel = "error"
						parts := strings.SplitN(msg.Message, ":", 2)
						if len(parts) >= 1 {
							addr := strings.TrimSpace(parts[0])

							// Check for associated diagnostic information
							m.applyState.mu.Lock()
							diagnostic := m.applyState.diagnostics[addr]

							errorSummary := ""
							errorDetail := ""
							if diagnostic != nil {
								errorSummary = diagnostic.Summary
								errorDetail = diagnostic.Detail
							} else {
								// Fallback: Collect error logs from around the same time
								var resourceErrorTime time.Time
								var errorLogs []string

								// Find when this resource failed
								for _, log := range m.applyState.logs {
									if log.Level == "error" && (log.Resource == addr || strings.Contains(log.Message, addr)) {
										resourceErrorTime = log.Timestamp
										break
									}
								}

								// Collect ALL error logs within 2 seconds
								if !resourceErrorTime.IsZero() {
									for _, log := range m.applyState.logs {
										if log.Level == "error" {
											timeDiff := log.Timestamp.Sub(resourceErrorTime)
											if timeDiff >= 0 && timeDiff <= 2*time.Second {
												logMsg := strings.TrimPrefix(log.Message, "❌ ")
												logMsg = strings.TrimSpace(logMsg)
												if !strings.Contains(logMsg, "errored after") && logMsg != "" {
													errorLogs = append(errorLogs, logMsg)
												}
											}
										}
									}
								}

								if len(errorLogs) > 0 {
									errorDetail = strings.Join(errorLogs, "\n\n")
								}
							}
							m.applyState.mu.Unlock()

							// Send a resource complete message with error
							m.sendMsg(resourceCompleteMsg{
								Address:      addr,
								Success:      false,
								Error:        msg.Message,
								ErrorSummary: errorSummary,
								ErrorDetail:  errorDetail,
								Duration:     0,
							})
						}
					}
				} else if strings.Contains(msg.Message, ": Still destroying...") ||
				          strings.Contains(msg.Message, ": Destroying...") ||
				          strings.Contains(msg.Message, ": Still creating...") ||
				          strings.Contains(msg.Message, ": Creating...") ||
				          strings.Contains(msg.Message, ": Still modifying...") ||
				          strings.Contains(msg.Message, ": Modifying...") {
					// Parse in-progress operations
					parts := strings.SplitN(msg.Message, ":", 2)
					if len(parts) >= 1 {
						addr := strings.TrimSpace(parts[0])
						// Remove any remote-exec or other provisioner suffixes
						if strings.Contains(addr, " (") {
							addr = strings.Split(addr, " (")[0]
						}
						
						// Determine action from message
						action := "update"
						if strings.Contains(msg.Message, "destroy") || strings.Contains(msg.Message, "Destroy") {
							action = "delete"
						} else if strings.Contains(msg.Message, "creat") || strings.Contains(msg.Message, "Creat") {
							action = "create"
						} else if strings.Contains(msg.Message, "modify") || strings.Contains(msg.Message, "Modify") {
							action = "update"
						}
						
						// Extract elapsed time if present (e.g., "[10s elapsed]")
						elapsedTime := ""
						if strings.Contains(msg.Message, "elapsed]") {
							// Find the pattern [XXs elapsed]
							startIdx := strings.LastIndex(msg.Message, "[")
							endIdx := strings.LastIndex(msg.Message, " elapsed]")
							if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
								elapsedTime = msg.Message[startIdx+1:endIdx]
							}
						}
						
						// Send start message if we haven't seen this resource yet
						if m.applyState != nil {
							if m.applyState.currentOp == nil || m.applyState.currentOp.Address != addr {
								m.sendMsg(resourceStartMsg{
									Address: addr,
									Action:  action,
								})
							} else if m.applyState.currentOp != nil && m.applyState.currentOp.Address == addr && elapsedTime != "" {
								// Update elapsed time for current operation
								m.applyState.currentOp.ElapsedTime = elapsedTime
							}
						}
					}
				}
				
				m.sendLogMessage(logLevel, msg.Message, "")
			}
		}
	}
	
	// Check if process completed
	if err := m.applyState.cmd.Wait(); err != nil {
		m.sendMsg(applyErrorMsg{err: err})
	} else {
		m.sendMsg(applyCompleteMsg{success: true})
	}
}

func (m *modernPlanModel) handleApplyStart(msg *TerraformJSONMessage) {
	if msg.Hook == nil || msg.Hook.Resource == nil {
		return
	}
	
	addr := msg.Hook.Resource.Addr
	action := msg.Hook.Action
	if action == "" {
		action = m.getResourceAction(addr)
	}
	
	m.sendMsg(resourceStartMsg{
		Address: addr,
		Action:  action,
	})
	m.sendLogMessage("info", fmt.Sprintf("🔄 Starting %s on %s...", action, addr), addr)
}

func (m *modernPlanModel) handleApplyProgress(msg *TerraformJSONMessage) {
	if msg.Hook == nil || msg.Hook.Resource == nil {
		return
	}
	
	addr := msg.Hook.Resource.Addr
	// Calculate progress based on elapsed time
	progress := 0.5
	if msg.Hook.ElapsedSeconds > 0 {
		// Assume most operations complete within 30 seconds
		progress = math.Min(msg.Hook.ElapsedSeconds/30.0, 0.9)
	}
	
	status := fmt.Sprintf("In progress (%.1fs)", msg.Hook.ElapsedSeconds)
	
	m.sendMsg(resourceProgressMsg{
		Address:  addr,
		Progress: progress,
		Status:   status,
	})
}

func (m *modernPlanModel) handleApplyComplete(msg *TerraformJSONMessage) {
	if msg.Hook == nil || msg.Hook.Resource == nil {
		return
	}
	
	addr := msg.Hook.Resource.Addr
	action := msg.Hook.Action
	if action == "" {
		action = m.getResourceAction(addr)
	}
	
	// Use elapsed seconds from the message
	duration := time.Duration(msg.Hook.ElapsedSeconds * float64(time.Second))
	
	m.sendMsg(resourceCompleteMsg{
		Address:  addr,
		Success:  true,
		Duration: duration,
	})
	
	// Include ID in log if available
	idInfo := ""
	if msg.Hook.IDKey != "" && msg.Hook.IDValue != "" {
		idInfo = fmt.Sprintf(" [%s=%s]", msg.Hook.IDKey, msg.Hook.IDValue)
	}
	
	m.sendLogMessage("info", fmt.Sprintf("✅ Completed %s on %s (%v)%s", action, addr, duration, idInfo), addr)
}

func (m *modernPlanModel) handleApplyError(msg *TerraformJSONMessage) {
	if msg.Hook == nil || msg.Hook.Resource == nil {
		return
	}

	addr := msg.Hook.Resource.Addr
	action := msg.Hook.Action
	if action == "" {
		action = m.getResourceAction(addr)
	}

	// Use elapsed seconds from the message
	duration := time.Duration(msg.Hook.ElapsedSeconds * float64(time.Second))

	// Error message is typically in the main message field
	errorMsg := msg.Message
	if errorMsg == "" {
		errorMsg = "Operation failed"
	}

	// Check for associated diagnostic information
	m.applyState.mu.Lock()
	diagnostic := m.applyState.diagnostics[addr]

	// Fallback: Extract error details from recent error logs if no diagnostic
	errorSummary := ""
	errorDetail := ""
	if diagnostic != nil {
		errorSummary = diagnostic.Summary
		errorDetail = diagnostic.Detail
	} else {
		// Strategy: Find the resource's error timestamp, then collect ALL errors near that time
		var resourceErrorTime time.Time
		var errorLogs []string

		// First pass: Find when this resource failed
		for _, log := range m.applyState.logs {
			if log.Level == "error" && (log.Resource == addr || strings.Contains(log.Message, addr)) {
				resourceErrorTime = log.Timestamp
				break
			}
		}

		// Second pass: Collect ALL error logs within 2 seconds of the resource failure
		if !resourceErrorTime.IsZero() {
			for _, log := range m.applyState.logs {
				if log.Level == "error" {
					// Check if this error is within 2 seconds of the resource error
					timeDiff := log.Timestamp.Sub(resourceErrorTime)
					if timeDiff >= 0 && timeDiff <= 2*time.Second {
						// Remove icon prefix for cleaner display
						msg := strings.TrimPrefix(log.Message, "❌ ")
						msg = strings.TrimSpace(msg)

						// Skip duplicate "errored after" messages (we already have that)
						if !strings.Contains(msg, "errored after") && msg != "" {
							errorLogs = append(errorLogs, msg)
						}
					}
				}
			}
		}

		// If we found error details, combine them
		if len(errorLogs) > 0 {
			errorDetail = strings.Join(errorLogs, "\n\n")
		}
	}
	m.applyState.mu.Unlock()

	m.sendMsg(resourceCompleteMsg{
		Address:      addr,
		Success:      false,
		Error:        errorMsg,
		ErrorSummary: errorSummary,
		ErrorDetail:  errorDetail,
		Duration:     duration,
	})

	m.sendLogMessage("error", fmt.Sprintf("❌ Failed %s on %s after %v: %s", action, addr, duration, errorMsg), addr)
}

func (m *modernPlanModel) handleDiagnostic(msg *TerraformJSONMessage) {
	if msg.Diagnostic == nil {
		return
	}

	level := "info"
	icon := "ℹ️"

	switch msg.Diagnostic.Severity {
	case "error":
		level = "error"
		icon = "❌"
		// Store diagnostic for later association with failed resource
		if msg.Diagnostic.Address != "" {
			m.applyState.mu.Lock()
			m.applyState.diagnostics[msg.Diagnostic.Address] = msg.Diagnostic
			m.applyState.mu.Unlock()
		}
	case "warning":
		level = "warning"
		icon = "⚠️"
	}

	// Log both summary and detail separately so they can be extracted
	if msg.Diagnostic.Summary != "" {
		summaryMsg := fmt.Sprintf("%s %s", icon, msg.Diagnostic.Summary)
		m.sendLogMessage(level, summaryMsg, msg.Diagnostic.Address)
	}

	// Log the detail as a separate message for better extraction
	if msg.Diagnostic.Detail != "" {
		m.sendLogMessage(level, msg.Diagnostic.Detail, msg.Diagnostic.Address)
	}
}

func (m *modernPlanModel) handleProvisionMessage(msg *TerraformJSONMessage) {
	switch msg.Type {
	case "provision_start":
		m.sendLogMessage("info", "🔧 Starting provisioner...", "")
	case "provision_progress":
		// Provisioner output is typically in the message field
		if msg.Message != "" {
			m.sendLogMessage("info", fmt.Sprintf("  %s", msg.Message), "")
		}
	case "provision_complete":
		m.sendLogMessage("info", "✅ Provisioner completed", "")
	case "provision_errored":
		m.sendLogMessage("error", "❌ Provisioner failed", "")
	}
}

func (m *modernPlanModel) sendLogMessage(level, message, resource string) {
	m.sendMsg(logMsg{
		Level:    level,
		Message:  message,
		Resource: resource,
	})
}

func (m *modernPlanModel) getResourceAction(address string) string {
	for _, provider := range m.providers {
		for _, service := range provider.services {
			for _, resource := range service.resources {
				if resource.Address == address && len(resource.Change.Actions) > 0 {
					return resource.Change.Actions[0]
				}
			}
		}
	}
	return "update"
}

// sendMsg sends a message through the program
func (m *modernPlanModel) sendMsg(msg tea.Msg) {
	if m.program != nil {
		m.program.Send(msg)
	}
}