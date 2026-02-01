// Hackathon Starter: Complete AI Financial Agent
// Build intelligent financial tools with nim-go-sdk + Liminal banking APIs
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/executor"
	"github.com/becomeliminal/nim-go-sdk/server"
	"github.com/becomeliminal/nim-go-sdk/tools"
	"github.com/joho/godotenv"
)

func main() {
	// ============================================================================
	// CONFIGURATION
	// ============================================================================
	// Load .env file if it exists (optional - will use system env vars if not found)
	_ = godotenv.Load()

	// Load configuration from environment variables
	// Create a .env file or export these in your shell

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Fatal("❌ ANTHROPIC_API_KEY environment variable is required")
	}

	liminalBaseURL := os.Getenv("LIMINAL_BASE_URL")
	if liminalBaseURL == "" {
		liminalBaseURL = "https://api.liminal.cash"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ============================================================================
	// LIMINAL EXECUTOR SETUP
	// ============================================================================
	// The HTTPExecutor handles all API calls to Liminal banking services.
	// Authentication flow: User logs in via frontend → JWT stored → WebSocket extracts JWT → Forwarded to Liminal API
	// No API key needed - JWT tokens from the frontend login flow (email/OTP) handle auth automatically

	liminalExecutor := executor.NewHTTPExecutor(executor.HTTPExecutorConfig{
		BaseURL: liminalBaseURL,
	})
	log.Println("✅ Liminal API configured")

	// ============================================================================
	// SERVER SETUP
	// ============================================================================
	// Create the nim-go-sdk server with Claude AI
	// The server handles WebSocket connections and manages conversations
	// Authentication is automatic: JWT tokens from the login flow are extracted
	// from WebSocket connections and forwarded to Liminal API calls

	srv, err := server.New(server.Config{
		AnthropicKey:    anthropicKey,
		SystemPrompt:    hackathonSystemPrompt,
		Model:           "claude-sonnet-4-20250514",
		MaxTokens:       4096,
		LiminalExecutor: liminalExecutor, // SDK automatically handles JWT extraction and forwarding
	})
	if err != nil {
		log.Fatal(err)
	}

	// ============================================================================
	// ADD LIMINAL BANKING TOOLS
	// ============================================================================
	// These are the 9 core Liminal tools that give your AI access to real banking:
	//
	// READ OPERATIONS (no confirmation needed):
	//   1. get_balance - Check wallet balance
	//   2. get_savings_balance - Check savings positions and APY
	//   3. get_vault_rates - Get current savings rates
	//   4. get_transactions - View transaction history
	//   5. get_profile - Get user profile info
	//   6. search_users - Find users by display tag
	//
	// WRITE OPERATIONS (require user confirmation):
	//   7. send_money - Send money to another user
	//   8. deposit_savings - Deposit funds into savings
	//   9. withdraw_savings - Withdraw funds from savings

	srv.AddTools(tools.LiminalTools(liminalExecutor)...)
	log.Println("✅ Added 9 Liminal banking tools")

	// ============================================================================
	// ADD CUSTOM TOOLS
	// ============================================================================
	// This is where you'll add your hackathon project's custom tools!
	// Below are example analyzer tools to get you started.

	srv.AddTool(createSpendingAnalyzerTool(liminalExecutor))
	log.Println("✅ Added custom spending analyzer tool")

	srv.AddTool(createSubscriptionAnalyzerTool(liminalExecutor))
	log.Println("✅ Added custom subscription analyzer tool")

	// TODO: Add more custom tools here!
	// Examples:
	//   - Savings goal tracker
	//   - Budget alerts
	//   - Spending category analyzer
	//   - Bill payment predictor
	//   - Cash flow forecaster

	// ============================================================================
	// START SERVER
	// ============================================================================

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚀 Hackathon Starter Server Running")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("Ready for connections! Start your frontend with: cd frontend && npm run dev")
	log.Println()

	if err := srv.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// ============================================================================
// SYSTEM PROMPT
// ============================================================================
// This prompt defines your AI agent's personality and behavior
// Customize this to match your hackathon project's focus!

const hackathonSystemPrompt = `You are Nim, a friendly AI financial assistant built for the Liminal Vibe Banking Hackathon.

WHAT YOU DO:
You help users manage their money using Liminal's stablecoin banking platform. You can check balances, review transactions, send money, and manage savings - all through natural conversation.

CONVERSATIONAL STYLE:
- Be warm, friendly, and conversational - not robotic
- Use casual language when appropriate, but stay professional about money
- Ask clarifying questions when something is unclear
- Remember context from earlier in the conversation
- Explain things simply without being condescending

WHEN TO USE TOOLS:
- Use tools immediately for simple queries ("what's my balance?")
- For actions, gather all required info first ("send $50 to @alice")
- Always confirm before executing money movements
- Don't use tools for general questions about how things work

MONEY MOVEMENT RULES (IMPORTANT):
- ALL money movements require explicit user confirmation
- Show a clear summary before confirming:
  * send_money: "Send $50 USD to @alice"
  * deposit_savings: "Deposit $100 USD into savings"
  * withdraw_savings: "Withdraw $50 USD from savings"
- Never assume amounts or recipients
- Always use the exact currency the user specified

AVAILABLE BANKING TOOLS:
- Check wallet balance (get_balance)
- Check savings balance and APY (get_savings_balance)
- View savings rates (get_vault_rates)
- View transaction history (get_transactions)
- Get profile info (get_profile)
- Search for users (search_users)
- Send money (send_money) - requires confirmation
- Deposit to savings (deposit_savings) - requires confirmation
- Withdraw from savings (withdraw_savings) - requires confirmation

CUSTOM ANALYTICAL TOOLS:
- Analyze spending patterns (analyze_spending)
- Detect subscriptions (analyze_subscriptions)

TIPS FOR GREAT INTERACTIONS:
- Proactively suggest relevant actions ("Want me to move some to savings?")
- Explain the "why" behind suggestions
- Celebrate financial wins ("Nice! Your savings earned $5 this month!")
- Be encouraging about savings goals
- Make finance feel less intimidating

Remember: You're here to make banking delightful and help users build better financial habits!`

// ============================================================================
// MOCK DATA GENERATORS
// ============================================================================

// generateMockTransactionsForAnalysis creates realistic transaction data for testing
// Useful for demo purposes without needing real user data
func generateMockTransactionsForAnalysis(days int) []map[string]interface{} {
	rand.Seed(time.Now().UnixNano())
	now := time.Now()
	transactions := []map[string]interface{}{}

	// Transaction templates - realistic merchant names and amounts
	templates := []struct {
		description string
		amount      float64
		txType      string
	}{
		// Food & Dining
		{"Starbucks Coffee", 8.50, "send"},
		{"Chipotle Mexican Grill", 15.75, "send"},
		{"Whole Foods Market", 67.30, "send"},
		{"DoorDash - Pizza Delivery", 32.50, "send"},
		{"Local Coffee Shop", 6.25, "send"},
		// Transportation
		{"Uber Ride", 18.50, "send"},
		{"Gas Station", 45.00, "send"},
		{"Lyft Ride", 22.75, "send"},
		{"Metro Card Reload", 30.00, "send"},
		// Shopping
		{"Amazon.com", 89.99, "send"},
		{"Target Store", 54.25, "send"},
		{"Nike Store", 125.00, "send"},
		// Entertainment
		{"Netflix Subscription", 15.99, "send"},
		{"Spotify Premium", 10.99, "send"},
		{"Movie Theater", 28.50, "send"},
		{"Steam Games", 59.99, "send"},
		// Bills
		{"Electric Bill Payment", 125.50, "send"},
		{"Internet Service", 79.99, "send"},
		{"Phone Bill", 65.00, "send"},
		// Income
		{"Payroll Deposit", 2500.00, "receive"},
		{"Freelance Payment", 450.00, "receive"},
		{"Refund from Amazon", 29.99, "receive"},
		{"Payment from @alice", 75.00, "receive"},
	}

	// Generate 30-40 transactions spread over the time period
	numTxs := 30 + rand.Intn(11)
	for i := 0; i < numTxs; i++ {
		template := templates[rand.Intn(len(templates))]
		daysAgo := rand.Intn(days)
		txDate := now.AddDate(0, 0, -daysAgo)

		// Add variance to amounts (80% - 120%) to make it more realistic
		variance := 0.8 + rand.Float64()*0.4
		amount := math.Round(template.amount*variance*100) / 100

		transactions = append(transactions, map[string]interface{}{
			"id":          fmt.Sprintf("tx_mock_%d", i),
			"type":        template.txType,
			"amount":      amount,
			"description": template.description,
			"date":        txDate.Format(time.RFC3339),
			"status":      "completed",
			"currency":    "USD",
		})
	}

	return transactions
}

// generateMockSubscriptionTransactions creates recurring payment patterns for subscription detection
func generateMockSubscriptionTransactions(months int) []map[string]interface{} {
	rand.Seed(time.Now().UnixNano())
	now := time.Now()
	transactions := []map[string]interface{}{}

	// Subscription templates with recurring patterns
	subscriptions := []struct {
		merchant  string
		amount    float64
		frequency int // days between payments
	}{
		{"Netflix Subscription", 15.99, 30},
		{"Spotify Premium", 10.99, 30},
		{"Amazon Prime", 14.99, 30},
		{"Adobe Creative Cloud", 54.99, 30},
		{"Planet Fitness", 24.99, 30},
		{"New York Times Digital", 17.00, 30},
		{"Hulu (No Ads)", 17.99, 30},
		{"iCloud Storage 200GB", 2.99, 30},
		{"GitHub Pro", 7.00, 30},
		{"Dropbox Plus", 11.99, 30},
	}

	// Add some irregular subscriptions
	irregularSubs := []struct {
		merchant  string
		amount    float64
		frequency int
	}{
		{"Annual Software License", 299.00, 365},
		{"Quarterly Insurance", 450.00, 90},
		{"Biweekly Meal Delivery", 89.99, 14},
	}

	subscriptions = append(subscriptions, irregularSubs...)

	// Select 5-8 random subscriptions for this user
	numSubs := 5 + rand.Intn(4)
	selectedSubs := make([]struct {
		merchant  string
		amount    float64
		frequency int
	}, numSubs)
	for i := 0; i < numSubs; i++ {
		selectedSubs[i] = subscriptions[rand.Intn(len(subscriptions))]
	}

	// Generate recurring transactions for each subscription
	daysToGenerate := months * 30
	for _, sub := range selectedSubs {
		numOccurrences := daysToGenerate / sub.frequency
		for j := 0; j < numOccurrences; j++ {
			daysAgo := j * sub.frequency
			if daysAgo > daysToGenerate {
				break
			}

			txDate := now.AddDate(0, 0, -daysAgo)
			// Add small variance to amounts (±2%) to simulate real-world pricing variations
			variance := 0.98 + rand.Float64()*0.04
			amount := math.Round(sub.amount*variance*100) / 100

			transactions = append(transactions, map[string]interface{}{
				"id":          fmt.Sprintf("tx_sub_%s_%d", sub.merchant, j),
				"type":        "send",
				"amount":      amount,
				"description": sub.merchant,
				"date":        txDate.Format(time.RFC3339),
				"status":      "completed",
				"currency":    "USD",
			})
		}
	}

	// Add some one-time purchases to make the data more realistic
	oneTimePurchases := []string{
		"Whole Foods Market",
		"Target Store",
		"Uber Ride",
		"Amazon.com",
		"Starbucks Coffee",
		"Gas Station",
	}

	for i := 0; i < 20; i++ {
		purchase := oneTimePurchases[rand.Intn(len(oneTimePurchases))]
		daysAgo := rand.Intn(daysToGenerate)
		txDate := now.AddDate(0, 0, -daysAgo)
		amount := 10.00 + rand.Float64()*90.00

		transactions = append(transactions, map[string]interface{}{
			"id":          fmt.Sprintf("tx_once_%d", i),
			"type":        "send",
			"amount":      math.Round(amount*100) / 100,
			"description": purchase,
			"date":        txDate.Format(time.RFC3339),
			"status":      "completed",
			"currency":    "USD",
		})
	}

	return transactions
}

// ============================================================================
// CUSTOM TOOL: SPENDING ANALYZER
// ============================================================================

// createSpendingAnalyzerTool builds a tool that analyzes spending patterns
// Returns insights on categories, velocity, and cash flow
func createSpendingAnalyzerTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("analyze_spending").
		Description("Analyze the user's spending patterns over a specified time period. Returns insights about spending velocity, categories, and trends. Uses mock data by default for demo purposes.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"days":     tools.IntegerProperty("Number of days to analyze (default: 30)"),
			"use_mock": tools.BoolProperty("Use mock data for testing (default: true)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
			var params struct {
				Days    int  `json:"days"`
				UseMock bool `json:"use_mock"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				// Default to mock mode
				params.UseMock = true
				params.Days = 30
			}

			// Default to 30 days if not specified
			if params.Days == 0 {
				params.Days = 30
			}

			var transactions []map[string]interface{}

			// STEP 1: Get transaction data (mock or real)
			if params.UseMock {
				// Generate mock transactions
				transactions = generateMockTransactionsForAnalysis(params.Days)
				log.Printf("📊 Generated %d mock transactions for analysis", len(transactions))
			} else {
				// Fetch real transactions from Liminal API
				txRequest := map[string]interface{}{
					"limit": 100,
				}
				txRequestJSON, _ := json.Marshal(txRequest)

				txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
					UserID:    toolParams.UserID,
					Tool:      "get_transactions",
					Input:     txRequestJSON,
					RequestID: toolParams.RequestID,
				})
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
					}, nil
				}

				if !txResponse.Success {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
					}, nil
				}

				// Parse transaction data
				var txData map[string]interface{}
				if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
					if txArray, ok := txData["transactions"].([]interface{}); ok {
						for _, tx := range txArray {
							if txMap, ok := tx.(map[string]interface{}); ok {
								transactions = append(transactions, txMap)
							}
						}
					}
				}
			}

			// STEP 2: Analyze the data
			analysis := analyzeTransactions(transactions, params.Days)

			// STEP 3: Return insights
			result := map[string]interface{}{
				"period_days":        params.Days,
				"total_transactions": len(transactions),
				"analysis":           analysis,
				"data_source":        map[string]bool{"is_mock": params.UseMock},
				"generated_at":       time.Now().Format(time.RFC3339),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// analyzeTransactions processes transaction data and returns spending insights
// Calculates totals, categories, velocity, and generates actionable insights
func analyzeTransactions(transactions []map[string]interface{}, days int) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{
			"summary": "No transactions found in the specified period",
		}
	}

	// Calculate basic metrics
	var totalSpent, totalReceived float64
	var spendCount, receiveCount int
	categorySpending := make(map[string]float64)
	categoryCount := make(map[string]int)

	for _, tx := range transactions {
		txType, _ := tx["type"].(string)
		amount, _ := tx["amount"].(float64)
		description, _ := tx["description"].(string)

		category := categorizeTransaction(description)

		switch txType {
		case "send":
			totalSpent += amount
			spendCount++
			categorySpending[category] += amount
			categoryCount[category]++
		case "receive":
			totalReceived += amount
			receiveCount++
		}
	}

	avgDailySpend := totalSpent / float64(days)
	netCashFlow := totalReceived - totalSpent

	// Find top spending categories
	type categoryInfo struct {
		name       string
		amount     float64
		count      int
		percentage float64
	}
	categories := []categoryInfo{}
	for name, amount := range categorySpending {
		percentage := 0.0
		if totalSpent > 0 {
			percentage = (amount / totalSpent) * 100
		}
		categories = append(categories, categoryInfo{
			name:       name,
			amount:     amount,
			count:      categoryCount[name],
			percentage: percentage,
		})
	}
	// Sort by amount (highest first)
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].amount > categories[j].amount
	})

	// Take top 5 categories
	topCategories := []map[string]interface{}{}
	for i := 0; i < len(categories) && i < 5; i++ {
		topCategories = append(topCategories, map[string]interface{}{
			"category":   categories[i].name,
			"amount":     fmt.Sprintf("%.2f", categories[i].amount),
			"count":      categories[i].count,
			"percentage": fmt.Sprintf("%.1f%%", categories[i].percentage),
		})
	}

	// Generate human-readable insights
	insights := []string{
		fmt.Sprintf("You made %d spending transactions over %d days", spendCount, days),
		fmt.Sprintf("Average daily spend: $%.2f", avgDailySpend),
	}

	if netCashFlow > 0 {
		insights = append(insights, fmt.Sprintf("Great! You're cash flow positive with $%.2f net income", netCashFlow))
	} else if netCashFlow < 0 {
		insights = append(insights, fmt.Sprintf("You spent $%.2f more than you received this period", math.Abs(netCashFlow)))
	}

	if len(topCategories) > 0 {
		topCat := categories[0]
		insights = append(insights, fmt.Sprintf("Your biggest spending category is %s (%.0f%% of spending)", topCat.name, topCat.percentage))
	}

	return map[string]interface{}{
		"total_spent":      fmt.Sprintf("%.2f", totalSpent),
		"total_received":   fmt.Sprintf("%.2f", totalReceived),
		"net_cash_flow":    fmt.Sprintf("%.2f", netCashFlow),
		"spend_count":      spendCount,
		"receive_count":    receiveCount,
		"avg_daily_spend":  fmt.Sprintf("%.2f", avgDailySpend),
		"velocity":         calculateVelocity(spendCount, days),
		"top_categories":   topCategories,
		"insights":         insights,
	}
}

// categorizeTransaction maps merchant descriptions to spending categories
// Uses keyword matching to classify transactions
func categorizeTransaction(description string) string {
	text := strings.ToLower(description)

	// Food & Dining
	if strings.Contains(text, "starbucks") || strings.Contains(text, "coffee") ||
		strings.Contains(text, "chipotle") || strings.Contains(text, "pizza") ||
		strings.Contains(text, "food") || strings.Contains(text, "doordash") ||
		strings.Contains(text, "restaurant") || strings.Contains(text, "cafe") {
		return "Food & Dining"
	}

	// Transportation
	if strings.Contains(text, "uber") || strings.Contains(text, "lyft") ||
		strings.Contains(text, "gas") || strings.Contains(text, "metro") ||
		strings.Contains(text, "parking") {
		return "Transportation"
	}

	// Shopping
	if strings.Contains(text, "amazon") || strings.Contains(text, "target") ||
		strings.Contains(text, "nike") || strings.Contains(text, "store") {
		return "Shopping"
	}

	// Entertainment
	if strings.Contains(text, "netflix") || strings.Contains(text, "spotify") ||
		strings.Contains(text, "movie") || strings.Contains(text, "steam") ||
		strings.Contains(text, "hulu") || strings.Contains(text, "disney") {
		return "Entertainment"
	}

	// Bills & Utilities
	if strings.Contains(text, "bill") || strings.Contains(text, "electric") ||
		strings.Contains(text, "internet") || strings.Contains(text, "phone") {
		return "Bills & Utilities"
	}

	return "Other"
}

// calculateVelocity determines spending frequency (low/moderate/high)
// Based on average transactions per week
func calculateVelocity(transactionCount, days int) string {
	txPerWeek := float64(transactionCount) / float64(days) * 7

	switch {
	case txPerWeek < 2:
		return "low"
	case txPerWeek < 7:
		return "moderate"
	default:
		return "high"
	}
}

// ============================================================================
// CUSTOM TOOL: SUBSCRIPTION ANALYZER
// ============================================================================

// createSubscriptionAnalyzerTool builds a tool that detects recurring payments
// Identifies subscriptions by finding payment patterns with regular intervals
func createSubscriptionAnalyzerTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("analyze_subscriptions").
		Description("Scan transaction history to identify recurring subscriptions and recurring payments. Returns subscription patterns, total monthly costs, and cancellation insights. Uses mock data by default for demo purposes.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"timeframe_months": tools.IntegerProperty("Number of months to analyze for recurring patterns (default: 6)"),
			"min_amount":       tools.NumberProperty("Minimum amount to be considered as subscription (default: 1.00)"),
			"max_amount":       tools.NumberProperty("Maximum amount to be considered as a subscription (default: 999.99)"),
			"use_mock":         tools.BoolProperty("Use mock data for testing (default: true)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				TimeframeMonths int     `json:"timeframe_months"`
				MinAmount       float64 `json:"min_amount"`
				MaxAmount       float64 `json:"max_amount"`
				UseMock         bool    `json:"use_mock"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				// Default to mock mode
				params.UseMock = true
				params.TimeframeMonths = 6
				params.MinAmount = 1.00
				params.MaxAmount = 999.99
			}

			// Set defaults
			if params.TimeframeMonths == 0 {
				params.TimeframeMonths = 6
			}
			if params.MinAmount == 0 {
				params.MinAmount = 1.00
			}
			if params.MaxAmount == 0 {
				params.MaxAmount = 999.99
			}

			var transactions []map[string]interface{}
			now := time.Now()
			cutoffDate := now.AddDate(0, -params.TimeframeMonths, 0)

			// Get transaction data (mock or real)
			if params.UseMock {
				// Generate mock subscription transactions
				transactions = generateMockSubscriptionTransactions(params.TimeframeMonths)
				log.Printf("📊 Generated %d mock subscription transactions", len(transactions))
			} else {
				// Fetch real transactions
				txRequest := map[string]interface{}{
					"limit":      500,
					"start_date": cutoffDate.Format("2006-01-02"),
				}
				txRequestJSON, _ := json.Marshal(txRequest)
				txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
					UserID:    toolParams.UserID,
					Tool:      "get_transactions",
					Input:     txRequestJSON,
					RequestID: toolParams.RequestID,
				})
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
					}, nil
				}
				if !txResponse.Success {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
					}, nil
				}

				var txData map[string]interface{}
				if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
					if txArray, ok := txData["transactions"].([]interface{}); ok {
						for _, tx := range txArray {
							if txMap, ok := tx.(map[string]interface{}); ok {
								transactions = append(transactions, txMap)
							}
						}
					}
				}
			}

			subscriptions := analyzeForSubscriptions(transactions, cutoffDate, params.MinAmount, params.MaxAmount)
			result := map[string]interface{}{
				"analysis_period":            fmt.Sprintf("%d months", params.TimeframeMonths),
				"total_transactions_scanned": len(transactions),
				"subscriptions_found":        len(subscriptions),
				"subscriptions":              subscriptions,
				"total_monthly_cost":         calculateTotalMonthlyCost(subscriptions),
				"warnings":                   generateWarnings(subscriptions),
				"data_source":                map[string]bool{"is_mock": params.UseMock},
				"generated_at":               now.Format(time.RFC3339),
			}
			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// analyzeForSubscriptions detects recurring payment patterns
// Groups transactions by merchant+amount, checks for regular intervals
func analyzeForSubscriptions(transactions []map[string]interface{}, cutoffDate time.Time, minAmount, maxAmount float64) []map[string]interface{} {
	if len(transactions) == 0 {
		return []map[string]interface{}{}
	}

	// Group transactions by merchant and amount
	type paymentKey struct {
		merchant string
		amount   string
	}
	paymentGroups := make(map[paymentKey][]time.Time)

	for _, tx := range transactions {
		txType, _ := tx["type"].(string)
		if txType != "send" { // Only look at outgoing payments
			continue
		}

		amount, _ := tx["amount"].(float64)
		if amount < minAmount || amount > maxAmount {
			continue
		}

		merchant := "Unknown"
		if desc, ok := tx["description"].(string); ok && desc != "" {
			merchant = desc
		} else if recipient, ok := tx["recipient"].(string); ok && recipient != "" {
			merchant = recipient
		}

		txDateStr, ok := tx["date"].(string)
		if !ok {
			continue
		}
		txDate, err := time.Parse(time.RFC3339, txDateStr)
		if err != nil {
			continue
		}
		if txDate.Before(cutoffDate) {
			continue
		}

		// Round amount to avoid floating point issues
		roundedAmount := fmt.Sprintf("%.2f", amount)
		key := paymentKey{merchant: merchant, amount: roundedAmount}
		paymentGroups[key] = append(paymentGroups[key], txDate)
	}

	var subscriptions []map[string]interface{}
	for key, dates := range paymentGroups {
		if len(dates) < 2 { // Need at least 2 occurrences to detect pattern
			continue
		}

		// Sort dates chronologically
		sort.Slice(dates, func(i, j int) bool {
			return dates[i].Before(dates[j])
		})

		// Calculate intervals between payments
		intervals := make([]int, 0)
		for i := 1; i < len(dates); i++ {
			daysBetween := int(dates[i].Sub(dates[i-1]).Hours() / 24)
			intervals = append(intervals, daysBetween)
		}

		// Check if intervals form a regular pattern
		if isRegularPattern(intervals) {
			amount, _ := strconv.ParseFloat(key.amount, 64)
			frequency := detectFrequency(intervals)
			subscription := map[string]interface{}{
				"merchant":        key.merchant,
				"amount":          amount,
				"frequency":       frequency,
				"occurrences":     len(dates),
				"last_occurrence": dates[len(dates)-1].Format("2006-01-02"),
				"estimated_next":  estimateNextPayment(dates[len(dates)-1], frequency),
				"total_paid":      amount * float64(len(dates)),
				"confidence":      calculateConfidence(len(dates), intervals),
			}
			subscriptions = append(subscriptions, subscription)
		}
	}

	return subscriptions
}

// isRegularPattern checks if payment intervals are consistent (within 20% tolerance)
// Returns true if 70% or more intervals fall within tolerance
func isRegularPattern(intervals []int) bool {
	if len(intervals) == 0 {
		return false
	}
	sum := 0
	for _, interval := range intervals {
		sum += interval
	}
	avg := float64(sum) / float64(len(intervals))

	withinTolerance := 0
	tolerance := avg * 0.2 // 20% tolerance
	for _, interval := range intervals {
		if math.Abs(float64(interval)-avg) <= tolerance {
			withinTolerance++
		}
	}
	return float64(withinTolerance)/float64(len(intervals)) >= 0.7
}

// detectFrequency classifies payment frequency based on average interval
func detectFrequency(intervals []int) string {
	if len(intervals) == 0 {
		return "unknown"
	}
	sum := 0
	for _, interval := range intervals {
		sum += interval
	}
	avgDays := float64(sum) / float64(len(intervals))

	switch {
	case avgDays >= 25 && avgDays <= 35:
		return "monthly"
	case avgDays >= 80 && avgDays <= 100:
		return "quarterly"
	case avgDays >= 170 && avgDays <= 190:
		return "semi-annual"
	case avgDays >= 350 && avgDays <= 380:
		return "annual"
	case avgDays >= 7 && avgDays <= 14:
		return "biweekly"
	case avgDays >= 1 && avgDays <= 7:
		return "weekly"
	default:
		return "irregular"
	}
}

// estimateNextPayment predicts the next payment date based on frequency
func estimateNextPayment(lastPayment time.Time, frequency string) string {
	switch frequency {
	case "monthly":
		return lastPayment.AddDate(0, 1, 0).Format("2006-01-02")
	case "quarterly":
		return lastPayment.AddDate(0, 3, 0).Format("2006-01-02")
	case "semi-annual":
		return lastPayment.AddDate(0, 6, 0).Format("2006-01-02")
	case "annual":
		return lastPayment.AddDate(1, 0, 0).Format("2006-01-02")
	case "biweekly":
		return lastPayment.AddDate(0, 0, 14).Format("2006-01-02")
	case "weekly":
		return lastPayment.AddDate(0, 0, 7).Format("2006-01-02")
	default:
		return "unknown"
	}
}

// calculateConfidence determines detection confidence based on occurrences and regularity
func calculateConfidence(occurrences int, intervals []int) string {
	if occurrences >= 4 && isRegularPattern(intervals) {
		return "high"
	} else if occurrences >= 3 {
		return "medium"
	} else {
		return "low"
	}
}

// calculateTotalMonthlyCost normalizes all subscriptions to monthly cost
// Converts quarterly, annual, etc. to equivalent monthly amount
func calculateTotalMonthlyCost(subscriptions []map[string]interface{}) float64 {
	var totalMonthly float64
	for _, sub := range subscriptions {
		amount, _ := sub["amount"].(float64)
		frequency, _ := sub["frequency"].(string)
		switch frequency {
		case "monthly":
			totalMonthly += amount
		case "quarterly":
			totalMonthly += amount / 3
		case "semi-annual":
			totalMonthly += amount / 6
		case "annual":
			totalMonthly += amount / 12
		case "biweekly":
			totalMonthly += amount * 2.167 // ~26 payments/year ÷ 12 months
		case "weekly":
			totalMonthly += amount * 4.333 // ~52 payments/year ÷ 12 months
		}
	}
	return math.Round(totalMonthly*100) / 100
}

// generateWarnings creates actionable insights about subscriptions
// Identifies duplicate categories, inactive subscriptions, and savings opportunities
func generateWarnings(subscriptions []map[string]interface{}) []string {
	warnings := make([]string, 0)
	if len(subscriptions) == 0 {
		warnings = append(warnings, "No subscriptions were detected in your transaction history.")
		return warnings
	}

	totalMonthly := calculateTotalMonthlyCost(subscriptions)
	warnings = append(warnings, fmt.Sprintf("You are spending approximately $%.2f per month on subscriptions.", totalMonthly))

	// Check for duplicate categories (e.g., multiple streaming services)
	merchantCategories := make(map[string][]string)
	knownPatterns := map[string][]string{
		"streaming": {"netflix", "hulu", "disney", "prime", "spotify", "hbo", "apple tv", "youtube premium"},
		"music":     {"spotify", "apple music", "youtube music", "tidal", "pandora"},
		"cloud":     {"dropbox", "google one", "icloud", "onedrive"},
		"fitness":   {"peloton", "classpass", "apple fitness", "strava", "planet fitness"},
		"software":  {"adobe", "github", "office"},
	}

	for _, sub := range subscriptions {
		merchant, _ := sub["merchant"].(string)
		merchantLower := strings.ToLower(merchant)
		for category, keywords := range knownPatterns {
			for _, keyword := range keywords {
				if strings.Contains(merchantLower, keyword) {
					merchantCategories[category] = append(merchantCategories[category], merchant)
					break
				}
			}
		}
	}

	// Warn about duplicate categories
	for category, merchants := range merchantCategories {
		if len(merchants) > 1 {
			warnings = append(warnings, fmt.Sprintf("You have multiple %s subscriptions: %s. Consider consolidating.", category, strings.Join(merchants, ", ")))
		}
	}

	// Check for potentially inactive subscriptions
	now := time.Now()
	for _, sub := range subscriptions {
		occurrences, _ := sub["occurrences"].(int)
		lastDateStr, _ := sub["last_occurrence"].(string)
		lastDate, err := time.Parse("2006-01-02", lastDateStr)
		if err == nil && occurrences < 3 && now.Sub(lastDate).Hours()/24 > 90 {
			merchant, _ := sub["merchant"].(string)
			warnings = append(warnings, fmt.Sprintf("Subscription to '%s' seems inactive (last paid %s). Consider cancelling if you no longer use it.", merchant, lastDateStr))
		}
	}

	// Suggest potential savings
	if totalMonthly > 50 {
		savings := math.Round(totalMonthly*0.1*100) / 100
		warnings = append(warnings, fmt.Sprintf("Tip: Cancelling just 10%% of your subscriptions could save you $%.2f monthly!", savings))
	}

	return warnings
}